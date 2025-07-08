package model

import (
	"time"

	"github.com/google/uuid"
)

type ExerciseSet struct {
	ID                       uuid.UUID  `db:"id"`
	WorkoutLogID             uuid.UUID  `db:"workout_log_id"`
	ExerciseID               uuid.UUID  `db:"exercise_id"`                 // The template exercise ID
	LoggedExerciseInstanceID uuid.UUID  `db:"logged_exercise_instance_id"` // New FK
	SetNumber                int        `db:"set_number"`
	Weight                   *float64   `db:"weight"`                     // Nullable
	Reps                     *int       `db:"reps" json:"reps,omitempty"` // Nullable, uses pointer
	FinishedAt               *time.Time `db:"finished_at"`                // Nullable
	Status                   int        `db:"status"`                     // e.g., 0=Pending, 1=Completed, 2=Skipped
	CreatedAt                time.Time  `db:"created_at"`
	UpdatedAt                time.Time  `db:"updated_at"`
	DeletedAt                *time.Time `db:"deleted_at"` // Nullable
}
