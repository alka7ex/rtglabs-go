package handler

import (
	"database/sql"
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/provider" // Assuming this contains NullTimeToTimePtr, NullInt64ToIntPtr, NullFloat64ToFloat64, NullInt64ToInt

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *WorkoutLogHandler) ShowWorkoutLog(c echo.Context) error {
	workoutLogIDStr := c.Param("id")
	workoutLogID, err := uuid.Parse(workoutLogIDStr)
	if err != nil {
		c.Logger().Warnf("ShowWorkoutLog: Invalid workout log ID format: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid workout log ID format"})
	}

	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		c.Logger().Error("ShowWorkoutLog: User ID not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}

	ctx := c.Request().Context()
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

	rows, queryErr := h.DB.QueryContext(ctx, query, workoutLogID, userID)
	if queryErr != nil {
		if queryErr == sql.ErrNoRows {
			c.Logger().Warnf("ShowWorkoutLog: Workout log not found for ID %s and UserID %s", workoutLogID, userID)
			return echo.NewHTTPError(http.StatusNotFound, "Workout log not found")
		}
		c.Logger().Errorf("ShowWorkoutLog: Failed to query workout log details: %v", queryErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout log details.")
	}
	defer rows.Close()

	var workoutLog dto.WorkoutLogResponse
	var exerciseInstances []dto.LoggedExerciseInstanceLog
	exerciseInstanceMap := make(map[uuid.UUID]int) // Maps UUID to slice index
	found := false

	for rows.Next() {
		found = true
		var (
			wlID, wlUserID                        uuid.UUID
			wlWorkoutID                           sql.Null[uuid.UUID] // Correct: use sql.Null for nullable UUID
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
			c.Logger().Errorf("ShowWorkoutLog: Failed to scan workout log row: %v", scanErr)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process workout log data.")
		}

		// Initialize workoutLog fields only once with the first row
		if workoutLog.ID == uuid.Nil { // Check if workoutLog is initialized
			workoutLog.ID = wlID
			// Handle wlWorkoutID: DTO is non-nullable UUID
			if wlWorkoutID.Valid {
				workoutLog.WorkoutID = wlWorkoutID.V
			} else {
				workoutLog.WorkoutID = uuid.Nil // Assign uuid.Nil if DB value is NULL
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

			// Populate nested WorkoutResponse and the top-level Name field
			if wID.Valid {
				workoutLog.Workout = dto.WorkoutResponse{
					ID:        wID.V,
					UserID:    wUserID.V,
					Name:      wName.String,
					CreatedAt: wCreatedAt.Time,
					UpdatedAt: wUpdatedAt.Time,
					DeletedAt: provider.NullTimeToTimePtr(wDeletedAt),
				}
				workoutLog.Name = wName.String // Assign the workout name to the top-level Name
			} else {
				// If the workout template itself is null/deleted, initialize Workout with zero values
				// and set the top-level Name to an empty string.
				workoutLog.Workout = dto.WorkoutResponse{}
				workoutLog.Name = "" // Set to empty string if no workout is associated
			}

			// Initialize the slice - will be populated below
			exerciseInstances = []dto.LoggedExerciseInstanceLog{}
		}

		// Handle logged exercise instances
		if leiID.Valid {
			leiUUID := leiID.V
			index, exists := exerciseInstanceMap[leiUUID]

			if !exists {
				// Create new exercise instance
				lei := dto.LoggedExerciseInstanceLog{
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
					ExerciseSets: []dto.ExerciseSetResponse{},
				}
				exerciseInstances = append(exerciseInstances, lei)
				exerciseInstanceMap[leiUUID] = len(exerciseInstances) - 1
				index = len(exerciseInstances) - 1
			}

			// Handle exercise sets
			if esID.Valid {
				// Add exercise set to the correct instance
				exerciseInstances[index].ExerciseSets = append(
					exerciseInstances[index].ExerciseSets,
					dto.ExerciseSetResponse{
						ID:                       esID.V,
						WorkoutLogID:             esWorkoutLogID.V,
						ExerciseID:               esExerciseID.V,
						LoggedExerciseInstanceID: esLeiID.V,
						SetNumber:                provider.NullInt64ToIntPtr(esSetNumber),
						Weight:                   provider.NullFloat64ToFloat64(esWeight),
						Reps:                     provider.NullInt64ToIntPtr(esReps),
						FinishedAt:               provider.NullTimeToTimePtr(esFinishedAt),
						Status:                   provider.NullInt64ToInt(esStatus),
						CreatedAt:                esCreatedAt.Time,
						UpdatedAt:                esUpdatedAt.Time,
						DeletedAt:                provider.NullTimeToTimePtr(esDeletedAt),
					})
			}
		}
	}

	if err := rows.Err(); err != nil {
		c.Logger().Errorf("ShowWorkoutLog: Rows iteration error for workout log details: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process workout log data.")
	}

	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "Workout log not found or not accessible.")
	}

	// Assign the ordered slice to the workout log
	workoutLog.LoggedExerciseInstances = exerciseInstances

	return c.JSON(http.StatusOK, workoutLog)
}
