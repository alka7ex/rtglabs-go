package workout_handler

import (
	"context"
	"fmt"
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"rtglabs-go/ent/exercise"
	"rtglabs-go/ent/user"
	"rtglabs-go/ent/workout"
	"rtglabs-go/ent/workoutexercise"
	"strconv"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type WorkoutHandler struct {
	Client *ent.Client
}

func NewWorkoutHandler(client *ent.Client) *WorkoutHandler {
	return &WorkoutHandler{Client: client}
}

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

func (h *WorkoutHandler) IndexWorkout(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
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

	query := h.Client.Workout.
		Query().
		Where(
			workout.DeletedAtIsNil(),
			workout.HasUserWith(user.IDEQ(userID)),
		).
		WithWorkoutExercises(func(wq *ent.WorkoutExerciseQuery) {
			wq.WithExercise()
			wq.WithWorkout() // 👈 ADD THIS
			wq.WithExerciseInstance(func(eiq *ent.ExerciseInstanceQuery) {
				eiq.WithExercise() // ✅ Preload fix
			})
			wq.Where(workoutexercise.DeletedAtIsNil())
		})

	totalCount, err := query.Count(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to count workouts")
	}

	workouts, err := query.
		Order(workout.ByCreatedAt(sql.OrderDesc())).
		Limit(limit).
		Offset(offset).
		All(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch workouts")
	}

	dtoWorkouts := make([]dto.WorkoutResponse, len(workouts))
	for i, w := range workouts {
		dtoWorkouts[i] = toWorkoutResponse(w)
	}

	lastPage := (totalCount + limit - 1) / limit
	baseURL := c.Request().URL.Path
	queryParams := c.Request().URL.Query()

	buildPageURL := func(p int) string {
		queryParams.Set("page", strconv.Itoa(p))
		queryParams.Set("limit", strconv.Itoa(limit))
		return fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())
	}

	var links []dto.Link
	for i := 1; i <= lastPage; i++ {
		url := buildPageURL(i)
		links = append(links, dto.Link{
			URL:    &url,
			Label:  strconv.Itoa(i),
			Active: i == page,
		})
	}

	var prevURL, nextURL *string
	if page > 1 {
		url := buildPageURL(page - 1)
		prevURL = &url
	}
	if page < lastPage {
		url := buildPageURL(page + 1)
		nextURL = &url
	}

	return c.JSON(http.StatusOK, dto.ListWorkoutResponse{
		CurrentPage:  page,
		Data:         dtoWorkouts,
		FirstPageURL: buildPageURL(1),
		From:         &offset,
		LastPage:     lastPage,
		LastPageURL:  buildPageURL(lastPage),
		Links:        links,
		NextPageURL:  nextURL,
		Path:         baseURL,
		PerPage:      limit,
		PrevPageURL:  prevURL,
		To:           &offset,
		Total:        totalCount,
	})
}

func (h *WorkoutHandler) GetWorkout(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}

	workoutID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid workout ID")
	}

	entWorkout, err := h.Client.Workout.
		Query().
		Where(
			workout.IDEQ(workoutID),
			workout.HasUserWith(user.IDEQ(userID)),
			workout.DeletedAtIsNil(),
		).
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
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "Workout not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout")
	}

	return c.JSON(http.StatusOK, toWorkoutResponse(entWorkout))
}

func toWorkoutResponse(w *ent.Workout) dto.WorkoutResponse {
	var deletedAt *time.Time
	if w.DeletedAt != nil {
		deletedAt = w.DeletedAt
	}

	var exercises []dto.WorkoutExerciseResponse
	for _, we := range w.Edges.WorkoutExercises {
		exercises = append(exercises, toWorkoutExerciseResponse(we))
	}

	var userID uuid.UUID
	if w.Edges.User != nil {
		userID = w.Edges.User.ID
	}

	return dto.WorkoutResponse{
		ID:               w.ID,
		UserID:           userID,
		Name:             w.Name,
		CreatedAt:        w.CreatedAt,
		UpdatedAt:        w.UpdatedAt,
		DeletedAt:        deletedAt,
		WorkoutExercises: exercises,
	}
}

func toWorkoutExerciseResponse(we *ent.WorkoutExercise) dto.WorkoutExerciseResponse {
	var deletedAt *time.Time
	if we.DeletedAt != nil {
		deletedAt = we.DeletedAt
	}

	var instanceID *uuid.UUID
	if we.Edges.ExerciseInstance != nil {
		instanceID = &we.Edges.ExerciseInstance.ID
	}

	var exerciseDTO *dto.ExerciseResponse
	if we.Edges.Exercise != nil {
		ex := toExerciseResponse(we.Edges.Exercise)
		exerciseDTO = &ex
	}

	var instanceDTO *dto.ExerciseInstanceResponse
	if we.Edges.ExerciseInstance != nil {
		inst := toExerciseInstanceResponse(we.Edges.ExerciseInstance)
		instanceDTO = &inst
	}

	var workoutID uuid.UUID
	if we.Edges.Workout != nil {
		workoutID = we.Edges.Workout.ID
	}

	var exerciseID uuid.UUID
	if we.Edges.Exercise != nil {
		exerciseID = we.Edges.Exercise.ID
	}

	return dto.WorkoutExerciseResponse{
		ID:                 we.ID,
		WorkoutID:          workoutID,
		ExerciseID:         exerciseID,
		ExerciseInstanceID: instanceID,
		Order:              we.Order,
		Sets:               we.Sets,
		Weight:             we.Weight,
		Reps:               we.Reps,
		CreatedAt:          we.CreatedAt,
		UpdatedAt:          we.UpdatedAt,
		DeletedAt:          deletedAt,
		Exercise:           exerciseDTO,
		ExerciseInstance:   instanceDTO,
	}
}

func toExerciseResponse(ex *ent.Exercise) dto.ExerciseResponse {
	var deletedAt *time.Time
	if ex.DeletedAt != nil {
		deletedAt = ex.DeletedAt
	}
	return dto.ExerciseResponse{
		ID:        ex.ID,
		Name:      ex.Name,
		CreatedAt: ex.CreatedAt,
		UpdatedAt: ex.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

func toExerciseInstanceResponse(ei *ent.ExerciseInstance) dto.ExerciseInstanceResponse {
	var workoutLogID *uuid.UUID
	if ei.Edges.WorkoutLog != nil {
		workoutLogID = &ei.Edges.WorkoutLog.ID
	}

	var deletedAt *time.Time
	if ei.DeletedAt != nil {
		deletedAt = ei.DeletedAt
	}

	var exerciseID uuid.UUID
	if ei.Edges.Exercise != nil {
		exerciseID = ei.Edges.Exercise.ID
	}

	return dto.ExerciseInstanceResponse{
		ID:           ei.ID,
		WorkoutLogID: workoutLogID,
		ExerciseID:   exerciseID,
		CreatedAt:    ei.CreatedAt,
		UpdatedAt:    ei.UpdatedAt,
		DeletedAt:    deletedAt,
	}
}

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

