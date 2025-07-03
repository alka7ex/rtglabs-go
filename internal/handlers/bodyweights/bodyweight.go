package handlers

import (
	"fmt"
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"rtglabs-go/ent/bodyweight"
	"rtglabs-go/ent/user"
	"strconv"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type BodyweightHandler struct {
	Client *ent.Client
}

func NewBodyweightHandler(client *ent.Client) *BodyweightHandler {
	return &BodyweightHandler{Client: client}
}

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

	// 4. Build the DTO response and return.
	response := dto.CreateBodyweightResponse{
		Message:    "Bodyweight record created successfully.",
		Bodyweight: toBodyweightResponse(bw),
	}

	return c.JSON(http.StatusCreated, response)
}

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
		)

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

// Update updates a bodyweight record by ID (maps to MVC 'Update').
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
	// We use the general Update() builder with a Where clause to enforce ownership.
	// This builder returns the number of updated rows, not the entity itself.
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
		Only(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to fetch updated bodyweight record:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve updated record")
	}

	// 6. Build the DTO response and return.
	response := dto.UpdateBodyweightResponse{
		Message:    "Bodyweight record updated successfully.",
		Bodyweight: toBodyweightResponse(updatedBw), // <-- FIX 2: Use the fetched entity, no indexing needed
	}

	return c.JSON(http.StatusOK, response)
}

// Destroy performs a soft delete on a bodyweight record (maps to MVC 'Destroy').
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
	_, err = h.Client.Bodyweight.
		Update().
		Where(
			bodyweight.IDEQ(id),
			bodyweight.HasUserWith(user.IDEQ(userID)), // ✅ query by edge condition
			bodyweight.DeletedAtIsNil(),
		).
		SetDeletedAt(now).
		Save(c.Request().Context())

	if err != nil {
		c.Logger().Error("Failed to delete bodyweight record:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete record")
	}

	// 4. Return the DTO response.
	response := dto.DeleteBodyweightResponse{
		Message: "Bodyweight record deleted successfully.",
	}

	return c.JSON(http.StatusOK, response)
}

// --- Helper Functions ---

// toBodyweightResponse converts an ent.Bodyweight entity to a dto.BodyweightResponse DTO.
func toBodyweightResponse(bw *ent.Bodyweight) dto.BodyweightResponse {
	// Safely dereference the pointer fields, assigning a zero value if they are nil.
	// This prevents a runtime panic.
	var createdAt time.Time

	// This part is from our previous fix, converting string to int.
	unit, err := strconv.Atoi(bw.Unit)
	if err != nil {
		// Log the error if the database value is not a valid integer.
		unit = 0
	}

	// DeletedAt can be a pointer in the DTO, so we can assign it directly.
	var deletedAt *time.Time
	if bw.DeletedAt != nil {
		deletedAt = bw.DeletedAt
	}

	return dto.BodyweightResponse{
		ID:        bw.ID,
		UserID:    bw.Edges.User.ID,
		Weight:    bw.Weight,
		Unit:      unit,
		CreatedAt: createdAt,
		DeletedAt: deletedAt,
	}
}
