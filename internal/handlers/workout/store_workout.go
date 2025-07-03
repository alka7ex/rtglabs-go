package handler

import (
	"context"
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"rtglabs-go/ent/exercise"
	"rtglabs-go/ent/workout"
	"rtglabs-go/ent/workoutexercise"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *WorkoutHandler) StoreWorkout(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		c.Logger().Error("User ID not found in context for StoreWorkout")
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	var req dto.CreateWorkoutRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx, err := h.Client.Tx(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create workout")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	exerciseInstanceMap := make(map[string]uuid.UUID)
	var createdWorkout *ent.Workout
	var errTx error

	errTx = func(ctx context.Context) error {
		createdWorkout, err = tx.Workout.
			Create().
			SetUserID(userID).
			SetName(req.Name).
			Save(ctx)
		if err != nil {
			return err
		}

		workoutExerciseBulk := make([]*ent.WorkoutExerciseCreate, 0, len(req.Exercises))

		for _, ex := range req.Exercises {
			var actualInstanceID uuid.UUID
			var createInstance bool

			if ex.ExerciseInstanceClientID != nil && *ex.ExerciseInstanceClientID != "" {
				if existing, found := exerciseInstanceMap[*ex.ExerciseInstanceClientID]; found {
					actualInstanceID = existing
				} else {
					createInstance = true
				}
			} else {
				createInstance = true
			}

			if createInstance {
				if _, err := tx.Exercise.Query().Where(exercise.IDEQ(ex.ExerciseID)).Only(ctx); err != nil {
					if ent.IsNotFound(err) {
						return echo.NewHTTPError(http.StatusBadRequest, "Invalid exercise ID")
					}
					return err
				}
				newInstance, err := tx.ExerciseInstance.Create().
					SetExerciseID(ex.ExerciseID).
					Save(ctx)
				if err != nil {
					return err
				}
				actualInstanceID = newInstance.ID
				if ex.ExerciseInstanceClientID != nil && *ex.ExerciseInstanceClientID != "" {
					exerciseInstanceMap[*ex.ExerciseInstanceClientID] = actualInstanceID
				}
			}

			we := tx.WorkoutExercise.Create().
				SetWorkoutID(createdWorkout.ID).
				SetExerciseID(ex.ExerciseID).
				SetExerciseInstanceID(actualInstanceID)

			if ex.Order != nil {
				we.SetOrder(uint(*ex.Order))
			}
			if ex.Sets != nil {
				we.SetSets(uint(*ex.Sets))
			}
			if ex.Weight != nil {
				we.SetWeight(*ex.Weight)
			}
			if ex.Reps != nil {
				we.SetReps(uint(*ex.Reps))
			}

			workoutExerciseBulk = append(workoutExerciseBulk, we)
		}

		if len(workoutExerciseBulk) > 0 {
			if _, err := tx.WorkoutExercise.CreateBulk(workoutExerciseBulk...).Save(ctx); err != nil {
				return err
			}
		}
		return nil
	}(c.Request().Context())

	if errTx != nil {
		tx.Rollback()
		if he, ok := errTx.(*echo.HTTPError); ok {
			return he
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create workout")
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to finalize workout creation")
	}

	finalWorkout, err := h.Client.Workout.
		Query().
		WithUser(). // 👈 ensure this
		Where(workout.IDEQ(createdWorkout.ID)).
		WithWorkoutExercises(func(wq *ent.WorkoutExerciseQuery) {
			wq.WithExercise()
			wq.WithWorkout() // 👈 ADD THIS
			wq.WithExerciseInstance(func(eiq *ent.ExerciseInstanceQuery) {
				eiq.WithExercise() // ✅ Preload fix
			})
			wq.Where(workoutexercise.DeletedAtIsNil())
		}).
		Only(c.Request().Context())

	if err != nil {
		return echo.NewHTTPError(http.StatusCreated, "Workout created, but failed to fetch details: "+err.Error())
	}

	return c.JSON(http.StatusCreated, dto.CreateWorkoutResponse{
		Message: "Workout created successfully.",
		Workout: toWorkoutResponse(finalWorkout),
	})
}
