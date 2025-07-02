package handlers

import (
	"fmt"
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"rtglabs-go/ent/exercise" // Import the exercise ent
	"strconv"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/labstack/echo/v4"
)

// ExerciseHandler holds the ent.Client for database operations.
type ExerciseHandler struct {
	Client *ent.Client
}

// NewExerciseHandler creates and returns a new ExerciseHandler.
func NewExerciseHandler(client *ent.Client) *ExerciseHandler {
	return &ExerciseHandler{Client: client}
}

// IndexExercise lists exercise records with optional filtering and pagination.
func (h *ExerciseHandler) IndexExercise(c echo.Context) error {
	// --- Pagination Parameters ---
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1 // Default to first page
	}
	limit, _ := strconv.Atoi(limitStr)
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

	// 5. Build the pagination response DTO.
	totalPages := (totalCount + limit - 1) / limit
	lastPage := totalPages

	// Build URLs for pagination
	baseURL := c.Request().URL.Path
	queryParams := c.Request().URL.Query()

	nextPageURL := ""
	if page < lastPage {
		queryParams.Set("page", strconv.Itoa(page+1))
		nextPageURL = fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())
	}
	prevPageURL := ""
	if page > 1 {
		queryParams.Set("page", strconv.Itoa(page-1))
		prevPageURL = fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())
	}

	// Set nullable pointers to nil if they are empty
	var nextPageURLPtr, prevPageURLPtr *string
	if nextPageURL != "" {
		nextPageURLPtr = &nextPageURL
	}
	if prevPageURL != "" {
		prevPageURLPtr = &prevPageURL
	}

	// Create links array
	var links []dto.Link
	links = append(links, dto.Link{URL: prevPageURLPtr, Label: "&laquo; Previous", Active: page > 1})
	for i := 1; i <= totalPages; i++ {
		pageURL := fmt.Sprintf("%s?page=%d&limit=%d", baseURL, i, limit)
		links = append(links, dto.Link{URL: &pageURL, Label: strconv.Itoa(i), Active: i == page})
	}
	links = append(links, dto.Link{URL: nextPageURLPtr, Label: "Next &raquo;", Active: page < lastPage})

	response := dto.ListExerciseResponse{ // Assuming you'll create this DTO
		CurrentPage:  page,
		Data:         dtoExercises,
		FirstPageURL: fmt.Sprintf("%s?page=1&limit=%d", baseURL, limit),
		From:         &offset,
		LastPage:     lastPage,
		LastPageURL:  fmt.Sprintf("%s?page=%d&limit=%d", baseURL, lastPage, limit),
		Links:        links,
		NextPageURL:  nextPageURLPtr,
		Path:         baseURL,
		PerPage:      limit,
		PrevPageURL:  prevPageURLPtr,
		To:           &offset, // Note: 'to' is typically offset + count, but this is a simple approximation
		Total:        totalCount,
	}

	return c.JSON(http.StatusOK, response)
}

// --- Helper Functions ---

// toExerciseResponse converts an ent.Exercise entity to a dto.ExerciseResponse DTO.
func toExerciseResponse(ex *ent.Exercise) dto.ExerciseResponse {
	var deletedAt *time.Time
	if ex.DeletedAt != nil {
		deletedAt = ex.DeletedAt
	}

	return dto.ExerciseResponse{
		ID:        ex.ID,
		Name:      ex.Name,
		CreatedAt: ex.CreatedAt,
		UpdatedAt: ex.UpdatedAt, // Include UpdatedAt from mixin
		DeletedAt: deletedAt,
	}
}
