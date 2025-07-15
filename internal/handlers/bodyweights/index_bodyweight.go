package handlers

import (
	"database/sql" // Import for sql.DB, sql.Null* types, sql.ErrNoRows
	"fmt"
	"net/http"
	"strconv"
	"strings" // Import for strings.ToLower and strings.Join

	"rtglabs-go/dto"
	"rtglabs-go/model"    // Import your model package (e.g., model.Bodyweight)
	"rtglabs-go/provider" // Import your pagination provider

	"github.com/Masterminds/squirrel" // Import squirrel
	"github.com/google/uuid"
	"github.com/labstack/echo/v4" // Import echo for context, logger, HTTP errors
)

// IndexBodyweight lists bodyweight records with optional filtering and pagination (maps to MVC 'Index').
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

	// --- Sorting Parameters ---
	sort := c.QueryParam("sort")   // e.g., "weight", "createdAt"
	order := c.QueryParam("order") // e.g., "asc", "desc"

	// Define allowed sortable columns
	allowedSortColumns := map[string]string{
		"weight":    "weight",
		"createdAt": "created_at",
		"updatedAt": "updated_at",
		// Add other sortable columns here if needed
	}

	// Default sort column and order
	defaultSortColumn := "created_at"
	defaultOrder := "DESC"

	// Validate and sanitize sort column
	if dbCol, ok := allowedSortColumns[sort]; ok {
		sort = dbCol // Use the actual database column name
	} else {
		sort = defaultSortColumn // Fallback to default
	}

	// Validate and sanitize order
	order = strings.ToUpper(order) // Convert to uppercase for consistency
	if order != "ASC" && order != "DESC" {
		order = defaultOrder // Fallback to default
	}
	// --- End Sorting Parameters ---

	ctx := c.Request().Context()

	// 2. Get total count BEFORE applying limit and offset.
	countQuery, countArgs, err := h.sq.Select("COUNT(*)").
		From("bodyweights").
		Where(squirrel.And{
			squirrel.Eq{"user_id": userID},
			squirrel.Expr("deleted_at IS NULL"),
		}).
		ToSql()

	if err != nil {
		c.Logger().Errorf("IndexBodyweight: Failed to build count query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve records count")
	}
	fmt.Printf("--- DIAGNOSTIC: IndexBodyweight Count Query Start ---\nSQL Count Query: %s\nArgs: %v\n--- DIAGNOSTIC: IndexBodyweight Count Query End ---\n", countQuery, countArgs)

	var totalCount int
	err = h.DB.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		c.Logger().Errorf("IndexBodyweight: Failed to count bodyweights: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve records count")
	}

	// 3. Fetch the paginated and sorted bodyweight records.
	// Start building the select query
	qb := h.sq.Select(
		"id", "user_id", "weight", "unit", "created_at", "updated_at", "deleted_at",
	).
		From("bodyweights").
		Where(squirrel.And{
			squirrel.Eq{"user_id": userID},
			squirrel.Expr("deleted_at IS NULL"),
		})

	// Apply sorting
	qb = qb.OrderBy(fmt.Sprintf("%s %s", sort, order)) // Use the sanitized sort and order

	// Apply pagination
	selectQuery, selectArgs, err := qb.Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()

	if err != nil {
		c.Logger().Errorf("IndexBodyweight: Failed to build select query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve records")
	}

	// Print the SQL query and arguments for debugging
	fmt.Printf("--- DIAGNOSTIC: IndexBodyweight Query Start ---\nSQL Query: %s\nArgs: %v\n--- DIAGNOSTIC: IndexBodyweight Query End ---\n", selectQuery, selectArgs)

	rows, err := h.DB.QueryContext(ctx, selectQuery, selectArgs...)
	if err != nil {
		c.Logger().Errorf("IndexBodyweight: Failed to query bodyweights: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve records")
	}
	defer rows.Close()

	var modelBodyweights []model.Bodyweight
	for rows.Next() {
		var bw model.Bodyweight
		var nullDeletedAt sql.NullTime // For nullable deleted_at column

		err := rows.Scan(
			&bw.ID,
			&bw.UserID,
			&bw.Weight,
			&bw.Unit,
			&bw.CreatedAt,
			&bw.UpdatedAt,
			&nullDeletedAt, // Scan into sql.NullTime
		)
		if err != nil {
			c.Logger().Errorf("IndexBodyweight: Failed to scan bodyweight row: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve records")
		}

		// Convert sql.NullTime to *time.Time
		if nullDeletedAt.Valid {
			bw.DeletedAt = &nullDeletedAt.Time
		} else {
			bw.DeletedAt = nil
		}
		modelBodyweights = append(modelBodyweights, bw)
	}

	// Check for any error during rows iteration
	if err = rows.Err(); err != nil {
		c.Logger().Errorf("IndexBodyweight: Rows iteration error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve records")
	}

	// 4. Convert model entities to DTOs.
	dtoBodyweights := make([]dto.BodyweightResponse, len(modelBodyweights))
	for i, bw := range modelBodyweights {
		dtoBodyweights[i] = toBodyweightResponse(&bw) // Pass address of bw
	}

	// 5. Use the new pagination utility function
	baseURL := c.Request().URL.Path
	queryParams := c.Request().URL.Query() // This already contains "sort" and "order"

	paginationData := provider.GeneratePaginationData(totalCount, page, limit, baseURL, queryParams)

	// Adjust the 'To' field in paginationData based on the actual number of items returned
	actualItemsCount := len(dtoBodyweights)
	if actualItemsCount > 0 {
		tempTo := offset + actualItemsCount
		paginationData.To = &tempTo
	} else {
		zero := 0 // If no items, 'to' should reflect that
		paginationData.To = &zero
	}

	// 6. Return the response using the embedded pagination DTO
	return c.JSON(http.StatusOK, dto.ListBodyweightResponse{
		Data:               dtoBodyweights,
		PaginationResponse: paginationData, // Embed the generated pagination data
	})
}

