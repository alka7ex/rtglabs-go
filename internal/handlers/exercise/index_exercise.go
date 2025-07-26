package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/model"
	"rtglabs-go/provider"
	"strconv"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/typesense/typesense-go/v3/typesense/api"
	"github.com/typesense/typesense-go/v3/typesense/api/pointer"
)

func (h *ExerciseHandler) IndexExercise(c echo.Context) error {
	ctx := c.Request().Context()

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

	searchName := strings.TrimSpace(c.QueryParam("name"))

	// If search is present, use Typesense
	if searchName != "" {
		// --- Typesense Search ---
		searchParams := &api.SearchCollectionParams{
			Q:       pointer.String(searchName),
			QueryBy: pointer.String("name"),
			Page:    pointer.Int(page),
			PerPage: pointer.Int(limit),
		}

		tsClient := h.TypesenseClient // Assumes you've added this to your handler
		searchRes, err := tsClient.Collection("exercises").Documents().Search(context.Background(), searchParams)
		if err != nil {
			c.Logger().Errorf("Typesense search failed: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search exercises")
		}

		// Parse IDs from hits
		var ids []uuid.UUID
		idIndex := make(map[uuid.UUID]int) // for ordering
		for i, hit := range searchRes.Hits {
			rawID := hit.Document["id"].(string)
			uid, err := uuid.Parse(rawID)
			if err != nil {
				c.Logger().Warnf("Invalid UUID in Typesense doc: %s", rawID)
				continue
			}
			ids = append(ids, uid)
			idIndex[uid] = i
		}

		if len(ids) == 0 {
			// No results
			return c.JSON(http.StatusOK, dto.ListExerciseResponse{
				Data:               []dto.ExerciseResponse{},
				PaginationResponse: provider.GeneratePaginationData(0, page, limit, c.Request().URL.Path, c.QueryParams()),
			})
		}

		// Query DB to get full rows
		query, args, err := h.sq.
			Select("id", "name", "created_at", "updated_at", "deleted_at").
			From("exercises").
			Where(squirrel.And{
				squirrel.Expr("id IN (?)", ids),
				squirrel.Expr("deleted_at IS NULL"),
			}).
			ToSql()
		if err != nil {
			c.Logger().Errorf("Failed to build DB query for exercises: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch exercises")
		}

		rows, err := h.DB.QueryContext(ctx, query, args...)
		if err != nil {
			c.Logger().Errorf("Failed to query exercises: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch exercises")
		}
		defer rows.Close()

		exMap := map[uuid.UUID]model.Exercise{}
		for rows.Next() {
			var ex model.Exercise
			var nullDeletedAt sql.NullTime
			if err := rows.Scan(&ex.ID, &ex.Name, &ex.CreatedAt, &ex.UpdatedAt, &nullDeletedAt); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read result row")
			}
			if nullDeletedAt.Valid {
				ex.DeletedAt = &nullDeletedAt.Time
			}
			exMap[ex.ID] = ex
		}

		// Preserve order based on Typesense hit order
		var ordered []dto.ExerciseResponse
		for _, id := range ids {
			if ex, ok := exMap[id]; ok {
				ordered = append(ordered, toExerciseResponse(&ex))
			}
		}

		// Pagination Response
		pagination := provider.GeneratePaginationData(int(*searchRes.Found), page, limit, c.Request().URL.Path, c.QueryParams())
		to := offset + len(ordered)
		pagination.To = &to

		return c.JSON(http.StatusOK, dto.ListExerciseResponse{
			Data:               ordered,
			PaginationResponse: pagination,
		})
	}

	// --- Fallback: Original SQL path (no search) ---
	where := squirrel.And{squirrel.Expr("deleted_at IS NULL")}

	// 1. Count
	countQuery, countArgs, err := h.sq.Select("COUNT(*)").
		From("exercises").
		Where(where).
		ToSql()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to count exercises")
	}
	var totalCount int
	if err := h.DB.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to count exercises")
	}

	// 2. Select
	selectQuery, selectArgs, err := h.sq.Select("id", "name", "created_at", "updated_at", "deleted_at").
		From("exercises").
		Where(where).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to build select query")
	}

	rows, err := h.DB.QueryContext(ctx, selectQuery, selectArgs...)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to query exercises")
	}
	defer rows.Close()

	var exercises []model.Exercise
	for rows.Next() {
		var ex model.Exercise
		var nullDeletedAt sql.NullTime
		if err := rows.Scan(&ex.ID, &ex.Name, &ex.CreatedAt, &ex.UpdatedAt, &nullDeletedAt); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to scan row")
		}
		if nullDeletedAt.Valid {
			ex.DeletedAt = &nullDeletedAt.Time
		}
		exercises = append(exercises, ex)
	}

	dtoExercises := make([]dto.ExerciseResponse, len(exercises))
	for i, ex := range exercises {
		dtoExercises[i] = toExerciseResponse(&ex)
	}

	pagination := provider.GeneratePaginationData(totalCount, page, limit, c.Request().URL.Path, c.QueryParams())
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
