package handlers

import (
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent/exercise"
	"rtglabs-go/provider"
	"strconv"

	"entgo.io/ent/dialect/sql"
	"github.com/labstack/echo/v4"
)

// IndexExercise lists exercise records with optional filtering and pagination.
func (h *ExerciseHandler) IndexExercise(c echo.Context) error {
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

	// 1. Base query for Exercise. No UserID filtering as it's assumed master data.
	query := h.Client.Exercise.
		Query().
		Where(exercise.DeletedAtIsNil()) // Filter out soft-deleted records

	// 2. Get total count BEFORE applying limit and offset.
	totalCount, err := query.Count(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to count exercises:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve records")
	}

	// 3. Fetch the paginated and sorted exercise records.
	entExercises, err := query.
		Order(exercise.ByCreatedAt(sql.OrderDesc())). // Always order for consistent pagination
		Limit(limit).
		Offset(offset).
		All(c.Request().Context())

	if err != nil {
		c.Logger().Error("Failed to list exercises:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve records")
	}

	// 4. Convert ent entities to DTOs.
	dtoExercises := make([]dto.ExerciseResponse, len(entExercises))
	for i, ex := range entExercises {
		dtoExercises[i] = toExerciseResponse(ex)
	}

	// 5. Use the new pagination utility function
	baseURL := c.Request().URL.Path
	queryParams := c.Request().URL.Query()

	paginationData := provider.GeneratePaginationData(totalCount, page, limit, baseURL, queryParams)

	// Adjust the 'To' field in paginationData based on the actual number of items returned
	actualItemsCount := len(dtoExercises)
	if actualItemsCount > 0 {
		tempTo := offset + actualItemsCount
		paginationData.To = &tempTo
	} else {
		zero := 0 // If no items, 'to' should be 0 or nil
		paginationData.To = &zero
	}

	// 6. Return the response using the embedded pagination DTO
	return c.JSON(http.StatusOK, dto.ListExerciseResponse{
		Data:               dtoExercises,
		PaginationResponse: paginationData, // Embed the generated pagination data
	})
}

// Assume toExerciseResponse function exists elsewhere in your handlers package
// func toExerciseResponse(entExercise *ent.Exercise) dto.ExerciseResponse {
//     // ... conversion logic
// }
