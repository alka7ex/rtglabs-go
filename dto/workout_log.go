// dto/workoutlog.go
package dto

import (
	"time"

	"github.com/google/uuid"
	"rtglabs-go/provider" // Assuming provider holds PaginationResponse and Link
)

// --- Request DTOs ---

// CreateWorkoutLogRequest defines the request body for starting a new workout log session.
type CreateWorkoutLogRequest struct {
	WorkoutID uuid.UUID `json:"workout_id" validate:"required"` // The ID of the workout template
}

// ListWorkoutLogRequest defines the query parameters for listing workout logs.
type ListWorkoutLogRequest struct {
	Page      int        `query:"page" validate:"min=1"`
	Limit     int        `query:"limit" validate:"min=1,max=100"`
	SortBy    string     `query:"sort_by"`    // e.g., "created_at", "started_at", "status"
	Order     string     `query:"order"`      // "asc" or "desc"
	WorkoutID *uuid.UUID `query:"workout_id"` // Optional: filter by workout template ID
	Status    *int       `query:"status"`     // Optional: filter by workout log status
}

// --- Response DTOs ---

// WorkoutLogResponse represents a single workout log entry for read operations.
// This is the core structure for a workout log when returned in full detail.
type WorkoutLogResponse struct {
	ID                         uuid.UUID             `json:"id"`
	UserID                     uuid.UUID             `json:"user_id"`
	WorkoutID                  uuid.UUID             `json:"workout_id"`
	StartedAt                  *time.Time            `json:"started_at,omitempty"`
	FinishedAt                 *time.Time            `json:"finished_at,omitempty"`
	Status                     int                   `json:"status"` // Assuming status is an int
	TotalActiveDurationSeconds uint                  `json:"total_active_duration_seconds"`
	TotalPauseDurationSeconds  uint                  `json:"total_pause_duration_seconds"`
	CreatedAt                  time.Time             `json:"created_at"`
	UpdatedAt                  time.Time             `json:"updated_at"`
	DeletedAt                  *time.Time            `json:"deleted_at,omitempty"`
	Workout                    WorkoutResponse       `json:"workout"`            // Nested workout object
	ExerciseInstances          []ExerciseInstanceLog `json:"exercise_instances"` // Array of exercise instances
}

// ExerciseInstanceLog represents an exercise instance within a workout log.
// This matches your Zod `exerciseInstanceLogSchema`
type ExerciseInstanceLog struct {
	ID                      uuid.UUID               `json:"id"`
	ExerciseID              uuid.UUID               `json:"exercise_id"`
	Exercise                ExerciseResponse        `json:"exercise"` // Your existing ExerciseResponse
	ExerciseSets            []ExerciseSetResponse   `json:"exercise_sets"`
	ExerciseInstanceDetails ExerciseInstanceDetails `json:"exercise_instance_details"` // Maps to exercise_instance_details
}

// ExerciseSetResponse for an individual set (read response)
type ExerciseSetResponse struct {
	ID                 uuid.UUID  `json:"id"`
	WorkoutLogID       uuid.UUID  `json:"workout_log_id"`
	ExerciseID         uuid.UUID  `json:"exercise_id"`
	ExerciseInstanceID *uuid.UUID `json:"exercise_instance_id,omitempty"` // Nullable in Entgo, pointer here
	Weight             float64    `json:"weight"`
	Reps               *int       `json:"reps,omitempty"` // Nullable in Entgo, pointer here
	SetNumber          int        `json:"set_number"`
	FinishedAt         *time.Time `json:"finished_at,omitempty"`
	Status             int        `json:"status"` // Added status field from schema
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

// ExerciseInstanceDetails to match the nested structure for exercise_instance_details
type ExerciseInstanceDetails struct {
	ID           uuid.UUID  `json:"id"`
	WorkoutLogID *uuid.UUID `json:"workout_log_id,omitempty"` // Nullable in your Zod schema
	ExerciseID   uuid.UUID  `json:"exercise_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// CreateWorkoutLogResponse for a successful creation
type CreateWorkoutLogResponse struct {
	Message    string             `json:"message"`
	WorkoutLog WorkoutLogResponse `json:"workout_log"`
}

// ListWorkoutLogResponse for paginated lists
type ListWorkoutLogResponse struct {
	Data                        []WorkoutLogResponse `json:"data"`
	provider.PaginationResponse                      // Embed common pagination fields
}

// WorkoutLogDetailResponse represents the response for fetching a single workout log.
// This is a type alias to WorkoutLogResponse, reflecting the Zod schema's definition.
type WorkoutLogDetailResponse WorkoutLogResponse
