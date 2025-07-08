package model

import (
	"time"

	"github.com/google/uuid"
)

// LoggedExerciseInstance represents an instance of an exercise performed within a specific workout log.
// This is essentially a log entry for a particular exercise within a workout session.
type LoggedExerciseInstance struct {
	ID           uuid.UUID  `db:"id"`
	WorkoutLogID uuid.UUID  `db:"workout_log_id"`
	ExerciseID   uuid.UUID  `db:"exercise_id"` // The original exercise template ID
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"` // Nullable
}
