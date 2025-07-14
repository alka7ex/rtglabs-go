package dto

import (
	"rtglabs-go/provider"
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
	DeletedAt *time.Time `json:"deleted_at"` // Use a pointer for nullable field
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

// ListBodyweightResponse defines the structure for a paginated list of bodyweight records.
// It now embeds the util.PaginationResponse.
type ListBodyweightResponse struct {
	Data                        []BodyweightResponse `json:"data"`
	provider.PaginationResponse                      // Embed the common pagination fields
}
