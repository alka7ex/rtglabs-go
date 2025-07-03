package handler

import (
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"rtglabs-go/ent/user"
	"rtglabs-go/ent/workout"
	"rtglabs-go/ent/workoutexercise"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *WorkoutHandler) UpdateWorkout(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}

	workoutID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid workout ID")
	}

	var req dto.CreateWorkoutRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Fetch workout & validate ownership
	existingWorkout, err := h.Client.Workout.
		Query().
		Where(
			workout.IDEQ(workoutID),
			workout.DeletedAtIsNil(),
			workout.HasUserWith(user.IDEQ(userID)),
		).
		WithWorkoutExercises(func(wq *ent.WorkoutExerciseQuery) {
			wq.Where(workoutexercise.DeletedAtIsNil())
		}).
		Only(c.Request().Context())

	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "Workout not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout")
	}

	ctx := c.Request().Context()
	tx, err := h.Client.Tx(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 1. Update name
	if _, err := tx.Workout.
		UpdateOneID(workoutID).
		SetName(req.Name).
		Save(ctx); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update workout name")
	}

	// 2. Diff workout_exercises
	existingIDs := map[uuid.UUID]bool{}
	for _, we := range existingWorkout.Edges.WorkoutExercises {
		existingIDs[we.ID] = true
	}

	incomingIDs := map[uuid.UUID]bool{}
	for _, ex := range req.Exercises {
		if ex.ID != nil {
			incomingIDs[*ex.ID] = true
		}
	}

	// 3. Soft delete removed ones
	for weID := range existingIDs {
		if !incomingIDs[weID] {
			if err := tx.WorkoutExercise.
				UpdateOneID(weID).
				SetDeletedAt(time.Now()).
				Exec(ctx); err != nil {
				tx.Rollback()
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete workout exercise")
			}
		}
	}

	exerciseInstanceMap := map[string]uuid.UUID{}

	// 4. Upsert workout_exercises
	for _, ex := range req.Exercises {
		var actualInstanceID uuid.UUID

		// Determine ExerciseInstance
		if ex.ExerciseInstanceClientID != nil && *ex.ExerciseInstanceClientID != "" {
			if id, ok := exerciseInstanceMap[*ex.ExerciseInstanceClientID]; ok {
				actualInstanceID = id
			} else {
				newInstance, err := tx.ExerciseInstance.
					Create().
					SetExerciseID(ex.ExerciseID).
					Save(ctx)
				if err != nil {
					tx.Rollback()
					return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create exercise instance")
				}
				actualInstanceID = newInstance.ID
				exerciseInstanceMap[*ex.ExerciseInstanceClientID] = actualInstanceID
			}
		} else {
			newInstance, err := tx.ExerciseInstance.
				Create().
				SetExerciseID(ex.ExerciseID).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create exercise instance")
			}
			actualInstanceID = newInstance.ID
		}

		// Create new WorkoutExercise
		weCreate := tx.WorkoutExercise.Create().
			SetWorkoutID(workoutID).
			SetExerciseID(ex.ExerciseID).
			SetExerciseInstanceID(actualInstanceID)

		if ex.Order != nil {
			weCreate.SetOrder(*ex.Order)
		}
		if ex.Sets != nil {
			weCreate.SetSets(*ex.Sets)
		}
		if ex.Weight != nil {
			weCreate.SetWeight(*ex.Weight)
		}
		if ex.Reps != nil {
			weCreate.SetReps(*ex.Reps)
		}

		if _, err := weCreate.Save(ctx); err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create workout exercise")
		}
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit update")
	}

	// Return latest data
	updated, err := h.Client.Workout.
		Query().
		Where(workout.IDEQ(workoutID)).
		WithWorkoutExercises(func(wq *ent.WorkoutExerciseQuery) {
			wq.Where(workoutexercise.DeletedAtIsNil())
			wq.WithExercise()
			wq.WithWorkout()
			wq.WithExerciseInstance(func(eiq *ent.ExerciseInstanceQuery) {
				eiq.WithExercise()
			})
		}).
		Only(ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusOK, "Updated, but failed to fetch: "+err.Error())
	}

	return c.JSON(http.StatusOK, dto.CreateWorkoutResponse{
		Message: "Workout updated successfully.",
		Workout: toWorkoutResponse(updated),
	})
}
