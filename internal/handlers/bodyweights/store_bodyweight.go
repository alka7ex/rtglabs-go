package handlers

import (
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent/bodyweight"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Store creates a new bodyweight record (maps to MVC 'Store' or 'Post').
func (h *BodyweightHandler) StoreBodyweight(c echo.Context) error {
	// 1. Get the authenticated user ID from the context.
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	// 2. Bind and validate the request body.
	var req dto.CreateBodyweightRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// 3. Create the bodyweight entity using the user ID from context.
	// This 'bw' will NOT have its edges loaded by default.
	bw, err := h.Client.Bodyweight.
		Create().
		SetUserID(userID). // <-- Security Fix: Use userID from context
		SetWeight(req.Weight).
		SetUnit(strconv.Itoa(*req.Unit)). // <-- FIX: Dereference the pointer with `*`
		Save(c.Request().Context())

	if err != nil {
		c.Logger().Error("Failed to create bodyweight record:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create record")
	}

	// --- NEW STEP: Fetch the created bodyweight again, with the User edge loaded ---
	// Use the ID of the newly created bodyweight to fetch it
	fetchedBw, err := h.Client.Bodyweight.
		Query().
		Where(bodyweight.IDEQ(bw.ID)). // Query by the ID of the just-created bodyweight
		WithUser().                    // Eager-load the User edge
		Only(c.Request().Context())
	if err != nil {
		// This error is less likely if the Save was successful, but handle it
		c.Logger().Error("Failed to fetch newly created bodyweight with user:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve created record details")
	}
	// --- END NEW STEP ---

	// 4. Build the DTO response and return, using the fetchedBw
	response := dto.CreateBodyweightResponse{
		Message:    "Bodyweight record created successfully.",
		Bodyweight: toBodyweightResponse(fetchedBw), // Use the eager-loaded 'fetchedBw'
	}

	return c.JSON(http.StatusCreated, response)
}

