package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	// !!! IMPORTANT: Add this import for PostgreSQL array support !!!
	"github.com/lib/pq"
	// Assuming rtglabs-go/provider has NullFloat64ToFloat64, NullInt64ToInt, NullTimeToTimePtr if you use them here
)

// ... (existing WorkoutLogHandler struct, StoreWorkoutLog, GetWorkoutLog functions)

// DestroyWorkoutLog performs a soft delete on a workout log and its associated
// logged exercise instances and exercise sets.
func (h *WorkoutLogHandler) DestroyWorkoutLog(c echo.Context) error {
	workoutLogIDStr := c.Param("id")
	workoutLogID, err := uuid.Parse(workoutLogIDStr)
	if err != nil {
		c.Logger().Warnf("DestroyWorkoutLog: Invalid workout log ID format: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid workout log ID format"})
	}

	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		c.Logger().Error("DestroyWorkoutLog: User ID not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}

	ctx := c.Request().Context()
	tx, err := h.DB.BeginTx(ctx, nil) // Start a transaction for atomicity
	if err != nil {
		c.Logger().Errorf("DestroyWorkoutLog: Failed to begin transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete workout log")
	}

	// IMPORTANT: Named return parameter 'err' is crucial for this defer to work correctly
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.Logger().Errorf("DestroyWorkoutLog: Recovered from panic during transaction, rolled back: %v", r)
			c.JSON(http.StatusInternalServerError, map[string]string{"message": "An unexpected error occurred."})
		} else if err != nil { // This 'err' is the named return variable for the function
			tx.Rollback()
			c.Logger().Errorf("DestroyWorkoutLog: Transaction rolled back due to error: %v", err)
		}
	}()

	now := time.Now().UTC()

	// 1. Soft delete the WorkoutLog entry
	// Changed log message prefix to DestroyWorkoutLog for consistency
	updateWorkoutLogQuery, argsWorkoutLog, buildErr := h.sq.Update("workout_logs").
		Set("deleted_at", now).
		Set("updated_at", now).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", workoutLogID, userID).
		ToSql()
	if buildErr != nil {
		c.Logger().Errorf("DestroyWorkoutLog: Failed to build update workout log query: %v", buildErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare deletion")
		return err
	}

	res, execErr := tx.ExecContext(ctx, updateWorkoutLogQuery, argsWorkoutLog...)
	if execErr != nil {
		c.Logger().Errorf("DestroyWorkoutLog: Failed to soft delete workout log %s: %v", workoutLogID, execErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete workout log")
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		// If no rows were affected, it means the workout log was not found,
		// or it was already deleted, or the user didn't own it.
		c.Logger().Warnf("DestroyWorkoutLog: No workout log found or authorized for ID %s and UserID %s to delete", workoutLogID, userID)
		err = echo.NewHTTPError(http.StatusNotFound, "Workout log not found or already deleted")
		return err
	}

	// 2. Get all LoggedExerciseInstance IDs associated with this WorkoutLog
	// We need these IDs to soft delete their related exercise_sets
	var loggedExerciseInstanceIDs []uuid.UUID
	// Using PostgreSQL's $1 placeholder, which Squirrel handles.
	// This query assumes `logged_exercise_instances.id` is type UUID in your DB.
	getLEIIDsQuery := `SELECT id FROM logged_exercise_instances WHERE workout_log_id = $1 AND deleted_at IS NULL;`
	leiRows, queryErr := tx.QueryContext(ctx, getLEIIDsQuery, workoutLogID)
	if queryErr != nil {
		c.Logger().Errorf("DestroyWorkoutLog: Failed to query logged exercise instance IDs for workout log %s: %v", workoutLogID, queryErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare deletion of related data")
		return err
	}
	defer leiRows.Close()

	for leiRows.Next() {
		var leiID uuid.UUID
		if scanErr := leiRows.Scan(&leiID); scanErr != nil {
			c.Logger().Errorf("DestroyWorkoutLog: Failed to scan logged exercise instance ID: %v", scanErr)
			err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to process related data for deletion")
			return err
		}
		loggedExerciseInstanceIDs = append(loggedExerciseInstanceIDs, leiID)
	}
	if err := leiRows.Err(); err != nil {
		c.Logger().Errorf("DestroyWorkoutLog: Rows iteration error for LEI IDs: %v", err)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to process related data for deletion")
		return err
	}

	// 3. Soft delete all LoggedExerciseInstance entries related to this WorkoutLog
	updateLEIQuery, argsLEI, buildErr := h.sq.Update("logged_exercise_instances").
		Set("deleted_at", now).
		Set("updated_at", now).
		Where("workout_log_id = ? AND deleted_at IS NULL", workoutLogID). // Use '?' for Squirrel
		ToSql()
	if buildErr != nil {
		c.Logger().Errorf("DestroyWorkoutLog: Failed to build update logged exercise instances query: %v", buildErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare deletion of related instances")
		return err
	}

	_, execErr = tx.ExecContext(ctx, updateLEIQuery, argsLEI...)
	if execErr != nil {
		c.Logger().Errorf("DestroyWorkoutLog: Failed to soft delete logged exercise instances for workout log %s: %v", workoutLogID, execErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete related exercise instances")
		return err
	}

	// 4. Soft delete all ExerciseSet entries related to the LoggedExerciseInstances
	// Use the collected loggedExerciseInstanceIDs for efficient deletion
	if len(loggedExerciseInstanceIDs) > 0 {
		updateESBuilder := h.sq.Update("exercise_sets").
			Set("deleted_at", now).
			Set("updated_at", now).
			// !!! IMPORTANT CHANGE HERE: Use pq.Array to wrap the slice !!!
			Where("logged_exercise_instance_id = ANY(?) AND deleted_at IS NULL", pq.Array(loggedExerciseInstanceIDs)) // Using PostgreSQL ANY operator with pq.Array

		updateESQuery, esArgs, buildErr := updateESBuilder.ToSql()
		if buildErr != nil {
			c.Logger().Errorf("DestroyWorkoutLog: Failed to build update exercise sets query: %v", buildErr)
			err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare deletion of exercise sets")
			return err
		}

		// --- DEBUGGING ADDITION ---
		c.Logger().Infof("DestroyWorkoutLog: Exercise Sets SQL: %s", updateESQuery)
		c.Logger().Infof("DestroyWorkoutLog: Exercise Sets Args: %v", esArgs)
		// --- END DEBUGGING ADDITION ---

		_, execErr = tx.ExecContext(ctx, updateESQuery, esArgs...)
		if execErr != nil {
			c.Logger().Errorf("DestroyWorkoutLog: Failed to soft delete exercise sets for workout log %s: %v", workoutLogID, execErr)
			c.Logger().Errorf("DestroyWorkoutLog: SQL error during ExerciseSet soft delete: %v", execErr) // More specific error logging
			err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete related exercise sets")
			return err
		}
	} else {
		c.Logger().Infof("DestroyWorkoutLog: No logged exercise instances found for workout log %s, skipping exercise set deletion.", workoutLogID)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		c.Logger().Errorf("DestroyWorkoutLog: Failed to commit transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to finalize workout log deletion")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Workout log and associated data soft-deleted successfully"})
}

