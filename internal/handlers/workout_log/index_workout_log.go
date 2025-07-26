package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings" // Import for strings.ToLower and strings.Contains

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"rtglabs-go/dto"
	"rtglabs-go/provider"
	"time" // Ensure time is imported
)

// Ensure your WorkoutLogHandler struct has a squirrel.StatementBuilderType field, e.g., `sq squirrel.StatementBuilderType`

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

	page := req.Page
	limit := req.Limit
	offset := (page - 1) * limit

	ctx := c.Request().Context()

	// --- Build Base Query Conditions for WorkoutLogs ---
	wlWhere := squirrel.And{
		squirrel.Eq{"wl.deleted_at": nil}, // Workout log not soft-deleted
		squirrel.Eq{"wl.user_id": userID}, // Owned by the current user
	}

	// --- Add Query Param Filtering for WorkoutLogs ---
	if req.WorkoutID != nil && *req.WorkoutID != uuid.Nil {
		wlWhere = append(wlWhere, squirrel.Eq{"wl.workout_id": *req.WorkoutID})
	}
	if req.Status != nil {
		wlWhere = append(wlWhere, squirrel.Eq{"wl.status": *req.Status})
	}

	// --- NEW: Handle filtering by workout name ---
	// If `name` filter is present, we need to join the `workouts` table
	// in both the count and ID selection queries.
	var nameFilterActive bool
	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		// Use ILIKE for case-insensitive partial match for PostgreSQL, or LIKE for MySQL/SQLite
		// For cross-database compatibility, convert the search term to lowercase and use LIKE.
		// Alternatively, if you know your DB, use specific functions like LOWER() or COLLATE.
		wlWhere = append(wlWhere, squirrel.Like{"LOWER(w.name)": "%" + strings.ToLower(strings.TrimSpace(*req.Name)) + "%"})
		nameFilterActive = true
	}

	// --- 1. Count Total Workout Logs (DISTINCT) ---
	countBuilder := h.sq.Select("COUNT(wl.id)").From("workout_logs AS wl").Where(wlWhere)
	if nameFilterActive { // Only join if name filter is active
		countBuilder = countBuilder.LeftJoin("workouts AS w ON wl.workout_id = w.id AND w.deleted_at IS NULL")
	}

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

	// If no workout logs found, return early
	if totalCount == 0 {
		paginationData := provider.GeneratePaginationData(totalCount, page, limit, c.Request().URL.Path, c.Request().URL.Query())
		return c.JSON(http.StatusOK, dto.ListWorkoutLogResponse{
			Data:               []dto.WorkoutLogResponse{},
			PaginationResponse: paginationData,
		})
	}

	// --- 2. Select PAGINATED Workout Log IDs ---
	wlIDsBuilder := h.sq.Select("wl.id").From("workout_logs AS wl").Where(wlWhere)
	if nameFilterActive { // Only join if name filter is active
		wlIDsBuilder = wlIDsBuilder.LeftJoin("workouts AS w ON wl.workout_id = w.id AND w.deleted_at IS NULL")
	}

	// Apply sorting to the primary workout logs
	orderCol := "wl.created_at" // Default sort column
	orderDir := "DESC"          // Default sort direction

	if req.SortBy != "" {
		switch strings.ToLower(req.SortBy) {
		case "created_at":
			orderCol = "wl.created_at"
		case "started_at":
			orderCol = "wl.started_at"
		case "status":
			orderCol = "wl.status"
		case "name": // NEW: Sort by workout name
			orderCol = "w.name"
			if !nameFilterActive { // If sorting by name, we MUST join even if not filtering by name
				wlIDsBuilder = wlIDsBuilder.LeftJoin("workouts AS w ON wl.workout_id = w.id AND w.deleted_at IS NULL")
			}
		default:
			orderCol = "wl.created_at" // Fallback to default
		}
	}

	if req.Order != "" && (strings.ToLower(req.Order) == "asc" || strings.ToLower(req.Order) == "desc") {
		orderDir = strings.ToUpper(req.Order)
	}
	wlIDsBuilder = wlIDsBuilder.OrderBy(orderCol + " " + orderDir)

	wlIDsBuilder = wlIDsBuilder.Limit(uint64(limit)).Offset(uint64(offset))

	wlIDsQuery, wlIDsArgs, err := wlIDsBuilder.ToSql()
	if err != nil {
		c.Logger().Errorf("IndexWorkoutLog: Failed to build workout log IDs query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workout log IDs")
	}

	wlIDsRows, err := h.DB.QueryContext(ctx, wlIDsQuery, wlIDsArgs...)
	if err != nil {
		c.Logger().Errorf("IndexWorkoutLog: Failed to query workout log IDs: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workout log IDs")
	}
	defer wlIDsRows.Close()

	var workoutLogIDs []uuid.UUID
	for wlIDsRows.Next() {
		var id uuid.UUID
		if err := wlIDsRows.Scan(&id); err != nil {
			c.Logger().Errorf("IndexWorkoutLog: Failed to scan workout log ID: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workout log IDs")
		}
		workoutLogIDs = append(workoutLogIDs, id)
	}
	if err = wlIDsRows.Err(); err != nil {
		c.Logger().Errorf("IndexWorkoutLog: Rows iteration error for workout log IDs: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workout log IDs")
	}

	// If no IDs were retrieved for the current page (e.g., page beyond total), return empty
	if len(workoutLogIDs) == 0 {
		paginationData := provider.GeneratePaginationData(totalCount, page, limit, c.Request().URL.Path, c.Request().URL.Query())
		return c.JSON(http.StatusOK, dto.ListWorkoutLogResponse{
			Data:               []dto.WorkoutLogResponse{},
			PaginationResponse: paginationData,
		})
	}

	// --- 3. Fetch all data for the retrieved Workout Log IDs ---
	// This struct is fine as is for scanning from the joined query
	type joinedWorkoutLogResult struct {
		ID                         uuid.UUID
		UserID                     uuid.UUID
		WorkoutID                  uuid.NullUUID
		StartedAt                  sql.NullTime
		FinishedAt                 sql.NullTime
		Status                     int
		TotalActiveDurationSeconds uint
		TotalPauseDurationSeconds  uint
		CreatedAt                  time.Time
		UpdatedAt                  time.Time
		DeletedAt                  sql.NullTime

		WID        uuid.NullUUID
		WUserID    uuid.NullUUID
		WName      sql.NullString // <--- This holds the workout name
		WCreatedAt sql.NullTime
		WUpdatedAt sql.NullTime
		WDeletedAt sql.NullTime

		LEIID           uuid.NullUUID
		LEIWorkoutLogID uuid.NullUUID
		LEIExerciseID   uuid.NullUUID
		LEICreatedAt    sql.NullTime
		LEIUpdatedAt    sql.NullTime
		LEIDeletedAt    sql.NullTime

		ExID        uuid.NullUUID
		ExName      sql.NullString
		ExCreatedAt sql.NullTime
		ExUpdatedAt sql.NullTime
		ExDeletedAt sql.NullTime

		ESID                       uuid.NullUUID
		ESWorkoutLogID             uuid.NullUUID
		ESExerciseID               uuid.NullUUID
		ESLoggedExerciseInstanceID uuid.NullUUID
		ESWeight                   sql.NullFloat64
		ESReps                     sql.NullInt64
		ESSetNumber                sql.NullInt64
		ESFinishedAt               sql.NullTime
		ESStatus                   sql.NullInt64
		ESCreatedAt                sql.NullTime
		ESUpdatedAt                sql.NullTime
		ESDeletedAt                sql.NullTime
	}

	// Build the main data query using the fetched workoutLogIDs
	// We always join 'workouts' here because we need its data for the response DTO
	mainSelectBuilder := h.sq.Select(
		"wl.id", "wl.user_id", "wl.workout_id", "wl.started_at", "wl.finished_at", "wl.status",
		"wl.total_active_duration_seconds", "wl.total_pause_duration_seconds",
		"wl.created_at", "wl.updated_at", "wl.deleted_at",
		"w.id AS w_id", "w.user_id AS w_user_id", "w.name AS w_name", // <--- Select w.name
		"w.created_at AS w_created_at", "w.updated_at AS w_updated_at", "w.deleted_at AS w_deleted_at",
		"lei.id AS lei_id", "lei.workout_log_id AS lei_workout_log_id", "lei.exercise_id AS lei_exercise_id",
		"lei.created_at AS lei_created_at", "lei.updated_at AS lei_updated_at", "lei.deleted_at AS lei_deleted_at",
		"ex.id AS ex_id", "ex.name AS ex_name", "ex.created_at AS ex_created_at", "ex.updated_at AS ex_updated_at", "ex.deleted_at AS ex_deleted_at",
		"es.id AS es_id", "es.workout_log_id AS es_workout_log_id", "es.exercise_id AS es_exercise_id", "es.logged_exercise_instance_id AS es_logged_exercise_instance_id",
		"es.weight AS es_weight", "es.reps AS es_reps", "es.set_number AS es_set_number", "es.finished_at AS es_finished_at", "es.status AS es_status",
		"es.created_at AS es_created_at", "es.updated_at AS es_updated_at", "es.deleted_at AS es_deleted_at",
	).
		From("workout_logs AS wl").
		LeftJoin("workouts AS w ON wl.workout_id = w.id AND w.deleted_at IS NULL"). // Always join for main data query
		LeftJoin("logged_exercise_instances AS lei ON wl.id = lei.workout_log_id AND lei.deleted_at IS NULL").
		LeftJoin("exercises AS ex ON lei.exercise_id = ex.id AND ex.deleted_at IS NULL").
		LeftJoin("exercise_sets AS es ON lei.id = es.logged_exercise_instance_id AND es.deleted_at IS NULL").
		// Crucially, filter by the IDs we just paginated
		Where(squirrel.Eq{"wl.id": workoutLogIDs}).
		// Maintain ordering for consistent aggregation
		OrderBy(orderCol+" "+orderDir, "lei.created_at ASC", "es.set_number ASC") // Apply same primary order as ID query

	mainQuery, mainArgs, err := mainSelectBuilder.ToSql()
	if err != nil {
		c.Logger().Errorf("IndexWorkoutLog: Failed to build main select query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workout logs details")
	}

	mainDataRows, err := h.DB.QueryContext(ctx, mainQuery, mainArgs...)
	if err != nil {
		c.Logger().Errorf("IndexWorkoutLog: Failed to query workout logs details: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workout logs details")
	}
	defer mainDataRows.Close()

	// Map to reconstruct the nested structure: WorkoutLog -> ExerciseInstanceLog -> ExerciseSet
	workoutLogsMap := make(map[uuid.UUID]*dto.WorkoutLogResponse)

	for mainDataRows.Next() {
		var jwlr joinedWorkoutLogResult

		err := mainDataRows.Scan(
			&jwlr.ID, &jwlr.UserID, &jwlr.WorkoutID, &jwlr.StartedAt, &jwlr.FinishedAt, &jwlr.Status,
			&jwlr.TotalActiveDurationSeconds, &jwlr.TotalPauseDurationSeconds,
			&jwlr.CreatedAt, &jwlr.UpdatedAt, &jwlr.DeletedAt,
			&jwlr.WID, &jwlr.WUserID, &jwlr.WName, &jwlr.WCreatedAt, &jwlr.WUpdatedAt, &jwlr.WDeletedAt, // <--- Scan w.name here
			&jwlr.LEIID, &jwlr.LEIWorkoutLogID, &jwlr.LEIExerciseID, &jwlr.LEICreatedAt, &jwlr.LEIUpdatedAt, &jwlr.LEIDeletedAt,
			&jwlr.ExID, &jwlr.ExName, &jwlr.ExCreatedAt, &jwlr.ExUpdatedAt, &jwlr.ExDeletedAt,
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

			// Handle nested Workout DTO and top-level Name
			if jwlr.WID.Valid { // Check if workout was joined successfully
				workoutLogDTO.Workout = dto.WorkoutResponse{
					ID:        jwlr.WID.UUID,
					UserID:    jwlr.WUserID.UUID, // Assuming WUserID is valid if WID is valid
					Name:      jwlr.WName.String,
					CreatedAt: jwlr.WCreatedAt.Time,
					UpdatedAt: jwlr.WUpdatedAt.Time,
					DeletedAt: provider.NullTimeToTimePtr(jwlr.WDeletedAt),
				}
				workoutLogDTO.Name = jwlr.WName.String // <--- Assign workout name to top-level Name
			} else {
				// If no workout is associated, set Workout to its zero value and Name to empty string.
				workoutLogDTO.Workout = dto.WorkoutResponse{}
				workoutLogDTO.Name = "" // Set to empty string if no workout is associated
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

	if err = mainDataRows.Err(); err != nil {
		c.Logger().Errorf("IndexWorkoutLog: Rows iteration error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workout logs")
	}

	// Convert map values to slice for final response.
	// IMPORTANT: Iterate over the workoutLogIDs slice to maintain the original pagination order.
	dtoWorkoutLogs := make([]dto.WorkoutLogResponse, 0, len(workoutLogIDs))
	for _, id := range workoutLogIDs {
		if wl, ok := workoutLogsMap[id]; ok {
			dtoWorkoutLogs = append(dtoWorkoutLogs, *wl)
		}
	}

	baseURL := c.Request().URL.Path
	queryParams := c.Request().URL.Query()

	// Ensure all relevant query parameters are included in the pagination links
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
	if req.Name != nil && strings.TrimSpace(*req.Name) != "" { // NEW: Add name to query params for pagination
		queryParams.Set("name", *req.Name)
	}

	paginationData := provider.GeneratePaginationData(totalCount, page, limit, baseURL, queryParams)

	return c.JSON(http.StatusOK, dto.ListWorkoutLogResponse{
		Data:               dtoWorkoutLogs,
		PaginationResponse: paginationData,
	})
}

