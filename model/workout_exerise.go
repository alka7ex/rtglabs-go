package model

import (
	"time"

	"github.com/google/uuid"
)

// WorkoutExercise represents a row in the 'workout_exercises' table.
// It maps directly to database columns using 'db' tags for SQLX.
//
// IMPORTANT NOTE: As per your Ent schema, all foreign keys (WorkoutID, ExerciseID, ExerciseInstanceID)
// have `Unique()` constraints. This implies a 1:1 relationship with each of Workout, Exercise,
// and ExerciseInstance. This is highly unusual for a "WorkoutExercise" entity which typically
// serves as a many-to-many join table. Your database DDL (below) will reflect these
// UNIQUE constraints on the foreign key columns.
type WorkoutExercise struct {
	ID                 uuid.UUID  `db:"id" json:"id"`                                             // From custommixin.UUID
	WorkoutOrder       *int       `db:"order" json:"order"`                             // Nullable, uses pointer
	Sets               *int       `db:"sets" json:"sets"`                               // Nullable, uses pointer
	Weight             *float64   `db:"weight" json:"weight"`                           // Nullable, uses pointer
	Reps               *int       `db:"reps" json:"reps"`                               // Nullable, uses pointer
	WorkoutID          uuid.UUID  `db:"workout_id" json:"workoutId"`                              // FK to workouts.id, UNIQUE and NOT NULL
	ExerciseID         uuid.UUID  `db:"exercise_id" json:"exerciseId"`                            // FK to exercises.id, UNIQUE and NOT NULL
	ExerciseInstanceID *uuid.UUID `db:"exercise_instance_id" json:"exerciseInstanceId"` // FK to exercise_instances.id, UNIQUE and NULLABLE
	CreatedAt          time.Time  `db:"created_at" json:"createdAt"`                              // From custommixin.Timestamps
	UpdatedAt          time.Time  `db:"updated_at" json:"updatedAt"`                              // From custommixin.Timestamps
	DeletedAt          *time.Time `db:"deleted_at" json:"deletedAt"`                    // From custommixin.Timestamps (for soft deletes), nullable
}

// NOTE ON EDGES (Relationships):
// These fields (`WorkoutID`, `ExerciseID`, `ExerciseInstanceID`) are the foreign keys.
// You would fetch related `Workout`, `Exercise`, or `ExerciseInstance` entities
// by performing SQL JOINs from the `workout_exercises` table to their respective tables.
