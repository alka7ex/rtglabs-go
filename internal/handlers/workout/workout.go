package workout_handler

import (
	"context"
	"fmt"
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"rtglabs-go/ent/exercise"
	"rtglabs-go/ent/exerciseinstance"
	"rtglabs-go/ent/user"
	"rtglabs-go/ent/workout"
	"rtglabs-go/ent/workoutexercise"
	"strconv"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// WorkoutHandler holds the ent.Client for database operations.
type WorkoutHandler struct {
	Client *ent.Client
}

// NewWorkoutHandler creates and returns a new WorkoutHandler.
func NewWorkoutHandler(client *ent.Client) *WorkoutHandler {
	return &WorkoutHandler{Client: client}
}

// StoreWorkout creates a new workout record and its associated exercises.
func (h *WorkoutHandler) StoreWorkout(c echo.Context) error {
	// 1. Get the authenticated user ID from the context.
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		c.Logger().Error("User ID not found in context for StoreWorkout")
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	// 2. Bind and validate the request body.
	var req dto.CreateWorkoutRequest
	if err := c.Bind(&req); err != nil {
		c.Logger().Warnf("Invalid request body for StoreWorkout: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}
	if err := c.Validate(&req); err != nil {
		c.Logger().Warnf("Validation failed for StoreWorkout: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Start a transaction
	tx, err := h.Client.Tx(c.Request().Context())
	if err != nil {
		c.Logger().Errorf("Failed to start transaction for StoreWorkout: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create workout")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// Map client-side instance IDs to actual UUIDs
	exerciseInstanceMap := make(map[string]uuid.UUID)

	var createdWorkout *ent.Workout
	var errTx error

	// Use a closure for transaction logic to simplify rollback
	errTx = func(ctx context.Context) error {
		// 3. Create the Workout entity.
		createdWorkout, err = tx.Workout.
			Create().
			SetUserID(userID).
			SetName(req.Name).
			Save(ctx)
		if err != nil {
			c.Logger().Errorf("Failed to create workout in transaction: %v", err)
			return err
		}

		workoutExerciseBulkCreators := make([]*ent.WorkoutExerciseCreate, 0, len(req.Exercises))

		for _, exerciseData := range req.Exercises {
			var actualExerciseInstanceID uuid.UUID
			var createExerciseInstance bool

			// Determine actual ExerciseInstance ID
			if exerciseData.ExerciseInstanceClientID != nil && *exerciseData.ExerciseInstanceClientID != "" {
				// If client provided a grouping ID, check if we already created an instance for it
				if existingID, found := exerciseInstanceMap[*exerciseData.ExerciseInstanceClientID]; found {
					actualExerciseInstanceID = existingID
				} else {
					// If not, mark to create a new ExerciseInstance for this group
					createExerciseInstance = true
				}
			} else {
				// If no client grouping ID, create a new unique ExerciseInstance for this workout exercise entry
				createExerciseInstance = true
			}

			if createExerciseInstance {
				// Ensure the exercise_id for the instance exists (optional, but good for data integrity)
				_, err = tx.Exercise.Query().Where(exercise.IDEQ(exerciseData.ExerciseID)).Only(ctx)
				if err != nil {
					if ent.IsNotFound(err) {
						c.Logger().Warnf("Exercise with ID %s not found for instance creation: %v", exerciseData.ExerciseID, err)
						return echo.NewHTTPError(http.StatusBadRequest, "Invalid exercise ID provided for an exercise instance")
					}
					c.Logger().Errorf("Database error querying Exercise for instance creation: %v", err)
					return err
				}

				newExerciseInstance, err := tx.ExerciseInstance.
					Create().
					SetExerciseID(exerciseData.ExerciseID).
					// WorkoutLogID is nullable and optional for now based on schema/Laravel logic
					// SetWorkoutLogID(...) // If linking to WorkoutLog, uncomment and provide ID
					Save(ctx)
				if err != nil {
					c.Logger().Errorf("Failed to create exercise instance in transaction: %v", err)
					return err
				}
				actualExerciseInstanceID = newExerciseInstance.ID

				if exerciseData.ExerciseInstanceClientID != nil && *exerciseData.ExerciseInstanceClientID != "" {
					exerciseInstanceMap[*exerciseData.ExerciseInstanceClientID] = actualExerciseInstanceID
				}
			}

			// Create the WorkoutExercise record
			wec := tx.WorkoutExercise.Create().
				SetWorkoutID(createdWorkout.ID).
				SetExerciseID(exerciseData.ExerciseID).
				SetExerciseInstanceID(actualExerciseInstanceID) // Link to the actual instance ID

			if exerciseData.Order != nil {
				wec.SetOrder(uint(*exerciseData.Order))
			}
			if exerciseData.Sets != nil {
				wec.SetSets(uint(*exerciseData.Sets))
			}
			if exerciseData.Weight != nil {
				wec.SetWeight(*exerciseData.Weight)
			}
			if exerciseData.Reps != nil {
				wec.SetReps(uint(*exerciseData.Reps))
			}

			workoutExerciseBulkCreators = append(workoutExerciseBulkCreators, wec)
		}

		// Bulk insert WorkoutExercise records
		if len(workoutExerciseBulkCreators) > 0 {
			_, err = tx.WorkoutExercise.CreateBulk(workoutExerciseBulkCreators...).Save(ctx)
			if err != nil {
				c.Logger().Errorf("Failed to bulk create workout exercises in transaction: %v", err)
				return err
			}
		}

		return nil
	}(c.Request().Context()) // Execute the transaction logic

	if errTx != nil {
		tx.Rollback() // Rollback on any error from the transaction logic
		// Check if the error is already an HTTPError, if not, make it a generic 500
		if he, ok := errTx.(*echo.HTTPError); ok {
			return he
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create workout and its exercises")
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		c.Logger().Errorf("Failed to commit transaction for StoreWorkout: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to finalize workout creation")
	}

	// 4. Reload workout with relationships for response.
	finalWorkout, err := h.Client.Workout.
		Query().
		Where(workout.IDEQ(createdWorkout.ID)).
		WithWorkoutExercises(func(wq *ent.WorkoutExerciseQuery) {
			wq.WithExercise()
			wq.WithExerciseInstance()
			wq.Where(workoutexercise.DeletedAtIsNil()) // Only load non-deleted pivot entries
		}).
		Only(c.Request().Context())

	if err != nil {
		c.Logger().Errorf("Failed to fetch created workout with relations: %v", err)
		// Even if fetching fails, the workout is created, so return 201 but with a warning.
		return echo.NewHTTPError(http.StatusCreated, "Workout created, but failed to retrieve full details: "+err.Error())
	}

	// 5. Build the DTO response and return.
	response := dto.CreateWorkoutResponse{
		Message: "Workout created successfully.",
		Workout: toWorkoutResponse(finalWorkout),
	}

	c.Logger().Infof("Workout and associated exercises created successfully. Workout ID: %s, User ID: %s", createdWorkout.ID, userID)
	return c.JSON(http.StatusCreated, response)
}

// IndexWorkout displays a listing of the user's workouts.
func (h *WorkoutHandler) IndexWorkout(c echo.Context) error {
	// 1. Get the authenticated user ID from the context.
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	// --- Pagination Parameters ---
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1 // Default to first page
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 15 // Default limit per page
	}
	if limit > 100 {
		limit = 100 // Cap the limit
	}
	offset := (page - 1) * limit
	// --- End Pagination Parameters ---

	// 2. Base query for workouts, filtered by authenticated user and not deleted.
	query := h.Client.Workout.
		Query().
		Where(
			workout.DeletedAtIsNil(),
			workout.HasUserWith(user.IDEQ(userID)), // Filter by authenticated user's ID
		)

	// Eager-load workoutExercises and their nested relationships for the index response
	query.WithWorkoutExercises(func(wq *ent.WorkoutExerciseQuery) {
		wq.WithExercise()
		wq.WithExerciseInstance()
		wq.Where(workoutexercise.DeletedAtIsNil()) // Only load non-deleted pivot entries
	})

	// 3. Get total count BEFORE applying limit and offset.
	totalCount, err := query.Count(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to count workouts:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workouts")
	}

	// 4. Fetch the paginated and sorted workout records.
	entWorkouts, err := query.
		Order(workout.ByCreatedAt(sql.OrderDesc())). // Order by creation date descending
		Limit(limit).
		Offset(offset).
		All(c.Request().Context())

	if err != nil {
		c.Logger().Error("Failed to list workouts:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workouts")
	}

	// 5. Convert ent entities to DTOs.
	dtoWorkouts := make([]dto.WorkoutResponse, len(entWorkouts))
	for i, w := range entWorkouts {
		dtoWorkouts[i] = toWorkoutResponse(w)
	}

	// 6. Build the pagination response DTO. (Same logic as Bodyweight handler)
	totalPages := (totalCount + limit - 1) / limit
	lastPage := totalPages

	baseURL := c.Request().URL.Path
	queryParams := c.Request().URL.Query()

	nextPageURL := ""
	if page < lastPage {
		queryParams.Set("page", strconv.Itoa(page+1))
		nextPageURL = fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())
	}
	prevPageURL := ""
	if page > 1 {
		queryParams.Set("page", strconv.Itoa(page-1))
		prevPageURL = fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())
	}

	var nextPageURLPtr, prevPageURLPtr *string
	if nextPageURL != "" {
		nextPageURLPtr = &nextPageURL
	}
	if prevPageURL != "" {
		prevPageURLPtr = &prevPageURL
	}

	var links []dto.Link
	links = append(links, dto.Link{URL: prevPageURLPtr, Label: "&laquo; Previous", Active: page > 1})
	for i := 1; i <= totalPages; i++ {
		pageURL := fmt.Sprintf("%s?page=%d&limit=%d", baseURL, i, limit)
		links = append(links, dto.Link{URL: &pageURL, Label: strconv.Itoa(i), Active: i == page})
	}
	links = append(links, dto.Link{URL: nextPageURLPtr, Label: "Next &raquo;", Active: page < lastPage})

	response := dto.ListWorkoutResponse{
		CurrentPage:  page,
		Data:         dtoWorkouts,
		FirstPageURL: fmt.Sprintf("%s?page=1&limit=%d", baseURL, limit),
		From:         &offset,
		LastPage:     lastPage,
		LastPageURL:  fmt.Sprintf("%s?page=%d&limit=%d", baseURL, lastPage, limit),
		Links:        links,
		NextPageURL:  nextPageURLPtr,
		Path:         baseURL,
		PerPage:      limit,
		PrevPageURL:  prevPageURLPtr,
		To:           &offset, // Note: 'to' is typically offset + count, but this is a simple approximation
		Total:        totalCount,
	}

	return c.JSON(http.StatusOK, response)
}

// GetWorkout displays the specified workout.
func (h *WorkoutHandler) GetWorkout(c echo.Context) error {
	// 1. Get the authenticated user ID from the context.
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	// 2. Parse the workout ID from the URL parameter.
	workoutID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid workout ID format")
	}

	// 3. Query the workout, ensuring it belongs to the authenticated user and is not soft-deleted.
	entWorkout, err := h.Client.Workout.
		Query().
		Where(
			workout.IDEQ(workoutID),
			workout.HasUserWith(user.IDEQ(userID)),
			workout.DeletedAtIsNil(), // Only retrieve non-deleted workouts
		).
		WithWorkoutExercises(func(wq *ent.WorkoutExerciseQuery) {
			wq.WithExercise()                          // Eager load related Exercise
			wq.WithExerciseInstance()                  // Eager load related ExerciseInstance
			wq.Where(workoutexercise.DeletedAtIsNil()) // Only load non-deleted pivot entries
		}).
		Only(c.Request().Context())

	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "Workout not found or you don't have access")
		}
		c.Logger().Errorf("Database query error for GetWorkout ID %s: %v", workoutID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workout")
	}

	// 4. Return the DTO response.
	return c.JSON(http.StatusOK, toWorkoutResponse(entWorkout))
}

// DestroyWorkout performs a soft delete on a workout and its associated records.

func (h *WorkoutHandler) DestroyWorkout(c echo.Context) error {
	// 1. Get the authenticated user ID from the context.
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	// 2. Parse the workout ID from the URL parameter.
	workoutID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid workout ID format")
	}

	// Start a transaction for atomicity
	tx, err := h.Client.Tx(c.Request().Context())
	if err != nil {
		c.Logger().Errorf("Failed to start transaction for DestroyWorkout: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete workout")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	errTx := func(ctx context.Context) error {
		// Verify workout ownership and existence before attempting to delete
		existingWorkout, err := tx.Workout.Query().
			Where(
				workout.IDEQ(workoutID),
				workout.HasUserWith(user.IDEQ(userID)),
				workout.DeletedAtIsNil(),
			).
			WithWorkoutExercises(). // Eager load workout exercises to get their IDs
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return echo.NewHTTPError(http.StatusNotFound, "Workout not found or you don't have access")
			}
			c.Logger().Errorf("Failed to retrieve workout for deletion: %v", err)
			return err
		}

		now := time.Now()

		// 3. Soft delete associated WorkoutExercise records
		var workoutExerciseIDs []uuid.UUID
		var instanceIDsToDeleteCandidates []uuid.UUID
		if len(existingWorkout.Edges.WorkoutExercises) > 0 {
			for _, we := range existingWorkout.Edges.WorkoutExercises {
				workoutExerciseIDs = append(workoutExerciseIDs, we.ID)
				// --- FIX START ---

				if we.Edges.ExerciseInstance != nil {
					instanceIDsToDeleteCandidates = append(instanceIDsToDeleteCandidates, we.Edges.ExerciseInstance.ID)
				}
				// --- FIX END ---
			}

			// Perform bulk soft delete for WorkoutExercise entries
			_, err = tx.WorkoutExercise.Update().
				Where(
					workoutexercise.IDIn(workoutExerciseIDs...),
					workoutexercise.HasWorkoutWith(workout.IDEQ(workoutID)),
					workoutexercise.DeletedAtIsNil(),
				).
				SetDeletedAt(now).
				Save(ctx)
			if err != nil {
				c.Logger().Errorf("Failed to soft delete WorkoutExercise records for Workout ID %s: %v", workoutID, err)
				return err
			}
			c.Logger().Infof("Soft deleted %d WorkoutExercise records for Workout ID: %s", len(workoutExerciseIDs), workoutID)
		}

		// 4. Determine and soft delete ExerciseInstance records that are no longer referenced.
		uniqueInstanceIDsToDeleteCandidates := make(map[uuid.UUID]struct{})
		for _, id := range instanceIDsToDeleteCandidates {
			uniqueInstanceIDsToDeleteCandidates[id] = struct{}{}
		}

		var actuallyDeleteInstanceIDs []uuid.UUID
		for instanceID := range uniqueInstanceIDsToDeleteCandidates {
			// Check if this ExerciseInstance is referenced by any other *non-deleted* WorkoutExercise
			// that does NOT belong to the workout being deleted.
			isReferencedByOtherWorkoutExercises, err := tx.WorkoutExercise.Query().
				Where(
					workoutexercise.HasExerciseInstanceWith(exerciseinstance.IDEQ(instanceID)),
					workoutexercise.Not(
						workoutexercise.HasWorkoutWith(workout.IDEQ(workoutID)),
					),
					workoutexercise.DeletedAtIsNil(), // And that reference is not soft-deleted
				).
				Exist(ctx)
			if err != nil {
				c.Logger().Errorf("Failed to check ExerciseInstance reference for ID %s: %v", instanceID, err)
				return err
			}

			if !isReferencedByOtherWorkoutExercises {
				actuallyDeleteInstanceIDs = append(actuallyDeleteInstanceIDs, instanceID)
			}
		}

		if len(actuallyDeleteInstanceIDs) > 0 {
			_, err = tx.ExerciseInstance.Update().
				Where(
					exerciseinstance.IDIn(actuallyDeleteInstanceIDs...),
					exerciseinstance.DeletedAtIsNil(), // Ensure we only soft delete non-deleted ones
				).
				SetDeletedAt(now).
				Save(ctx)
			if err != nil {
				c.Logger().Errorf("Failed to soft delete ExerciseInstance records: %v", err)
				return err
			}
			c.Logger().Infof("Soft deleted %d ExerciseInstance records for Workout ID: %s. IDs: %v", len(actuallyDeleteInstanceIDs), workoutID, actuallyDeleteInstanceIDs)
		}

		// 5. Soft delete the Workout itself
		_, err = tx.Workout.Update().
			Where(
				workout.IDEQ(workoutID),
				workout.HasUserWith(user.IDEQ(userID)),
				workout.DeletedAtIsNil(),
			).
			SetDeletedAt(now).
			Save(ctx)
		if err != nil {
			c.Logger().Errorf("Failed to soft delete Workout ID %s: %v", workoutID, err)
			return err
		}
		c.Logger().Infof("Workout soft deleted successfully. ID: %s", workoutID)

		return nil
	}(c.Request().Context())

	if errTx != nil {
		tx.Rollback()
		if he, ok := errTx.(*echo.HTTPError); ok {
			return he
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete workout and its associated records")
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		c.Logger().Errorf("Failed to commit transaction for DestroyWorkout: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to finalize workout deletion")
	}

	// 6. Return the DTO response.
	response := dto.DeleteWorkoutResponse{
		Message: "Workout soft deleted successfully.",
	}

	return c.JSON(http.StatusOK, response)
}

