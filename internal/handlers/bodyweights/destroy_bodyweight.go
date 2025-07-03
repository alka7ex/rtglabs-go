package handlers

import (
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent/bodyweight"
	"rtglabs-go/ent/user"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *BodyweightHandler) DestroyBodyweight(c echo.Context) error {
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

	// 3. Perform the soft delete, ensuring it belongs to the authenticated user.
	now := time.Now()
	rowsAffected, err := h.Client.Bodyweight. // Capture rowsAffected
							Update().
							Where(
			bodyweight.IDEQ(id),
			bodyweight.HasUserWith(user.IDEQ(userID)), // ✅ query by edge condition
			bodyweight.DeletedAtIsNil(),               // Crucial: only update if not already soft-deleted
		).
		SetDeletedAt(now).
		Save(c.Request().Context())

	if err != nil {
		c.Logger().Error("Failed to delete bodyweight record:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete record")
	}

	// --- MODIFICATION START ---

	// 4. Check rowsAffected to determine the appropriate response.
	if rowsAffected == 0 {
		// If 0 rows were affected, it means:
		// a) The bodyweight with the given ID was not found for the current user.
		// b) The bodyweight was found for the current user, but it was already soft-deleted
		//    (because of bodyweight.DeletedAtIsNil() in the Where clause).
		// We can return a specific message for this scenario.
		return c.JSON(http.StatusNotFound, dto.DeleteBodyweightResponse{
			Message: "Bodyweight record not found, not accessible, or already soft-deleted.",
		})
	}

	// --- MODIFICATION END ---

	// 5. If rowsAffected > 0, it means the record was successfully soft-deleted.
	response := dto.DeleteBodyweightResponse{
		Message: "Bodyweight record deleted successfully.",
	}

	return c.JSON(http.StatusOK, response)
}
