package handlers

import (
	"database/sql"
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/model"
	"rtglabs-go/provider"
	"strconv"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/labstack/echo/v4"
)

// IndexExercise lists exercise records with optional filtering and pagination.
func (h *ExerciseHandler) IndexExercise(c echo.Context) error {
	// --- Pagination Parameters ---
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 {
		limit = 15
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit
	// --- End Pagination Parameters ---

	ctx := c.Request().Context()
	searchName := c.QueryParam("name")

	// --- WHERE Clause Construction ---
	where := squirrel.And{
		squirrel.Expr("deleted_at IS NULL"),
	}
	if searchName != "" {
		// ILIKE for case-insensitive PostgreSQL, fallback to LOWER(name) LIKE for SQLite
		where = append(where, squirrel.Expr("LOWER(name) LIKE ?", "%"+strings.ToLower(searchName)+"%"))
	}
	// --- End WHERE Clause ---

	// 1. Count Query
	countQuery, countArgs, err := h.sq.Select("COUNT(*)").
		From("exercises").
		Where(where).
		ToSql()
	if err != nil {
		c.Logger().Errorf("IndexExercise: Failed to build count query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to count exercises")
	}

	var totalCount int
	err = h.DB.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		c.Logger().Errorf("IndexExercise: Failed to execute count query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to count exercises")
	}

	// 2. Select Query
	selectQuery, selectArgs, err := h.sq.Select("id", "name", "created_at", "updated_at", "deleted_at").
		From("exercises").
		Where(where).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		c.Logger().Errorf("IndexExercise: Failed to build select query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list exercises")
	}

	rows, err := h.DB.QueryContext(ctx, selectQuery, selectArgs...)
	if err != nil {
		c.Logger().Errorf("IndexExercise: Failed to query exercises: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list exercises")
	}
	defer rows.Close()

	var exercises []model.Exercise
	for rows.Next() {
		var ex model.Exercise
		var nullDeletedAt sql.NullTime

		if err := rows.Scan(&ex.ID, &ex.Name, &ex.CreatedAt, &ex.UpdatedAt, &nullDeletedAt); err != nil {
			c.Logger().Errorf("IndexExercise: Failed to scan row: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list exercises")
		}

		if nullDeletedAt.Valid {
			ex.DeletedAt = &nullDeletedAt.Time
		}

		exercises = append(exercises, ex)
	}

	if err := rows.Err(); err != nil {
		c.Logger().Errorf("IndexExercise: Row iteration error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list exercises")
	}

	// 3. Convert to DTOs
	dtoExercises := make([]dto.ExerciseResponse, len(exercises))
	for i, ex := range exercises {
		dtoExercises[i] = toExerciseResponse(&ex)
	}

	// 4. Pagination metadata
	baseURL := c.Request().URL.Path
	queryParams := c.Request().URL.Query()
	if searchName != "" {
		queryParams.Set("name", searchName)
	}

	pagination := provider.GeneratePaginationData(totalCount, page, limit, baseURL, queryParams)

	if len(dtoExercises) > 0 {
		tempTo := offset + len(dtoExercises)
		pagination.To = &tempTo
	} else {
		zero := 0
		pagination.To = &zero
	}

	return c.JSON(http.StatusOK, dto.ListExerciseResponse{
		Data:               dtoExercises,
		PaginationResponse: pagination,
	})
}
