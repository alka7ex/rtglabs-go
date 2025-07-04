package handler

import (
	"fmt"
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"rtglabs-go/ent/exercise"
	"rtglabs-go/ent/exerciseset"
	"rtglabs-go/ent/workoutlog"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *WorkoutHandler) UpdateWorkoutLog(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	logID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid workout log ID")
	}

	var req dto.UpdateWorkoutLogRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	seenClientIDs := map[string]bool{}
	for _, inst := range req.ExerciseInstances {
		if inst.ExerciseInstanceClientID != nil {
			if seenClientIDs[*inst.ExerciseInstanceClientID] {
				return echo.NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("Duplicate exercise_instance_client_id: %s", *inst.ExerciseInstanceClientID))
			}
			seenClientIDs[*inst.ExerciseInstanceClientID] = true
		}
	}

	tx, err := h.Client.Tx(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start transaction")
	}
	defer tx.Rollback()

	logTx, err := tx.WorkoutLog.Query().
		Where(workoutlog.IDEQ(logID)).
		WithUser().
		WithExerciseSets(func(q *ent.ExerciseSetQuery) {
			q.WithExerciseInstance()
		}).
		Only(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Workout log not found")
	}

	if logTx.Edges.User.ID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "You are not authorized to modify this log")
	}

	if req.FinishedAt != nil {
		if _, err := tx.WorkoutLog.UpdateOneID(logID).SetFinishedAt(*req.FinishedAt).Save(ctx); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update finished_at")
		}
	}

	// Map current sets
	currentSets := make(map[uuid.UUID]*ent.ExerciseSet)
	for _, set := range logTx.Edges.ExerciseSets {
		currentSets[set.ID] = set
	}

	// Collect incoming set IDs
	incomingSetIDs := map[uuid.UUID]bool{}
	for _, inst := range req.ExerciseInstances {
		for _, s := range inst.ExerciseSets {
			if s.ID != nil {
				incomingSetIDs[*s.ID] = true
			}
		}
	}

	// Soft-delete removed sets
	for id := range currentSets {
		if !incomingSetIDs[id] {
			if err := tx.ExerciseSet.DeleteOneID(id).Exec(ctx); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete exercise set")
			}
		}
	}

	// Map for clientID → server UUID
	clientToServerInstanceID := make(map[string]uuid.UUID)

	// Process instances and sets
	for _, inst := range req.ExerciseInstances {
		var instanceID uuid.UUID

		if inst.ID != nil {
			instanceID = *inst.ID
		} else {
			// Ensure the exercise_id exists
			exists, err := tx.Exercise.Query().
				Where(exercise.IDEQ(inst.ExerciseID)).
				Exist(ctx)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify exercise_id")
			}
			if !exists {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Exercise ID %s does not exist", inst.ExerciseID))
			}

			// Create new instance
			newInst, err := tx.ExerciseInstance.Create().
				SetExerciseID(inst.ExerciseID).
				SetWorkoutLogID(logID).
				Save(ctx)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create exercise instance")
			}
			instanceID = newInst.ID

			// Track mapping
			if inst.ExerciseInstanceClientID != nil {
				clientToServerInstanceID[*inst.ExerciseInstanceClientID] = instanceID
			}
		}

		// Create/update sets
		for _, s := range inst.ExerciseSets {
			setNumber := 1
			if s.SetNumber > 0 {
				setNumber = s.SetNumber
			}

			if s.ID != nil {
				oldSet, exists := currentSets[*s.ID]
				if !exists {
					return echo.NewHTTPError(http.StatusNotFound,
						fmt.Sprintf("Exercise set %s not found in current workout log", s.ID.String()))
				}

				if _, err := tx.ExerciseSet.UpdateOneID(oldSet.ID).
					SetExerciseID(s.ExerciseID).
					SetWorkoutLogID(logID).
					SetExerciseInstanceID(instanceID).
					SetNillableWeight(s.Weight).
					SetNillableReps(s.Reps).
					SetSetNumber(setNumber).
					SetNillableFinishedAt(s.FinishedAt).
					Save(ctx); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update exercise set")
				}
			} else {
				finalInstanceID := instanceID
				if s.ExerciseInstanceID == nil && inst.ExerciseInstanceClientID != nil {
					if resolvedID, ok := clientToServerInstanceID[*inst.ExerciseInstanceClientID]; ok {
						finalInstanceID = resolvedID
					}
				}

				create := tx.ExerciseSet.Create().
					SetExerciseID(s.ExerciseID).
					SetWorkoutLogID(logID).
					SetExerciseInstanceID(finalInstanceID).
					SetSetNumber(setNumber).
					SetNillableWeight(s.Weight).
					SetNillableReps(s.Reps).
					SetNillableFinishedAt(s.FinishedAt).
					SetStatus(0)
				if _, err := create.Save(ctx); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create exercise set")
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Transaction commit failed")
	}

	// Reload updated workout log
	updatedLog, err := h.Client.WorkoutLog.Query().
		Where(workoutlog.IDEQ(logID)).
		WithWorkout(func(wq *ent.WorkoutQuery) { // Add a function to WithWorkout
			wq.WithUser() // Eager load the User from the Workout
		}).
		WithUser(). // Already there for the WorkoutLog's direct user
		WithExerciseSets(func(q *ent.ExerciseSetQuery) {
			q.WithExercise().
				WithExerciseInstance(func(qi *ent.ExerciseInstanceQuery) {
					qi.WithExercise()
				}).
				WithWorkoutLog(). // Eager load the WorkoutLog from the ExerciseSet
				Order(ent.Asc(exerciseset.FieldSetNumber))
		}).
		Only(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reload updated workout log")
	}
	return c.JSON(http.StatusOK, dto.UpdateWorkoutLogResponse{
		Message:    "Workout log updated successfully.",
		WorkoutLog: toWorkoutLogResponse(updatedLog),
	})
}
