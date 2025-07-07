package handlers

import (
	"database/sql" // Import for *sql.DB
	"time"

	"rtglabs-go/dto"
	// "rtglabs-go/ent" // We will remove this import as we are moving away from ent
	"github.com/Masterminds/squirrel" // Import squirrel
	"rtglabs-go/model"                // We'll assume you have a model.Exercise struct now
)

// ExerciseHandler holds the database client and squirrel statement builder.
type ExerciseHandler struct {
	DB *sql.DB
	sq squirrel.StatementBuilderType
}

// NewExerciseHandler creates and returns a new ExerciseHandler.
// It now takes *sql.DB and initializes squirrel with the appropriate placeholder format.
func NewExerciseHandler(db *sql.DB) *ExerciseHandler {
	return &ExerciseHandler{
		DB: db,
		sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar), // Or squirrel.Question for '?' placeholders
	}
}

// --- Helper Functions ---

// toExerciseResponse converts a model.Exercise entity to a dto.ExerciseResponse DTO.
// This function will now accept model.Exercise, not ent.Exercise.
func toExerciseResponse(ex *model.Exercise) dto.ExerciseResponse { // Changed parameter type
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
