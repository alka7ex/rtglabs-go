package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateWorkoutRequest represents the request body for creating a new workout.
type CreateWorkoutRequest struct {
	Name      string                         `json:"name" validate:"required,max=255"`
	Exercises []CreateWorkoutExerciseRequest `json:"exercises" validate:"required,min=1,dive"` // `dive` validates each element in the slice
}

// CreateWorkoutExerciseRequest represents a single exercise within a workout creation request.
type CreateWorkoutExerciseRequest struct {
	ID                       *uuid.UUID `json:"id" validate:"omitempty,uuid"`
	ExerciseID               uuid.UUID  `json:"exercise_id" validate:"required,uuid"`
	Order                    *uint      `json:"order" validate:"omitempty,min=1"`
	Sets                     *uint      `json:"sets" validate:"omitempty,min=0"`
	Weight                   *float64   `json:"weight" validate:"omitempty,min=0"`
	Reps                     *uint      `json:"reps" validate:"omitempty,min=0"`
	ExerciseInstanceClientID *string    `json:"exercise_instance_client_id,omitempty"` // Client-side temp ID for grouping
}

// WorkoutResponse represents the full workout details to be returned in a response.
type WorkoutResponse struct {
	ID               uuid.UUID                 `json:"id"`
	UserID           uuid.UUID                 `json:"user_id"`
	Name             string                    `json:"name"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
	DeletedAt        *time.Time                `json:"deleted_at,omitempty"`
	WorkoutExercises []WorkoutExerciseResponse `json:"workout_exercises,omitempty"`
}

// WorkoutExerciseResponse represents a single exercise associated with a workout (pivot data).
type WorkoutExerciseResponse struct {
	ID                 uuid.UUID                 `json:"id"`
	WorkoutID          uuid.UUID                 `json:"workout_id"`
	ExerciseID         uuid.UUID                 `json:"exercise_id"`
	ExerciseInstanceID *uuid.UUID                `json:"exercise_instance_id,omitempty"`
	Order              *uint                     `json:"order,omitempty"`
	Sets               *uint                     `json:"sets,omitempty"`
	Weight             *float64                  `json:"weight,omitempty"`
	Reps               *uint                     `json:"reps,omitempty"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
	DeletedAt          *time.Time                `json:"deleted_at,omitempty"`
	Exercise           *ExerciseResponse         `json:"exercise,omitempty"`          // Eager loaded
	ExerciseInstance   *ExerciseInstanceResponse `json:"exercise_instance,omitempty"` // Eager loaded
}

// ExerciseInstanceResponse represents an ExerciseInstance record.
type ExerciseInstanceResponse struct {
	ID           uuid.UUID  `json:"id"`
	WorkoutLogID *uuid.UUID `json:"workout_log_id,omitempty"`
	ExerciseID   uuid.UUID  `json:"exercise_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// CreateWorkoutResponse is the response for a successful workout creation.
type CreateWorkoutResponse struct {
	Message string          `json:"message"`
	Workout WorkoutResponse `json:"workout"`
}

// ListWorkoutResponse represents the paginated list response for workouts.
type ListWorkoutResponse struct {
	CurrentPage  int               `json:"current_page"`
	Data         []WorkoutResponse `json:"data"`
	FirstPageURL string            `json:"first_page_url"`
	From         *int              `json:"from"`
	LastPage     int               `json:"last_page"`
	LastPageURL  string            `json:"last_page_url"`
	Links        []Link            `json:"links"`
	NextPageURL  *string           `json:"next_page_url"`
	Path         string            `json:"path"`
	PerPage      int               `json:"per_page"`
	PrevPageURL  *string           `json:"prev_page_url"`
	To           *int              `json:"to"`
	Total        int               `json:"total"`
}

// DeleteWorkoutResponse is the response for a successful workout deletion.
type DeleteWorkoutResponse struct {
	Message string `json:"message"`
}
