package handlers

import (
	"database/sql"
	"rtglabs-go/dto"
	"rtglabs-go/model"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/typesense/typesense-go/v3/typesense"
)

type ExerciseHandler struct {
	DB              *sql.DB
	sq              squirrel.StatementBuilderType
	TypesenseClient *typesense.Client
}

func NewExerciseHandler(db *sql.DB, tsClient *typesense.Client) *ExerciseHandler {
	return &ExerciseHandler{
		DB:              db,
		sq:              squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		TypesenseClient: tsClient,
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
