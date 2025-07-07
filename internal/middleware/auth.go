package middleware

import (
	"database/sql" // <--- NEW: Import for standard SQL DB
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/squirrel" // <--- NEW: Import squirrel
	"github.com/labstack/echo/v4"
	// REMOVE: "rtglabs-go/ent" and "rtglabs-go/ent/session" as we are no longer using Ent
)

// User represents the structure of your users table.
// You should define this struct according to your actual database schema.
type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	// Add other fields you might need from the user table
}

// Session represents the structure of your sessions table.
// You should define this struct according to your actual database schema.
type Session struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	// Add other fields from the session table
}

// AuthMiddleware provides authentication based on a Bearer token.
// It now accepts *sql.DB instead of *ent.Client.
func AuthMiddleware(db *sql.DB) echo.MiddlewareFunc { // <--- PARAMETER CHANGE: *ent.Client to *sql.DB
	// Initialize squirrel StatementBuilder (recommended for consistency)
	// Use squirrel.Question for MySQL/SQLite, squirrel.Dollar for PostgreSQL
	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question) // Adjust based on your DB

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing bearer token")
			}
			token := strings.TrimPrefix(auth, "Bearer ")

			var sess Session
			var user User

			// 1. Query the session using squirrel
			// Join sessions with users to get user details in one go
			query, args, err := sq.Select(
				"s.id", "s.user_id", "s.token", "s.expires_at",
				"u.id", "u.email", // Select user fields
			).
				From("sessions s").                  // Assuming your session table is named 'sessions'
				Join("users u ON s.user_id = u.id"). // Assuming your users table is named 'users'
				Where(squirrel.Eq{"s.token": token}).
				ToSql()

			if err != nil {
				c.Logger().Errorf("AuthMiddleware: Failed to build SQL query: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}

			row := db.QueryRowContext(c.Request().Context(), query, args...)

			err = row.Scan(
				&sess.ID, &sess.UserID, &sess.Token, &sess.ExpiresAt,
				&user.ID, &user.Email, // Scan user fields
			)

			if err != nil {
				if err == sql.ErrNoRows {
					// Token not found or no matching user
					return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
				}
				c.Logger().Errorf("AuthMiddleware: Failed to scan session/user row: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}

			// 2. Check token expiry
			if sess.ExpiresAt.Before(time.Now()) {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			// 3. Set user and token in context
			c.Set("user", user) // Set the retrieved User struct
			c.Set("token", token)

			return next(c)
		}
	}
}

