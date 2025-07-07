package dto

import (
	"rtglabs-go/provider"
	"time"

	"github.com/google/uuid"
)

// ExerciseResponse represents the response structure for a single exercise.
// You need to define this struct, as it's used in WorkoutExerciseResponse.

// ExerciseInstanceResponse represents an ExerciseInstance record.
type ExerciseInstanceResponse struct {
	ID           uuid.UUID  `json:"id"`
	WorkoutLogID *uuid.UUID `json:"workout_log_id,omitempty"`
	ExerciseID   uuid.UUID  `json:"exercise_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// --- End Common Reusables ---

// CreateWorkoutRequest represents the request body for creating a new workout.
type CreateWorkoutRequest struct {
	Name      string                         `json:"name" validate:"required,max=255"`
	Exercises []CreateWorkoutExerciseRequest `json:"exercises" validate:"required,min=1,dive"` // `dive` validates each element in the slice
}

// CreateWorkoutExerciseRequest represents a single exercise within a workout creation request.
type CreateWorkoutExerciseRequest struct {
	ID                       *uuid.UUID `json:"id,omitempty" validate:"omitempty,uuid"`
	ExerciseID               uuid.UUID  `json:"exercise_id" validate:"required,uuid"`
	WorkoutOrder             *uint      `json:"order,omitempty" validate:"omitempty,min=1"`  // Pointers
	Sets                     *uint      `json:"sets,omitempty" validate:"omitempty,min=0"`   // Pointers
	Weight                   *float64   `json:"weight,omitempty" validate:"omitempty,min=0"` // Pointers
	Reps                     *uint      `json:"reps,omitempty" validate:"omitempty,min=0"`   // Pointers
	ExerciseInstanceClientID *string    `json:"exercise_instance_client_id,omitempty"`       // As per your original DTO
}

// WorkoutResponse represents the full workout details to be returned in a response.
type WorkoutResponse struct {
	ID               uuid.UUID                 `json:"id"`
	UserID           uuid.UUID                 `json:"user_id"`
	Name             string                    `json:"name"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
	DeletedAt        *time.Time                `json:"deleted_at,omitempty"`
	WorkoutExercises []WorkoutExerciseResponse `json:"workout_exercises,omitempty"`
}

// WorkoutExerciseResponse represents a single exercise associated with a workout (pivot data).
type WorkoutExerciseResponse struct {
	ID                 uuid.UUID                 `json:"id"`
	WorkoutID          uuid.UUID                 `json:"workout_id"`
	ExerciseID         uuid.UUID                 `json:"exercise_id"`
	ExerciseInstanceID *uuid.UUID                `json:"exercise_instance_id,omitempty"` // Pointer
	WorkoutOrder       *uint                     `json:"order,omitempty"`                // Pointer
	Sets               *uint                     `json:"sets,omitempty"`                 // Pointer
	Weight             *float64                  `json:"weight,omitempty"`               // Pointer
	Reps               *uint                     `json:"reps,omitempty"`                 // Pointer
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
	DeletedAt          *time.Time                `json:"deleted_at,omitempty"`
	Exercise           *ExerciseResponse         `json:"exercise,omitempty"`
	ExerciseInstance   *ExerciseInstanceResponse `json:"exercise_instance,omitempty"`
}

// CreateWorkoutResponse is the response for a successful workout creation.
type CreateWorkoutResponse struct {
	Message string          `json:"message"`
	Workout WorkoutResponse `json:"workout"`
}

// ListWorkoutResponse represents the paginated list response for workouts.
// It now embeds the util.PaginationResponse.
type ListWorkoutResponse struct {
	Data []WorkoutResponse `json:"data"`
	provider.PaginationResponse
}

// GenericListResponse can be used for any paginated list, making it even more reusable.
type GenericListResponse[T any] struct {
	Data []T `json:"data"`
	provider.PaginationResponse
}

// DeleteWorkoutResponse is the response for a successful workout deletion.
type DeleteWorkoutResponse struct {
	Message string `json:"message"`
}

// UpdateWorkoutRequest represents the request body for updating a workout.
type UpdateWorkoutRequest struct {
	Name      string                         `json:"name" validate:"required,max=255"`
	Exercises []UpdateWorkoutExerciseRequest `json:"exercises" validate:"required,dive"`
}

// UpdateWorkoutExerciseRequest represents a single exercise within a workout update request.
type UpdateWorkoutExerciseRequest struct {
	ID                       *uuid.UUID `json:"id,omitempty" validate:"omitempty,uuid"`
	ExerciseID               uuid.UUID  `json:"exercise_id" validate:"required,uuid"`
	WorkoutOrder             *uint      `json:"order,omitempty" validate:"omitempty,min=1"`
	Sets                     *uint      `json:"sets,omitempty" validate:"omitempty,min=0"`
	Weight                   *float64   `json:"weight,omitempty" validate:"omitempty,min=0"`
	Reps                     *uint      `json:"reps,omitempty" validate:"omitempty,min=0"`
	ExerciseInstanceID       *uuid.UUID `json:"exercise_instance_id,omitempty" validate:"omitempty,uuid"`
	ExerciseInstanceClientID *string    `json:"exercise_instance_client_id,omitempty"` // Use string as per your original DTO
}
