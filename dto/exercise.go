package dto

import (
	"time"

	"rtglabs-go/provider"

	"github.com/google/uuid"
)

// ExerciseResponse represents the response structure for a single exercise.
type ExerciseResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ListExerciseResponse represents the paginated list response for exercises.
// It now embeds the util.PaginationResponse.
type ListExerciseResponse struct {
	Data                        []ExerciseResponse `json:"data"`
	provider.PaginationResponse                    // Embed the common pagination fields
}
type CreateExerciseRequest struct {
	Exercises []ExerciseNameOnly `json:"exercises" validate:"required,dive"`
}

type ExerciseNameOnly struct {
	Name string `json:"name" validate:"required"`
}

type CreateExerciseResponse struct {
	Message   string             `json:"message"`
	Exercises []ExerciseResponse `json:"exercises"`
}
