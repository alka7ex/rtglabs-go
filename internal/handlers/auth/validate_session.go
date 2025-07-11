package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"rtglabs-go/model"

	"github.com/Masterminds/squirrel"
	"github.com/labstack/echo/v4"
)

// ValidateSession validates the current session token from the Authorization header.
// Assumes AuthHandler has DB *sql.DB and sq squirrel.StatementBuilderType
func (h *AuthHandler) ValidateSession(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Authorization header missing")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization header format. Expected 'Bearer <token>'")
	}

	token := parts[1]
	if token == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Token missing from Authorization header")
	}

	ctx := c.Request().Context()

	var entSession model.Session

	sessionQuery := h.sq.Select("id", "user_id", "token", "expires_at", "created_at").
		From("sessions").
		Where(squirrel.Eq{"token": token}).
		Limit(1)

	sqlQuery, args, err := sessionQuery.ToSql()
	if err != nil {
		c.Logger().Errorf("ValidateSession: Failed to build SQL query for session validation: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to validate session")
	}

	row := h.DB.QueryRowContext(ctx, sqlQuery, args...)

	err = row.Scan(
		&entSession.ID,
		&entSession.UserID,
		&entSession.Token,
		&entSession.ExpiresAt,
		&entSession.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
		}
		c.Logger().Errorf("ValidateSession: Database query error: %v. Query: %s, Args: %v", err, sqlQuery, args)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to validate session")
	}

	if time.Now().After(entSession.ExpiresAt) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Token expired")
	}

	// 3. If everything is valid, return a success JSON message
	return c.JSON(http.StatusOK, map[string]string{"message": "Session is valid"})
}

