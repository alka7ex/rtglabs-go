package handler

import (
	"context"
	"database/sql"
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/model" // Assuming this contains your WorkoutLogStatusCompleted
	"rtglabs-go/provider"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *WorkoutLogHandler) UpdateWorkoutLog(c echo.Context) error {
	workoutLogIDStr := c.Param("id")
	workoutLogID, err := uuid.Parse(workoutLogIDStr)
	if err != nil {
		c.Logger().Warnf("UpdateWorkoutLog: Invalid workout log ID format: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid workout log ID format"})
	}

	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		c.Logger().Error("UpdateWorkoutLog: User ID not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}

	var req dto.UpdateWorkoutLogRequest
	if err := c.Bind(&req); err != nil {
		c.Logger().Warnf("UpdateWorkoutLog: Failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request payload"})
	}

	// Validate the request payload.
	// Ensure that if a new LEI is being created (ID is nil), ExerciseID is present.
	for _, leiReq := range req.LoggedExerciseInstances {
		if leiReq.ID == nil || *leiReq.ID == uuid.Nil {
			if leiReq.ExerciseID == nil || *leiReq.ExerciseID == uuid.Nil {
				c.Logger().Warn("UpdateWorkoutLog: New LoggedExerciseInstance without ExerciseID")
				return echo.NewHTTPError(http.StatusBadRequest, "New logged exercise instance requires an ExerciseID")
			}
		}
	}

	if err := c.Validate(req); err != nil {
		c.Logger().Warnf("UpdateWorkoutLog: Request validation failed: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	ctx := c.Request().Context()
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		c.Logger().Errorf("UpdateWorkoutLog: Failed to begin transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update workout log")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.Logger().Errorf("UpdateWorkoutLog: Recovered from panic during transaction, rolled back: %v", r)
			c.JSON(http.StatusInternalServerError, map[string]string{"message": "An unexpected error occurred."})
		} else if err != nil { // This 'err' is the named return variable
			tx.Rollback()
			c.Logger().Errorf("UpdateWorkoutLog: Transaction rolled back due to error: %v", err)
		}
	}()

	now := time.Now().UTC()

	// --- 1. Update the main WorkoutLog entry ---
	updateWLBuilder := h.sq.Update("workout_logs").
		Set("updated_at", now).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", workoutLogID, userID)

	// Based on new DTO, UpdateWorkoutLogRequest no longer has a 'Status' field.
	// If FinishedAt is provided, the status is not automatically set to Completed here.
	// If you need to update status, you must add 'Status *int' to UpdateWorkoutLogRequest DTO.
	if req.FinishedAt != nil {
		updateWLBuilder = updateWLBuilder.Set("finished_at", req.FinishedAt)
	}

	updateWLQuery, argsWL, buildErr := updateWLBuilder.ToSql()
	if buildErr != nil {
		c.Logger().Errorf("UpdateWorkoutLog: Failed to build workout log update query: %v", buildErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare workout log update")
		return err
	}

	res, execErr := tx.ExecContext(ctx, updateWLQuery, argsWL...)
	if execErr != nil {
		c.Logger().Errorf("UpdateWorkoutLog: Failed to update workout log %s: %v", workoutLogID, execErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to update workout log")
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		c.Logger().Warnf("UpdateWorkoutLog: No workout log found or authorized for ID %s and UserID %s to update", workoutLogID, userID)
		err = echo.NewHTTPError(http.StatusNotFound, "Workout log not found or already deleted")
		return err
	}

	// --- 2. Process LoggedExerciseInstances (create, update, soft-delete) ---

	// Fetch existing active LEI IDs for this workout log
	existingLEIIDs := make(map[uuid.UUID]struct{})
	rows, queryErr := tx.QueryContext(ctx, "SELECT id FROM logged_exercise_instances WHERE workout_log_id = $1 AND deleted_at IS NULL", workoutLogID)
	if queryErr != nil {
		c.Logger().Errorf("UpdateWorkoutLog: Failed to query existing logged exercise instances: %v", queryErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to update workout log (fetch existing LEI)")
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		if scanErr := rows.Scan(&id); scanErr != nil {
			c.Logger().Errorf("UpdateWorkoutLog: Failed to scan existing LEI ID: %v", scanErr)
			err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to update workout log (scan existing LEI)")
			return err
		}
		existingLEIIDs[id] = struct{}{}
	}
	if rows.Err() != nil {
		c.Logger().Errorf("UpdateWorkoutLog: Rows iteration error for existing LEI IDs: %v", rows.Err())
		err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to update workout log (rows error existing LEI)")
		return err
	}

	requestedLEIIDs := make(map[uuid.UUID]struct{})
	for _, leiReq := range req.LoggedExerciseInstances {
		// If ID is provided and valid, mark it as requested
		if leiReq.ID != nil && *leiReq.ID != uuid.Nil {
			requestedLEIIDs[*leiReq.ID] = struct{}{}
		}

		// The DTO UpdateLoggedExerciseInstanceRequest no longer has IsDeleted.
		// So explicit deletion from request is removed.
		// Deletion will be handled by comparing existing vs. requested IDs at the end of the loop.

		if leiReq.ID == nil || *leiReq.ID == uuid.Nil {
			// --- CREATE NEW LoggedExerciseInstance ---
			// Validation for missing ExerciseID for new LEI is handled at the top of the function.
			newLEIID := uuid.New()
			insertLEIBuilder := h.sq.Insert("logged_exercise_instances").
				Columns("id", "workout_log_id", "exercise_id", "created_at", "updated_at").
				Values(newLEIID, workoutLogID, *leiReq.ExerciseID, now, now) // Dereference leiReq.ExerciseID

			insertLEIQuery, argsInsertLEI, buildErr := insertLEIBuilder.ToSql()
			if buildErr != nil {
				c.Logger().Errorf("UpdateWorkoutLog: Failed to build insert LEI query: %v", buildErr)
				err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to create new exercise instance")
				return err
			}
			_, execErr := tx.ExecContext(ctx, insertLEIQuery, argsInsertLEI...)
			if execErr != nil {
				c.Logger().Errorf("UpdateWorkoutLog: Failed to insert new LEI %s: %v", newLEIID, execErr)
				err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to create new exercise instance")
				return err
			}

			// Process exercise sets for the newly created LEI
			err = h.processExerciseSets(tx, ctx, c.Logger(), workoutLogID, newLEIID, leiReq.ExerciseID, leiReq.ExerciseSets, now) // Pass *uuid.UUID
			if err != nil {
				return err
			}

		} else {
			// --- UPDATE EXISTING LoggedExerciseInstance (and its sets) ---
			currentLEIID := *leiReq.ID

			updateLEIBuilder := h.sq.Update("logged_exercise_instances").
				Set("updated_at", now).
				Where("id = ? AND workout_log_id = ? AND deleted_at IS NULL", currentLEIID, workoutLogID)

			// If ExerciseID is provided in the request for an existing LEI, update it.
			if leiReq.ExerciseID != nil {
				updateLEIBuilder = updateLEIBuilder.Set("exercise_id", *leiReq.ExerciseID)
			}

			updateLEIQuery, argsUpdateLEI, buildErr := updateLEIBuilder.ToSql()
			if buildErr != nil {
				c.Logger().Errorf("UpdateWorkoutLog: Failed to build update existing LEI query: %v", buildErr)
				err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to update exercise instance")
				return err
			}
			_, execErr := tx.ExecContext(ctx, updateLEIQuery, argsUpdateLEI...)
			if execErr != nil {
				c.Logger().Errorf("UpdateWorkoutLog: Failed to update existing LEI %s: %v", currentLEIID, execErr)
				err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to update exercise instance")
				return err
			}

			// Process exercise sets for the existing LEI
			err = h.processExerciseSets(tx, ctx, c.Logger(), workoutLogID, currentLEIID, leiReq.ExerciseID, leiReq.ExerciseSets, now) // Pass *uuid.UUID
			if err != nil {
				return err
			}
		}
	}

	// Soft delete any existing LEIs that were not present in the request
	for existingLEIID := range existingLEIIDs {
		if _, found := requestedLEIIDs[existingLEIID]; !found {
			// This means the existing LEI was removed from the request, so soft-delete it.
			err = h.softDeleteLoggedExerciseInstance(tx, ctx, c.Logger(), workoutLogID, existingLEIID, now)
			if err != nil {
				return err
			}
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		c.Logger().Errorf("UpdateWorkoutLog: Failed to commit transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to finalize workout log update")
	}

	// Fetch the updated workout log details for the response
	// After tx.Commit(), the transaction object is no longer valid for queries.
	// We need to use the main database connection (h.DB) for fetching.
	updatedWorkoutLog, fetchErr := h.fetchWorkoutLogDetails(ctx, c.Logger(), workoutLogID, userID)
	if fetchErr != nil {
		c.Logger().Errorf("UpdateWorkoutLog: Failed to fetch updated workout log details after commit: %v", fetchErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Workout log updated, but failed to retrieve full details.")
	}

	return c.JSON(http.StatusOK, dto.UpdateWorkoutLogResponse{
		Message:    "Workout log updated successfully!",
		WorkoutLog: updatedWorkoutLog,
	})
}

// Helper function to process exercise sets (create, update, soft-delete)
func (h *WorkoutLogHandler) processExerciseSets(tx *sql.Tx, ctx context.Context, logger echo.Logger,
	workoutLogID, loggedExerciseInstanceID uuid.UUID,
	// exerciseID is now *uuid.UUID (nullable) according to the new DTO
	exerciseID *uuid.UUID,
	setRequests []dto.UpdateExerciseSetRequest, now time.Time) error {

	existingSetIDs := make(map[uuid.UUID]struct{})
	rows, queryErr := tx.QueryContext(ctx, "SELECT id FROM exercise_sets WHERE logged_exercise_instance_id = $1 AND deleted_at IS NULL", loggedExerciseInstanceID)
	if queryErr != nil {
		logger.Errorf("processExerciseSets: Failed to query existing exercise sets for LEI %s: %v", loggedExerciseInstanceID, queryErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update exercise sets (fetch existing sets)")
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		if scanErr := rows.Scan(&id); scanErr != nil {
			logger.Errorf("processExerciseSets: Failed to scan existing set ID for LEI %s: %v", loggedExerciseInstanceID, scanErr)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update exercise sets (scan existing sets)")
		}
		existingSetIDs[id] = struct{}{}
	}
	if rows.Err() != nil {
		logger.Errorf("processExerciseSets: Rows iteration error for existing set IDs for LEI %s: %v", rows.Err())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update exercise sets (rows error existing sets)")
	}

	requestedSetIDs := make(map[uuid.UUID]struct{})
	for i, setReq := range setRequests {
		if setReq.ID != nil && *setReq.ID != uuid.Nil {
			requestedSetIDs[*setReq.ID] = struct{}{}
		}

		// The DTO UpdateExerciseSetRequest no longer has IsDeleted.
		// So explicit deletion from request is removed.
		// Deletion will be handled by comparing existing vs. requested IDs at the end of the loop.

		if setReq.ID == nil || *setReq.ID == uuid.Nil {
			// --- CREATE NEW ExerciseSet ---
			newSetID := uuid.New()
			// setReq.ExerciseID is now non-nullable uuid.UUID in DTO, use directly.
			// setReq.Weight is non-nullable float64 in DTO, use directly.
			// setReq.Reps, setReq.SetNumber, setReq.Status are nullable.
			insertESBuilder := h.sq.Insert("exercise_sets").
				Columns("id", "workout_log_id", "exercise_id", "logged_exercise_instance_id",
														"set_number", "weight", "reps", "finished_at", "status", "created_at", "updated_at").
				Values(newSetID, workoutLogID, setReq.ExerciseID, loggedExerciseInstanceID, // Use setReq.ExerciseID directly
					provider.IntPtrToInt(setReq.SetNumber, i+1), // Default to i+1 if SetNumber is nil
					setReq.Weight,                        // Direct use for non-nullable float64
					provider.IntPtrToInt(setReq.Reps, 0), // Default to 0 if Reps is nil
					setReq.FinishedAt,
					provider.IntPtrToInt(setReq.Status, model.ExerciseSetStatusPending), // Default status
					now, now)

			insertESQuery, argsInsertES, buildErr := insertESBuilder.ToSql()
			if buildErr != nil {
				logger.Errorf("processExerciseSets: Failed to build insert ES query: %v", buildErr)
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create new exercise set")
			}
			_, execErr := tx.ExecContext(ctx, insertESQuery, argsInsertES...)
			if execErr != nil {
				logger.Errorf("processExerciseSets: Failed to insert new ES %s for LEI %s: %v", newSetID, loggedExerciseInstanceID, execErr)
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create new exercise set")
			}

		} else {
			// --- UPDATE EXISTING ExerciseSet ---
			currentSetID := *setReq.ID
			updateESBuilder := h.sq.Update("exercise_sets").
				Set("updated_at", now).
				Where("id = ? AND logged_exercise_instance_id = ? AND deleted_at IS NULL", currentSetID, loggedExerciseInstanceID)

			// setReq.ExerciseID is non-nullable, but it's probably not intended to be updated after creation for a set.
			// If it needs to be updated:
			// updateESBuilder = updateESBuilder.Set("exercise_id", setReq.ExerciseID)

			// setReq.Weight is non-nullable (float64) in the DTO, so no nil check needed.
			updateESBuilder = updateESBuilder.Set("weight", setReq.Weight)

			// setReq.Reps, setReq.Status, setReq.SetNumber are nullable in DTO.
			if setReq.Reps != nil {
				updateESBuilder = updateESBuilder.Set("reps", *setReq.Reps)
			}
			if setReq.Status != nil {
				updateESBuilder = updateESBuilder.Set("status", *setReq.Status)
			}
			if setReq.SetNumber != nil {
				updateESBuilder = updateESBuilder.Set("set_number", *setReq.SetNumber)
			}
			if setReq.FinishedAt != nil {
				updateESBuilder = updateESBuilder.Set("finished_at", setReq.FinishedAt)
			}

			updateESQuery, argsUpdateES, buildErr := updateESBuilder.ToSql()
			if buildErr != nil {
				logger.Errorf("processExerciseSets: Failed to build update existing ES query: %v", buildErr)
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update exercise set")
			}
			res, execErr := tx.ExecContext(ctx, updateESQuery, argsUpdateES...)
			if execErr != nil {
				logger.Errorf("processExerciseSets: Failed to update existing ES %s for LEI %s: %v", currentSetID, loggedExerciseInstanceID, execErr)
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update exercise set")
			}
			rowsAffected, _ := res.RowsAffected()
			if rowsAffected == 0 {
				logger.Warnf("processExerciseSets: No exercise set found for ID %s under LEI %s to update", currentSetID, loggedExerciseInstanceID)
			}
		}
	}

	// Soft delete any existing sets that were not present in the request
	for existingSetID := range existingSetIDs {
		if _, found := requestedSetIDs[existingSetID]; !found {
			// This means the existing set was removed from the request, so soft-delete it.
			err := h.softDeleteExerciseSet(tx, ctx, logger, loggedExerciseInstanceID, existingSetID, now)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// softDeleteLoggedExerciseInstance is a helper to soft delete a logged exercise instance and its sets.
func (h *WorkoutLogHandler) softDeleteLoggedExerciseInstance(tx *sql.Tx, ctx context.Context, logger echo.Logger, workoutLogID uuid.UUID, leiID uuid.UUID, now time.Time) error {
	if leiID == uuid.Nil { // Check against uuid.Nil directly
		logger.Warn("softDeleteLoggedExerciseInstance: Attempted to soft delete nil LoggedExerciseInstance ID")
		return nil
	}

	updateLEIQuery, argsLEI, buildErr := h.sq.Update("logged_exercise_instances").
		Set("deleted_at", now).
		Set("updated_at", now).
		Where("id = ? AND workout_log_id = ? AND deleted_at IS NULL", leiID, workoutLogID). // Use leiID directly
		ToSql()
	if buildErr != nil {
		logger.Errorf("softDeleteLoggedExerciseInstance: Failed to build delete LEI query for %s: %v", leiID, buildErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare deletion of exercise instance")
	}

	res, execErr := tx.ExecContext(ctx, updateLEIQuery, argsLEI...)
	if execErr != nil {
		logger.Errorf("softDeleteLoggedExerciseInstance: Failed to soft delete LEI %s: %v", leiID, execErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete exercise instance")
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		logger.Warnf("softDeleteLoggedExerciseInstance: No LEI %s found for workout log %s to soft delete, or already deleted", leiID, workoutLogID)
	}

	// Also soft-delete all associated exercise sets
	updateESQuery, argsES, buildErr := h.sq.Update("exercise_sets").
		Set("deleted_at", now).
		Set("updated_at", now).
		Where("logged_exercise_instance_id = ? AND deleted_at IS NULL", leiID). // Use leiID directly
		ToSql()
	if buildErr != nil {
		logger.Errorf("softDeleteLoggedExerciseInstance: Failed to build delete ES query for LEI %s: %v", leiID, buildErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare deletion of exercise sets")
	}

	_, execErr = tx.ExecContext(ctx, updateESQuery, argsES...)
	if execErr != nil {
		logger.Errorf("softDeleteLoggedExerciseInstance: Failed to soft delete sets for LEI %s: %v", leiID, execErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete exercise sets")
	}
	return nil
}

// softDeleteExerciseSet is a helper to soft delete a single exercise set.
func (h *WorkoutLogHandler) softDeleteExerciseSet(tx *sql.Tx, ctx context.Context, logger echo.Logger, loggedExerciseInstanceID uuid.UUID, setID uuid.UUID, now time.Time) error {
	if setID == uuid.Nil { // Check against uuid.Nil directly
		logger.Warn("softDeleteExerciseSet: Attempted to soft delete nil ExerciseSet ID")
		return nil
	}

	updateESQuery, argsES, buildErr := h.sq.Update("exercise_sets").
		Set("deleted_at", now).
		Set("updated_at", now).
		Where("id = ? AND logged_exercise_instance_id = ? AND deleted_at IS NULL", setID, loggedExerciseInstanceID). // Use setID directly
		ToSql()
	if buildErr != nil {
		logger.Errorf("softDeleteExerciseSet: Failed to build delete ES query for %s: %v", setID, buildErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare deletion of set")
	}

	res, execErr := tx.ExecContext(ctx, updateESQuery, argsES...)
	if execErr != nil {
		logger.Errorf("softDeleteExerciseSet: Failed to soft delete ES %s: %v", setID, execErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete set")
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		logger.Warnf("softDeleteExerciseSet: No ES %s found under LEI %s to soft delete, or already deleted", setID, loggedExerciseInstanceID)
	}
	return nil
}

// fetchWorkoutLogDetails is a helper to fetch the complete workout log details after an update.
// Removed the 'tx *sql.Tx' parameter as it's not needed after tx.Commit().
func (h *WorkoutLogHandler) fetchWorkoutLogDetails(ctx context.Context, logger echo.Logger, workoutLogID, userID uuid.UUID) (dto.WorkoutLogResponse, error) {
	var workoutLog dto.WorkoutLogResponse
	leiPointerMap := make(map[uuid.UUID]*dto.LoggedExerciseInstanceLog)

	// Always use h.DB (the main database connection pool) for fetching details after a transaction is committed.
	queryRunner := h.DB

	query := `
		SELECT
			wl.id, wl.workout_id, wl.user_id, wl.started_at, wl.finished_at, wl.status,
			wl.total_active_duration_seconds, wl.total_pause_duration_seconds,
			wl.created_at, wl.updated_at, wl.deleted_at,
			w.id AS w_id, w.user_id AS w_user_id, w.name AS w_name, w.created_at AS w_created_at, w.updated_at AS w_updated_at, w.deleted_at AS w_deleted_at,
			lei.id AS lei_id, lei.workout_log_id AS lei_workout_log_id, lei.exercise_id AS lei_exercise_id,
			lei.created_at AS lei_created_at, lei.updated_at AS lei_updated_at, lei.deleted_at AS lei_deleted_at,
			e.id AS e_id, e.name AS e_name, e.created_at AS e_created_at, e.updated_at AS e_updated_at, e.deleted_at AS e_deleted_at,
			es.id AS es_id, es.workout_log_id AS es_workout_log_id, es.exercise_id AS es_exercise_id,
			es.logged_exercise_instance_id AS es_lei_id, es.set_number, es.weight, es.reps,
			es.finished_at AS es_finished_at, es.status AS es_status, es.created_at AS es_created_at, es.updated_at AS es_updated_at, es.deleted_at AS es_deleted_at
		FROM workout_logs AS wl
		LEFT JOIN workouts AS w ON wl.workout_id = w.id
		LEFT JOIN logged_exercise_instances AS lei ON wl.id = lei.workout_log_id AND lei.deleted_at IS NULL
		LEFT JOIN exercises AS e ON lei.exercise_id = e.id AND e.deleted_at IS NULL
		LEFT JOIN exercise_sets AS es ON lei.id = es.logged_exercise_instance_id AND es.deleted_at IS NULL
		WHERE wl.id = $1 AND wl.user_id = $2 AND wl.deleted_at IS NULL
		ORDER BY lei.created_at ASC, es.set_number ASC;
	`
	rows, queryErr := queryRunner.QueryContext(ctx, query, workoutLogID, userID)
	if queryErr != nil {
		if queryErr == sql.ErrNoRows {
			logger.Warnf("fetchWorkoutLogDetails: Workout log not found for ID %s and UserID %s", workoutLogID, userID)
			return workoutLog, echo.NewHTTPError(http.StatusNotFound, "Workout log not found")
		}
		logger.Errorf("fetchWorkoutLogDetails: Failed to query workout log details: %v", queryErr)
		return workoutLog, echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout log details.")
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var (
			wlID, wlUserID                        uuid.UUID
			wlWorkoutID                           sql.Null[uuid.UUID]
			wlStartedAt, wlFinishedAt             sql.NullTime
			wlStatus                              int
			wlTotalActiveDurationSeconds          uint
			wlTotalPauseDurationSeconds           uint
			wlCreatedAt, wlUpdatedAt, wlDeletedAt sql.NullTime

			wID, wUserID                       sql.Null[uuid.UUID]
			wName                              sql.NullString
			wCreatedAt, wUpdatedAt, wDeletedAt sql.NullTime

			leiID, leiWorkoutLogID, leiExerciseID    sql.Null[uuid.UUID]
			leiCreatedAt, leiUpdatedAt, leiDeletedAt sql.NullTime

			eID                                sql.Null[uuid.UUID]
			eName                              sql.NullString
			eCreatedAt, eUpdatedAt, eDeletedAt sql.NullTime

			esID, esWorkoutLogID, esExerciseID, esLeiID sql.Null[uuid.UUID]
			esSetNumber                                 sql.NullInt64
			esWeight                                    sql.NullFloat64
			esReps                                      sql.NullInt64
			esFinishedAt                                sql.NullTime
			esStatus                                    sql.NullInt64
			esCreatedAt, esUpdatedAt, esDeletedAt       sql.NullTime
		)

		scanErr := rows.Scan(
			&wlID, &wlWorkoutID, &wlUserID, &wlStartedAt, &wlFinishedAt, &wlStatus,
			&wlTotalActiveDurationSeconds, &wlTotalPauseDurationSeconds,
			&wlCreatedAt, &wlUpdatedAt, &wlDeletedAt,
			&wID, &wUserID, &wName, &wCreatedAt, &wUpdatedAt, &wDeletedAt,
			&leiID, &leiWorkoutLogID, &leiExerciseID, &leiCreatedAt, &leiUpdatedAt, &leiDeletedAt,
			&eID, &eName, &eCreatedAt, &eUpdatedAt, &eDeletedAt,
			&esID, &esWorkoutLogID, &esExerciseID, &esLeiID, &esSetNumber, &esWeight, &esReps,
			&esFinishedAt, &esStatus, &esCreatedAt, &esUpdatedAt, &esDeletedAt,
		)
		if scanErr != nil {
			logger.Errorf("fetchWorkoutLogDetails: Failed to scan workout log row: %v", scanErr)
			return workoutLog, echo.NewHTTPError(http.StatusInternalServerError, "Failed to process workout log data.")
		}

		// Only populate main workoutLog fields once
		if workoutLog.ID == uuid.Nil { // Check if ID is still the zero value
			workoutLog.ID = wlID
			// WorkoutID in DTO is non-nullable uuid.UUID
			if wlWorkoutID.Valid {
				workoutLog.WorkoutID = wlWorkoutID.V
			} else {
				workoutLog.WorkoutID = uuid.Nil // Assign zero UUID if null
			}
			workoutLog.UserID = wlUserID
			workoutLog.StartedAt = provider.NullTimeToTimePtr(wlStartedAt)
			workoutLog.FinishedAt = provider.NullTimeToTimePtr(wlFinishedAt)
			workoutLog.Status = wlStatus
			workoutLog.TotalActiveDurationSeconds = wlTotalActiveDurationSeconds
			workoutLog.TotalPauseDurationSeconds = wlTotalPauseDurationSeconds
			workoutLog.CreatedAt = wlCreatedAt.Time
			workoutLog.UpdatedAt = wlUpdatedAt.Time
			workoutLog.DeletedAt = provider.NullTimeToTimePtr(wlDeletedAt)
			workoutLog.LoggedExerciseInstances = []dto.LoggedExerciseInstanceLog{} // Correct field name

			if wID.Valid {
				workoutLog.Workout = dto.WorkoutResponse{
					ID:        wID.V,
					UserID:    wUserID.V,
					Name:      wName.String,
					CreatedAt: wCreatedAt.Time,
					UpdatedAt: wUpdatedAt.Time,
					DeletedAt: provider.NullTimeToTimePtr(wDeletedAt),
				}
			} else {
				// Fallback if workout template isn't found
				workoutLog.Workout = dto.WorkoutResponse{
					ID:        uuid.Nil,
					UserID:    uuid.Nil, // Assuming UserID for workout can be nil if workout is nil
					Name:      "",
					CreatedAt: time.Time{},
					UpdatedAt: time.Time{},
					DeletedAt: nil,
				}
			}
		}

		if leiID.Valid {
			leiUUID := leiID.V
			lei, exists := leiPointerMap[leiUUID]
			if !exists {
				// Create a new LoggedExerciseInstanceLog DTO
				newLei := dto.LoggedExerciseInstanceLog{
					ID:           leiUUID,
					WorkoutLogID: leiWorkoutLogID.V,
					ExerciseID:   leiExerciseID.V, // ExerciseID is uuid.UUID here (non-nullable from DB)
					CreatedAt:    leiCreatedAt.Time,
					UpdatedAt:    leiUpdatedAt.Time,
					DeletedAt:    provider.NullTimeToTimePtr(leiDeletedAt),
					ExerciseSets: []dto.ExerciseSetResponse{},
				}
				// Populate nested ExerciseResponse
				if eID.Valid {
					newLei.Exercise = dto.ExerciseResponse{
						ID:        eID.V,
						Name:      eName.String,
						CreatedAt: eCreatedAt.Time,
						UpdatedAt: eUpdatedAt.Time,
						DeletedAt: provider.NullTimeToTimePtr(eDeletedAt),
					}
				}
				// Append the new LoggedExerciseInstanceLog to the main workout log's slice
				workoutLog.LoggedExerciseInstances = append(workoutLog.LoggedExerciseInstances, newLei)
				// Get a pointer to the newly appended element in the slice
				lei = &workoutLog.LoggedExerciseInstances[len(workoutLog.LoggedExerciseInstances)-1]
				leiPointerMap[leiUUID] = lei // Store the pointer in the map for future set additions
			}

			// Handle ExerciseSets for this LoggedExerciseInstance
			if esID.Valid {
				esDTO := dto.ExerciseSetResponse{
					ID:                       esID.V,
					WorkoutLogID:             esWorkoutLogID.V,
					ExerciseID:               esExerciseID.V,
					LoggedExerciseInstanceID: esLeiID.V,
					SetNumber:                provider.NullInt64ToIntPtr(esSetNumber),
					Weight:                   provider.NullFloat64ToFloat64(esWeight),
					Reps:                     provider.NullInt64ToIntPtr(esReps),
					FinishedAt:               provider.NullTimeToTimePtr(esFinishedAt),
					Status:                   provider.NullInt64ToInt(esStatus), // Status is non-nullable int in DTO
					CreatedAt:                esCreatedAt.Time,
					UpdatedAt:                esUpdatedAt.Time,
					DeletedAt:                provider.NullTimeToTimePtr(esDeletedAt),
				}
				// Append the set to the correct LoggedExerciseInstance using the pointer
				lei.ExerciseSets = append(lei.ExerciseSets, esDTO)
			}
		}
	}

	if err := rows.Err(); err != nil {
		logger.Errorf("fetchWorkoutLogDetails: Rows iteration error for workout log details: %v", err)
		return workoutLog, echo.NewHTTPError(http.StatusInternalServerError, "Failed to process workout log data.")
	}

	if !found {
		return workoutLog, echo.NewHTTPError(http.StatusNotFound, "Workout log not found or not accessible.")
	}

	return workoutLog, nil
}
