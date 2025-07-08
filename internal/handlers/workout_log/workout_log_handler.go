package handler

import (
	"database/sql" // For *sql.DB, sql.Null* types

	"github.com/Masterminds/squirrel" // Import squirrel
)

// WorkoutHandler holds the database client and squirrel statement builder.
type WorkoutLogHandler struct {
	DB *sql.DB
	sq squirrel.StatementBuilderType
}

// NewWorkoutHandler creates and returns a new WorkoutHandler.
// It now takes *sql.DB and initializes squirrel with the appropriate placeholder format.
func NewWorkoutLogHandler(db *sql.DB) *WorkoutLogHandler {
	return &WorkoutLogHandler{
		DB: db,
		sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar), // Or squirrel.Question for '?'
	}
}

