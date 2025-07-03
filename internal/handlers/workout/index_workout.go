package handler

import (
	"fmt"
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"rtglabs-go/ent/user"
	"rtglabs-go/ent/workout"
	"rtglabs-go/ent/workoutexercise"
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
			wq.WithWorkout() // 👈 ADD THIS
			wq.WithExerciseInstance(func(eiq *ent.ExerciseInstanceQuery) {
				eiq.WithExercise() // ✅ Preload fix
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

	lastPage := (totalCount + limit - 1) / limit
	baseURL := c.Request().URL.Path
	queryParams := c.Request().URL.Query()

	buildPageURL := func(p int) string {
		queryParams.Set("page", strconv.Itoa(p))
		queryParams.Set("limit", strconv.Itoa(limit))
		return fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())
	}

	var links []dto.Link
	for i := 1; i <= lastPage; i++ {
		url := buildPageURL(i)
		links = append(links, dto.Link{
			URL:    &url,
			Label:  strconv.Itoa(i),
			Active: i == page,
		})
	}

	var prevURL, nextURL *string
	if page > 1 {
		url := buildPageURL(page - 1)
		prevURL = &url
	}
	if page < lastPage {
		url := buildPageURL(page + 1)
		nextURL = &url
	}

	return c.JSON(http.StatusOK, dto.ListWorkoutResponse{
		CurrentPage:  page,
		Data:         dtoWorkouts,
		FirstPageURL: buildPageURL(1),
		From:         &offset,
		LastPage:     lastPage,
		LastPageURL:  buildPageURL(lastPage),
		Links:        links,
		NextPageURL:  nextURL,
		Path:         baseURL,
		PerPage:      limit,
		PrevPageURL:  prevURL,
		To:           &offset,
		Total:        totalCount,
	})
}
