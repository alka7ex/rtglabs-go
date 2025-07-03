package handlers

import (
	"net/http"
	"rtglabs-go/ent"
	"rtglabs-go/ent/bodyweight"
	"rtglabs-go/ent/user"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Get retrieves a single bodyweight record by ID (maps to MVC 'Get' or 'Show').
func (h *BodyweightHandler) GetBodyweight(c echo.Context) error {
	// 1. Get the authenticated user ID from the context.
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	// 2. Parse the ID from the URL parameter.
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID format")
	}

	// 3. Query the bodyweight record, ensuring it belongs to the authenticated user.
	bw, err := h.Client.Bodyweight.
		Query().
		Where(
			bodyweight.IDEQ(id),
			bodyweight.HasUserWith(user.IDEQ(userID)), // ✅ query by edge condition
			bodyweight.DeletedAtIsNil(),
		).
		WithUser().
		Only(c.Request().Context())
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "Bodyweight not found or you don't have access")
		}
		c.Logger().Error("Database query error:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve record")
	}

	// 4. Return the DTO response.
	return c.JSON(http.StatusOK, toBodyweightResponse(bw))
}
