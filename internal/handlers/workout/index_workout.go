package handler

import (
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"rtglabs-go/ent/user"
	"rtglabs-go/ent/workout"
	"rtglabs-go/ent/workoutexercise"
	"rtglabs-go/provider"
	"strconv"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *WorkoutHandler) IndexWorkout(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}

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

	query := h.Client.Workout.
		Query().
		Where(
			workout.DeletedAtIsNil(),
			workout.HasUserWith(user.IDEQ(userID)),
		).
		WithWorkoutExercises(func(wq *ent.WorkoutExerciseQuery) {
			wq.WithExercise()
			wq.WithWorkout()
			wq.WithExerciseInstance(func(eiq *ent.ExerciseInstanceQuery) {
				eiq.WithExercise()
			})
			wq.Where(workoutexercise.DeletedAtIsNil())
		})

	totalCount, err := query.Count(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to count workouts")
	}

	workouts, err := query.
		Order(workout.ByCreatedAt(sql.OrderDesc())).
		Limit(limit).
		Offset(offset).
		All(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workouts")
	}

	dtoWorkouts := make([]dto.WorkoutResponse, len(workouts))
	for i, w := range workouts {
		dtoWorkouts[i] = toWorkoutResponse(w)
	}

	baseURL := c.Request().URL.Path
	queryParams := c.Request().URL.Query()

	paginationData := provider.GeneratePaginationData(totalCount, page, limit, baseURL, queryParams)

	// Update the 'To' field based on the actual number of items in the current response
	actualItemsCount := len(dtoWorkouts)
	if actualItemsCount > 0 {
		tempTo := offset + actualItemsCount
		paginationData.To = &tempTo
	} else {
		zero := 0
		paginationData.To = &zero // If no items, 'to' should be 0 or nil
	}

	return c.JSON(http.StatusOK, dto.ListWorkoutResponse{
		Data:               dtoWorkouts,
		PaginationResponse: paginationData, // Embed the pagination data
	})
}
