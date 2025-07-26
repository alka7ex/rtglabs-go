package dto

import (
	"time"

	"rtglabs-go/provider" // Assuming this contains your PaginationResponse

	"github.com/google/uuid"
)

// ExerciseResponse represents the response structure for a single exercise.
// This is the authoritative definition for ExerciseResponse.

// ExerciseResponse represents the response structure for a single exercise.
type ExerciseResponse struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`   // <--- ADD THIS
	Position     string     `json:"position"`      // <--- ADD THIS
	ForceType    string     `json:"force_type"`    // <--- ADD THIS
	Difficulty   string     `json:"difficulty"`    // <--- ADD THIS
	MovementType string     `json:"movement_type"` // <--- ADD THIS
	MuscleGroup  string     `json:"muscle_group"`  // <--- ADD THIS
	Equipment    string     `json:"equipment"`     // <--- ADD THIS
	Bodypart     string     `json:"bodypart"`      // <--- ADD THIS
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

// ListExerciseResponse represents the paginated list response for exercises.
type ListExerciseResponse struct {
	Data []ExerciseResponse `json:"data"`
	provider.PaginationResponse
}

// CreateExerciseRequest defines the request for creating multiple exercises.
type CreateExerciseRequest struct {
	Exercises []ExerciseNameOnly `json:"exercises" validate:"required,dive"`
}

// ExerciseNameOnly is a helper struct for creating exercises by name.
type ExerciseNameOnly struct {
	Name string `json:"name" validate:"required"`
}

// CreateExerciseResponse is the response for a successful exercise creation.
type CreateExerciseResponse struct {
	Message   string             `json:"message"`
	Exercises []ExerciseResponse `json:"exercises"`
}
