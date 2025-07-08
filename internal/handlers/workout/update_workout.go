package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"rtglabs-go/dto"
	"rtglabs-go/model"
	"rtglabs-go/provider" // Import the provider package for helper functions
)

// UpdateWorkout updates an existing workout and its associated workout exercises.
func (h *WorkoutHandler) UpdateWorkout(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		c.Logger().Error("UpdateWorkout: User ID not found in context (auth middleware missing or failed)")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required: User ID not found.")
	}

	workoutID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Logger().Warnf("UpdateWorkout: Invalid workout ID parameter format: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid workout ID format provided in URL.")
	}

	var req dto.UpdateWorkoutRequest
	if err := c.Bind(&req); err != nil {
		c.Logger().Warnf("UpdateWorkout: Invalid request body received: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body format. Please check JSON syntax.")
	}
	if err := c.Validate(&req); err != nil {
		c.Logger().Warnf("UpdateWorkout: Request validation failed: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Validation error: %v", err.Error()))
	}

	ctx := c.Request().Context()

	// --- 1. Fetch existing Workout and its WorkoutExercises to validate ownership and diff ---
	var existingWorkout model.Workout
	var existingWorkoutExercises []model.WorkoutExercise

	// Fetch workout
	workoutSelectQuery, workoutSelectArgs, buildErr := h.sq.Select(
		"w.id", "w.user_id", "w.name", "w.created_at", "w.updated_at", "w.deleted_at",
	).From("workouts AS w").
		Where(squirrel.Eq{"w.id": workoutID, "w.deleted_at": nil}).
		Where(squirrel.Eq{"w.user_id": userID}). // Validate ownership
		ToSql()
	if buildErr != nil {
		c.Logger().Errorf("UpdateWorkout: Failed to build workout select query: %v", buildErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal error: Could not build query to retrieve workout.")
	}

	var nullWorkoutDeletedAt sql.NullTime
	queryRowErr := h.DB.QueryRowContext(ctx, workoutSelectQuery, workoutSelectArgs...).Scan(
		&existingWorkout.ID, &existingWorkout.UserID, &existingWorkout.Name,
		&existingWorkout.CreatedAt, &existingWorkout.UpdatedAt, &nullWorkoutDeletedAt,
	)
	if queryRowErr != nil {
		if queryRowErr == sql.ErrNoRows {
			// Specific error for not found or not owned
			return echo.NewHTTPError(http.StatusNotFound, "Workout not found or you do not have permission to update it.")
		}
		c.Logger().Errorf("UpdateWorkout: Database error fetching existing workout %s: %v", workoutID, queryRowErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error: Failed to retrieve workout details.")
	}
	// Using provider helper
	existingWorkout.DeletedAt = provider.NullTimeToTimePtr(nullWorkoutDeletedAt)

	// Fetch existing workout exercises
	weSelectQuery, weSelectArgs, buildErr := h.sq.Select(
		"id", "workout_id", "exercise_id", "exercise_instance_id",
		"workout_order", "sets", "weight", "reps",
		"created_at", "updated_at", "deleted_at",
	).From("workout_exercises").
		Where(squirrel.Eq{"workout_id": workoutID, "deleted_at": nil}).
		ToSql()
	if buildErr != nil {
		c.Logger().Errorf("UpdateWorkout: Failed to build workout exercises select query: %v", buildErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal error: Could not build query to retrieve workout exercises.")
	}

	// Declare `rows` outside the loop, as it's used with `defer`
	rows, queryErr := h.DB.QueryContext(ctx, weSelectQuery, weSelectArgs...)
	if queryErr != nil {
		c.Logger().Errorf("UpdateWorkout: Database error querying existing workout exercises for workout %s: %v", workoutID, queryErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error: Failed to retrieve existing workout exercises.")
	}
	defer rows.Close()

	for rows.Next() {
		var we model.WorkoutExercise
		var weDeletedAt sql.NullTime
		var weOrder, weSets, weReps sql.NullInt64
		var weWeight sql.NullFloat64
		var weExerciseInstanceID sql.Null[uuid.UUID]

		scanErr := rows.Scan(
			&we.ID, &we.WorkoutID, &we.ExerciseID, &weExerciseInstanceID,
			&weOrder, &weSets, &weWeight, &weReps,
			&we.CreatedAt, &we.UpdatedAt, &weDeletedAt,
		)
		if scanErr != nil {
			c.Logger().Errorf("UpdateWorkout: Database scan error for existing workout exercise row for workout %s: %v", workoutID, scanErr)
			return echo.NewHTTPError(http.StatusInternalServerError, "Database error: Failed to process existing workout exercises.")
		}

		// --- FIX START: Use provider functions for type conversion ---
		we.DeletedAt = provider.NullTimeToTimePtr(weDeletedAt)
		we.WorkoutOrder = provider.NullInt64ToIntPtr(weOrder)
		we.Sets = provider.NullInt64ToIntPtr(weSets)
		we.Weight = provider.NullFloat64ToFloat64Ptr(weWeight)
		we.Reps = provider.NullInt64ToIntPtr(weReps)
		// --- FIX END ---

		if weExerciseInstanceID.Valid {
			we.ExerciseInstanceID = &weExerciseInstanceID.V
		} else {
			we.ExerciseInstanceID = nil
		}

		existingWorkoutExercises = append(existingWorkoutExercises, we)
	}
	if err = rows.Err(); err != nil { // Check for errors during rows iteration
		c.Logger().Errorf("UpdateWorkout: Rows iteration error for existing workout exercises for workout %s: %v", workoutID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error: Failed to read all existing workout exercises.")
	}

	// --- Start Transaction ---
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		c.Logger().Errorf("UpdateWorkout: Failed to begin transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error: Could not start transaction.")
	}
	// IMPORTANT: Named return parameter 'err' is crucial for this defer to work correctly
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.Logger().Errorf("UpdateWorkout: Recovered from panic during transaction, rolled back: %v", r)
			c.JSON(http.StatusInternalServerError, map[string]string{"message": "An unexpected error occurred."})
			// panic(r) // Re-panic if you want to crash the application
		} else if err != nil { // This 'err' is the named return variable for the function
			tx.Rollback()
			c.Logger().Errorf("UpdateWorkout: Transaction rolled back due to error: %v", err)
		}
	}()

	now := time.Now().UTC()

	// --- 2. Update Workout Name ---
	updateWorkoutBuilder := h.sq.Update("workouts").
		Set("name", req.Name).
		Set("updated_at", now).
		Where(squirrel.Eq{"id": workoutID})
	updateWorkoutQuery, updateWorkoutArgs, buildErr := updateWorkoutBuilder.ToSql()
	if buildErr != nil {
		err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build workout name update query: %v", buildErr))
		return err
	}
	_, execErr := tx.ExecContext(ctx, updateWorkoutQuery, updateWorkoutArgs...)
	if execErr != nil {
		c.Logger().Errorf("UpdateWorkout: Database error updating workout name for workout %s: %v", workoutID, execErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Database error: Failed to update workout name.")
		return err
	}

	// --- 3. Diff workout_exercises and handle updates/deletes/creations ---
	existingWEIDs := make(map[uuid.UUID]model.WorkoutExercise)
	for _, we := range existingWorkoutExercises {
		existingWEIDs[we.ID] = we
	}

	incomingWEsMap := make(map[uuid.UUID]dto.UpdateWorkoutExerciseRequest) // For quick lookup of incoming by ID
	var newOrUpdatedWEs []dto.UpdateWorkoutExerciseRequest                 // Collect all WE DTOs to process (new and updated)

	for _, exReq := range req.Exercises {
		if exReq.ID != nil {
			// If an ID is provided, check if it's a valid UUID first
			if _, parseErr := uuid.Parse(exReq.ID.String()); parseErr != nil {
				err = echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid WorkoutExercise ID format '%s'. Please provide a valid UUID.", exReq.ID))
				return err
			}
			incomingWEsMap[*exReq.ID] = exReq
			newOrUpdatedWEs = append(newOrUpdatedWEs, exReq) // Add to process as update
		} else {
			// This is a new workout exercise (no ID)
			newOrUpdatedWEs = append(newOrUpdatedWEs, exReq) // Add to process as new
		}
	}

	// Soft delete removed ones (present in existing, but not in incoming request)
	for weID := range existingWEIDs { // Iterate over keys only
		if _, found := incomingWEsMap[weID]; !found { // If existing ID is NOT in incoming
			softDeleteQuery, softDeleteArgs, buildErr := h.sq.Update("workout_exercises").
				Set("deleted_at", now).
				Where(squirrel.Eq{"id": weID}).ToSql()
			if buildErr != nil {
				err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build soft delete query for WorkoutExercise ID %s: %v", weID, buildErr))
				return err
			}
			_, execErr = tx.ExecContext(ctx, softDeleteQuery, softDeleteArgs...)
			if execErr != nil {
				c.Logger().Errorf("UpdateWorkout: Database error soft deleting workout exercise ID %s: %v", weID, execErr)
				err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Database error: Failed to remove workout exercise with ID %s.", weID))
				return err
			}
		}
	}

	exerciseInstanceMap := map[string]uuid.UUID{} // Map for client-side instance IDs for this request

	// Process New/Updated Workout Exercises
	for i, exReq := range newOrUpdatedWEs {
		var actualInstanceID uuid.UUID

		// Validate base ExerciseID exists BEFORE creating/updating instances or WE
		if _, parseErr := uuid.Parse(exReq.ExerciseID.String()); parseErr != nil {
			err = echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid exercise ID format for exercise #%d: %v", i+1, parseErr))
			return err
		}
		checkExQuery, checkExArgs, buildErr := h.sq.Select("id").From("exercises").
			Where(squirrel.Eq{"id": exReq.ExerciseID}).ToSql()
		if buildErr != nil {
			err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build exercise existence check query for exercise #%d.", i+1))
			return err
		}
		var existsID uuid.UUID
		checkErr := tx.QueryRowContext(ctx, checkExQuery, checkExArgs...).Scan(&existsID)
		if checkErr != nil {
			if checkErr == sql.ErrNoRows {
				err = echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid exercise ID '%s' found for exercise #%d. Exercise does not exist.", exReq.ExerciseID, i+1))
				return err
			}
			c.Logger().Errorf("UpdateWorkout: Database error checking exercise existence for ID %s (exercise #%d): %v", exReq.ExerciseID, i+1, checkErr)
			err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Database error checking exercise for exercise #%d.", i+1))
			return err
		}

		// Determine ExerciseInstance
		if exReq.ExerciseInstanceClientID != nil && *exReq.ExerciseInstanceClientID != "" {
			// Client provided a client ID for grouping
			if id, ok := exerciseInstanceMap[*exReq.ExerciseInstanceClientID]; ok {
				// Re-use existing instance ID from this request's map
				actualInstanceID = id
			} else {
				// Create new ExerciseInstance and store in map
				newInstanceID := uuid.New()
				insertInstanceBuilder := h.sq.Insert("exercise_instances").
					Columns("id", "exercise_id", "created_at", "updated_at").
					Values(newInstanceID, exReq.ExerciseID, now, now)

				insertInstanceQuery, insertInstanceArgs, buildErr := insertInstanceBuilder.ToSql()
				if buildErr != nil {
					err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build new exercise instance query for exercise #%d: %v", i+1, buildErr))
					return err
				}
				_, execErr = tx.ExecContext(ctx, insertInstanceQuery, insertInstanceArgs...)
				if execErr != nil {
					c.Logger().Errorf("UpdateWorkout: Database error inserting new exercise instance for exercise #%d: %v", i+1, execErr)
					err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Database error: Could not create exercise instance for exercise #%d.", i+1))
					return err
				}
				actualInstanceID = newInstanceID
				exerciseInstanceMap[*exReq.ExerciseInstanceClientID] = actualInstanceID
			}
		} else if exReq.ExerciseInstanceID != nil {
			// Client provided an existing ExerciseInstanceID.
			// Validate if this instance ID exists AND is tied to the correct base ExerciseID
			checkInstanceQuery, checkInstanceArgs, buildErr := h.sq.Select("id").From("exercise_instances").
				Where(squirrel.Eq{"id": *exReq.ExerciseInstanceID, "exercise_id": exReq.ExerciseID}).ToSql()
			if buildErr != nil {
				err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build exercise instance check query for exercise #%d.", i+1))
				return err
			}
			var existingInstanceID uuid.UUID
			checkInstanceErr := tx.QueryRowContext(ctx, checkInstanceQuery, checkInstanceArgs...).Scan(&existingInstanceID)
			if checkInstanceErr != nil {
				if checkInstanceErr == sql.ErrNoRows {
					err = echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid ExerciseInstance ID '%s' for exercise '%s' (exercise #%d). Instance does not exist or is not linked to this exercise.", *exReq.ExerciseInstanceID, exReq.ExerciseID, i+1))
					return err
				}
				c.Logger().Errorf("UpdateWorkout: Database error checking ExerciseInstance existence for ID %s (exercise #%d): %v", *exReq.ExerciseInstanceID, i+1, checkInstanceErr)
				err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Database error checking ExerciseInstance for exercise #%d.", i+1))
				return err
			}
			actualInstanceID = *exReq.ExerciseInstanceID
		} else {
			// No client ID or existing ID, create new ExerciseInstance
			newInstanceID := uuid.New()
			insertInstanceBuilder := h.sq.Insert("exercise_instances").
				Columns("id", "exercise_id", "created_at", "updated_at").
				Values(newInstanceID, exReq.ExerciseID, now, now)

			insertInstanceQuery, insertInstanceArgs, buildErr := insertInstanceBuilder.ToSql()
			if buildErr != nil {
				err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build new exercise instance query (no client ID) for exercise #%d: %v", i+1, buildErr))
				return err
			}
			_, execErr = tx.ExecContext(ctx, insertInstanceQuery, insertInstanceArgs...)
			if execErr != nil {
				c.Logger().Errorf("UpdateWorkout: Database error inserting new exercise instance (no client ID) for exercise #%d: %v", i+1, execErr)
				err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Database error: Could not create exercise instance for exercise #%d.", i+1))
				return err
			}
			actualInstanceID = newInstanceID
		}

		// Prepare values for common WE fields
		weValues := squirrel.Eq{
			"workout_id":           workoutID,
			"exercise_id":          exReq.ExerciseID,
			"exercise_instance_id": actualInstanceID,
			"updated_at":           now,
			"deleted_at":           nil, // Ensure it's not soft-deleted when updated/created
		}
		// Handle nullable fields for workout exercise details
		if exReq.WorkoutOrder != nil {
			weValues["workout_order"] = *exReq.WorkoutOrder
		} else {
			weValues["workout_order"] = nil
		}
		if exReq.Sets != nil {
			weValues["sets"] = *exReq.Sets
		} else {
			weValues["sets"] = nil
		}
		if exReq.Weight != nil {
			weValues["weight"] = *exReq.Weight
		} else {
			weValues["weight"] = nil
		}
		if exReq.Reps != nil {
			weValues["reps"] = *exReq.Reps
		} else { // Typo fix: exReq.Rps should be exReq.Reps
			weValues["reps"] = nil
		}

		if exReq.ID != nil && existingWEIDs[*exReq.ID].ID != uuid.Nil {
			// This is an update to an existing WorkoutExercise
			// Ensure the ID being updated actually belongs to THIS workout to prevent tampering
			if existingWEIDs[*exReq.ID].WorkoutID != workoutID {
				err = echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("WorkoutExercise ID '%s' does not belong to workout '%s'.", *exReq.ID, workoutID))
				return err
			}

			updateWEBuilder := h.sq.Update("workout_exercises").
				SetMap(weValues).
				Where(squirrel.Eq{"id": *exReq.ID, "workout_id": workoutID}) // Add workout_id to WHERE for safety

			updateWEQuery, updateWEArgs, buildErr := updateWEBuilder.ToSql()
			if buildErr != nil {
				err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build update query for WorkoutExercise ID %s: %v", *exReq.ID, buildErr))
				return err
			}
			_, execErr = tx.ExecContext(ctx, updateWEQuery, updateWEArgs...)
			if execErr != nil {
				c.Logger().Errorf("UpdateWorkout: Database error updating workout exercise ID %s: %v", *exReq.ID, execErr)
				err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Database error: Failed to update workout exercise with ID %s.", *exReq.ID))
				return err
			}
		} else {
			// This is a new WorkoutExercise (no ID provided in the request body for this specific WE)
			weValues["id"] = uuid.New()
			weValues["created_at"] = now // Set creation time for new records

			insertWEBuilder := h.sq.Insert("workout_exercises").SetMap(weValues)
			insertWEQuery, insertWEArgs, buildErr := insertWEBuilder.ToSql()
			if buildErr != nil {
				err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build insert query for new WorkoutExercise (exercise #%d): %v", i+1, buildErr))
				return err
			}
			_, execErr = tx.ExecContext(ctx, insertWEQuery, insertWEArgs...)
			if execErr != nil {
				c.Logger().Errorf("UpdateWorkout: Database error inserting new workout exercise (exercise #%d): %v", i+1, execErr)
				err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Database error: Failed to create new workout exercise (exercise #%d).", i+1))
				return err
			}
		}
	}

	// --- Commit the transaction ---
	if err = tx.Commit(); err != nil {
		c.Logger().Errorf("UpdateWorkout: Failed to commit transaction for workout %s: %v", workoutID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error: Failed to finalize workout update.")
	}

	// --- Fetch the final updated workout details for the response ---
	updatedWorkoutModel := &model.Workout{}
	workoutSelectQuery, workoutSelectArgs, buildErr = h.sq.Select(
		"id", "user_id", "name", "created_at", "updated_at", "deleted_at",
	).From("workouts").Where(squirrel.Eq{"id": workoutID}).ToSql()
	if buildErr != nil {
		err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build final workout select query for workout %s: %v", workoutID, buildErr))
		return err
	}

	var finalNullWorkoutDeletedAt sql.NullTime
	queryRowErr = h.DB.QueryRowContext(ctx, workoutSelectQuery, workoutSelectArgs...).Scan(
		&updatedWorkoutModel.ID, &updatedWorkoutModel.UserID, &updatedWorkoutModel.Name,
		&updatedWorkoutModel.CreatedAt, &updatedWorkoutModel.UpdatedAt, &finalNullWorkoutDeletedAt,
	)
	if queryRowErr != nil {
		c.Logger().Errorf("UpdateWorkout: Database error fetching final workout %s after update: %v", workoutID, queryRowErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Workout updated, but failed to retrieve its updated details.")
		return err
	}
	// Using provider helper
	updatedWorkoutModel.DeletedAt = provider.NullTimeToTimePtr(finalNullWorkoutDeletedAt)

	// Re-fetch workout exercises with joins, scan into joined struct
	type joinedWorkoutExercise struct {
		model.WorkoutExercise
		Exercise         model.Exercise
		ExerciseInstance model.ExerciseInstance
	}

	var joinedWorkoutExercises []joinedWorkoutExercise

	selectJoinedQuery, selectJoinedArgs, buildErr := h.sq.Select(
		// WorkoutExercise fields (aliased as we)
		"we.id", "we.workout_id", "we.exercise_id", "we.exercise_instance_id",
		"we.workout_order", "we.sets", "we.weight", "we.reps",
		"we.created_at", "we.updated_at", "we.deleted_at",
		// Exercise fields (aliased as e) - Ensure these aliases match the Scan order
		"e.id", "e.name", "e.created_at", "e.updated_at", "e.deleted_at",
		// ExerciseInstance fields (aliased as ei) - Ensure these aliases match the Scan order
		"ei.id", "ei.workout_log_id", "ei.exercise_id", "ei.created_at", "ei.updated_at", "ei.deleted_at",
	).
		From("workout_exercises AS we").
		LeftJoin("exercises AS e ON we.exercise_id = e.id").
		LeftJoin("exercise_instances AS ei ON we.exercise_instance_id = ei.id").
		Where(squirrel.Eq{"we.workout_id": workoutID}).
		Where(squirrel.Expr("we.deleted_at IS NULL")).
		OrderBy("we.created_at ASC"). // Consistent order for response
		ToSql()
	if buildErr != nil {
		err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build final joined workout exercises query: %v", buildErr))
		return err
	}

	// FIX: Declare `joinedRows` here so it's in scope for the rest of the function
	var joinedRows *sql.Rows
	joinedRows, queryErr = h.DB.QueryContext(ctx, selectJoinedQuery, selectJoinedArgs...)
	if queryErr != nil {
		c.Logger().Errorf("UpdateWorkout: Database error querying final joined WEs for workout %s: %v", workoutID, queryErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Workout updated, but failed to retrieve its updated exercise details.")
		return err
	}
	defer joinedRows.Close()

	for joinedRows.Next() {
		var weModel model.WorkoutExercise  // Scan directly into a model struct
		var exModel model.Exercise         // Scan directly into a model struct
		var eiModel model.ExerciseInstance // Scan directly into a model struct

		var weDeletedAt, exDeletedAt, eiDeletedAt sql.NullTime
		var weOrder, weSets, weReps sql.NullInt64
		var weWeight sql.NullFloat64
		var eiWorkoutLogID sql.Null[uuid.UUID]
		var weExerciseInstanceID sql.Null[uuid.UUID]

		scanErr := joinedRows.Scan(
			&weModel.ID, &weModel.WorkoutID, &weModel.ExerciseID, &weExerciseInstanceID,
			&weOrder, &weSets, &weWeight, &weReps,
			&weModel.CreatedAt, &weModel.UpdatedAt, &weDeletedAt,
			&exModel.ID, &exModel.Name, &exModel.CreatedAt, &exModel.UpdatedAt, &exDeletedAt,
			&eiModel.ID, &eiWorkoutLogID, &eiModel.ExerciseID, &eiModel.CreatedAt, &eiModel.UpdatedAt, &eiDeletedAt,
		)
		if scanErr != nil {
			c.Logger().Errorf("UpdateWorkout: Database scan error for final joined workout exercise row for workout %s: %v", workoutID, scanErr)
			return echo.NewHTTPError(http.StatusInternalServerError, "Workout updated, but failed to retrieve its exercise details accurately.")
		}

		// --- FIX START: Use provider functions for type conversion ---
		weModel.DeletedAt = provider.NullTimeToTimePtr(weDeletedAt)
		weModel.WorkoutOrder = provider.NullInt64ToIntPtr(weOrder)
		weModel.Sets = provider.NullInt64ToIntPtr(weSets)
		weModel.Weight = provider.NullFloat64ToFloat64Ptr(weWeight)
		weModel.Reps = provider.NullInt64ToIntPtr(weReps)
		// --- FIX END ---

		if weExerciseInstanceID.Valid {
			weModel.ExerciseInstanceID = &weExerciseInstanceID.V
		} else {
			weModel.ExerciseInstanceID = nil
		}

		exModel.DeletedAt = provider.NullTimeToTimePtr(exDeletedAt)

		if eiWorkoutLogID.Valid {
			eiModel.WorkoutLogID = &eiWorkoutLogID.V
		} else {
			eiModel.WorkoutLogID = nil
		}
		eiModel.DeletedAt = provider.NullTimeToTimePtr(eiDeletedAt)

		joinedWorkoutExercises = append(joinedWorkoutExercises, joinedWorkoutExercise{
			WorkoutExercise:  weModel,
			Exercise:         exModel,
			ExerciseInstance: eiModel,
		})
	}

	if err = joinedRows.Err(); err != nil {
		c.Logger().Errorf("UpdateWorkout: Final joined rows iteration error for workout exercises for workout %s: %v", workoutID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Workout updated, but encountered an error processing exercise details for response.")
	}

	// Now convert the joined data to the final DTOs
	finalWorkoutExercisesDTO := make([]dto.WorkoutExerciseResponse, len(joinedWorkoutExercises))
	for i, jwe := range joinedWorkoutExercises {
		finalWorkoutExercisesDTO[i] = toWorkoutExerciseResponse(
			&jwe.WorkoutExercise,
			&jwe.Exercise,
			&jwe.ExerciseInstance,
		)
	}

	// Create the final workout response using the populated DTOs
	finalWorkoutResponse := toWorkoutResponse(updatedWorkoutModel, finalWorkoutExercisesDTO)

	return c.JSON(http.StatusOK, dto.CreateWorkoutResponse{ // Changed to StatusOK as it's an update
		Message: "Workout updated successfully.",
		Workout: finalWorkoutResponse,
	})
}

