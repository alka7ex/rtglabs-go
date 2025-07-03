package handlers

import (
	"rtglabs-go/dto"
	"rtglabs-go/ent" // Import the exercise ent
	"time"
)

// ExerciseHandler holds the ent.Client for database operations.
type ExerciseHandler struct {
	Client *ent.Client
}

// NewExerciseHandler creates and returns a new ExerciseHandler.
func NewExerciseHandler(client *ent.Client) *ExerciseHandler {
	return &ExerciseHandler{Client: client}
}

// --- Helper Functions ---

// toExerciseResponse converts an ent.Exercise entity to a dto.ExerciseResponse DTO.
func toExerciseResponse(ex *ent.Exercise) dto.ExerciseResponse {
	var deletedAt *time.Time
	if ex.DeletedAt != nil {
		deletedAt = ex.DeletedAt
	}

	return dto.ExerciseResponse{
		ID:        ex.ID,
		Name:      ex.Name,
		CreatedAt: ex.CreatedAt,
		UpdatedAt: ex.UpdatedAt, // Include UpdatedAt from mixin
		DeletedAt: deletedAt,
	}
}
