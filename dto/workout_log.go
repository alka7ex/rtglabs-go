package dto

import (
	"time"

	"github.com/google/uuid"
	"rtglabs-go/provider" // Assuming this contains your PaginationResponse
)

// --- Request DTOs ---

// ListWorkoutLogRequest defines the query parameters for listing workout logs.
type ListWorkoutLogRequest struct {
	Page      int        `query:"page" validate:"min=1"`
	Limit     int        `query:"limit" validate:"min=1,max=100"`
	SortBy    string     `query:"sort_by"`    // e.g., "created_at", "started_at", "status"
	Order     string     `query:"order"`      // "asc" or "desc"
	WorkoutID *uuid.UUID `query:"workout_id"` // Optional: filter by workout template ID
	Status    *int       `query:"status"`     // Optional: filter by workout log status
}

// CreateWorkoutLogRequest defines the request body for creating a workout log.
type CreateWorkoutLogRequest struct {
	WorkoutID uuid.UUID `json:"workout_id" validate:"required,uuid"` // The ID of the workout template to base this log on
}

// UpdateWorkoutLogRequest defines the request body for updating a workout log.
// Aligned with Zod's `workoutLogUpdateRequestSchema`
type UpdateWorkoutLogRequest struct {
	FinishedAt              *time.Time                            `json:"finished_at"`
	LoggedExerciseInstances []UpdateLoggedExerciseInstanceRequest `json:"logged_exercise_instances"`
}

// UpdateLoggedExerciseInstanceRequest defines an individual exercise instance in an update request.
// Aligned with Zod's `workoutLogUpdateExerciseInstanceRequestSchema`
type UpdateLoggedExerciseInstanceRequest struct {
	ID                             *uuid.UUID                 `json:"id"`                                         // Null for new, ID for existing
	ExerciseID                     *uuid.UUID                 `json:"exercise_id" validate:"required_without=ID"` // Required if creating new LEI
	LoggedExerciseInstanceClientID *string                    `json:"logged_exercise_instance_client_id"`         // Added for client-side ID mapping of new instances
	ExerciseSets                   []UpdateExerciseSetRequest `json:"exercise_sets"`
}

// UpdateExerciseSetRequest defines an individual exercise set in an update request.
// Aligned with Zod's `workoutLogUpdateSetRequestSchema` (after adding `status` to Zod)
type UpdateExerciseSetRequest struct {
	ID                       *uuid.UUID `json:"id"`                   // Null for new, ID for existing
	WorkoutLogID             uuid.UUID  `json:"workout_log_id"`       // **Matches Zod**
	ExerciseID               uuid.UUID  `json:"exercise_id"`          // **Matches Zod**
	LoggedExerciseInstanceID *uuid.UUID `json:"exercise_instance_id"` // **Matches Zod's `exercise_instance_id`**
	SetNumber                *int       `json:"set_number"`           // Nullable
	Weight                   float64    `json:"weight"`               // Non-nullable (assuming `weight` in Zod is `z.number()`)
	Reps                     *int       `json:"reps"`                 // Nullable
	FinishedAt               *time.Time `json:"finished_at"`
	Status                   *int       `json:"status"` // Nullable (matches Zod after adjustment)
}

// --- Response DTOs ---

// CreateWorkoutLogResponse defines the response body for creating a workout log.
type CreateWorkoutLogResponse struct {
	Message    string             `json:"message"`
	WorkoutLog WorkoutLogResponse `json:"workout_log"`
}

// UpdateWorkoutLogResponse will likely just return the updated WorkoutLogResponse
type UpdateWorkoutLogResponse struct {
	Message    string             `json:"message"`
	WorkoutLog WorkoutLogResponse `json:"workout_log"`
}

// ListWorkoutLogResponse for paginated lists of workout logs
type ListWorkoutLogResponse struct {
	Data []WorkoutLogResponse `json:"data"`
	provider.PaginationResponse
}

// WorkoutLogResponse defines the structure for a single workout log in responses.
// Aligned with Zod's `workoutLogSchema` (after adding `status`, `total_active_duration_seconds`, `total_pause_duration_seconds` to Zod)
type WorkoutLogResponse struct {
	ID                         uuid.UUID                   `json:"id"`
	WorkoutID                  uuid.UUID                   `json:"workout_id"` // **Matches Zod (non-nullable)**
	UserID                     uuid.UUID                   `json:"user_id"`
	Workout                    WorkoutResponse             `json:"workout"` // Nested workout object
	StartedAt                  *time.Time                  `json:"started_at"`
	FinishedAt                 *time.Time                  `json:"finished_at"`
	Status                     int                         `json:"status"`                        // **Matches Zod (non-nullable)**
	TotalActiveDurationSeconds uint                        `json:"total_active_duration_seconds"` // **Matches Zod**
	TotalPauseDurationSeconds  uint                        `json:"total_pause_duration_seconds"`  // **Matches Zod**
	CreatedAt                  time.Time                   `json:"created_at"`
	UpdatedAt                  time.Time                   `json:"updated_at"`
	DeletedAt                  *time.Time                  `json:"deleted_at"`
	LoggedExerciseInstances    []LoggedExerciseInstanceLog `json:"logged_exercise_instances"`
}

// LoggedExerciseInstanceLog defines the structure for an exercise instance within a workout log response.
// Aligned with Zod's `exerciseInstanceLogSchema` (after adding `workout_log_id`, `created_at`, `updated_at`, `deleted_at` to Zod and removing `exercise_instance_details` from Zod)
type LoggedExerciseInstanceLog struct {
	ID           uuid.UUID             `json:"id"`
	WorkoutLogID uuid.UUID             `json:"workout_log_id"` // **Matches Zod (non-nullable)**
	ExerciseID   uuid.UUID             `json:"exercise_id"`
	Exercise     ExerciseResponse      `json:"exercise"`
	ExerciseSets []ExerciseSetResponse `json:"exercise_sets"`
	CreatedAt    time.Time             `json:"created_at"` // **Matches Zod**
	UpdatedAt    time.Time             `json:"updated_at"` // **Matches Zod**
	DeletedAt    *time.Time            `json:"deleted_at"` // **Matches Zod**
}

// ExerciseSetResponse defines the structure for an individual exercise set in responses.
// Aligned with Zod's `exerciseSetSchema` (after adding `logged_exercise_instance_id` and `status` to Zod)
type ExerciseSetResponse struct {
	ID                       uuid.UUID  `json:"id"`
	WorkoutLogID             uuid.UUID  `json:"workout_log_id"`
	ExerciseID               uuid.UUID  `json:"exercise_id"`
	LoggedExerciseInstanceID uuid.UUID  `json:"logged_exercise_instance_id"` // **Matches Zod**
	SetNumber                *int       `json:"set_number"`
	Weight                   float64    `json:"weight"`
	Reps                     *int       `json:"reps"`
	FinishedAt               *time.Time `json:"finished_at"`
	Status                   int        `json:"status"` // **Matches Zod (non-nullable)**
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
	DeletedAt                *time.Time `json:"deleted_at"`
}
