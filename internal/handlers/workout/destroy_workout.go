package handler

import (
	"database/sql" // For sql.DB, sql.Tx, sql.Null* types
	"net/http"
	"time"

	"github.com/Masterminds/squirrel" // Import squirrel
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// containsUUID is a helper to check if a slice of UUIDs contains a specific UUID.
func containsUUID(s []uuid.UUID, e uuid.UUID) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// DestroyWorkout soft deletes a workout and its associated workout exercises and exercise instances.
func (h *WorkoutHandler) DestroyWorkout(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		c.Logger().Error("DestroyWorkout: User ID not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}

	workoutID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Logger().Errorf("DestroyWorkout: Invalid workout ID param: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid workout ID")
	}

	ctx := c.Request().Context()
	tx, err := h.DB.BeginTx(ctx, nil) // Use h.DB for the transaction
	if err != nil {
		c.Logger().Errorf("DestroyWorkout: Failed to begin transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete workout (transaction error)")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.Logger().Errorf("DestroyWorkout: Recovered from panic, transaction rolled back: %v", r)
			panic(r)
		} else if err != nil { // Check if an error occurred in the main function body
			tx.Rollback()
			c.Logger().Errorf("DestroyWorkout: Transaction rolled back due to error: %v", err)
		}
	}()

	now := time.Now().UTC()

	// --- 1. Verify workout ownership and existence, and collect ExerciseInstance IDs ---
	// Need to select workout ID, user ID, and also the exercise_instance_ids from associated workout_exercises.
	// We'll join and scan to get a single workout's data, plus all relevant exercise instance IDs.

	// First, verify the workout exists and belongs to the user.
	var existingWorkoutID uuid.UUID
	checkWorkoutBuilder := h.sq.Select("id").From("workouts").
		Where(squirrel.Eq{"id": workoutID, "deleted_at": nil, "user_id": userID})
	checkWorkoutQuery, checkWorkoutArgs, err := checkWorkoutBuilder.ToSql()
	if err != nil {
		c.Logger().Errorf("DestroyWorkout: Failed to build check workout query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout for deletion")
	}

	err = tx.QueryRowContext(ctx, checkWorkoutQuery, checkWorkoutArgs...).Scan(&existingWorkoutID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusNotFound, "Workout not found or already deleted or unauthorized")
		}
		c.Logger().Errorf("DestroyWorkout: Failed to check workout existence: %v", err)
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout for deletion")
	}

	// Now, collect the exercise_instance_ids from the associated non-deleted workout_exercises
	exerciseInstanceIDsToDelete := make([]uuid.UUID, 0)
	collectInstancesBuilder := h.sq.Select("exercise_instance_id").
		From("workout_exercises").
		Where(squirrel.Eq{"workout_id": workoutID, "deleted_at": nil}).
		Where(squirrel.Expr("exercise_instance_id IS NOT NULL")) // Only consider non-null instance IDs

	collectInstancesQuery, collectInstancesArgs, err := collectInstancesBuilder.ToSql()
	if err != nil {
		c.Logger().Errorf("DestroyWorkout: Failed to build collect instances query: %v", err)
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare for deletion")
	}

	rows, err := tx.QueryContext(ctx, collectInstancesQuery, collectInstancesArgs...)
	if err != nil {
		c.Logger().Errorf("DestroyWorkout: Failed to query exercise instances for deletion: %v", err)
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare for deletion")
	}
	defer rows.Close()

	for rows.Next() {
		var instanceID sql.Null[uuid.UUID] // Use sql.Null[T] for nullable UUIDs
		if err := rows.Scan(&instanceID); err != nil {
			c.Logger().Errorf("DestroyWorkout: Failed to scan exercise instance ID: %v", err)
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare for deletion")
		}
		if instanceID.Valid && instanceID.V != uuid.Nil {
			// Ensure unique IDs
			if !containsUUID(exerciseInstanceIDsToDelete, instanceID.V) {
				exerciseInstanceIDsToDelete = append(exerciseInstanceIDsToDelete, instanceID.V)
			}
		}
	}
	if err = rows.Err(); err != nil {
		c.Logger().Errorf("DestroyWorkout: Rows iteration error for collecting instance IDs: %v", err)
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare for deletion")
	}

	// --- 2. Soft delete the workout itself ---
	updateWorkoutBuilder := h.sq.Update("workouts").
		Set("deleted_at", now).
		Where(squirrel.Eq{"id": workoutID})
	updateWorkoutQuery, updateWorkoutArgs, err := updateWorkoutBuilder.ToSql()
	if err != nil {
		c.Logger().Errorf("DestroyWorkout: Failed to build soft delete workout query: %v", err)
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to soft delete workout")
	}
	_, err = tx.ExecContext(ctx, updateWorkoutQuery, updateWorkoutArgs...)
	if err != nil {
		c.Logger().Errorf("DestroyWorkout: Failed to soft delete workout: %v", err)
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to soft delete workout")
	}

	// --- 3. Soft delete all associated WorkoutExercises ---
	updateWEBuilder := h.sq.Update("workout_exercises").
		Set("deleted_at", now).
		Where(squirrel.Eq{"workout_id": workoutID, "deleted_at": nil}) // Only soft-delete non-deleted WEs
	updateWEQuery, updateWEArgs, err := updateWEBuilder.ToSql()
	if err != nil {
		c.Logger().Errorf("DestroyWorkout: Failed to build soft delete workout exercises query: %v", err)
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to soft delete associated workout exercises")
	}
	_, err = tx.ExecContext(ctx, updateWEQuery, updateWEArgs...)
	if err != nil {
		c.Logger().Errorf("DestroyWorkout: Failed to soft delete associated workout exercises: %v", err)
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to soft delete associated workout exercises")
	}

	// --- 4. Soft delete associated ExerciseInstances ---
	if len(exerciseInstanceIDsToDelete) > 0 {
		updateEIBuilder := h.sq.Update("exercise_instances").
			Set("deleted_at", now).
			Where(squirrel.Eq{"id": exerciseInstanceIDsToDelete, "deleted_at": nil}) // Only soft-delete non-deleted EIs
		updateEIQuery, updateEIArgs, err := updateEIBuilder.ToSql()
		if err != nil {
			c.Logger().Errorf("DestroyWorkout: Failed to build soft delete exercise instances query: %v", err)
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to soft delete associated exercise instances")
		}
		_, err = tx.ExecContext(ctx, updateEIQuery, updateEIArgs...)
		if err != nil {
			c.Logger().Errorf("DestroyWorkout: Failed to soft delete associated exercise instances: %v", err)
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to soft delete associated exercise instances")
		}
	}

	// --- Commit the transaction ---
	if err = tx.Commit(); err != nil {
		c.Logger().Errorf("DestroyWorkout: Failed to commit transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to finalize workout deletion")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Workout, associated exercises, and exercise instances deleted successfully.",
	})
}
