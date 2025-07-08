package dto

import (
	"time"

	"github.com/google/uuid"
	"rtglabs-go/provider" // Assuming this contains your PaginationResponse
)

// ListWorkoutLogRequest defines the query parameters for listing workout logs.
type ListWorkoutLogRequest struct {
	Page      int        `query:"page" validate:"min=1"`
	Limit     int        `query:"limit" validate:"min=1,max=100"`
	SortBy    string     `query:"sort_by"`    // e.g., "created_at", "started_at", "status"
	Order     string     `query:"order"`      // "asc" or "desc"
	WorkoutID *uuid.UUID `query:"workout_id"` // Optional: filter by workout template ID
	Status    *int       `query:"status"`     // Optional: filter by workout log status
}

// ListWorkoutLogResponse for paginated lists of workout logs
type ListWorkoutLogResponse struct {
	Data []WorkoutLogResponse `json:"data"`
	provider.PaginationResponse
}

// WorkoutLogResponse represents the full details of a workout log for API responses.
// It reuses the WorkoutResponse from workout.go.
type WorkoutLogResponse struct {
	ID                         uuid.UUID                   `json:"id"`
	WorkoutID                  uuid.UUID                   `json:"workout_id"`
	UserID                     uuid.UUID                   `json:"user_id"`
	StartedAt                  *time.Time                  `json:"started_at"`
	FinishedAt                 *time.Time                  `json:"finished_at"`
	Status                     int                         `json:"status"`
	TotalActiveDurationSeconds uint                        `json:"total_active_duration_seconds"`
	TotalPauseDurationSeconds  uint                        `json:"total_pause_duration_seconds"`
	CreatedAt                  time.Time                   `json:"created_at"`
	UpdatedAt                  time.Time                   `json:"updated_at"`
	DeletedAt                  *time.Time                  `json:"deleted_at"`
	Workout                    WorkoutResponse             `json:"workout"`                   // Reusing WorkoutResponse from workout.go
	LoggedExerciseInstances    []LoggedExerciseInstanceLog `json:"logged_exercise_instances"` // Nested *logged* exercise instances
}

// ExerciseInstanceDetails represents the *logged* details of an exercise instance
// from the NEW 'logged_exercise_instances' table.
type ExerciseInstanceDetails struct {
	ID           uuid.UUID `json:"id"`
	WorkoutLogID uuid.UUID `json:"workout_log_id"`
	ExerciseID   uuid.UUID `json:"exercise_id"`
	// Add other fields from 'logged_exercise_instances' table here
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ExerciseInstanceLog represents an individual exercise performed within a workout log.
// Its ID is the ID from the 'logged_exercise_instances' table.
type LoggedExerciseInstanceLog struct {
	// These fields are directly from 'logged_exercise_instances'
	ID           uuid.UUID `json:"id"`
	WorkoutLogID uuid.UUID `json:"workout_log_id"`
	ExerciseID   uuid.UUID `json:"exercise_id"`
	// The actual base exercise details
	Exercise     ExerciseResponse      `json:"exercise"`      // Reusing ExerciseResponse from exercise.go
	ExerciseSets []ExerciseSetResponse `json:"exercise_sets"` // Nested actual performed sets
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	DeletedAt    *time.Time            `json:"deleted_at,omitempty"`
}

// ExerciseSetResponse represents a single set performed for an exercise instance within a workout log.
type ExerciseSetResponse struct {
	ID           uuid.UUID  `json:"id"`
	WorkoutLogID uuid.UUID  `json:"workout_log_id"`
	ExerciseID   uuid.UUID  `json:"exercise_id"`
	Weight       float64    `json:"weight"`
	Reps         int        `json:"reps"`
	SetNumber    int        `json:"set_number"`
	FinishedAt   *time.Time `json:"finished_at"`
	Status       int        `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
} // Add this to dto/workout_log.go
type CreateWorkoutLogRequest struct {
	WorkoutID uuid.UUID `json:"workout_id" validate:"required,uuid"` // The ID of the workout template to base this log on
}

// And a response DTO
type CreateWorkoutLogResponse struct {
	Message    string             `json:"message"`
	WorkoutLog WorkoutLogResponse `json:"workout_log"`
}
