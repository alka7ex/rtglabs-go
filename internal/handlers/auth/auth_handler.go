package handlers

import (
	"context"
	"database/sql" // <--- NEW: Import for standard SQL DB
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel" // <--- NEW: Import squirrel
	"github.com/google/uuid"
	// REMOVE: "rtglabs-go/ent" and "rtglabs-go/ent/session"
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
	DB *sql.DB                       // Your standard SQL database client
	sq squirrel.StatementBuilderType // <--- ADD THIS FIELD: squirrel query builder
	// ... potentially a logger, or other dependencies
}

// NewAuthHandler creates a new AuthHandler instance.
// It now accepts a *sql.DB instance.
func NewAuthHandler(db *sql.DB) *AuthHandler { // Parameter changed
	// Initialize squirrel with the appropriate placeholder format for your DB
	// squirrel.Question for MySQL/SQLite, squirrel.Dollar for PostgreSQL
	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question) // Or squirrel.Dollar if you use PostgreSQL

	return &AuthHandler{
		DB: db,
		sq: sq, // <--- INITIALIZE THE FIELD HERE
	}
}

// ValidateToken checks if the token is valid and not expired.
// It returns the user ID if valid, or an error otherwise.
// This function is intended to be used by middleware.
func (h *AuthHandler) ValidateToken(token string) (uuid.UUID, error) {
	ctx := context.Background() // Or pass the context from the HTTP request if available

	// Initialize squirrel StatementBuilder (adjust placeholder for your DB: squirrel.Question for MySQL/SQLite, squirrel.Dollar for PostgreSQL)
	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question) // Assuming MySQL/SQLite

	var userID uuid.UUID
	var expiresAt time.Time

	// Build the SQL query to select session and associated user ID
	query, args, err := sq.Select("s.expires_at", "u.id").
		From("sessions s").                  // Assuming your sessions table is named 'sessions'
		Join("users u ON s.user_id = u.id"). // Assuming your users table is named 'users'
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
