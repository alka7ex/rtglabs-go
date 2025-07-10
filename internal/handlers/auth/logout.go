package handlers

import (
	"net/http"
	"rtglabs-go/dto"

	"github.com/Masterminds/squirrel"
	"github.com/labstack/echo/v4"
)

// DestroySession handles user logout by invalidating a specific session token.
func (h *AuthHandler) DestroySession(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Authorization header required")
	}

	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization header format")
	}

	if tokenString == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Session token is required")
	}

	ctx := c.Request().Context()

	// Build the DELETE query for the session
	deleteSessionQuery, deleteSessionArgs, err := h.sq.Delete("sessions").
		Where(squirrel.Eq{"token": tokenString}).
		ToSql()
	if err != nil {
		c.Logger().Errorf("DestroySession: Failed to build delete session query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to destroy session")
	}

	result, err := h.DB.ExecContext(ctx, deleteSessionQuery, deleteSessionArgs...)
	if err != nil {
		c.Logger().Errorf("DestroySession: Failed to delete session: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to destroy session")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.Logger().Errorf("DestroySession: Failed to get rows affected: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify session destruction")
	}

	if rowsAffected == 0 {
		// This means the token was not found or already deleted
		return echo.NewHTTPError(http.StatusNotFound, "Session not found or already invalidated")
	}

	return c.JSON(http.StatusOK, dto.LogoutResponse{
		Message: "Session destroyed successfully",
	})
}
