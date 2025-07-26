package dto

import (
	"time"

	"rtglabs-go/provider" // Assuming this contains your PaginationResponse
)

// ExerciseResponse represents the response structure for a single exercise.
// This is the authoritative definition for ExerciseResponse.
type ExerciseResponse struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
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
