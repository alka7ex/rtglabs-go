package model

import (
	"time"

	"github.com/google/uuid"
)

// Workout represents a row in the 'workouts' table.
// It maps directly to database columns using 'db' tags for SQLX.
// As per the Ent schema's `Unique()` edge.From, each user can have only one such workout.
type Workout struct {
	ID        uuid.UUID  `db:"id" json:"id"`          // From custommixin.UUID
	UserID    uuid.UUID  `db:"user_id" json:"userId"` // Foreign Key to users.id. UNIQUE and NOT NULL in DB.
	Name      string     `db:"name" json:"name"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`           // From custommixin.Timestamps
	UpdatedAt time.Time  `db:"updated_at" json:"updatedAt"`           // From custommixin.Timestamps
	DeletedAt *time.Time `db:"deleted_at" json:"deletedAt"` // From custommixin.Timestamps (for soft deletes), nullable
}

// NOTE ON EDGES (Relationships):
// - The 1:1 relationship with User (via UserID foreign key) is explained above.
// - The `edge.To` relations (to WorkoutExercise, WorkoutLog) mean that:
//   - The 'workout_exercises' table will have a 'workout_id' foreign key.
//   - The 'workout_logs' table will have a 'workout_id' foreign key.
//
// These related entities are NOT directly embedded in the Workout struct.
// You'll fetch them via separate queries, often using SQL JOINs.
