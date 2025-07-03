package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"rtglabs-go/ent"
	"rtglabs-go/ent/exerciseset"
	"rtglabs-go/ent/user"
	"rtglabs-go/ent/workout"
	"rtglabs-go/ent/workoutlog"
)

// ShowWorkoutLog retrieves a single workout log by its ID for the authenticated user.
func (h *WorkoutHandler) ShowWorkoutLog(c echo.Context) error {
	ctx := c.Request().Context()

	// 1. Get authenticated user ID from context
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	// 2. Get workout_log ID from URL parameter
	workoutLogIDStr := c.Param("id")
	workoutLogID, err := uuid.Parse(workoutLogIDStr)
	if err != nil {
		c.Logger().Errorf("Invalid workout log ID format: %s, Error: %v", workoutLogIDStr, err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid workout log ID format")
	}

	// 3. Query the WorkoutLog, ensuring it belongs to the authenticated user
	// and eager-loading all necessary relationships for the DTO.
	entWorkoutLog, err := h.Client.WorkoutLog.Query().
		Where(
			workoutlog.IDEQ(workoutLogID),
			workoutlog.DeletedAtIsNil(),
			workoutlog.HasUserWith(user.IDEQ(userID)),
		).
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
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.Logger().Warnf("WorkoutLog not found or not accessible to user %s: %s", userID, workoutLogID)
			return echo.NewHTTPError(http.StatusNotFound, "Workout log not found or unauthorized")
		}
		c.Logger().Error("Failed to retrieve workout log:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout log")
	}

	// 4. Convert the Entgo entity to the DTO
	dtoWorkoutLog := toWorkoutLogResponse(entWorkoutLog)

	// 5. Return the response
	return c.JSON(http.StatusOK, dtoWorkoutLog) // CORRECTED: Pass the DTO directly
}
