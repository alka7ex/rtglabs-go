package handlers

import (
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent/bodyweight"
	"rtglabs-go/ent/user"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *BodyweightHandler) UpdateBodyweight(c echo.Context) error {
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

	// 3. Bind and validate the request body.
	var req dto.UpdateBodyweightRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// 4. Update the record, ensuring it belongs to the authenticated user.
	rowsAffected, err := h.Client.Bodyweight.
		Update().
		Where(
			bodyweight.IDEQ(id),
			bodyweight.HasUserWith(user.IDEQ(userID)), // ✅ query by edge condition
		).
		SetWeight(req.Weight).
		SetUnit(strconv.Itoa(*req.Unit)). // <-- FIX: Dereference the pointer and convert to string
		SetUpdatedAt(time.Now()).
		Save(c.Request().Context())

	if err != nil {
		c.Logger().Error("Failed to update bodyweight record:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update record")
	}

	// Check if any rows were affected (i.e., the record was found and updated)
	if rowsAffected == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "Bodyweight not found or you don't have access")
	}

	// 5. Fetch the updated record from the database to return it in the response.
	// This is a separate query, as the Update() builder doesn't return the entity.
	updatedBw, err := h.Client.Bodyweight.
		Query().
		Where(
			bodyweight.IDEQ(id),
			bodyweight.HasUserWith(user.IDEQ(userID)), // ✅ query by edge condition
		).
		WithUser(). // <--- ADD THIS LINE TO EAGER-LOAD THE USER
		Only(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to fetch updated bodyweight record:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve updated record")
	}

	// 6. Build the DTO response and return.
	response := dto.UpdateBodyweightResponse{
		Message:    "Bodyweight record updated successfully.",
		Bodyweight: toBodyweightResponse(updatedBw), // This 'updatedBw' will now have its User edge loaded
	}

	return c.JSON(http.StatusOK, response)
}

