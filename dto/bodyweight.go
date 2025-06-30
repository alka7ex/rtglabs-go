package dto

import (
	"time"

	"github.com/google/uuid"
)

// --- Request DTOs ---

// CreateBodyweightRequest defines the request body for creating a new bodyweight record.
// Note: The UserID is typically retrieved from the authentication context, not the request body,
// for security and to prevent users from creating records for other users.
type CreateBodyweightRequest struct {
	Weight float64 `json:"weight" validate:"required,gt=0"`
	Unit   *int    `json:"unit" validate:"required"` // Changed to int to match your JSON example
}

// UpdateBodyweightRequest defines the request body for updating an existing bodyweight record.
type UpdateBodyweightRequest struct {
	Weight float64 `json:"weight" validate:"required,gt=0"`
	Unit   *int    `json:"unit" validate:"required"` // Changed to int to match your JSON example
}

// --- Response DTOs ---

// BodyweightResponse is the base DTO for a single bodyweight record.
// It is used in show, create, and update responses.
type BodyweightResponse struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Weight    float64    `json:"weight"`
	Unit      int        `json:"unit"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"` // Use a pointer for nullable field
}

// CreateBodyweightResponse defines the structure for a successful creation response.
type CreateBodyweightResponse struct {
	Message    string             `json:"message"`
	Bodyweight BodyweightResponse `json:"bodyweight"`
}

// UpdateBodyweightResponse defines the structure for a successful update response.
type UpdateBodyweightResponse struct {
	Message    string             `json:"message"`
	Bodyweight BodyweightResponse `json:"bodyweight"`
}

// DeleteBodyweightResponse defines the structure for a successful deletion response.
type DeleteBodyweightResponse struct {
	Message string `json:"message"`
}

// --- Pagination DTOs ---

// Link defines the structure for a pagination link.
type Link struct {
	URL    *string `json:"url"` // Use a pointer for nullable URLs
	Label  string  `json:"label"`
	Active bool    `json:"active"`
}

// ListBodyweightResponse defines the structure for a paginated list of bodyweight records.
type ListBodyweightResponse struct {
	CurrentPage  int                  `json:"current_page"`
	Data         []BodyweightResponse `json:"data"`
	FirstPageURL string               `json:"first_page_url"`
	From         *int                 `json:"from"` // Use a pointer for nullable field
	LastPage     int                  `json:"last_page"`
	LastPageURL  string               `json:"last_page_url"`
	Links        []Link               `json:"links"`
	NextPageURL  *string              `json:"next_page_url"` // Use a pointer for nullable field
	Path         string               `json:"path"`
	PerPage      int                  `json:"per_page"`
	PrevPageURL  *string              `json:"prev_page_url"` // Use a pointer for nullable field
	To           *int                 `json:"to"`            // Use a pointer for nullable field
	Total        int                  `json:"total"`
}

