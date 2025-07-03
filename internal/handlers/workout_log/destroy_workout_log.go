package handler

import (
	"net/http"
	"rtglabs-go/ent"
	"rtglabs-go/ent/exerciseset" // For exerciseset predicates
	"rtglabs-go/ent/user"        // For user predicates
	"rtglabs-go/ent/workoutlog"  // For workoutlog predicates
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// DestroyWorkoutLog soft-deletes a workout log and its associated exercise sets.
func (h *WorkoutHandler) DestroyWorkoutLog(c echo.Context) error {
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

	// Start a transaction for atomicity
	tx, err := h.Client.Tx(ctx)
	if err != nil {
		c.Logger().Error("Failed to begin transaction for soft delete:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete workout log")
	}
	defer tx.Rollback() // Rollback on error

	// 3. Query the WorkoutLog to ensure it exists and belongs to the authenticated user.
	// We use `workoutlog.IDEQ` and `workoutlog.HasUserWith` for authorization.
	// Also ensure it's not already deleted.
	workoutLogToDestroy, err := tx.WorkoutLog.Query().
		Where(
			workoutlog.IDEQ(workoutLogID),
			workoutlog.DeletedAtIsNil(),               // Only delete if not already soft-deleted
			workoutlog.HasUserWith(user.IDEQ(userID)), // Ownership check
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.Logger().Warnf("WorkoutLog not found or not accessible to user %s for deletion: %s", userID, workoutLogID)
			return echo.NewHTTPError(http.StatusNotFound, "Workout log not found or unauthorized")
		}
		c.Logger().Error("Failed to query workout log for deletion:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete workout log")
	}

	// 4. Soft delete the WorkoutLog
	now := time.Now()
	_, err = workoutLogToDestroy.Update().
		SetDeletedAt(now).
		Save(ctx)
	if err != nil {
		c.Logger().Error("Failed to soft delete WorkoutLog:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete workout log")
	}
	c.Logger().Infof("Soft deleted WorkoutLog: %s", workoutLogID)

	// 5. Soft delete all associated ExerciseSets for this WorkoutLog
	// We use a bulk update for efficiency.
	rowsAffected, err := tx.ExerciseSet.Update().
		SetDeletedAt(now).
		Where(
			exerciseset.HasWorkoutLogWith(workoutlog.IDEQ(workoutLogID)), // Ensures sets belong to this specific workout log
			exerciseset.DeletedAtIsNil(),                                 // Only soft-delete sets that are not already soft-deleted
		).
		Save(ctx)
	if err != nil {
		c.Logger().Error("Failed to soft delete associated ExerciseSets:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete workout log's sets")
	}
	c.Logger().Infof("Soft deleted %d ExerciseSets for WorkoutLog: %s", rowsAffected, workoutLogID)

	// 6. Commit the transaction
	if err = tx.Commit(); err != nil {
		c.Logger().Error("Failed to commit transaction for soft delete:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to complete deletion")
	}

	// 7. Return a success response
	return c.JSON(http.StatusOK, map[string]string{"message": "Workout log and its sets soft-deleted successfully"})
	// Or return http.StatusNoContent if you prefer no body for a successful delete
	// return c.NoContent(http.StatusNoContent)
}

// (Other handler functions and helper conversion functions would remain in this file)
