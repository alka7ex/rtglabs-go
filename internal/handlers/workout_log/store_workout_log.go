package handler

import (
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"rtglabs-go/ent/exerciseset"
	"rtglabs-go/ent/workout"
	"rtglabs-go/ent/workoutexercise"
	"rtglabs-go/ent/workoutlog"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// WorkoutHandler struct (assuming it contains h.Client *ent.Client)
type WorkoutHandler struct {
	Client *ent.Client
	// ... other dependencies
}

// StoreWorkoutLog creates a new workout log session based on a workout template.
func (h *WorkoutHandler) StoreWorkoutLog(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	req := new(dto.CreateWorkoutLogRequest)
	if err := c.Bind(req); err != nil {
		c.Logger().Error("Failed to bind CreateWorkoutLogRequest:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		c.Logger().Error("Validation failed for CreateWorkoutLogRequest:", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx, err := h.Client.Tx(ctx)
	if err != nil {
		c.Logger().Error("Failed to begin transaction:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start workout log session")
	}
	defer tx.Rollback()

	workoutTemplate, err := tx.Workout.Get(ctx, req.WorkoutID)
	if ent.IsNotFound(err) {
		return echo.NewHTTPError(http.StatusNotFound, "Workout template not found")
	}
	if err != nil {
		c.Logger().Error("Failed to retrieve workout template:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start workout log session")
	}

	// Create the new WorkoutLog
	// Note: Entgo automatically handles the foreign key column.
	// You set the edge, and Entgo sets the ID.
	workoutLog, err := tx.WorkoutLog.
		Create().
		SetUserID(userID).                // Correct way to set the user edge by ID
		SetWorkoutID(workoutTemplate.ID). // This field is correctly directly on schema (nullable workout_id)
		SetStartedAt(time.Now()).
		SetStatus(0).
		Save(ctx)
	if err != nil {
		c.Logger().Error("Failed to create WorkoutLog:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start workout log session")
	}

	c.Logger().Infof("WorkoutLog created: %s for user %s based on workout template ID: %s",
		workoutLog.ID, userID, workoutTemplate.ID)

	workoutExercises, err := workoutTemplate.QueryWorkoutExercises().
		WithExercise().
		WithExerciseInstance(func(eiq *ent.ExerciseInstanceQuery) {
			eiq.WithExercise()
		}).
		Where(workoutexercise.DeletedAtIsNil()).
		All(ctx)
	if err != nil {
		c.Logger().Error("Failed to query workout exercises from template:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start workout log session")
	}

	var exerciseSetCreators []*ent.ExerciseSetCreate
	exerciseInstanceSetCounters := make(map[uuid.UUID]int)

	for _, we := range workoutExercises {
		if we.Edges.Exercise == nil || we.Edges.ExerciseInstance == nil {
			c.Logger().Warnf("Skipping workout exercise %s due to missing related exercise or exercise instance data.", we.ID)
			continue
		}

		templateWeight := we.Weight
		var templateReps *int
		if we.Reps != nil {
			val := int(*we.Reps)
			templateReps = &val
		}

		numSets := 1
		if we.Sets != nil {
			numSets = int(*we.Sets)
		}

		for i := 1; i <= numSets; i++ {
			instanceID := we.Edges.ExerciseInstance.ID

			exerciseInstanceSetCounters[instanceID]++
			currentSetNumber := exerciseInstanceSetCounters[instanceID]

			creator := tx.ExerciseSet.Create().
				SetWorkoutLog(workoutLog).
				SetExercise(we.Edges.Exercise).
				SetExerciseInstance(we.Edges.ExerciseInstance).
				SetSetNumber(currentSetNumber).
				SetStatus(0)

			if templateWeight != nil {
				creator.SetWeight(*templateWeight)
			}
			if templateReps != nil {
				creator.SetReps(*templateReps)
			}

			exerciseSetCreators = append(exerciseSetCreators, creator)
		}
	}

	if len(exerciseSetCreators) > 0 {
		_, err = tx.ExerciseSet.CreateBulk(exerciseSetCreators...).Save(ctx)
		if err != nil {
			c.Logger().Error("Failed to create bulk ExerciseSets:", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start workout log session")
		}
		c.Logger().Infof("Copied %d initial exercise sets from workout template %s to WorkoutLog %s",
			len(exerciseSetCreators), workoutTemplate.ID, workoutLog.ID)
	} else {
		c.Logger().Info("No initial exercise sets generated for WorkoutLog %s (template had no exercises or zero sets defined).", workoutLog.ID)
	}

	if err = tx.Commit(); err != nil {
		c.Logger().Error("Failed to commit transaction:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start workout log session")
	}

	// Reload the workout log with its new relationships for the response payload
	finalWorkoutLog, err := h.Client.WorkoutLog.Query().
		Where(workoutlog.IDEQ(workoutLog.ID)).
		WithWorkout(func(wq *ent.WorkoutQuery) {
			wq.Where(workout.DeletedAtIsNil())
		}).
		WithUser(). // Eager load the User edge to access User.ID later
		WithExerciseSets(func(esq *ent.ExerciseSetQuery) {
			esq.WithExercise()
			esq.WithExerciseInstance(func(eiq *ent.ExerciseInstanceQuery) {
				eiq.WithExercise()
			}).
				Where(exerciseset.DeletedAtIsNil()).
				Order(ent.Asc(exerciseset.FieldSetNumber))
		}).
		Only(ctx)
	if err != nil {
		c.Logger().Error("Failed to load final WorkoutLog for response:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve created workout log")
	}

	transformedWorkoutLog := toWorkoutLogResponse(finalWorkoutLog)

	return c.JSON(http.StatusCreated, dto.CreateWorkoutLogResponse{
		Message:    "Workout log session started successfully based on template.",
		WorkoutLog: transformedWorkoutLog,
	})
}

// --- Helper Conversion Functions ---

// toWorkoutLogResponse converts an *ent.WorkoutLog to dto.WorkoutLogResponse
func toWorkoutLogResponse(wl *ent.WorkoutLog) dto.WorkoutLogResponse {
	resp := dto.WorkoutLogResponse{
		ID:                         wl.ID,
		StartedAt:                  wl.StartedAt,
		FinishedAt:                 wl.FinishedAt,
		Status:                     wl.Status,
		TotalActiveDurationSeconds: wl.TotalActiveDurationSeconds,
		TotalPauseDurationSeconds:  wl.TotalPauseDurationSeconds,
		CreatedAt:                  wl.CreatedAt,
		UpdatedAt:                  wl.UpdatedAt,
		DeletedAt:                  wl.DeletedAt,
	}

	// Corrected: Access UserID via the eager-loaded Edge
	if wl.Edges.User != nil {
		resp.UserID = wl.Edges.User.ID
	}
	// Corrected: Access WorkoutID via the eager-loaded Edge or the direct field if it's there
	// Based on your schema, workout_id is a field, not an edge in workoutLog directly.
	// If it's an edge, you'd use wl.Edges.Workout.ID, but your schema defined `workout_id` as a field.
	// So, I'll assume `SetWorkoutID` was correct. If your schema has `edge.From("workout",...)` it's `wl.Edges.Workout.ID`
	if wl.Edges.Workout != nil { // Check if the edge was loaded
		resp.WorkoutID = wl.Edges.Workout.ID
	}

	if wl.Edges.Workout != nil {
		resp.Workout = toWorkoutResponse(wl.Edges.Workout)
	}

	exerciseInstancesMap := make(map[uuid.UUID]dto.ExerciseInstanceLog)
	for _, es := range wl.Edges.ExerciseSets {
		if es.Edges.ExerciseInstance == nil || es.Edges.Exercise == nil {
			continue
		}

		instanceID := es.Edges.ExerciseInstance.ID
		if _, exists := exerciseInstancesMap[instanceID]; !exists {
			instanceLog := dto.ExerciseInstanceLog{
				ID:         instanceID,
				ExerciseID: es.Edges.Exercise.ID, // Exercise ID of the instance
				Exercise:   toExerciseResponse(es.Edges.Exercise),
				ExerciseInstanceDetails: dto.ExerciseInstanceDetails{
					ID:           es.Edges.ExerciseInstance.ID,
					WorkoutLogID: &wl.ID,
					ExerciseID:   es.Edges.ExerciseInstance.Edges.Exercise.ID,
					CreatedAt:    es.Edges.ExerciseInstance.CreatedAt,
					UpdatedAt:    es.Edges.ExerciseInstance.UpdatedAt,
					DeletedAt:    es.Edges.ExerciseInstance.DeletedAt,
				},
				ExerciseSets: []dto.ExerciseSetResponse{},
			}
			exerciseInstancesMap[instanceID] = instanceLog
		}

		instanceEntry := exerciseInstancesMap[instanceID]
		instanceEntry.ExerciseSets = append(instanceEntry.ExerciseSets, toExerciseSetResponse(es))
		exerciseInstancesMap[instanceID] = instanceEntry
	}

	for _, eiLog := range exerciseInstancesMap {
		resp.ExerciseInstances = append(resp.ExerciseInstances, eiLog)
	}

	return resp
}

// toExerciseSetResponse converts an *ent.ExerciseSet to dto.ExerciseSetResponse
func toExerciseSetResponse(es *ent.ExerciseSet) dto.ExerciseSetResponse {
	resp := dto.ExerciseSetResponse{
		ID:         es.ID,
		SetNumber:  es.SetNumber,
		FinishedAt: es.FinishedAt,
		Status:     es.Status,
		CreatedAt:  es.CreatedAt,
		UpdatedAt:  es.UpdatedAt,
		DeletedAt:  es.DeletedAt,
	}

	// Corrected: Access WorkoutLogID via the eager-loaded Edge
	if es.Edges.WorkoutLog != nil {
		resp.WorkoutLogID = es.Edges.WorkoutLog.ID
	}
	// Corrected: Access ExerciseID via the eager-loaded Edge
	if es.Edges.Exercise != nil {
		resp.ExerciseID = es.Edges.Exercise.ID
	}
	// Corrected: Dereference pointer for Weight
	if es.Weight != nil {
		resp.Weight = *es.Weight // Dereference the pointer
	}
	// Corrected: Access ExerciseInstanceID via the eager-loaded Edge
	if es.Edges.ExerciseInstance != nil {
		resp.ExerciseInstanceID = &es.Edges.ExerciseInstance.ID // DTO expects *uuid.UUID, so take address of ID
	}

	if es.Reps != nil {
		resp.Reps = es.Reps
	}

	return resp
}

// toExerciseResponse converts an *ent.Exercise to dto.ExerciseResponse
func toExerciseResponse(e *ent.Exercise) dto.ExerciseResponse {
	return dto.ExerciseResponse{
		ID:        e.ID,
		Name:      e.Name,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
		DeletedAt: e.DeletedAt,
	}
}

// toWorkoutResponse converts an *ent.Workout to dto.WorkoutResponse
func toWorkoutResponse(w *ent.Workout) dto.WorkoutResponse {
	resp := dto.WorkoutResponse{
		ID:        w.ID,
		Name:      w.Name,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
		DeletedAt: w.DeletedAt,
	}
	// Corrected: Access UserID via the eager-loaded Edge
	if w.Edges.User != nil {
		resp.UserID = w.Edges.User.ID
	}
	return resp
}

