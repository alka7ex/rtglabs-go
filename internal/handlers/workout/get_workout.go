package handler // Renamed to handlers for consistency with other files

import (
	"database/sql" // For sql.DB, sql.Null* types
	"net/http"
	"time"

	"github.com/Masterminds/squirrel" // Import squirrel
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"rtglabs-go/dto"
	"rtglabs-go/model" // Import your model package (Workout, WorkoutExercise, Exercise, ExerciseInstance)
)

// GetWorkout retrieves a single workout by ID for a specific user, with eager loaded details.
func (h *WorkoutHandler) GetWorkout(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		c.Logger().Error("GetWorkout: User ID not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}

	workoutID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Logger().Errorf("GetWorkout: Invalid workout ID param: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid workout ID")
	}

	ctx := c.Request().Context()

	// Define a struct to scan the joined results into
	// This structure captures all the fields from the main workout and its eagerly loaded children.
	type joinedWorkoutResult struct {
		model.Workout // Embedding the Workout model directly for its fields

		// Fields for WorkoutExercise (aliased in query)
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

		// Fields for Exercise (aliased in query, related to workout_exercise)
		ExID        uuid.NullUUID  // exercises.id
		ExName      sql.NullString // exercises.name
		ExCreatedAt sql.NullTime   // exercises.created_at
		ExUpdatedAt sql.NullTime   // exercises.updated_at
		ExDeletedAt sql.NullTime   // exercises.deleted_at

		// Fields for ExerciseInstance (aliased in query)
		EiID           uuid.NullUUID // exercise_instances.id
		EiWorkoutLogID uuid.NullUUID // exercise_instances.workout_log_id
		EiExerciseID   uuid.NullUUID // exercise_instances.exercise_id (this is the exercise_id on the instance itself)
		EiCreatedAt    sql.NullTime  // exercise_instances.created_at
		EiUpdatedAt    sql.NullTime  // exercise_instances.updated_at
		EiDeletedAt    sql.NullTime  // exercise_instances.deleted_at
	}

	// Build the complex SELECT query with LEFT JOINs
	selectBuilder := h.sq.Select(
		// Workout fields (aliased as w)
		"w.id", "w.user_id", "w.name", "w.created_at", "w.updated_at", "w.deleted_at",
		// WorkoutExercise fields (aliased as we)
		"we.id AS we_id", "we.workout_id AS we_workout_id", "we.exercise_id AS we_exercise_id", "we.exercise_instance_id AS we_exercise_instance_id",
		"we.workout_order AS we_order", "we.sets AS we_sets", "we.weight AS we_weight", "we.reps AS we_reps",
		"we.created_at AS we_created_at", "we.updated_at AS we_updated_at", "we.deleted_at AS we_deleted_at",
		// Exercise fields (aliased as e, related to we.exercise_id)
		"e.id AS ex_id", "e.name AS ex_name", "e.created_at AS ex_created_at", "e.updated_at AS ex_updated_at", "e.deleted_at AS ex_deleted_at",
		// ExerciseInstance fields (aliased as ei, related to we.exercise_instance_id)
		"ei.id AS ei_id", "ei.workout_log_id AS ei_workout_log_id", "ei.exercise_id AS ei_exercise_id", "ei.created_at AS ei_created_at", "ei.updated_at AS ei_updated_at", "ei.deleted_at AS ei_deleted_at",
	).
		From("workouts AS w").
		LeftJoin("workout_exercises AS we ON w.id = we.workout_id AND we.deleted_at IS NULL"). // Only non-deleted workout exercises
		LeftJoin("exercises AS e ON we.exercise_id = e.id").                                   // Exercise for WorkoutExercise
		LeftJoin("exercise_instances AS ei ON we.exercise_instance_id = ei.id").
		Where(squirrel.Eq{
			"w.id":         workoutID,
			"w.user_id":    userID,
			"w.deleted_at": nil, // Workout must not be soft-deleted
		}).
		OrderBy("we.workout_order ASC", "we.created_at ASC") // Order workout exercises consistently

	selectQuery, selectArgs, err := selectBuilder.ToSql()
	if err != nil {
		c.Logger().Errorf("GetWorkout: Failed to build select query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout")
	}

	rows, err := h.DB.QueryContext(ctx, selectQuery, selectArgs...)
	if err != nil {
		c.Logger().Errorf("GetWorkout: Failed to query workout: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout")
	}
	defer rows.Close()

	// Use a map to reconstruct the single workout and its nested exercises
	// This is important because a LEFT JOIN will return multiple rows for the same workout
	// if it has multiple workout exercises.
	var workoutDTO *dto.WorkoutResponse // Will store the single workout DTO

	for rows.Next() {
		var jwr joinedWorkoutResult
		var workoutDeletedAt sql.NullTime

		err := rows.Scan(
			// Workout fields (w)
			&jwr.ID, &jwr.UserID, &jwr.Name, &jwr.CreatedAt, &jwr.UpdatedAt, &workoutDeletedAt,
			// WorkoutExercise fields (we)
			&jwr.WEID, &jwr.WEWorkoutID, &jwr.WEExerciseID, &jwr.WEExerciseInstanceID,
			&jwr.WEOrder, &jwr.WESets, &jwr.WEWeight, &jwr.WEReps,
			&jwr.WECreatedAt, &jwr.WEUpdatedAt, &jwr.WEDeletedAt,
			// Exercise fields (e)
			&jwr.ExID, &jwr.ExName, &jwr.ExCreatedAt, &jwr.ExUpdatedAt, &jwr.ExDeletedAt,
			// ExerciseInstance fields (ei)
			&jwr.EiID, &jwr.EiWorkoutLogID, &jwr.EiExerciseID, &jwr.EiCreatedAt, &jwr.EiUpdatedAt, &jwr.EiDeletedAt,
		)
		if err != nil {
			c.Logger().Errorf("GetWorkout: Failed to scan row: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout")
		}

		// Initialize the main workout DTO if this is the first row for this workout
		if workoutDTO == nil {
			workoutDTO = &dto.WorkoutResponse{
				ID:        jwr.ID,
				UserID:    jwr.UserID,
				Name:      jwr.Name,
				CreatedAt: jwr.CreatedAt,
				UpdatedAt: jwr.UpdatedAt,
				// Populate DeletedAt
				DeletedAt: func() *time.Time {
					if workoutDeletedAt.Valid {
						return &workoutDeletedAt.Time
					}
					return nil
				}(),
				WorkoutExercises: []dto.WorkoutExerciseResponse{}, // Initialize slice
			}
		}

		// If there's a WorkoutExercise for this row (LEFT JOIN might return NULLs if no WE for the workout)
		if jwr.WEID.Valid { // Check if workout_exercise fields are present
			var weModel model.WorkoutExercise
			weModel.ID = jwr.WEID.UUID
			weModel.WorkoutID = jwr.WEWorkoutID.UUID
			weModel.ExerciseID = jwr.WEExerciseID.UUID
			if jwr.WEExerciseInstanceID.Valid {
				weModel.ExerciseInstanceID = &jwr.WEExerciseInstanceID.UUID
			} else {
				weModel.ExerciseInstanceID = nil
			}
			if jwr.WEOrder.Valid {
				val := uint(jwr.WEOrder.Int64)
				weModel.WorkoutOrder = &val
			} else {
				weModel.WorkoutOrder = nil
			}
			if jwr.WESets.Valid {
				val := uint(jwr.WESets.Int64)
				weModel.Sets = &val
			} else {
				weModel.Sets = nil
			}
			if jwr.WEWeight.Valid {
				weModel.Weight = &jwr.WEWeight.Float64
			} else {
				weModel.Weight = nil
			}
			if jwr.WEReps.Valid {
				val := uint(jwr.WEReps.Int64)
				weModel.Reps = &val
			} else {
				weModel.Reps = nil
			}
			weModel.CreatedAt = jwr.WECreatedAt.Time
			weModel.UpdatedAt = jwr.WECreatedAt.Time // Use CreatedAt as UpdatedAt if not explicitly updated in WE
			if jwr.WEDeletedAt.Valid {
				weModel.DeletedAt = &jwr.WEDeletedAt.Time
			} else {
				weModel.DeletedAt = nil
			}

			var exModel model.Exercise
			// Check if exercise fields are present (from LEFT JOIN)
			if jwr.ExID.Valid {
				exModel.ID = jwr.ExID.UUID
				exModel.Name = jwr.ExName.String
				exModel.CreatedAt = jwr.ExCreatedAt.Time
				exModel.UpdatedAt = jwr.ExUpdatedAt.Time
				if jwr.ExDeletedAt.Valid {
					exModel.DeletedAt = &jwr.ExDeletedAt.Time
				} else {
					exModel.DeletedAt = nil
				}
			}

			var eiModel model.ExerciseInstance
			// Check if exercise_instance fields are present (from LEFT JOIN)
			if jwr.EiID.Valid {
				eiModel.ID = jwr.EiID.UUID
				if jwr.EiWorkoutLogID.Valid {
					eiModel.WorkoutLogID = &jwr.EiWorkoutLogID.UUID
				} else {
					eiModel.WorkoutLogID = nil
				}
				eiModel.ExerciseID = jwr.EiExerciseID.UUID
				eiModel.CreatedAt = jwr.EiCreatedAt.Time
				eiModel.UpdatedAt = jwr.EiUpdatedAt.Time
				if jwr.EiDeletedAt.Valid {
					eiModel.DeletedAt = &jwr.EiDeletedAt.Time
				} else {
					eiModel.DeletedAt = nil
				}
			}

			// Convert to WorkoutExerciseResponse DTO with nested Exercise and ExerciseInstance
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
		c.Logger().Errorf("GetWorkout: Rows iteration error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout")
	}

	// If workoutDTO is still nil after iterating, it means no workout was found.
	if workoutDTO == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Workout not found")
	}

	return c.JSON(http.StatusOK, workoutDTO) // Return the single WorkoutResponse DTO
}

