package model

import (
	"time"

	"github.com/google/uuid"
)

// WorkoutLog represents a row in the 'workout_logs' table.
// It maps directly to database columns using 'db' tags for SQLX.
//
// IMPORTANT NOTE: As per your Ent schema, the `Unique()` constraint on the 'user' and 'workout' edges
// implies 1:1 relationships from User/Workout to WorkoutLog. This is highly unusual for a log entity,
// which typically has a 1:N relationship (one user has many logs, one workout has many logs).
// Your database DDL (below) will reflect these UNIQUE constraints on the foreign key columns.
type WorkoutLog struct {
	ID                         uuid.UUID  `db:"id" json:"id"`
	UserID                     uuid.UUID  `db:"user_id" json:"userId"`                                           // FK to users.id, UNIQUE and NOT NULL
	WorkoutID                  *uuid.UUID `db:"workout_id" json:"workoutId,omitempty"`                           // FK to workouts.id, UNIQUE and NULLABLE
	StartedAt                  *time.Time `db:"started_at" json:"startedAt,omitempty"`                           // Nullable, uses pointer
	FinishedAt                 *time.Time `db:"finished_at" json:"finishedAt,omitempty"`                         // Nullable, uses pointer
	Status                     int        `db:"status" json:"status"`                                            // Default 0
	TotalActiveDurationSeconds uint       `db:"total_active_duration_seconds" json:"totalActiveDurationSeconds"` // Default 0
	TotalPauseDurationSeconds  uint       `db:"total_pause_duration_seconds" json:"totalPauseDurationSeconds"`   // Default 0
	CreatedAt                  time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt                  time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt                  *time.Time `db:"deleted_at" json:"deletedAt,omitempty"` // For soft deletes, nullable
}

const (
	WorkoutLogStatusInProgress = 0
	WorkoutLogStatusCompleted  = 1
	WorkoutLogStatusPaused     = 2
	// Add other statuses if needed
)

// NOTE ON EDGES (Relationships):
// - The 1:1 relationships with User and Workout (via UserID and WorkoutID foreign keys) are explained above.
// - The `edge.To` relations (to ExerciseSet, ExerciseInstance) mean that:
//   - The 'exercise_sets' table will have a 'workout_log_id' foreign key.
//   - The 'exercise_instances' table will have a 'workout_log_id' foreign key.
//
// These related entities are NOT directly embedded in the WorkoutLog struct.
// You'll fetch them via separate queries, often using SQL JOINs.
