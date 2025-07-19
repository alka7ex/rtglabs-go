package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// User represents the structure of your users table.
// You MUST ensure this struct matches your actual 'users' table columns.
type User struct {
	ID    uuid.UUID
	Email string
	// Add other fields from your user table if needed for this handler
}

// AuthHandler holds the standard SQL DB client.
type AuthHandler struct {
	DB             *sql.DB
	sq             squirrel.StatementBuilderType
	GoogleClientID string
	// ... potentially a logger, or other dependencies
}

// NewAuthHandler creates a new AuthHandler instance.
// It now accepts a *sql.DB instance.
func NewAuthHandler(db *sql.DB, googleClientID string) *AuthHandler {
	// FIX: Use squirrel.Dollar for PostgreSQL
	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar) // <--- CHANGED THIS LINE

	return &AuthHandler{
		DB:             db,
		sq:             sq,
		GoogleClientID: googleClientID, // ✅ set it here
	}
} // ValidateToken checks if the token is valid and not expired.
// It returns the user ID if valid, or an error otherwise.
// This function is intended to be used by middleware.
func (h *AuthHandler) ValidateToken(token string) (uuid.UUID, error) {
	ctx := context.Background() // Or pass the context from the HTTP request if available

	// NOTE: You are also re-initializing `sq` here locally with squirrel.Question.
	// You should ideally use `h.sq` which is already configured for Dollar.
	// If you intend to use this local `sq` for some reason, ensure it also uses Dollar.
	// For consistency and avoiding re-initialization overhead, it's better to just use `h.sq`.
	// For now, I'm assuming the error is from the *insert* query, not this validation.
	sqLocal := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar) // <--- Also change this if you keep it. Recommended: remove and use h.sq

	var userID uuid.UUID
	var expiresAt time.Time

	// Build the SQL query to select session and associated user ID
	query, args, err := sqLocal.Select("s.expires_at", "u.id"). // Changed to sqLocal, or simply use h.sq here
									From("sessions s").
									Join("users u ON s.user_id = u.id").
									Where(squirrel.Eq{"s.token": token}).
									ToSql()

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	// Execute the query
	row := h.DB.QueryRowContext(ctx, query, args...)

	// Scan the results into variables
	err = row.Scan(&expiresAt, &userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, errors.New("invalid or expired session token")
		}
		return uuid.Nil, fmt.Errorf("database query failed: %w", err)
	}

	// Check token expiry
	if expiresAt.Before(time.Now()) {
		return uuid.Nil, errors.New("invalid or expired session token")
	}

	// Return the user's ID
	return userID, nil
}