// --- Helper Functions for DTO Conversion ---

func toWorkoutResponse(w *ent.Workout) dto.WorkoutResponse {
	var deletedAt *time.Time
	if w.DeletedAt != nil {
		deletedAt = w.DeletedAt
	}

	workoutExercises := make([]dto.WorkoutExerciseResponse, 0)
	if w.Edges.WorkoutExercises != nil {
		for _, we := range w.Edges.WorkoutExercises {
			workoutExercises = append(workoutExercises, toWorkoutExerciseResponse(we))
		}
	}

	return dto.WorkoutResponse{
		ID:               w.ID,
		UserID:           w.Edges.User.ID,
		Name:             w.Name,
		CreatedAt:        w.CreatedAt,
		UpdatedAt:        w.UpdatedAt,
		DeletedAt:        deletedAt,
		WorkoutExercises: workoutExercises,
	}
}

func toWorkoutExerciseResponse(we *ent.WorkoutExercise) dto.WorkoutExerciseResponse {
	var deletedAt *time.Time
	if we.DeletedAt != nil {
		deletedAt = we.DeletedAt
	}

	var exerciseInstanceID *uuid.UUID
	// --- FIX START ---
	if we.ExerciseInstanceID != nil { // Check if the pointer is not nil
		exerciseInstanceID = we.ExerciseInstanceID // Already a *uuid.UUID, so just assign
	}
	// --- FIX END ---

	var exerciseDTO *dto.ExerciseResponse
	if we.Edges.Exercise != nil {
		ex := toExerciseResponse(we.Edges.Exercise)
		exerciseDTO = &ex
	}

	var exerciseInstanceDTO *dto.ExerciseInstanceResponse
	if we.Edges.ExerciseInstance != nil {
		ei := toExerciseInstanceResponse(we.Edges.ExerciseInstance)
		exerciseInstanceDTO = &ei
	}

	return dto.WorkoutExerciseResponse{
		ID:                 we.ID,
		WorkoutID:          we.Edges.Workout.ID,
		ExerciseID:         we.Edges.Exercise.ID,
		ExerciseInstanceID: exerciseInstanceID,
		Order:              we.Order,
		Sets:               we.Sets,
		Weight:             we.Weight,
		Reps:               we.Reps,
		CreatedAt:          we.CreatedAt,
		UpdatedAt:          we.UpdatedAt,
		DeletedAt:          deletedAt,
		Exercise:           exerciseDTO,
		ExerciseInstance:   exerciseInstanceDTO,
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

	return dto.ExerciseInstanceResponse{
		ID:           ei.ID,
		WorkoutLogID: workoutLogID,
		ExerciseID:   ei.Edges.Exercise.ID,
		CreatedAt:    ei.CreatedAt,
		UpdatedAt:    ei.UpdatedAt,
		DeletedAt:    deletedAt,
	}
}
