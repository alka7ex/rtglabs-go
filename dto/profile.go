package dto

import "github.com/google/uuid"

// GetProfileResponse represents the response body for retrieving a user's profile.
type GetProfileResponse struct {
	UserID  uuid.UUID        `json:"user_id"`
	Email   string           `json:"email"`
	Name    string           `json:"name"`
	Profile *ProfileResponse `json:"profile,omitempty"`
}

// Adjusted ProfileResponse for consistency
type ProfileResponse struct {
	ID        uuid.UUID  `json:"id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Units     int        `json:"units"`
	Gender    int        `json:"gender"`
	Age       int        `json:"age"`
	Height    float64    `json:"height"` // Change back to float64 to match your request
	Weight    float64    `json:"weight"` // Change back to float64 to match your request
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

// UpdateProfileRequest represents the request body for updating a profile.
type UpdateProfileRequest struct {
	Units  int     `json:"units"`
	Age    int     `json:"age"`
	Height float64 `json:"height"`
	Gender int     `json:"gender"`
	Weight float64 `json:"weight"`
}

// UpdateProfileResponse represents the response body for a successful profile update.
type UpdateProfileResponse struct {
	Message string          `json:"message"`
	Profile ProfileResponse `json:"profile"`
}
