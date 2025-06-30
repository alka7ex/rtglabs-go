package handlers

import (
	"net/http"
	"time"

	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"rtglabs-go/ent/user"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Login handles user login and session creation (maps to MVC 'Store' or 'Post').
func (h *AuthHandler) StoreLogin(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	entUser, err := h.Client.User.Query().
		Where(user.EmailEQ(req.Email)).
		WithProfile(). // Eager-load the profile
		Only(c.Request().Context())
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
		}
		c.Logger().Error("Database query error:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to authenticate")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(entUser.Password), []byte(req.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
	}

	token := uuid.New().String()
	expiry := time.Now().Add(7 * 24 * time.Hour)

	_, err = h.Client.Session.Create().
		SetToken(token).
		SetExpiresAt(expiry).
		SetUser(entUser).
		Save(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to create session:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create session")
	}

	// Build the response DTO from the fetched Ent entity.
	responseUser := dto.UserWithProfileResponse{
		BaseUserResponse: dto.BaseUserResponse{
			ID:              entUser.ID,
			Name:            entUser.Name,
			Email:           entUser.Email,
			EmailVerifiedAt: entUser.EmailVerifiedAt,
			CreatedAt:       entUser.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:       entUser.UpdatedAt.Format(time.RFC3339Nano),
		},
	}

	// Check if the profile edge was loaded and populate the Profile DTO.
	if profile, err := entUser.Edges.ProfileOrErr(); err == nil && profile != nil {
		responseUser.Profile = &dto.ProfileResponse{
			ID:        profile.ID,
			UserID:    profile.UserID,
			Units:     profile.Units,
			Gender:    profile.Gender,
			Age:       profile.Age,
			Height:    profile.Height,
			Weight:    profile.Weight,
			CreatedAt: profile.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt: profile.UpdatedAt.Format(time.RFC3339Nano),
		}
	}

	// Create the final response object.
	response := dto.LoginResponse{
		Message:   "Logged in successfully!",
		User:      responseUser,
		Token:     token,
		ExpiresAt: expiry.Format("2006-01-02 15:04:05"),
	}

	return c.JSON(http.StatusOK, response)
}
