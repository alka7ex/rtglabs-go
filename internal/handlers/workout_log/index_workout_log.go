package handler

import (
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"rtglabs-go/ent/exerciseset" // Need for ordering ExerciseSets
	"rtglabs-go/ent/user"
	"rtglabs-go/ent/workout"
	"rtglabs-go/ent/workoutlog"
	"rtglabs-go/provider"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// IndexWorkoutLog retrieves a paginated list of workout logs for the authenticated user.
func (h *WorkoutHandler) IndexWorkoutLog(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	req := new(dto.ListWorkoutLogRequest)
	if err := c.Bind(req); err != nil {
		c.Logger().Error("Failed to bind ListWorkoutLogRequest:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	page := req.Page
	limit := req.Limit
	offset := (page - 1) * limit

	query := h.Client.WorkoutLog.Query().
		Where(
			workoutlog.DeletedAtIsNil(),
			workoutlog.HasUserWith(user.IDEQ(userID)),
		)

	if req.WorkoutID != nil {
		query = query.Where(workoutlog.HasWorkoutWith(workout.IDEQ(*req.WorkoutID)))
	}
	if req.Status != nil {
		query = query.Where(workoutlog.StatusEQ(*req.Status))
	}

	totalCount, err := query.Clone().Count(ctx)
	if err != nil {
		c.Logger().Error("Failed to count workout logs:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout logs")
	}

	if req.SortBy != "" {
		orderFunc := ent.Asc
		if req.Order == "desc" {
			orderFunc = ent.Desc
		}

		switch req.SortBy {
		case "created_at":
			query = query.Order(orderFunc(workoutlog.FieldCreatedAt))
		case "started_at":
			query = query.Order(orderFunc(workoutlog.FieldStartedAt))
		case "status":
			query = query.Order(orderFunc(workoutlog.FieldStatus))
		default:
			query = query.Order(ent.Desc(workoutlog.FieldCreatedAt))
		}
	} else {
		query = query.Order(ent.Desc(workoutlog.FieldCreatedAt))
	}

	entWorkoutLogs, err := query.
		Offset(offset).
		Limit(limit).
		WithWorkout(func(wq *ent.WorkoutQuery) {
			wq.Where(workout.DeletedAtIsNil())
		}).
		WithUser().
		WithExerciseSets(func(esq *ent.ExerciseSetQuery) {
			esq.WithExercise()
			esq.WithExerciseInstance(func(eiq *ent.ExerciseInstanceQuery) {
				eiq.WithExercise()
			}).
				Where(exerciseset.DeletedAtIsNil()).
				Order(ent.Asc(exerciseset.FieldSetNumber))
		}).
		All(ctx)
	if err != nil {
		c.Logger().Error("Failed to retrieve workout logs:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout logs")
	}

	dtoWorkoutLogs := make([]dto.WorkoutLogResponse, 0)

	for _, wl := range entWorkoutLogs {
		dtoWorkoutLogs = append(dtoWorkoutLogs, toWorkoutLogResponse(wl))
	}

	baseURL := c.Request().URL.Path
	queryParams := c.Request().URL.Query()

	if req.WorkoutID != nil {
		queryParams.Set("workout_id", req.WorkoutID.String())
	}
	if req.Status != nil {
		queryParams.Set("status", strconv.Itoa(*req.Status))
	}
	if req.SortBy != "" {
		queryParams.Set("sort_by", req.SortBy)
	}
	if req.Order != "" {
		queryParams.Set("order", req.Order)
	}

	paginationData := provider.GeneratePaginationData(totalCount, page, limit, baseURL, queryParams)

	actualItemsCount := len(dtoWorkoutLogs)
	if actualItemsCount > 0 {
		tempTo := offset + actualItemsCount
		paginationData.To = &tempTo
	} else {
		// This block is actually fine as is, but if 'To' was being set to null, this ensures it's 0.
		// Given that 'from' and 'to' are 0 when no results, this matches current behavior for empty.
		zero := 0
		paginationData.To = &zero
	}

	return c.JSON(http.StatusOK, dto.ListWorkoutLogResponse{
		Data:               dtoWorkoutLogs, // <- Make sure this is an initialized slice, not nil
		PaginationResponse: paginationData,
	})
}

