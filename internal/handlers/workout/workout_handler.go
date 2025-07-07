package handler

import (
	"database/sql" // For *sql.DB, sql.Null* types
	"time"

	"github.com/Masterminds/squirrel" // Import squirrel
	"rtglabs-go/dto"
	"rtglabs-go/model" // Import your model package (Workout, WorkoutExercise etc.)
)

// WorkoutHandler holds the database client and squirrel statement builder.
type WorkoutHandler struct {
	DB *sql.DB
	sq squirrel.StatementBuilderType
}

// NewWorkoutHandler creates and returns a new WorkoutHandler.
// It now takes *sql.DB and initializes squirrel with the appropriate placeholder format.
func NewWorkoutHandler(db *sql.DB) *WorkoutHandler {
	return &WorkoutHandler{
		DB: db,
		sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar), // Or squirrel.Question for '?'
	}
}

// --- Helper Functions (Adjusted for `model` structs) ---

// toWorkoutResponse converts a model.Workout entity to a dto.WorkoutResponse DTO.
// It will need to fetch WorkoutExercises separately, as SQL won't load edges automatically.

func toWorkoutResponse(w *model.Workout, workoutExercisesDTO []dto.WorkoutExerciseResponse) dto.WorkoutResponse { // <--- MODIFIED HERE
	var deletedAt *time.Time
	if w.DeletedAt != nil {
		deletedAt = w.DeletedAt
	}

	return dto.WorkoutResponse{
		ID:               w.ID,
		UserID:           w.UserID,
		Name:             w.Name,
		CreatedAt:        w.CreatedAt,
		UpdatedAt:        w.UpdatedAt,
		DeletedAt:        deletedAt,
		WorkoutExercises: workoutExercisesDTO, // <--- DIRECTLY USE THE PASSED DTO SLICE
	}
}

// toWorkoutExerciseResponse converts a model.WorkoutExercise entity to a dto.WorkoutExerciseResponse DTO.
// It now also takes optional related Exercise and ExerciseInstance models for nesting.
func toWorkoutExerciseResponse(
	we *model.WorkoutExercise,
	ex *model.Exercise, // Optional related Exercise
	ei *model.ExerciseInstance, // Optional related ExerciseInstance
) dto.WorkoutExerciseResponse {
	var deletedAt *time.Time
	if we.DeletedAt != nil {
		deletedAt = we.DeletedAt
	}

	var exerciseDTO *dto.ExerciseResponse
	if ex != nil { // Check if related exercise data was provided
		tempEx := toExerciseResponse(ex)
		exerciseDTO = &tempEx
	}

	var instanceDTO *dto.ExerciseInstanceResponse
	if ei != nil { // Check if related exercise instance data was provided
		tempEi := toExerciseInstanceResponse(ei)
		instanceDTO = &tempEi
	}

	// Assuming WorkoutID and ExerciseID are always present on model.WorkoutExercise
	workoutID := we.WorkoutID
	exerciseID := we.ExerciseID

	return dto.WorkoutExerciseResponse{
		ID:                 we.ID,
		WorkoutID:          workoutID,
		ExerciseID:         exerciseID,
		ExerciseInstanceID: we.ExerciseInstanceID, // Already a pointer in model
		WorkoutOrder:       we.WorkoutOrder,       // Already a pointer in model
		Sets:               we.Sets,               // Already a pointer in model
		Weight:             we.Weight,             // Already a pointer in model
		Reps:               we.Reps,               // Already a pointer in model
		CreatedAt:          we.CreatedAt,
		UpdatedAt:          we.UpdatedAt,
		DeletedAt:          deletedAt,
		Exercise:           exerciseDTO,
		ExerciseInstance:   instanceDTO,
	}
}

// These helper functions should already be in your `handlers` package from previous refactors
// but included here for completeness of context for this file.

// toExerciseResponse converts a model.Exercise entity to a dto.ExerciseResponse DTO.
func toExerciseResponse(ex *model.Exercise) dto.ExerciseResponse {
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

// toExerciseInstanceResponse converts a model.ExerciseInstance entity to a dto.ExerciseInstanceResponse DTO.
func toExerciseInstanceResponse(ei *model.ExerciseInstance) dto.ExerciseInstanceResponse {
	var deletedAt *time.Time
	if ei.DeletedAt != nil {
		deletedAt = ei.DeletedAt
	}

	// Assuming ExerciseID is always present on ExerciseInstance
	exerciseID := ei.ExerciseID

	return dto.ExerciseInstanceResponse{
		ID:           ei.ID,
		WorkoutLogID: ei.WorkoutLogID, // This assumes WorkoutLogID is directly on model.ExerciseInstance, already a pointer
		ExerciseID:   exerciseID,
		CreatedAt:    ei.CreatedAt,
		UpdatedAt:    ei.UpdatedAt,
		DeletedAt:    deletedAt,
	}
}
