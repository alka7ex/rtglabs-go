package handlers

import (
	"database/sql"
	"rtglabs-go/dto"
	"rtglabs-go/model"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/typesense/typesense-go/v3/typesense"
)

// ExerciseHandler struct (no change)
type ExerciseHandler struct {
	DB              *sql.DB
	sq              squirrel.StatementBuilderType
	TypesenseClient *typesense.Client
}

// NewExerciseHandler function (no change)
func NewExerciseHandler(db *sql.DB, tsClient *typesense.Client) *ExerciseHandler {
	return &ExerciseHandler{
		DB:              db,
		sq:              squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		TypesenseClient: tsClient,
	}
}

// toExerciseResponse (no change, assuming it's correctly defined elsewhere or here)
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
