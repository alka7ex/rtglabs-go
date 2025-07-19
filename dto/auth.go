package dto

import (
	"time"

	"github.com/google/uuid"
)

// --- Requests ---

// RegisterRequest represents the request body for user registration.
type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// GoogleLoginRequest represents the payload for Google login.
type GoogleLoginRequest struct {
	IDToken string `json:"id_token" validate:"required"` // Google ID token (JWT)
}

// ForgotPasswordRequest for the forgot password endpoint
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest for the reset password endpoint
type ResetPasswordRequest struct {
	Token              string `json:"token" validate:"required"`
	NewPassword        string `json:"new_password" validate:"required,min=8"` // Example validation
	ConfirmNewPassword string `json:"confirm_new_password" validate:"required,eqfield=NewPassword"`
}

// LogoutRequest represents the request body for logging out a specific session.
type LogoutRequest struct {
	Token string `json:"token" validate:"required"`
}

// --- Responses ---

// BaseUserResponse represents the core user data fields for a response.
// This can be embedded in other response structs to avoid duplication.
type BaseUserResponse struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       string     `json:"created_at"` // Using string for formatted time
	UpdatedAt       string     `json:"updated_at"` // Using string for formatted time
}

// UserWithProfileResponse embeds BaseUserResponse and includes the Profile.
type UserWithProfileResponse struct {
	BaseUserResponse
	Profile *ProfileResponse `json:"profile"`
}

// RegisterResponse represents the response body for a successful registration.
// NOW INCLUDES TOKEN AND EXPIRES_AT
type RegisterResponse struct {
	Message   string                  `json:"message"`
	User      UserWithProfileResponse `json:"user"`
	Token     string                  `json:"token"`
	ExpiresAt string                  `json:"expires_at"` // Using a string to format the time
}

// LoginResponse represents the full response body for a successful login.
type LoginResponse struct {
	Message   string                  `json:"message"`
	User      UserWithProfileResponse `json:"user"`
	Token     string                  `json:"token"`
	ExpiresAt string                  `json:"expires_at"` // Using a string to format the time
}

// GoogleLoginResponse represents the full response body for a successful Google login.
type GoogleLoginResponse struct {
	Message   string                  `json:"message"`    // e.g. "Login successful"
	User      UserWithProfileResponse `json:"user"`       // Full user object
	Token     string                  `json:"token"`      // JWT or session token
	ExpiresAt string                  `json:"expires_at"` // Expiration time as string (e.g., ISO 8601)
}

// LogoutResponse represents the response body for a successful logout.
type LogoutResponse struct {
	Message string `json:"message"`
}
