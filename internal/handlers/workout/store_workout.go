package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"unsafe" // Consider removing unsafe if possible with different pointer casting or explicit type conversions

	"rtglabs-go/dto"
	"rtglabs-go/model"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// StoreWorkout creates a new workout record and its associated workout exercises.
func (h *WorkoutHandler) StoreWorkout(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		// This should ideally be caught by middleware before hitting the handler if user_id is mandatory.
		c.Logger().Error("StoreWorkout: User ID not found in context (auth middleware missing or failed)")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required: User ID not found.")
	}

	var req dto.CreateWorkoutRequest
	if err := c.Bind(&req); err != nil {
		c.Logger().Warnf("StoreWorkout: Invalid request body received: %v", err) // Use Warn for client errors
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body format. Please check JSON syntax.")
	}
	if err := c.Validate(&req); err != nil {
		c.Logger().Warnf("StoreWorkout: Request validation failed: %v", err)
		// Assuming c.Validate returns a user-friendly error message from validator
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Validation error: %v", err.Error()))
	}

	ctx := c.Request().Context()

	// Start a SQL transaction
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		c.Logger().Errorf("StoreWorkout: Failed to begin transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error: Could not start transaction.")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.Logger().Errorf("StoreWorkout: Recovered from panic during transaction, rolled back: %v", r)
			// Return a generic internal server error for panics to avoid leaking internal details
			c.JSON(http.StatusInternalServerError, map[string]string{"message": "An unexpected error occurred."})
			// Re-panic if you want the program to crash, or just log if you want it to recover gracefully
			// panic(r)
		} else if err != nil { // This 'err' is the named return variable for the function
			tx.Rollback()
			c.Logger().Errorf("StoreWorkout: Transaction rolled back due to error: %v", err)
			// The error is already an echo.HTTPError, so it will be handled by Echo's error handler
		}
	}()

	exerciseInstanceMap := make(map[string]uuid.UUID)
	var createdWorkoutID uuid.UUID

	// --- 1. Create the Workout ---
	now := time.Now().UTC()
	createdWorkoutID = uuid.New()

	insertWorkoutBuilder := h.sq.Insert("workouts").
		Columns("id", "user_id", "name", "created_at", "updated_at").
		Values(createdWorkoutID, userID, req.Name, now, now)

	insertWorkoutQuery, insertWorkoutArgs, buildErr := insertWorkoutBuilder.ToSql()
	if buildErr != nil {
		err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build workout insert query: %v", buildErr))
		return err
	}

	_, execErr := tx.ExecContext(ctx, insertWorkoutQuery, insertWorkoutArgs...)
	if execErr != nil {
		// This is where a duplicate user_id would show up if the UNIQUE constraint was still there
		c.Logger().Errorf("StoreWorkout: Failed to insert workout: %v", execErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Database error: Could not save workout.")
		return err
	}

	// --- 2. Process Workout Exercises and Exercise Instances ---
	var workoutExerciseColumns []string
	var workoutExerciseValues [][]interface{}

	workoutExerciseColumns = []string{
		"id", "workout_id", "exercise_id", "exercise_instance_id",
		"workout_order", "sets", "weight", "reps",
		"created_at", "updated_at",
	}

	for i, exReq := range req.Exercises {
		var actualInstanceID uuid.UUID
		var createInstance bool

		// Check if ExerciseID is valid UUID. This prevents SQL errors if it's not a UUID.
		if _, parseErr := uuid.Parse(exReq.ExerciseID.String()); parseErr != nil {
			err = echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid exercise ID format for exercise #%d: %v", i+1, parseErr))
			return err
		}

		if exReq.ExerciseInstanceClientID != nil && *exReq.ExerciseInstanceClientID != "" {
			if existing, found := exerciseInstanceMap[*exReq.ExerciseInstanceClientID]; found {
				actualInstanceID = existing
				createInstance = false
			} else {
				createInstance = true
			}
		} else {
			createInstance = true
		}

		if createInstance {
			// Check if ExerciseID exists (important for foreign key integrity and user feedback)
			checkExQuery, checkExArgs, buildErr := h.sq.Select("id").From("exercises").
				Where(squirrel.Eq{"id": exReq.ExerciseID}).ToSql()
			if buildErr != nil {
				err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build exercise check query for exercise #%d.", i+1))
				return err
			}
			var existsID uuid.UUID
			checkErr := tx.QueryRowContext(ctx, checkExQuery, checkExArgs...).Scan(&existsID)
			if checkErr != nil {
				if checkErr == sql.ErrNoRows {
					err = echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid exercise ID '%s' found for exercise #%d. Exercise does not exist.", exReq.ExerciseID, i+1))
					return err
				}
				c.Logger().Errorf("StoreWorkout: Database error checking exercise existence for ID %s: %v", exReq.ExerciseID, checkErr)
				err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Database error checking exercise for exercise #%d.", i+1))
				return err
			}

			// Create new ExerciseInstance
			newInstanceID := uuid.New()
			insertInstanceBuilder := h.sq.Insert("exercise_instances").
				Columns("id", "exercise_id", "created_at", "updated_at").
				Values(newInstanceID, exReq.ExerciseID, now, now)

			insertInstanceQuery, insertInstanceArgs, buildErr := insertInstanceBuilder.ToSql()
			if buildErr != nil {
				err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build exercise instance insert query for exercise #%d.", i+1))
				return err
			}
			_, execErr = tx.ExecContext(ctx, insertInstanceQuery, insertInstanceArgs...)
			if execErr != nil {
				c.Logger().Errorf("StoreWorkout: Failed to insert exercise instance for exercise #%d: %v", i+1, execErr)
				err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Database error: Could not create exercise instance for exercise #%d.", i+1))
				return err
			}
			actualInstanceID = newInstanceID

			if exReq.ExerciseInstanceClientID != nil && *exReq.ExerciseInstanceClientID != "" {
				exerciseInstanceMap[*exReq.ExerciseInstanceClientID] = actualInstanceID
			}
		}

		weID := uuid.New()
		var order, sets, reps sql.NullInt64
		var weight sql.NullFloat64

		if exReq.WorkoutOrder != nil {
			order.Valid = true
			order.Int64 = int64(*exReq.WorkoutOrder)
		}
		if exReq.Sets != nil {
			sets.Valid = true
			sets.Int64 = int64(*exReq.Sets)
		}
		if exReq.Weight != nil {
			weight.Valid = true
			weight.Float64 = *exReq.Weight
		}
		if exReq.Reps != nil {
			reps.Valid = true
			reps.Int64 = int64(*exReq.Reps)
		}

		workoutExerciseValues = append(workoutExerciseValues, []interface{}{
			weID, createdWorkoutID, exReq.ExerciseID, actualInstanceID,
			order, sets, weight, reps,
			now, now,
		})
	}

	if len(workoutExerciseValues) > 0 {
		insertWeBuilder := h.sq.Insert("workout_exercises").Columns(workoutExerciseColumns...)
		for _, vals := range workoutExerciseValues {
			insertWeBuilder = insertWeBuilder.Values(vals...)
		}
		insertWeQuery, insertWeArgs, buildErr := insertWeBuilder.ToSql()
		if buildErr != nil {
			err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build workout exercises insert query: %v", buildErr))
			return err
		}
		_, execErr = tx.ExecContext(ctx, insertWeQuery, insertWeArgs...)
		if execErr != nil {
			c.Logger().Errorf("StoreWorkout: Failed to insert workout exercises: %v", execErr)
			err = echo.NewHTTPError(http.StatusInternalServerError, "Database error: Could not save workout exercises.")
			return err
		}
	}

	// --- 3. Commit the transaction ---
	if err = tx.Commit(); err != nil {
		c.Logger().Errorf("StoreWorkout: Failed to commit transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error: Failed to finalize workout creation.")
	}

	// --- 4. Fetch the final workout details for the response ---
	workoutModel := &model.Workout{}
	workoutSelectQuery, workoutSelectArgs, buildErr := h.sq.Select(
		"id", "user_id", "name", "created_at", "updated_at", "deleted_at",
	).From("workouts").Where(squirrel.Eq{"id": createdWorkoutID}).ToSql()
	if buildErr != nil {
		err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build final workout select query: %v", buildErr))
		return err
	}

	var nullDeletedAt sql.NullTime
	queryRowErr := h.DB.QueryRowContext(ctx, workoutSelectQuery, workoutSelectArgs...).Scan(
		&workoutModel.ID, &workoutModel.UserID, &workoutModel.Name,
		&workoutModel.CreatedAt, &workoutModel.UpdatedAt, &nullDeletedAt,
	)
	if queryRowErr != nil {
		c.Logger().Errorf("StoreWorkout: Failed to fetch final workout: %v", queryRowErr)
		// This should ideally not happen after a successful commit, but good to handle
		err = echo.NewHTTPError(http.StatusInternalServerError, "Workout created, but failed to retrieve its details.")
		return err
	}
	if nullDeletedAt.Valid {
		workoutModel.DeletedAt = &nullDeletedAt.Time
	} else {
		workoutModel.DeletedAt = nil
	}

	type joinedWorkoutExercise struct {
		model.WorkoutExercise
		Exercise         model.Exercise
		ExerciseInstance model.ExerciseInstance
	}

	var joinedWorkoutExercises []joinedWorkoutExercise

	selectJoinedQuery, selectJoinedArgs, buildErr := h.sq.Select(
		"we.id", "we.workout_id", "we.exercise_id", "we.exercise_instance_id",
		"we.workout_order", "we.sets", "we.weight", "we.reps",
		"we.created_at", "we.updated_at", "we.deleted_at",
		"e.id", "e.name", "e.created_at", "e.updated_at", "e.deleted_at",
		"ei.id", "ei.workout_log_id", "ei.exercise_id", "ei.created_at", "ei.updated_at", "ei.deleted_at",
	).
		From("workout_exercises AS we").
		LeftJoin("exercises AS e ON we.exercise_id = e.id").
		LeftJoin("exercise_instances AS ei ON we.exercise_instance_id = ei.id").
		Where(squirrel.Eq{"we.workout_id": createdWorkoutID}).
		Where(squirrel.Expr("we.deleted_at IS NULL")).
		ToSql()
	if buildErr != nil {
		err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Internal error: Could not build joined workout exercises query: %v", buildErr))
		return err
	}

	joinedRows, queryErr := h.DB.QueryContext(ctx, selectJoinedQuery, selectJoinedArgs...)
	if queryErr != nil {
		c.Logger().Errorf("StoreWorkout: Failed to query joined WEs: %v", queryErr)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Workout created, but failed to retrieve its exercise details.")
		return err
	}
	defer joinedRows.Close()

	for joinedRows.Next() {
		var jwe joinedWorkoutExercise
		var weDeletedAt, exDeletedAt, eiDeletedAt sql.NullTime
		var weOrder, weSets, weReps sql.NullInt64
		var weWeight sql.NullFloat64
		var eiWorkoutLogID sql.Null[uuid.UUID]
		var weExerciseInstanceID sql.Null[uuid.UUID]

		scanErr := joinedRows.Scan(
			&jwe.ID, &jwe.WorkoutID, &jwe.ExerciseID, &weExerciseInstanceID,
			&weOrder, &weSets, &weWeight, &weReps,
			&jwe.CreatedAt, &jwe.UpdatedAt, &weDeletedAt,
			&jwe.Exercise.ID, &jwe.Exercise.Name, &jwe.Exercise.CreatedAt, &jwe.Exercise.UpdatedAt, &exDeletedAt,
			&jwe.ExerciseInstance.ID, &eiWorkoutLogID, &jwe.ExerciseInstance.ExerciseID, &jwe.ExerciseInstance.CreatedAt, &jwe.ExerciseInstance.UpdatedAt, &eiDeletedAt,
		)
		if scanErr != nil {
			c.Logger().Errorf("StoreWorkout: Failed to scan joined workout exercise row: %v", scanErr)
			err = echo.NewHTTPError(http.StatusInternalServerError, "Workout created, but failed to retrieve its exercise details accurately.")
			return err
		}

		// Map nullable values to pointers in model structs
		if weDeletedAt.Valid {
			jwe.DeletedAt = &weDeletedAt.Time
		} else {
			jwe.DeletedAt = nil
		}
		if weOrder.Valid {
			// UNSAFE: Reconsider this conversion. A safer way is to use a direct uint64 field
			// in your model and cast it, or handle it as an int64 and then convert safely.
			jwe.WorkoutOrder = (*uint)(unsafe.Pointer(&weOrder.Int64))
		} else {
			jwe.WorkoutOrder = nil
		}
		if weSets.Valid {
			jwe.Sets = (*uint)(unsafe.Pointer(&weSets.Int64))
		} else {
			jwe.Sets = nil
		}
		if weWeight.Valid {
			jwe.Weight = &weWeight.Float64
		} else {
			jwe.Weight = nil
		}
		if weReps.Valid {
			jwe.Reps = (*uint)(unsafe.Pointer(&weReps.Int64))
		} else {
			jwe.Reps = nil
		}
		if weExerciseInstanceID.Valid {
			jwe.ExerciseInstanceID = &weExerciseInstanceID.V
		} else {
			jwe.ExerciseInstanceID = nil
		}

		if exDeletedAt.Valid {
			jwe.Exercise.DeletedAt = &exDeletedAt.Time
		} else {
			jwe.Exercise.DeletedAt = nil
		}
		if eiWorkoutLogID.Valid {
			jwe.ExerciseInstance.WorkoutLogID = &eiWorkoutLogID.V
		} else {
			jwe.ExerciseInstance.WorkoutLogID = nil
		}
		if eiDeletedAt.Valid {
			jwe.ExerciseInstance.DeletedAt = &eiDeletedAt.Time
		} else {
			jwe.ExerciseInstance.DeletedAt = nil
		}

		joinedWorkoutExercises = append(joinedWorkoutExercises, jwe)
	}

	if err = joinedRows.Err(); err != nil {
		c.Logger().Errorf("StoreWorkout: Rows iteration error for joined workout exercises: %v", err)
		err = echo.NewHTTPError(http.StatusInternalServerError, "Workout created, but encountered an error processing exercise details.")
		return err
	}

	finalWorkoutExercisesDTO := make([]dto.WorkoutExerciseResponse, len(joinedWorkoutExercises))
	for i, jwe := range joinedWorkoutExercises {
		// Ensure to pass the pointers from the `jwe` struct directly
		finalWorkoutExercisesDTO[i] = toWorkoutExerciseResponse(&jwe.WorkoutExercise, &jwe.Exercise, &jwe.ExerciseInstance)
	}

	finalWorkoutResponse := toWorkoutResponse(workoutModel, finalWorkoutExercisesDTO)

	return c.JSON(http.StatusCreated, dto.CreateWorkoutResponse{
		Message: "Workout created successfully.",
		Workout: finalWorkoutResponse,
	})
}
