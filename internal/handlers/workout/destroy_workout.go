package handler

import (
	"net/http"
	"rtglabs-go/ent"
	"rtglabs-go/ent/exerciseinstance"
	"rtglabs-go/ent/user"
	"rtglabs-go/ent/workout"
	"rtglabs-go/ent/workoutexercise"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *WorkoutHandler) DestroyWorkout(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}

	workoutID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid workout ID")
	}

	ctx := c.Request().Context()
	tx, err := h.Client.Tx(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 1. Verify workout ownership and existence, and preload WorkoutExercises
	existingWorkout, err := tx.Workout.
		Query().
		Where(
			workout.IDEQ(workoutID),
			workout.DeletedAtIsNil(),
			workout.HasUserWith(user.IDEQ(userID)),
		).
		WithWorkoutExercises(func(wq *ent.WorkoutExerciseQuery) {
			wq.Where(workoutexercise.DeletedAtIsNil()) // Only get non-deleted workout exercises
			wq.WithExerciseInstance()                  // <--- ADD THIS BACK! It's crucial for `we.Edges.ExerciseInstance.ID`
		}).
		Only(ctx)

	if err != nil {
		tx.Rollback()
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "Workout not found or already deleted")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout for deletion")
	}

	// Collect ExerciseInstance IDs to be soft-deleted
	exerciseInstanceIDsToDelete := make([]uuid.UUID, 0)
	for _, we := range existingWorkout.Edges.WorkoutExercises {
		// Access the related ExerciseInstance via Edges.
		// We now perform a nil check on Edges.ExerciseInstance itself.
		if we.Edges.ExerciseInstance != nil {
			// Now it's safe to access we.Edges.ExerciseInstance.ID
			// Also ensure the ID itself is not the zero UUID if the field is optional
			if we.Edges.ExerciseInstance.ID != uuid.Nil {
				if !containsUUID(exerciseInstanceIDsToDelete, we.Edges.ExerciseInstance.ID) {
					exerciseInstanceIDsToDelete = append(exerciseInstanceIDsToDelete, we.Edges.ExerciseInstance.ID)
				}
			}
		}
	}

	// 2. Soft delete the workout itself
	if _, err := tx.Workout.
		UpdateOneID(workoutID).
		SetDeletedAt(time.Now()).
		Save(ctx); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to soft delete workout")
	}

	// 3. Soft delete all associated WorkoutExercises
	if _, err := tx.WorkoutExercise.Update().
		Where(workoutexercise.HasWorkoutWith(workout.IDEQ(workoutID))).
		SetDeletedAt(time.Now()).
		Save(ctx); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to soft delete associated workout exercises")
	}

	// 4. Soft delete associated ExerciseInstances
	// Only soft delete if there are instances to delete
	if len(exerciseInstanceIDsToDelete) > 0 {
		if _, err := tx.ExerciseInstance.Update().
			Where(exerciseinstance.IDIn(exerciseInstanceIDsToDelete...)). // CORRECTED PREDICATE
			SetDeletedAt(time.Now()).
			Save(ctx); err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to soft delete associated exercise instances")
		}
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit workout deletion")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Workout, associated exercises, and exercise instances deleted successfully.",
	})
}

// containsUUID is a helper to check if a slice of UUIDs contains a specific UUID.
func containsUUID(s []uuid.UUID, e uuid.UUID) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
