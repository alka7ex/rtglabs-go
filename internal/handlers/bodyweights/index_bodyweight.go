package handlers

import (
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent/bodyweight"
	"rtglabs-go/ent/user"
	"rtglabs-go/provider"
	"strconv"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Index lists bodyweight records with optional filtering and pagination (maps to MVC 'Index').
func (h *BodyweightHandler) IndexBodyweight(c echo.Context) error {
	// 1. Get the authenticated user ID from the context.
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	// --- Pagination Parameters ---
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1 // Default to first page
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 {
		limit = 15 // Default limit per page
	}
	if limit > 100 {
		limit = 100 // Cap the limit
	}
	offset := (page - 1) * limit
	// --- End Pagination Parameters ---

	// 2. Base query filtered by the authenticated user's ID.
	query := h.Client.Bodyweight.
		Query().
		Where(
			bodyweight.DeletedAtIsNil(),
			bodyweight.HasUserWith(user.IDEQ(userID)), // ✅ query by edge condition
		).WithUser()

	// 3. Get total count BEFORE applying limit and offset.
	totalCount, err := query.Count(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to count bodyweights:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve records")
	}

	// 4. Fetch the paginated and sorted bodyweight records.
	entBodyweights, err := query.
		Order(bodyweight.ByCreatedAt(sql.OrderDesc())). // Always order for consistent pagination
		Limit(limit).
		Offset(offset).
		All(c.Request().Context())

	if err != nil {
		c.Logger().Error("Failed to list bodyweights:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve records")
	}

	// 5. Convert ent entities to DTOs.
	dtoBodyweights := make([]dto.BodyweightResponse, len(entBodyweights))
	for i, bw := range entBodyweights {
		dtoBodyweights[i] = toBodyweightResponse(bw)
	}

	// 6. Use the new pagination utility function
	baseURL := c.Request().URL.Path
	queryParams := c.Request().URL.Query()

	paginationData := provider.GeneratePaginationData(totalCount, page, limit, baseURL, queryParams)

	// Adjust the 'To' field in paginationData based on the actual number of items returned
	// The 'From' field calculation is already handled inside GeneratePaginationData to be 1-based index
	actualItemsCount := len(dtoBodyweights)
	if actualItemsCount > 0 {
		tempTo := offset + actualItemsCount
		paginationData.To = &tempTo
	} else {
		zero := 0 // If no items, 'to' should reflect that
		paginationData.To = &zero
	}

	// 7. Return the response using the embedded pagination DTO
	return c.JSON(http.StatusOK, dto.ListBodyweightResponse{
		Data:               dtoBodyweights,
		PaginationResponse: paginationData, // Embed the generated pagination data
	})
}

// Assume toBodyweightResponse function exists elsewhere in your handlers package
// func toBodyweightResponse(entBodyweight *ent.Bodyweight) dto.BodyweightResponse {
//     // ... conversion logic
// }

