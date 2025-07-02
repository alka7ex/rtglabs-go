package dto

import (
	"time"

	"github.com/google/uuid"
)

// ExerciseResponse represents the response structure for a single exercise.
type ExerciseResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ListExerciseResponse represents the paginated list response for exercises.
type ListExerciseResponse struct {
	CurrentPage  int                `json:"current_page"`
	Data         []ExerciseResponse `json:"data"`
	FirstPageURL string             `json:"first_page_url"`
	From         *int               `json:"from"`
	LastPage     int                `json:"last_page"`
	LastPageURL  string             `json:"last_page_url"`
	Links        []Link             `json:"links"`
	NextPageURL  *string            `json:"next_page_url"`
	Path         string             `json:"path"`
	PerPage      int                `json:"per_page"`
	PrevPageURL  *string            `json:"prev_page_url"`
	To           *int               `json:"to"`
	Total        int                `json:"total"`
}

// Link is a helper for pagination links.
type Link struct {
	URL    *string `json:"url"`
	Label  string  `json:"label"`
	Active bool    `json:"active"`
}
