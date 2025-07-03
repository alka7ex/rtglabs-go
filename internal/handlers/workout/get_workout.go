package handler

import (
	"net/http"
	"rtglabs-go/ent"
	"rtglabs-go/ent/user"
	"rtglabs-go/ent/workout"
	"rtglabs-go/ent/workoutexercise"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *WorkoutHandler) GetWorkout(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}

	workoutID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid workout ID")
	}

	entWorkout, err := h.Client.Workout.
		Query().
		Where(
			workout.IDEQ(workoutID),
			workout.HasUserWith(user.IDEQ(userID)),
			workout.DeletedAtIsNil(),
		).
		WithWorkoutExercises(func(wq *ent.WorkoutExerciseQuery) {
			wq.WithExercise()
			wq.WithWorkout() // 👈 ADD THIS
			wq.WithExerciseInstance(func(eiq *ent.ExerciseInstanceQuery) {
				eiq.WithExercise() // ✅ Preload fix
			})
			wq.Where(workoutexercise.DeletedAtIsNil())
		}).
		Only(c.Request().Context())

	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "Workout not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout")
	}

	return c.JSON(http.StatusOK, toWorkoutResponse(entWorkout))
}
