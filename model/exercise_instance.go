package model

import (
	"time"

	"github.com/google/uuid"
)

// ExerciseInstance represents a row in the 'exercise_instances' table.
// It maps directly to database columns using 'db' tags for SQLX.
// Both Exercise and WorkoutLog relationships are 1:1 here due to `Unique()` constraints in the Ent schema.
type ExerciseInstance struct {
	ID           uuid.UUID  `db:"id" json:"id"`                                 // From custommixin.UUID
	ExerciseID   uuid.UUID  `db:"exercise_id" json:"exerciseId"`                // Foreign Key to exercises.id, UNIQUE and NOT NULL
	WorkoutLogID *uuid.UUID `db:"workout_log_id" json:"workoutLogId,omitempty"` // Foreign Key to workout_logs.id, UNIQUE and NULLABLE
	CreatedAt    time.Time  `db:"created_at" json:"createdAt"`                  // From custommixin.Timestamps
	UpdatedAt    time.Time  `db:"updated_at" json:"updatedAt"`                  // From custommixin.Timestamps
	DeletedAt    *time.Time `db:"deleted_at" json:"deletedAt,omitempty"`        // From custommixin.Timestamps (for soft deletes), nullable
}

// NOTE ON EDGES (Relationships):
// - The 1:1 relationships with Exercise and WorkoutLog (via their respective IDs) are explained above.
// - The `edge.To` relations (to WorkoutExercise, ExerciseSet) mean that:
//   - The 'workout_exercises' table will have an 'exercise_instance_id' foreign key.
//   - The 'exercise_sets' table will have an 'exercise_instance_id' foreign key.
//
// These related entities are NOT directly embedded in the ExerciseInstance struct.
// You'll fetch them via separate queries, often using SQL JOINs.
