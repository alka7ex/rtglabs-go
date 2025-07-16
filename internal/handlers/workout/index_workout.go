package handler

import (
	"database/sql" // For sql.DB, sql.Null* types
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"rtglabs-go/dto"
	"rtglabs-go/model"    // Import your model package
	"rtglabs-go/provider" // Import your pagination provider and NullInt64ToIntPtr

	"github.com/Masterminds/squirrel" // Import squirrel
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// IndexWorkout retrieves a paginated list of workouts for a specific user, with optional filtering and sorting.
func (h *WorkoutHandler) IndexWorkout(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		c.Logger().Error("IndexWorkout: User ID not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 {
		limit = 15
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	ctx := c.Request().Context()

	// --- Build Base Query Conditions ---
	baseWhere := squirrel.And{
		squirrel.Eq{"w.deleted_at": nil}, // Workout not soft-deleted
		squirrel.Eq{"w.user_id": userID}, // Owned by the current user
	}

	// --- Add Query Param Filtering for 'name' ---
	searchName := c.QueryParam("name")
	if searchName != "" {
		// Apply a case-insensitive 'contains' filter on the 'name' field of the workout
		// Using ILike for case-insensitive LIKE
		baseWhere = append(baseWhere, squirrel.ILike{"w.name": "%" + searchName + "%"})
	}

	// --- Add Query Param Sorting ---
	sortBy := c.QueryParam("sort")
	orderBy := c.QueryParam("order")

	// Define allowed sortable columns and their corresponding SQL expressions
	// Map frontend sort names to database column names (with aliases if applicable)
	allowedSortColumns := map[string]string{
		"name":       "w.name",
		"created_at": "w.created_at",
		"updated_at": "w.updated_at", // Added updated_at as a sortable column
	}

	// Default sorting for SQL query
	sqlSortClause := []string{"w.created_at DESC", "w.id ASC"} // Default sort by creation descending, then ID ascending for stability

	if sqlColumn, ok := allowedSortColumns[sortBy]; ok {
		orderDirection := "ASC"
		if strings.ToLower(orderBy) == "desc" {
			orderDirection = "DESC"
		}
		// When dynamically sorting, ensure the secondary sort (w.id ASC) is always present for consistent pagination
		sqlSortClause = []string{fmt.Sprintf("%s %s", sqlColumn, orderDirection), "w.id ASC"}
	}

	// --- 1. Count Total Workouts ---
	countBuilder := h.sq.Select("COUNT(w.id)").From("workouts AS w").Where(baseWhere)
	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		c.Logger().Errorf("IndexWorkout: Failed to build count query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to count workouts")
	}

	var totalCount int
	err = h.DB.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		c.Logger().Errorf("IndexWorkout: Failed to count workouts: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to count workouts")
	}

	// --- 2. Fetch Workouts with Eager Loaded Relationships ---
	// Define a struct to scan the joined results into
	type joinedWorkoutResult struct {
		model.Workout
		WEID                 uuid.NullUUID   // workout_exercises.id
		WEWorkoutID          uuid.NullUUID   // workout_exercises.workout_id
		WEExerciseID         uuid.NullUUID   // workout_exercises.exercise_id
		WEExerciseInstanceID uuid.NullUUID   // workout_exercises.exercise_instance_id
		WEOrder              sql.NullInt64   // workout_exercises.workout_order
		WESets               sql.NullInt64   // workout_exercises.sets
		WEWeight             sql.NullFloat64 // workout_exercises.weight
		WEReps               sql.NullInt64   // workout_exercises.reps
		WECreatedAt          sql.NullTime    // workout_exercises.created_at
		WEUpdatedAt          sql.NullTime    // workout_exercises.updated_at
		WEDeletedAt          sql.NullTime    // workout_exercises.deleted_at

		ExID        uuid.NullUUID  // exercises.id
		ExName      sql.NullString // exercises.name
		ExCreatedAt sql.NullTime   // exercises.created_at
		ExUpdatedAt sql.NullTime   // exercises.updated_at
		ExDeletedAt sql.NullTime   // exercises.deleted_at

		EiID           uuid.NullUUID // exercise_instances.id
		EiWorkoutLogID uuid.NullUUID // exercise_instances.workout_log_id
		EiExerciseID   uuid.NullUUID // exercise_instances.exercise_id
		EiCreatedAt    sql.NullTime  // exercise_instances.created_at
		EiUpdatedAt    sql.NullTime  // exercise_instances.updated_at
		EiDeletedAt    sql.NullTime  // exercise_instances.deleted_at
	}

	selectBuilder := h.sq.Select(
		// Workout fields (aliased as w)
		"w.id", "w.user_id", "w.name", "w.created_at", "w.updated_at", "w.deleted_at",
		// WorkoutExercise fields (aliased as we)
		"we.id AS we_id", "we.workout_id AS we_workout_id", "we.exercise_id AS we_exercise_id", "we.exercise_instance_id AS we_exercise_instance_id",
		"we.workout_order AS we_order", "we.sets AS we_sets", "we.weight AS we_weight", "we.reps AS we_reps",
		"we.created_at AS we_created_at", "we.updated_at AS we_updated_at", "we.deleted_at AS we_deleted_at",
		// Exercise fields (aliased as e)
		"e.id AS ex_id", "e.name AS ex_name", "e.created_at AS ex_created_at", "e.updated_at AS ex_updated_at", "e.deleted_at AS ex_deleted_at",
		// ExerciseInstance fields (aliased as ei)
		"ei.id AS ei_id", "ei.workout_log_id AS ei_workout_log_id", "ei.exercise_id AS ei_exercise_id", "ei.created_at AS ei_created_at", "ei.updated_at AS ei_updated_at", "ei.deleted_at AS ei_deleted_at",
	).
		From("workouts AS w").
		LeftJoin("workout_exercises AS we ON w.id = we.workout_id AND we.deleted_at IS NULL"). // Only non-deleted workout exercises
		LeftJoin("exercises AS e ON we.exercise_id = e.id").
		LeftJoin("exercise_instances AS ei ON we.exercise_instance_id = ei.id").
		Where(baseWhere).
		OrderBy(sqlSortClause...). // Apply dynamic SQL sorting
		Limit(uint64(limit)).
		Offset(uint64(offset))

	selectQuery, selectArgs, err := selectBuilder.ToSql()
	if err != nil {
		c.Logger().Errorf("IndexWorkout: Failed to build select query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workouts")
	}

	rows, err := h.DB.QueryContext(ctx, selectQuery, selectArgs...)
	if err != nil {
		c.Logger().Errorf("IndexWorkout: Failed to query workouts: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workouts")
	}
	defer rows.Close()

	// Map to reconstruct the nested structure
	workoutsMap := make(map[uuid.UUID]*dto.WorkoutResponse)

	for rows.Next() {
		var jwr joinedWorkoutResult
		var workoutDeletedAt sql.NullTime

		err := rows.Scan(
			// Workout fields
			&jwr.ID, &jwr.UserID, &jwr.Name, &jwr.CreatedAt, &jwr.UpdatedAt, &workoutDeletedAt,
			// WorkoutExercise fields
			&jwr.WEID, &jwr.WEWorkoutID, &jwr.WEExerciseID, &jwr.WEExerciseInstanceID,
			&jwr.WEOrder, &jwr.WESets, &jwr.WEWeight, &jwr.WEReps,
			&jwr.WECreatedAt, &jwr.WEUpdatedAt, &jwr.WEDeletedAt,
			// Exercise fields
			&jwr.ExID, &jwr.ExName, &jwr.ExCreatedAt, &jwr.ExUpdatedAt, &jwr.ExDeletedAt,
			// ExerciseInstance fields
			&jwr.EiID, &jwr.EiWorkoutLogID, &jwr.EiExerciseID, &jwr.EiCreatedAt, &jwr.EiUpdatedAt, &jwr.EiDeletedAt,
		)
		if err != nil {
			c.Logger().Errorf("IndexWorkout: Failed to scan row: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workouts")
		}

		// Handle nullable fields for Workout
		jwr.DeletedAt = provider.NullTimeToTimePtr(workoutDeletedAt)

		// Get or create the Workout DTO
		workoutDTO, exists := workoutsMap[jwr.ID]
		if !exists {
			workoutDTO = &dto.WorkoutResponse{
				ID:               jwr.ID,
				UserID:           jwr.UserID,
				Name:             jwr.Name,
				CreatedAt:        jwr.CreatedAt,
				UpdatedAt:        jwr.UpdatedAt,
				DeletedAt:        jwr.DeletedAt,
				WorkoutExercises: []dto.WorkoutExerciseResponse{}, // Initialize slice
			}
			workoutsMap[jwr.ID] = workoutDTO
		}

		// If there's a WorkoutExercise for this row (LEFT JOIN might return NULLs if no WE)
		if jwr.WEID.Valid { // Check if workout_exercise fields are present
			var weModel model.WorkoutExercise
			weModel.ID = jwr.WEID.UUID
			weModel.WorkoutID = jwr.WEWorkoutID.UUID
			weModel.ExerciseID = jwr.WEExerciseID.UUID
			weModel.ExerciseInstanceID = provider.NullUUIDToUUIDPtr(jwr.WEExerciseInstanceID) // Using a provider helper for NullUUID

			weModel.WorkoutOrder = provider.NullInt64ToIntPtr(jwr.WEOrder)
			weModel.Sets = provider.NullInt64ToIntPtr(jwr.WESets)
			weModel.Weight = provider.NullFloat64ToFloat64Ptr(jwr.WEWeight)
			weModel.Reps = provider.NullInt64ToIntPtr(jwr.WEReps)

			weModel.CreatedAt = jwr.WECreatedAt.Time
			// UpdatedAt for WorkoutExercise can be null in DB, ensure it's handled properly
			if jwr.WEUpdatedAt.Valid {
				weModel.UpdatedAt = jwr.WEUpdatedAt.Time
			} else {
				weModel.UpdatedAt = jwr.WECreatedAt.Time // Fallback if UpdatedAt is null in DB
			}
			weModel.DeletedAt = provider.NullTimeToTimePtr(jwr.WEDeletedAt)

			var exModel model.Exercise
			// Check if exercise fields are present (from LEFT JOIN)
			if jwr.ExID.Valid {
				exModel.ID = jwr.ExID.UUID
				exModel.Name = jwr.ExName.String
				exModel.CreatedAt = jwr.ExCreatedAt.Time
				exModel.UpdatedAt = jwr.ExUpdatedAt.Time
				exModel.DeletedAt = provider.NullTimeToTimePtr(jwr.ExDeletedAt)
			}

			var eiModel model.ExerciseInstance
			// Check if exercise_instance fields are present (from LEFT JOIN)
			if jwr.EiID.Valid {
				eiModel.ID = jwr.EiID.UUID
				eiModel.WorkoutLogID = provider.NullUUIDToUUIDPtr(jwr.EiWorkoutLogID) // Using a provider helper for NullUUID
				eiModel.ExerciseID = jwr.EiExerciseID.UUID                            // Assuming always valid if EiID is valid
				eiModel.CreatedAt = jwr.EiCreatedAt.Time
				eiModel.UpdatedAt = eiModel.CreatedAt // Fallback if UpdatedAt is null in DB
				if jwr.EiUpdatedAt.Valid {            // Handle actual UpdatedAt
					eiModel.UpdatedAt = jwr.EiUpdatedAt.Time
				}
				eiModel.DeletedAt = provider.NullTimeToTimePtr(jwr.EiDeletedAt)
			}

			// Convert to WorkoutExerciseResponse DTO with nested Exercise and ExerciseInstance
			// Pass nil if the joined ID was not valid, to indicate no related entity
			weDTO := toWorkoutExerciseResponse(
				&weModel,
				func() *model.Exercise {
					if jwr.ExID.Valid {
						return &exModel
					}
					return nil
				}(),
				func() *model.ExerciseInstance {
					if jwr.EiID.Valid {
						return &eiModel
					}
					return nil
				}(),
			)
			workoutDTO.WorkoutExercises = append(workoutDTO.WorkoutExercises, weDTO)
		}
	}

	if err = rows.Err(); err != nil {
		c.Logger().Errorf("IndexWorkout: Rows iteration error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workouts")
	}

	// Convert map values to slice for final response
	dtoWorkouts := make([]dto.WorkoutResponse, 0, len(workoutsMap))
	for _, w := range workoutsMap {
		dtoWorkouts = append(dtoWorkouts, *w)
	}

	// --- IMPORTANT: In-memory sorting after map reconstruction ---
	// The map iteration order is NOT guaranteed. We must sort the final slice
	// based on the requested sort parameters.
	sort.Slice(dtoWorkouts, func(i, j int) bool {
		a := dtoWorkouts[i]
		b := dtoWorkouts[j]

		// Apply sorting based on sortBy and orderBy
		if sortBy == "name" {
			if orderBy == "desc" {
				return a.Name > b.Name
			}
			return a.Name < b.Name
		} else if sortBy == "created_at" {
			if orderBy == "desc" {
				return a.CreatedAt.After(b.CreatedAt)
			}
			return a.CreatedAt.Before(b.CreatedAt)
		} else if sortBy == "updated_at" {
			// Handle nulls for updated_at carefully.
			// If both are null, consider them equal.
			// If one is null, the non-null one comes first (for asc) or last (for desc).
			// This logic might need adjustment based on specific null handling requirements.
			if a.UpdatedAt.IsZero() && b.UpdatedAt.IsZero() {
				return false // Considered equal if both are zero/null
			}
			if a.UpdatedAt.IsZero() { // a is null, b is not
				return orderBy == "desc" // If desc, null (a) goes last, so false (a not less than b)
			}
			if b.UpdatedAt.IsZero() { // b is null, a is not
				return orderBy == "asc" // If asc, null (b) goes last, so true (a less than b)
			}
			if orderBy == "desc" {
				return a.UpdatedAt.After(b.UpdatedAt)
			}
			return a.UpdatedAt.Before(b.UpdatedAt)
		}

		// Fallback to default sort if sortBy is not recognized or not provided
		// Default: created_at DESC, then id ASC
		if a.CreatedAt.After(b.CreatedAt) {
			return true
		}
		if a.CreatedAt.Before(b.CreatedAt) {
			return false
		}
		// If CreatedAt is the same, sort by ID for stability
		return a.ID.String() < b.ID.String()
	})

	baseURL := c.Request().URL.Path
	queryParams := c.Request().URL.Query()

	// Ensure filtering and sorting query parameters are included in the queryParams for pagination links
	if searchName != "" {
		queryParams.Set("name", searchName)
	}
	if sortBy != "" {
		queryParams.Set("sort", sortBy)
	}
	if orderBy != "" {
		queryParams.Set("order", orderBy)
	}

	paginationData := provider.GeneratePaginationData(totalCount, page, limit, baseURL, queryParams)

	// Update the 'To' field based on the actual number of items in the current response
	actualItemsCount := len(dtoWorkouts)
	if actualItemsCount > 0 {
		tempTo := offset + actualItemsCount
		paginationData.To = &tempTo
	} else {
		zero := 0 // If no items, 'to' should be 0 or nil
		paginationData.To = &zero
	}

	return c.JSON(http.StatusOK, dto.ListWorkoutResponse{
		Data:               dtoWorkouts,
		PaginationResponse: paginationData, // Embed the pagination data
	})
}
