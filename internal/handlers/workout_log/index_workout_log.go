package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"rtglabs-go/dto"
	"rtglabs-go/provider" // For NullUUID and NullTimeToTimePtr, etc.
)

func (h *WorkoutLogHandler) IndexWorkoutLog(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		c.Logger().Error("IndexWorkoutLog: User ID not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}

	req := new(dto.ListWorkoutLogRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid query parameters for list workout log"})
	}

	// --- Apply defaults BEFORE validation ---
	if req.Page == 0 {
		req.Page = 1 // Default page
	}
	if req.Limit == 0 {
		req.Limit = 15 // Default limit
	}
	if req.Limit > 100 { // Max limit
		req.Limit = 100
	}
	// --- END FIX ---

	// Now validate the request after defaults are applied
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	// Use the potentially updated values from req directly
	page := req.Page
	limit := req.Limit
	offset := (page - 1) * limit

	ctx := c.Request().Context()

	// --- Build Base Query Conditions ---
	baseWhere := squirrel.And{
		squirrel.Eq{"wl.deleted_at": nil}, // Workout log not soft-deleted
		squirrel.Eq{"wl.user_id": userID}, // Owned by the current user
	}

	// --- Add Query Param Filtering ---
	if req.WorkoutID != nil && *req.WorkoutID != uuid.Nil {
		baseWhere = append(baseWhere, squirrel.Eq{"wl.workout_id": *req.WorkoutID})
	}
	if req.Status != nil {
		baseWhere = append(baseWhere, squirrel.Eq{"wl.status": *req.Status})
	}

	// --- 1. Count Total Workout Logs ---
	countBuilder := h.sq.Select("COUNT(DISTINCT wl.id)").From("workout_logs AS wl").Where(baseWhere)
	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		c.Logger().Errorf("IndexWorkoutLog: Failed to build count query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to count workout logs")
	}

	var totalCount int
	err = h.DB.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		c.Logger().Errorf("IndexWorkoutLog: Failed to count workout logs: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to count workout logs")
	}

	type joinedWorkoutLogResult struct {
		// WorkoutLog fields
		ID                         uuid.UUID
		UserID                     uuid.UUID
		WorkoutID                  uuid.NullUUID // Use NullUUID for nullable FK
		StartedAt                  sql.NullTime
		FinishedAt                 sql.NullTime
		Status                     int
		TotalActiveDurationSeconds uint
		TotalPauseDurationSeconds  uint
		CreatedAt                  time.Time
		UpdatedAt                  time.Time
		DeletedAt                  sql.NullTime

		// Workout fields (template)
		WID        uuid.NullUUID
		WUserID    uuid.NullUUID
		WName      sql.NullString
		WCreatedAt sql.NullTime
		WUpdatedAt sql.NullTime
		WDeletedAt sql.NullTime

		// LoggedExerciseInstance fields (aliased as lei)
		LEIID           uuid.NullUUID
		LEIWorkoutLogID uuid.NullUUID
		LEIExerciseID   uuid.NullUUID
		LEICreatedAt    sql.NullTime
		LEIUpdatedAt    sql.NullTime
		LEIDeletedAt    sql.NullTime

		// Exercise fields for LEI's exercise relationship (aliased as ex)
		ExID        uuid.NullUUID
		ExName      sql.NullString
		ExCreatedAt sql.NullTime
		ExUpdatedAt sql.NullTime
		ExDeletedAt sql.NullTime

		// ExerciseSet fields (aliased as es)
		ESID                       uuid.NullUUID
		ESWorkoutLogID             uuid.NullUUID
		ESExerciseID               uuid.NullUUID
		ESLoggedExerciseInstanceID uuid.NullUUID // Now links to logged_exercise_instances
		ESWeight                   sql.NullFloat64
		ESReps                     sql.NullInt64
		ESSetNumber                sql.NullInt64
		ESFinishedAt               sql.NullTime
		ESStatus                   sql.NullInt64
		ESCreatedAt                sql.NullTime
		ESUpdatedAt                sql.NullTime
		ESDeletedAt                sql.NullTime
	}
	// Update the SELECT statement to use 'logged_exercise_instances' and its alias 'lei'
	selectBuilder := h.sq.Select(
		// WorkoutLog fields (aliased as wl)
		"wl.id", "wl.user_id", "wl.workout_id", "wl.started_at", "wl.finished_at", "wl.status",
		"wl.total_active_duration_seconds", "wl.total_pause_duration_seconds",
		"wl.created_at", "wl.updated_at", "wl.deleted_at",
		// Workout fields (aliased as w) - selected with distinct aliases
		"w.id AS w_id", "w.user_id AS w_user_id", "w.name AS w_name",
		"w.created_at AS w_created_at", "w.updated_at AS w_updated_at", "w.deleted_at AS w_deleted_at",
		// LoggedExerciseInstance fields (aliased as lei)
		"lei.id AS lei_id", "lei.workout_log_id AS lei_workout_log_id", "lei.exercise_id AS lei_exercise_id",
		"lei.created_at AS lei_created_at", "lei.updated_at AS lei_updated_at", "lei.deleted_at AS lei_deleted_at",
		// Exercise fields (aliased as ex) for the *logged instance's* exercise
		"ex.id AS ex_id", "ex.name AS ex_name", "ex.created_at AS ex_created_at", "ex.updated_at AS ex_updated_at", "ex.deleted_at AS ex_deleted_at",
		// ExerciseSet fields (aliased as es) - now linked to logged_exercise_instances
		"es.id AS es_id", "es.workout_log_id AS es_workout_log_id", "es.exercise_id AS es_exercise_id", "es.logged_exercise_instance_id AS es_logged_exercise_instance_id",
		"es.weight AS es_weight", "es.reps AS es_reps", "es.set_number AS es_set_number", "es.finished_at AS es_finished_at", "es.status AS es_status",
		"es.created_at AS es_created_at", "es.updated_at AS es_updated_at", "es.deleted_at AS es_deleted_at",
	).
		From("workout_logs AS wl").
		LeftJoin("workouts AS w ON wl.workout_id = w.id").
		// Corrected join: now joining to 'logged_exercise_instances'
		LeftJoin("logged_exercise_instances AS lei ON wl.id = lei.workout_log_id AND lei.deleted_at IS NULL").
		// Exercise for the logged instance
		LeftJoin("exercises AS ex ON lei.exercise_id = ex.id AND ex.deleted_at IS NULL").
		// Sets for the logged instance - ExerciseSet needs FK to logged_exercise_instances
		LeftJoin("exercise_sets AS es ON lei.id = es.logged_exercise_instance_id AND es.deleted_at IS NULL").
		Where(baseWhere)

	// Apply sorting
	if req.SortBy != "" {
		order := "ASC"
		if req.Order != "" && (req.Order == "desc" || req.Order == "DESC") {
			order = "DESC"
		}
		switch req.SortBy {
		case "created_at":
			selectBuilder = selectBuilder.OrderBy("wl.created_at " + order)
		case "started_at":
			selectBuilder = selectBuilder.OrderBy("wl.started_at " + order)
		case "status":
			selectBuilder = selectBuilder.OrderBy("wl.status " + order)
		default:
			selectBuilder = selectBuilder.OrderBy("wl.created_at DESC") // Default sort
		}
	} else {
		selectBuilder = selectBuilder.OrderBy("wl.created_at DESC") // Default sort
	}

	// Add secondary ordering for consistent nested results (crucial for aggregation)
	selectBuilder = selectBuilder.OrderBy("lei.created_at ASC", "es.set_number ASC") // Sort by logged instance then set number

	selectBuilder = selectBuilder.Limit(uint64(limit)).Offset(uint64(offset))

	selectQuery, selectArgs, err := selectBuilder.ToSql()
	if err != nil {
		c.Logger().Errorf("IndexWorkoutLog: Failed to build select query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workout logs")
	}

	rows, err := h.DB.QueryContext(ctx, selectQuery, selectArgs...)
	if err != nil {
		c.Logger().Errorf("IndexWorkoutLog: Failed to query workout logs: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workout logs")
	}
	defer rows.Close()

	// Map to reconstruct the nested structure: WorkoutLog -> ExerciseInstanceLog -> ExerciseSet
	workoutLogsMap := make(map[uuid.UUID]*dto.WorkoutLogResponse)
	// No need for a separate map for loggedExerciseInstancesMap in this structure
	// We will create/retrieve it directly from workoutLogsMap entry's slice.

	for rows.Next() {
		var jwlr joinedWorkoutLogResult

		err := rows.Scan(
			// WorkoutLog fields
			&jwlr.ID, &jwlr.UserID, &jwlr.WorkoutID, &jwlr.StartedAt, &jwlr.FinishedAt, &jwlr.Status,
			&jwlr.TotalActiveDurationSeconds, &jwlr.TotalPauseDurationSeconds,
			&jwlr.CreatedAt, &jwlr.UpdatedAt, &jwlr.DeletedAt,
			// Workout fields (from JOIN)
			&jwlr.WID, &jwlr.WUserID, &jwlr.WName, &jwlr.WCreatedAt, &jwlr.WUpdatedAt, &jwlr.WDeletedAt,
			// LoggedExerciseInstance fields (from JOIN)
			&jwlr.LEIID, &jwlr.LEIWorkoutLogID, &jwlr.LEIExerciseID, &jwlr.LEICreatedAt, &jwlr.LEIUpdatedAt, &jwlr.LEIDeletedAt,
			// Exercise fields (from JOIN)
			&jwlr.ExID, &jwlr.ExName, &jwlr.ExCreatedAt, &jwlr.ExUpdatedAt, &jwlr.ExDeletedAt,
			// ExerciseSet fields (from JOIN)
			&jwlr.ESID, &jwlr.ESWorkoutLogID, &jwlr.ESExerciseID, &jwlr.ESLoggedExerciseInstanceID,
			&jwlr.ESWeight, &jwlr.ESReps, &jwlr.ESSetNumber, &jwlr.ESFinishedAt, &jwlr.ESStatus,
			&jwlr.ESCreatedAt, &jwlr.ESUpdatedAt, &jwlr.ESDeletedAt,
		)
		if err != nil {
			c.Logger().Errorf("IndexWorkoutLog: Failed to scan row: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workout logs")
		}

		// Get or create the WorkoutLog DTO
		workoutLogDTO, exists := workoutLogsMap[jwlr.ID]
		if !exists {
			workoutLogDTO = &dto.WorkoutLogResponse{
				ID:                         jwlr.ID,
				UserID:                     jwlr.UserID,
				Status:                     jwlr.Status,
				TotalActiveDurationSeconds: jwlr.TotalActiveDurationSeconds,
				TotalPauseDurationSeconds:  jwlr.TotalPauseDurationSeconds,
				CreatedAt:                  jwlr.CreatedAt,
				UpdatedAt:                  jwlr.UpdatedAt,
				LoggedExerciseInstances:    []dto.LoggedExerciseInstanceLog{}, // Initialize slice
			}

			// Assign nullable time fields for WorkoutLog
			workoutLogDTO.StartedAt = provider.NullTimeToTimePtr(jwlr.StartedAt)
			workoutLogDTO.FinishedAt = provider.NullTimeToTimePtr(jwlr.FinishedAt)
			workoutLogDTO.DeletedAt = provider.NullTimeToTimePtr(jwlr.DeletedAt)

			// Handle WorkoutID (non-nullable UUID in DTO)
			if jwlr.WorkoutID.Valid {
				workoutLogDTO.WorkoutID = jwlr.WorkoutID.UUID
			} else {
				workoutLogDTO.WorkoutID = uuid.Nil // Assign zero UUID if DB value is NULL
			}

			// Handle nested Workout DTO
			if jwlr.WID.Valid { // Check if workout was joined successfully
				workoutLogDTO.Workout = dto.WorkoutResponse{
					ID:        jwlr.WID.UUID,
					UserID:    jwlr.WUserID.UUID, // Assuming WUserID is valid if WID is valid
					Name:      jwlr.WName.String,
					CreatedAt: jwlr.WCreatedAt.Time,
					UpdatedAt: jwlr.WUpdatedAt.Time,
					DeletedAt: provider.NullTimeToTimePtr(jwlr.WDeletedAt),
				}
			} else {
				// If no workout is associated, set Workout to its zero value.
				workoutLogDTO.Workout = dto.WorkoutResponse{}
			}

			workoutLogsMap[jwlr.ID] = workoutLogDTO
		}

		// If there's a LoggedExerciseInstance for this row (LEFT JOIN might return NULLs)
		if jwlr.LEIID.Valid {
			leiUUID := jwlr.LEIID.UUID
			// Find or create the LoggedExerciseInstance within the workoutLogDTO's slice
			var leiDTO *dto.LoggedExerciseInstanceLog
			foundLei := false
			for i := range workoutLogDTO.LoggedExerciseInstances {
				if workoutLogDTO.LoggedExerciseInstances[i].ID == leiUUID {
					leiDTO = &workoutLogDTO.LoggedExerciseInstances[i]
					foundLei = true
					break
				}
			}

			if !foundLei { // If not found, create a new one and append
				newLei := dto.LoggedExerciseInstanceLog{
					ID:           leiUUID,
					WorkoutLogID: jwlr.LEIWorkoutLogID.UUID,
					ExerciseID:   jwlr.LEIExerciseID.UUID,
					CreatedAt:    jwlr.LEICreatedAt.Time, // Direct .Time access
					UpdatedAt:    jwlr.LEIUpdatedAt.Time, // Direct .Time access
					DeletedAt:    provider.NullTimeToTimePtr(jwlr.LEIDeletedAt),
					ExerciseSets: []dto.ExerciseSetResponse{}, // Initialize slice
				}

				// Populate Exercise field
				if jwlr.ExID.Valid {
					newLei.Exercise = dto.ExerciseResponse{
						ID:        jwlr.ExID.UUID,
						Name:      jwlr.ExName.String,
						CreatedAt: jwlr.ExCreatedAt.Time,
						UpdatedAt: jwlr.ExUpdatedAt.Time,
						DeletedAt: provider.NullTimeToTimePtr(jwlr.ExDeletedAt),
					}
				}
				workoutLogDTO.LoggedExerciseInstances = append(workoutLogDTO.LoggedExerciseInstances, newLei)
				// Get a pointer to the newly appended element for further modification
				leiDTO = &workoutLogDTO.LoggedExerciseInstances[len(workoutLogDTO.LoggedExerciseInstances)-1]
			}

			// If there's an ExerciseSet for this row (LEFT JOIN might return NULLs)
			if jwlr.ESID.Valid {
				esDTO := dto.ExerciseSetResponse{
					ID:                       jwlr.ESID.UUID,
					WorkoutLogID:             jwlr.ESWorkoutLogID.UUID,
					ExerciseID:               jwlr.ESExerciseID.UUID,
					LoggedExerciseInstanceID: jwlr.ESLoggedExerciseInstanceID.UUID,
					Status:                   provider.NullInt64ToInt(jwlr.ESStatus), // Non-nullable int
					CreatedAt:                jwlr.ESCreatedAt.Time,                  // Direct .Time access
					UpdatedAt:                jwlr.ESUpdatedAt.Time,                  // Direct .Time access
				}
				// Handle nullable fields from DB scan results, converting to *int
				esDTO.SetNumber = provider.NullInt64ToIntPtr(jwlr.ESSetNumber)
				esDTO.Reps = provider.NullInt64ToIntPtr(jwlr.ESReps)
				esDTO.Weight = provider.NullFloat64ToFloat64(jwlr.ESWeight) // Non-nullable float64
				esDTO.FinishedAt = provider.NullTimeToTimePtr(jwlr.ESFinishedAt)
				esDTO.DeletedAt = provider.NullTimeToTimePtr(jwlr.ESDeletedAt)

				leiDTO.ExerciseSets = append(leiDTO.ExerciseSets, esDTO)
			}
		}
	}

	if err = rows.Err(); err != nil {
		c.Logger().Errorf("IndexWorkoutLog: Rows iteration error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workout logs")
	}

	// Convert map values to slice for final response
	dtoWorkoutLogs := make([]dto.WorkoutLogResponse, 0, len(workoutLogsMap))
	for _, wl := range workoutLogsMap {
		dtoWorkoutLogs = append(dtoWorkoutLogs, *wl)
	}

	baseURL := c.Request().URL.Path
	queryParams := c.Request().URL.Query()

	// Ensure the query parameters are included in the pagination links
	if req.WorkoutID != nil {
		queryParams.Set("workout_id", req.WorkoutID.String())
	}
	if req.Status != nil {
		queryParams.Set("status", strconv.Itoa(*req.Status))
	}
	if req.SortBy != "" {
		queryParams.Set("sort_by", req.SortBy)
	}
	if req.Order != "" {
		queryParams.Set("order", req.Order)
	}

	paginationData := provider.GeneratePaginationData(totalCount, page, limit, baseURL, queryParams)

	// Update the 'To' field based on the actual number of items in the current response
	actualItemsCount := len(dtoWorkoutLogs)
	if actualItemsCount > 0 {
		tempTo := offset + actualItemsCount
		paginationData.To = &tempTo
	} else {
		zero := 0 // If no items, 'to' should be 0 or nil
		paginationData.To = &zero
	}

	return c.JSON(http.StatusOK, dto.ListWorkoutLogResponse{
		Data:               dtoWorkoutLogs,
		PaginationResponse: paginationData,
	})
}
