package dto

import (
	"time"

	"github.com/google/uuid"
	"rtglabs-go/provider" // Assuming this contains your PaginationResponse
)

// ExerciseInstanceResponse represents an ExerciseInstance record from the 'exercise_instances' table.
// This is for the *template* exercise instance used in workout blueprints.
type ExerciseInstanceResponse struct {
	ID           uuid.UUID  `json:"id"`
	WorkoutLogID *uuid.UUID `json:"workout_log_id"` // Removed ""
	ExerciseID   uuid.UUID  `json:"exercise_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"` // Removed ""
}

// CreateWorkoutRequest represents the request body for creating a new workout template.
type CreateWorkoutRequest struct {
	Name      string                         `json:"name" validate:"required,max=255"`
	Exercises []CreateWorkoutExerciseRequest `json:"exercises" validate:"required,min=1,dive"`
}

// CreateWorkoutExerciseRequest represents a single exercise within a workout template creation request.
type CreateWorkoutExerciseRequest struct {
	ID                       *uuid.UUID `json:"id" validate:"omitempty,uuid"`
	ExerciseID               uuid.UUID  `json:"exercise_id" validate:"required,uuid"`
	WorkoutOrder             *uint      `json:"order" validate:"omitempty,min=1"`
	Sets                     *uint      `json:"sets" validate:"omitempty,min=0"`
	Weight                   *float64   `json:"weight" validate:"omitempty,min=0"`
	Reps                     *uint      `json:"reps" validate:"omitempty,min=0"`
	ExerciseInstanceClientID *string    `json:"exercise_instance_client_id"`
}

// WorkoutExerciseResponse represents a single exercise associated with a workout (pivot data).
// This is used for defining workout *templates*.
type WorkoutExerciseResponse struct {
	ID                 uuid.UUID                 `json:"id"`
	WorkoutID          uuid.UUID                 `json:"workout_id"`
	ExerciseID         uuid.UUID                 `json:"exercise_id"`
	ExerciseInstanceID *uuid.UUID                `json:"exercise_instance_id"` // Removed ""
	WorkoutOrder       *uint                     `json:"order"`                // Removed ""
	Sets               *uint                     `json:"sets"`                 // Removed ""
	Weight             *float64                  `json:"weight"`               // Removed ""
	Reps               *uint                     `json:"reps"`                 // Removed ""
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
	DeletedAt          *time.Time                `json:"deleted_at"`        // Removed ""
	Exercise           *ExerciseResponse         `json:"exercise"`          // Removed ""
	ExerciseInstance   *ExerciseInstanceResponse `json:"exercise_instance"` // Removed ""
}

// WorkoutResponse represents the full workout details to be returned in a response.
// This is the authoritative definition for WorkoutResponse.
type WorkoutResponse struct {
	ID               uuid.UUID                 `json:"id"`
	UserID           uuid.UUID                 `json:"user_id"`
	Name             string                    `json:"name"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
	DeletedAt        *time.Time                `json:"deleted_at"`        // Removed ""
	WorkoutExercises []WorkoutExerciseResponse `json:"workout_exercises"` // Removed "" - For slice, an empty slice [] will be []
}

// CreateWorkoutResponse is the response for a successful workout template creation.
type CreateWorkoutResponse struct {
	Message string          `json:"message"`
	Workout WorkoutResponse `json:"workout"`
}

// ListWorkoutResponse represents the paginated list response for workout templates.
type ListWorkoutResponse struct {
	Data []WorkoutResponse `json:"data"`
	provider.PaginationResponse
}

// DeleteWorkoutResponse is the response for a successful workout template deletion.
type DeleteWorkoutResponse struct {
	Message string `json:"message"`
}

// UpdateWorkoutRequest represents the request body for updating a workout template.
type UpdateWorkoutRequest struct {
	Name      string                         `json:"name" validate:"required,max=255"`
	Exercises []UpdateWorkoutExerciseRequest `json:"exercises" validate:"required,dive"`
}

// UpdateWorkoutExerciseRequest represents a single exercise within a workout template update request.
type UpdateWorkoutExerciseRequest struct {
	ID                       *uuid.UUID `json:"id" validate:"omitempty,uuid"`
	ExerciseID               uuid.UUID  `json:"exercise_id" validate:"required,uuid"`
	WorkoutOrder             *uint      `json:"order" validate:"omitempty,min=1"`
	Sets                     *uint      `json:"sets" validate:"omitempty,min=0"`
	Weight                   *float64   `json:"weight" validate:"omitempty,min=0"`
	Reps                     *uint      `json:"reps" validate:"omitempty,min=0"`
	ExerciseInstanceID       *uuid.UUID `json:"exercise_instance_id" validate:"omitempty,uuid"`
	ExerciseInstanceClientID *string    `json:"exercise_instance_client_id"`
}
