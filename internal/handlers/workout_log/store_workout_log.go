package handler

import (
	"database/sql"
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/model"
	"rtglabs-go/provider" // Ensure this is imported for helper functions like NullTimeToTimePtr
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *WorkoutLogHandler) StoreWorkoutLog(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		c.Logger().Error("StoreWorkoutLog: User ID not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}

	req := new(dto.CreateWorkoutLogRequest)
	if err := c.Bind(req); err != nil {
		c.Logger().Warnf("StoreWorkoutLog: Invalid request payload: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request payload"})
	}
	if err := c.Validate(req); err != nil {
		c.Logger().Warnf("StoreWorkoutLog: Request validation failed: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	ctx := c.Request().Context()
	tx, err := h.DB.BeginTx(ctx, nil) // Start a transaction for atomicity
	if err != nil {
		c.Logger().Errorf("StoreWorkoutLog: Failed to begin transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create workout log")
	}
	// IMPORTANT: Named return parameter 'err' is crucial for this defer to work correctly
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.Logger().Errorf("StoreWorkoutLog: Recovered from panic during transaction, rolled back: %v", r)
			c.JSON(http.StatusInternalServerError, map[string]string{"message": "An unexpected error occurred."})
			// panic(r) // Re-panic if you want to crash the application
		} else if err != nil { // This 'err' is the named return variable for the function
			tx.Rollback()
			c.Logger().Errorf("StoreWorkoutLog: Transaction rolled back due to error: %v", err)
		}
	}()

	now := time.Now().UTC()

	// 1. Fetch the workout template and its associated workout exercises
	var workoutTemplate model.Workout // This is the variable that caused the "unused" warning
	workoutExercises := make([]model.WorkoutExercise, 0)
	exercisesMap := make(map[uuid.UUID]model.Exercise) // To store exercise details for efficiency

	// Fetch Workout Template and its WorkoutExercises, along with Exercise details
	query := `
		SELECT
			w.id, w.user_id, w.name, w.created_at, w.updated_at, w.deleted_at,
			we.id AS we_id, we.workout_id AS we_workout_id, we.exercise_id AS we_exercise_id,
			we.exercise_instance_id AS we_exercise_instance_id, -- Keep this for template reference
			we.workout_order, we.sets, we.weight, we.reps, we.created_at AS we_created_at, we.updated_at AS we_updated_at, we.deleted_at AS we_deleted_at,
			e.id AS e_id, e.name AS e_name, e.created_at AS e_created_at, e.updated_at AS e_updated_at, e.deleted_at AS e_deleted_at
		FROM workouts AS w
		LEFT JOIN workout_exercises AS we ON w.id = we.workout_id AND we.deleted_at IS NULL
		LEFT JOIN exercises AS e ON we.exercise_id = e.id AND e.deleted_at IS NULL
		WHERE w.id = $1 AND w.user_id = $2 AND w.deleted_at IS NULL
		ORDER BY we.workout_order ASC;
	`
	rows, queryErr := tx.QueryContext(ctx, query, req.WorkoutID, userID)
	if queryErr != nil {
		c.Logger().Errorf("StoreWorkoutLog: Failed to query workout template: %v", queryErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout template")
		return err
	}
	defer rows.Close()

	workoutFetched := false
	for rows.Next() {
		var (
			wID, wUserID                       uuid.UUID
			wName                              sql.NullString
			wCreatedAt, wUpdatedAt, wDeletedAt sql.NullTime

			weID, weWorkoutID, weExerciseID       sql.Null[uuid.UUID]
			weExerciseInstanceID                  sql.Null[uuid.UUID]
			weWorkoutOrder, weSets, weReps        sql.NullInt64
			weWeight                              sql.NullFloat64
			weCreatedAt, weUpdatedAt, weDeletedAt sql.NullTime

			eID                                sql.Null[uuid.UUID]
			eName                              sql.NullString
			eCreatedAt, eUpdatedAt, eDeletedAt sql.NullTime
		)

		scanErr := rows.Scan(
			&wID, &wUserID, &wName, &wCreatedAt, &wUpdatedAt, &wDeletedAt,
			&weID, &weWorkoutID, &weExerciseID, &weExerciseInstanceID, &weWorkoutOrder,
			&weSets, &weWeight, &weReps, &weCreatedAt, &weUpdatedAt, &weDeletedAt,
			&eID, &eName, &eCreatedAt, &eUpdatedAt, &eDeletedAt,
		)
		if scanErr != nil {
			c.Logger().Errorf("StoreWorkoutLog: Failed to scan workout template row: %v", scanErr)
			err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to process workout template data")
			return err
		}

		// This populates `workoutTemplate` which will be used later
		if !workoutFetched {
			workoutTemplate = model.Workout{
				ID:        wID,
				UserID:    wUserID,
				Name:      wName.String,
				CreatedAt: wCreatedAt.Time,
				UpdatedAt: wUpdatedAt.Time,
				DeletedAt: provider.NullTimeToTimePtr(wDeletedAt),
			}
			workoutFetched = true
		}

		if weID.Valid {
			we := model.WorkoutExercise{
				ID:           weID.V,
				WorkoutID:    weWorkoutID.V,
				ExerciseID:   weExerciseID.V,
				WorkoutOrder: provider.NullInt64ToIntPtr(weWorkoutOrder),
				Sets:         provider.NullInt64ToIntPtr(weSets),
				Weight:       provider.NullFloat64ToFloat64Ptr(weWeight),
				Reps:         provider.NullInt64ToIntPtr(weReps),
				CreatedAt:    weCreatedAt.Time,
				UpdatedAt:    weUpdatedAt.Time,
				DeletedAt:    provider.NullTimeToTimePtr(weDeletedAt),
			}
			if weExerciseInstanceID.Valid {
				we.ExerciseInstanceID = &weExerciseInstanceID.V
			}
			workoutExercises = append(workoutExercises, we)
		}

		if eID.Valid {
			exercisesMap[eID.V] = model.Exercise{
				ID:        eID.V,
				Name:      eName.String,
				CreatedAt: eCreatedAt.Time,
				UpdatedAt: eUpdatedAt.Time,
				DeletedAt: provider.NullTimeToTimePtr(eDeletedAt),
			}
		}
	}

	if err = rows.Err(); err != nil {
		c.Logger().Errorf("StoreWorkoutLog: Rows iteration error for template: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process workout template data")
	}

	if !workoutFetched {
		return echo.NewHTTPError(http.StatusBadRequest, "Workout template not found or not accessible")
	}

	// 2. Create the new WorkoutLog entry
	newWorkoutLogID := uuid.New()
	workoutLog := model.WorkoutLog{
		ID:                         newWorkoutLogID,
		WorkoutID:                  &req.WorkoutID, // Link to the template workout
		UserID:                     userID,
		StartedAt:                  &now, // Set to current time as it's "created successfully"
		Status:                     0,    // e.g., 0 for "Planned" or "Active"
		TotalActiveDurationSeconds: 0,
		TotalPauseDurationSeconds:  0,
		CreatedAt:                  now,
		UpdatedAt:                  now,
	}

	insertWorkoutLogQuery, args, buildErr := h.sq.Insert("workout_logs").SetMap(map[string]interface{}{
		"id":                            workoutLog.ID,
		"workout_id":                    workoutLog.WorkoutID,
		"user_id":                       workoutLog.UserID,
		"started_at":                    workoutLog.StartedAt,
		"status":                        workoutLog.Status,
		"total_active_duration_seconds": workoutLog.TotalActiveDurationSeconds,
		"total_pause_duration_seconds":  workoutLog.TotalPauseDurationSeconds,
		"created_at":                    workoutLog.CreatedAt,
		"updated_at":                    workoutLog.UpdatedAt,
	}).ToSql()
	if buildErr != nil {
		c.Logger().Errorf("StoreWorkoutLog: Failed to build insert workout log query: %v", buildErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare workout log insertion")
		return err
	}

	_, execErr := tx.ExecContext(ctx, insertWorkoutLogQuery, args...)
	if execErr != nil {
		c.Logger().Errorf("StoreWorkoutLog: Failed to insert workout log: %v", execErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to create workout log entry")
		return err
	}

	// 3. Create logged_exercise_instances for each exercise from the template
	loggedExerciseInstances := make([]model.LoggedExerciseInstance, 0, len(workoutExercises))
	for _, we := range workoutExercises {
		newLoggedExerciseInstanceID := uuid.New()
		loggedInstance := model.LoggedExerciseInstance{
			ID:           newLoggedExerciseInstanceID,
			WorkoutLogID: newWorkoutLogID,
			ExerciseID:   we.ExerciseID, // Use the base exercise ID from workout_exercises
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		loggedExerciseInstances = append(loggedExerciseInstances, loggedInstance)
	}

	if len(loggedExerciseInstances) > 0 {
		insertLEIBuilder := h.sq.Insert("logged_exercise_instances").Columns(
			"id", "workout_log_id", "exercise_id", "created_at", "updated_at", "deleted_at",
		)
		for _, lei := range loggedExerciseInstances {
			insertLEIBuilder = insertLEIBuilder.Values(
				lei.ID, lei.WorkoutLogID, lei.ExerciseID, lei.CreatedAt, lei.UpdatedAt, lei.DeletedAt,
			)
		}
		insertLEIQuery, leiArgs, buildErr := insertLEIBuilder.ToSql()
		if buildErr != nil {
			c.Logger().Errorf("StoreWorkoutLog: Failed to build insert logged exercise instances query: %v", buildErr)
			err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare logged exercise instances")
			return err
		}
		_, execErr = tx.ExecContext(ctx, insertLEIQuery, leiArgs...)
		if execErr != nil {
			c.Logger().Errorf("StoreWorkoutLog: Failed to insert logged exercise instances: %v", execErr)
			c.Logger().Errorf("StoreWorkoutLog: SQL error during LEI insertion: %v", execErr)
			err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to create logged exercise instances")
			return err
		}
	}

	// 4. Create initial ExerciseSets for each LoggedExerciseInstance based on template defaults
	workoutExerciseMapByExerciseID := make(map[uuid.UUID]model.WorkoutExercise)
	for _, we := range workoutExercises {
		workoutExerciseMapByExerciseID[we.ExerciseID] = we
	}

	exerciseSetsToInsert := make([]model.ExerciseSet, 0)
	for _, lei := range loggedExerciseInstances {
		if we, found := workoutExerciseMapByExerciseID[lei.ExerciseID]; found {
			numSets := 0
			if we.Sets != nil {
				numSets = int(*we.Sets)
			}

			for i := 1; i <= numSets; i++ {
				newExerciseSetID := uuid.New()
				exerciseSet := model.ExerciseSet{
					ID:                       newExerciseSetID,
					WorkoutLogID:             lei.WorkoutLogID,
					ExerciseID:               lei.ExerciseID,
					LoggedExerciseInstanceID: lei.ID,
					SetNumber:                i,
					Weight:                   we.Weight,
					Reps:                     we.Reps,
					FinishedAt:               nil,
					Status:                   0, // Pending/Not Started
					CreatedAt:                now,
					UpdatedAt:                now,
					DeletedAt:                nil,
				}
				exerciseSetsToInsert = append(exerciseSetsToInsert, exerciseSet)
			}
		}
	}

	if len(exerciseSetsToInsert) > 0 {
		insertESBuilder := h.sq.Insert("exercise_sets").Columns(
			"id", "workout_log_id", "exercise_id", "logged_exercise_instance_id", "set_number",
			"weight", "reps", "finished_at", "status", "created_at", "updated_at", "deleted_at",
		)
		for _, es := range exerciseSetsToInsert {
			insertESBuilder = insertESBuilder.Values(
				es.ID, es.WorkoutLogID, es.ExerciseID, es.LoggedExerciseInstanceID, es.SetNumber,
				es.Weight, es.Reps, es.FinishedAt, es.Status, es.CreatedAt, es.UpdatedAt, es.DeletedAt,
			)
		}
		insertESQuery, esArgs, buildErr := insertESBuilder.ToSql()
		if buildErr != nil {
			c.Logger().Errorf("StoreWorkoutLog: Failed to build insert exercise sets query: %v", buildErr)
			err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare exercise sets")
			return err
		}
		_, execErr = tx.ExecContext(ctx, insertESQuery, esArgs...)
		if execErr != nil {
			c.Logger().Errorf("StoreWorkoutLog: Failed to insert exercise sets: %v", execErr)
			c.Logger().Errorf("StoreWorkoutLog: SQL error during ExerciseSet insertion: %v", execErr)
			err = echo.NewHTTPError(http.StatusInternalServerError, "Failed to create exercise sets")
			return err
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil { // Assign to 'err' for the defer to catch it
		c.Logger().Errorf("StoreWorkoutLog: Failed to commit transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to finalize workout log creation")
	}

	// --- FINAL STEP: Fetch the complete Workout Log with all its nested data for the response ---
	// This query will join workout_logs -> logged_exercise_instances -> exercises -> exercise_sets
	finalQuery := `
		SELECT
			wl.id, wl.workout_id, wl.user_id, wl.started_at, wl.finished_at, wl.status,
			wl.total_active_duration_seconds, wl.total_pause_duration_seconds,
			wl.created_at, wl.updated_at, wl.deleted_at,
			lei.id AS lei_id, lei.workout_log_id AS lei_workout_log_id, lei.exercise_id AS lei_exercise_id,
			lei.created_at AS lei_created_at, lei.updated_at AS lei_updated_at, lei.deleted_at AS lei_deleted_at,
			e.id AS e_id, e.name AS e_name, e.created_at AS e_created_at, e.updated_at AS e_updated_at, e.deleted_at AS e_deleted_at,
			es.id AS es_id, es.workout_log_id AS es_workout_log_id, es.exercise_id AS es_exercise_id,
			es.logged_exercise_instance_id AS es_lei_id, es.set_number, es.weight, es.reps,
			es.finished_at AS es_finished_at, es.status AS es_status, es.created_at AS es_created_at, es.updated_at AS es_updated_at, es.deleted_at AS es_deleted_at
		FROM workout_logs AS wl
		LEFT JOIN logged_exercise_instances AS lei ON wl.id = lei.workout_log_id AND lei.deleted_at IS NULL
		LEFT JOIN exercises AS e ON lei.exercise_id = e.id AND e.deleted_at IS NULL
		LEFT JOIN exercise_sets AS es ON lei.id = es.logged_exercise_instance_id AND es.deleted_at IS NULL
		WHERE wl.id = $1
		ORDER BY lei.created_at ASC, es.set_number ASC;
	`
	finalRows, queryErr := h.DB.QueryContext(ctx, finalQuery, newWorkoutLogID)
	if queryErr != nil {
		c.Logger().Errorf("StoreWorkoutLog: Failed to fetch final workout log details: %v", queryErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve created workout log details.")
	}
	defer finalRows.Close()

	var finalWorkoutLog dto.WorkoutLogResponse
	var currentWorkoutLog *model.WorkoutLog
	loggedExerciseInstanceMap := make(map[uuid.UUID]dto.LoggedExerciseInstanceLog) // Map to build nested LEIs

	for finalRows.Next() {
		var (
			wlID, wlWorkoutID, wlUserID           uuid.UUID
			wlStartedAt, wlFinishedAt             sql.NullTime
			wlStatus                              int
			wlTotalActiveDurationSeconds          uint
			wlTotalPauseDurationSeconds           uint
			wlCreatedAt, wlUpdatedAt, wlDeletedAt sql.NullTime

			// Removed workout (w) fields from scan, as we use `workoutTemplate`
			leiID, leiWorkoutLogID, leiExerciseID    sql.Null[uuid.UUID]
			leiCreatedAt, leiUpdatedAt, leiDeletedAt sql.NullTime

			eID                                sql.Null[uuid.UUID]
			eName                              sql.NullString
			eCreatedAt, eUpdatedAt, eDeletedAt sql.NullTime

			esID, esWorkoutLogID, esExerciseID, esLeiID sql.Null[uuid.UUID]
			esSetNumber                                 int
			esWeight                                    sql.NullFloat64
			esReps                                      sql.NullInt64
			esFinishedAt                                sql.NullTime
			esStatus                                    sql.NullInt64 // status can be 0, 1 etc.
			esCreatedAt, esUpdatedAt, esDeletedAt       sql.NullTime
		)

		scanErr := finalRows.Scan(
			&wlID, &wlWorkoutID, &wlUserID, &wlStartedAt, &wlFinishedAt, &wlStatus,
			&wlTotalActiveDurationSeconds, &wlTotalPauseDurationSeconds,
			&wlCreatedAt, &wlUpdatedAt, &wlDeletedAt,
			&leiID, &leiWorkoutLogID, &leiExerciseID, &leiCreatedAt, &leiUpdatedAt, &leiDeletedAt,
			&eID, &eName, &eCreatedAt, &eUpdatedAt, &eDeletedAt,
			&esID, &esWorkoutLogID, &esExerciseID, &esLeiID, &esSetNumber, &esWeight, &esReps,
			&esFinishedAt, &esStatus, &esCreatedAt, &esUpdatedAt, &esDeletedAt,
		)
		if scanErr != nil {
			c.Logger().Errorf("StoreWorkoutLog: Failed to scan final workout log row: %v", scanErr)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process created workout log details.")
		}

		// Initialize top-level WorkoutLog if it hasn't been yet
		if currentWorkoutLog == nil {
			currentWorkoutLog = &model.WorkoutLog{
				ID:                         wlID,
				UserID:                     wlUserID,
				TotalActiveDurationSeconds: wlTotalActiveDurationSeconds,
				TotalPauseDurationSeconds:  wlTotalPauseDurationSeconds,
				Status:                     wlStatus,
				CreatedAt:                  wlCreatedAt.Time,
				UpdatedAt:                  wlUpdatedAt.Time,
				StartedAt:                  provider.NullTimeToTimePtr(wlStartedAt),
				FinishedAt:                 provider.NullTimeToTimePtr(wlFinishedAt),
				DeletedAt:                  provider.NullTimeToTimePtr(wlDeletedAt),
			}
			// Use the workout ID from the workout log itself, which should match req.WorkoutID
			finalWorkoutLog.WorkoutID = wlWorkoutID

			// Populate the nested Workout DTO using the already fetched `workoutTemplate`
			finalWorkoutLog.Workout = dto.WorkoutResponse{
				ID:        workoutTemplate.ID,
				UserID:    workoutTemplate.UserID,
				Name:      workoutTemplate.Name,
				CreatedAt: workoutTemplate.CreatedAt,
				UpdatedAt: workoutTemplate.UpdatedAt,
				DeletedAt: workoutTemplate.DeletedAt,
			}

			finalWorkoutLog.ID = currentWorkoutLog.ID
			finalWorkoutLog.UserID = currentWorkoutLog.UserID
			finalWorkoutLog.StartedAt = currentWorkoutLog.StartedAt
			finalWorkoutLog.FinishedAt = currentWorkoutLog.FinishedAt
			finalWorkoutLog.Status = currentWorkoutLog.Status
			finalWorkoutLog.TotalActiveDurationSeconds = currentWorkoutLog.TotalActiveDurationSeconds
			finalWorkoutLog.TotalPauseDurationSeconds = currentWorkoutLog.TotalPauseDurationSeconds
			finalWorkoutLog.CreatedAt = currentWorkoutLog.CreatedAt
			finalWorkoutLog.UpdatedAt = currentWorkoutLog.UpdatedAt
			finalWorkoutLog.DeletedAt = currentWorkoutLog.DeletedAt
		}

		// Populate LoggedExerciseInstances
		if leiID.Valid {
			leiUUID := leiID.V
			lei, exists := loggedExerciseInstanceMap[leiUUID]
			if !exists {
				lei = dto.LoggedExerciseInstanceLog{
					ID:           leiUUID,
					WorkoutLogID: leiWorkoutLogID.V,
					ExerciseID:   leiExerciseID.V,
					CreatedAt:    leiCreatedAt.Time,
					UpdatedAt:    leiUpdatedAt.Time,
					DeletedAt:    provider.NullTimeToTimePtr(leiDeletedAt),
					Exercise: dto.ExerciseResponse{
						ID:        eID.V,
						Name:      eName.String,
						CreatedAt: eCreatedAt.Time,
						UpdatedAt: eUpdatedAt.Time,
						DeletedAt: provider.NullTimeToTimePtr(eDeletedAt),
					},
					ExerciseSets: []dto.ExerciseSetResponse{}, // Initialize slice
				}
			}

			// Populate ExerciseSets for this LoggedExerciseInstance
			if esID.Valid {
				lei.ExerciseSets = append(lei.ExerciseSets, dto.ExerciseSetResponse{
					ID:           esID.V,
					WorkoutLogID: esWorkoutLogID.V,
					ExerciseID:   esExerciseID.V,
					SetNumber:    esSetNumber,
					Weight:       provider.NullFloat64ToFloat64(esWeight),
					Reps:         provider.NullInt64ToInt(esReps),
					FinishedAt:   provider.NullTimeToTimePtr(esFinishedAt),
					Status:       provider.NullInt64ToInt(esStatus),
					CreatedAt:    esCreatedAt.Time,
					UpdatedAt:    esUpdatedAt.Time,
					DeletedAt:    provider.NullTimeToTimePtr(esDeletedAt),
				})
			}
			loggedExerciseInstanceMap[leiUUID] = lei // Update map with appended set
		}
	}

	if err = finalRows.Err(); err != nil {
		c.Logger().Errorf("StoreWorkoutLog: Final rows iteration error for workout log details: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process created workout log details.")
	}

	// Convert map values to slice for the DTO
	for _, lei := range loggedExerciseInstanceMap {
		finalWorkoutLog.LoggedExerciseInstances = append(finalWorkoutLog.LoggedExerciseInstances, lei)
	}

	return c.JSON(http.StatusCreated, dto.CreateWorkoutLogResponse{
		Message:    "Workout log created successfully!",
		WorkoutLog: finalWorkoutLog,
	})
}

// Add these helper functions to your provider package if they don't exist:
// func NullFloat64ToFloat64(nf sql.NullFloat64) float64 {
// 	if nf.Valid {
// 		return nf.Float64
// 	}
// 	return 0.0 // Default or handle as appropriate for your application
// }

// func NullInt64ToInt(ni sql.NullInt64) int {
// 	if ni.Valid {
// 		return int(ni.Int64)
// 	}
// 	return 0 // Default or handle as appropriate for your application
// }
