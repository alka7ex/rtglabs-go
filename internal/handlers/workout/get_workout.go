package handler // Renamed to workout_handlers for consistency

import (
	"net/http"
	"rtglabs-go/ent"
	"rtglabs-go/ent/user"
	"rtglabs-go/ent/workout"
	"rtglabs-go/ent/workoutexercise" // Needed for the WorkoutexerciseQuery

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
		WithUser(). // <--- ADDED: To ensure w.Edges.User is populated for userID mapping
		WithWorkoutExercises(func(wq *ent.WorkoutExerciseQuery) {
			wq.WithExercise()
			// wq.WithWorkout() // <--- REMOVED: Redundant as workout is the parent
			wq.WithExerciseInstance(func(eiq *ent.ExerciseInstanceQuery) {
				eiq.WithExercise() // Preload the Exercise for each ExerciseInstance
				// This assumes ExerciseInstance also has an Exercise edge.
				// If ExerciseInstance only has exercise_id, then this `WithExercise()` might be for an outer ExerciseInstance edge.
				// Based on your Zod `exercise_instance: exerciseInstanceSchema`, which has `exercise_id` but not `exercise` object
				// you might not need to load exercise FROM exercise_instance if it's not mapped.
				// But your DTO `ExerciseInstanceResponse` doesn't have an `ExerciseResponse` field.
				// So, `eiq.WithExercise()` here only makes sense if you also map `ei.Edges.Exercise` to a field in `ExerciseInstanceResponse`.
				// For now, let's keep it as it matches your original `WithExercise()` assumption.
			})
			wq.Where(workoutexercise.DeletedAtIsNil()) // Only include non-soft-deleted workout exercises
		}).
		Only(c.Request().Context())

	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "Workout not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout")
	}

	// This function will now receive an `entWorkout` that has `Edges.User` and
	// `Edges.WorkoutExercises` (with their nested `Edges.Exercise` and `Edges.ExerciseInstance`) loaded.
	return c.JSON(http.StatusOK, toWorkoutResponse(entWorkout))
}
