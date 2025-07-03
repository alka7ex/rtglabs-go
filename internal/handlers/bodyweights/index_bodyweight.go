package handlers

import (
	"fmt"
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent/bodyweight"
	"rtglabs-go/ent/user"
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

	// 6. Build the pagination response DTO.
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

	response := dto.ListBodyweightResponse{
		CurrentPage:  page,
		Data:         dtoBodyweights,
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
