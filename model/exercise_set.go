package model

import (
	"time"

	"github.com/google/uuid"
)

// ExerciseSet represents a row in the 'exercise_sets' table.
// It maps directly to database columns using 'db' tags for SQLX.
//
// IMPORTANT NOTE: As per your Ent schema, the `Unique()` constraint on all three `edge.From`
// relations (workout_log, exercise, exercise_instance) implies a 1:1 relationship with each.
// This is highly unusual for an "ExerciseSet" which would typically be a part of many,
// and would primarily use a composite key (workout_log_id, exercise_id, set_number)
// or (exercise_instance_id, set_number).
// Your database DDL (below) will reflect these UNIQUE constraints on the foreign key columns.
type ExerciseSet struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	Weight             *float64   `db:"weight" json:"weight,omitempty"`                           // Nullable, uses pointer, decimal(8,2) in DB
	Reps               *int       `db:"reps" json:"reps,omitempty"`                               // Nullable, uses pointer
	SetNumber          int        `db:"set_number" json:"setNumber"`                              // NOT NULL as no Optional() or Nillable()
	FinishedAt         *time.Time `db:"finished_at" json:"finishedAt,omitempty"`                  // Nullable, uses pointer
	Status             int        `db:"status" json:"status"`                                     // Default 0
	WorkoutLogID       uuid.UUID  `db:"workout_log_id" json:"workoutLogId"`                       // FK to workout_logs.id, UNIQUE and NOT NULL
	ExerciseID         uuid.UUID  `db:"exercise_id" json:"exerciseId"`                            // FK to exercises.id, UNIQUE and NOT NULL
	ExerciseInstanceID *uuid.UUID `db:"exercise_instance_id" json:"exerciseInstanceId,omitempty"` // FK to exercise_instances.id, UNIQUE and NULLABLE
	CreatedAt          time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt          *time.Time `db:"deleted_at" json:"deletedAt,omitempty"` // For soft deletes, nullable
}

// NOTE ON EDGES (Relationships):
// These fields (`WorkoutLogID`, `ExerciseID`, `ExerciseInstanceID`) are the foreign keys.
// You would fetch related `WorkoutLog`, `Exercise`, or `ExerciseInstance` entities
// by performing SQL JOINs from the `exercise_sets` table to their respective tables.
