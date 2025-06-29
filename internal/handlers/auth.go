package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"rtglabs-go/ent/profile"
	"rtglabs-go/ent/session"
	"rtglabs-go/ent/user"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Client *ent.Client
}

func NewAuthHandler(client *ent.Client) *AuthHandler {
	return &AuthHandler{Client: client}
}

// ValidateToken checks if the token is valid and not expired.
// Returns the user ID if valid, or error if invalid/expired.
func (h *AuthHandler) ValidateToken(token string) (uuid.UUID, error) {
	ctx := context.Background()

	session, err := h.Client.Session.
		Query().
		Where(
			session.TokenEQ(token),
			session.ExpiresAtGTE(time.Now()),
		).
		Only(ctx)

	if err != nil {
		return uuid.Nil, errors.New("invalid or expired session token")
	}

	return session.ID, nil
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	user, err := h.Client.User.Create().
		SetName(req.Name).
		SetEmail(req.Email).
		SetPassword(string(hashed)).
		Save(c.Request().Context())

	if err != nil {
		return echo.NewHTTPError(http.StatusConflict, "Email already registered")
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "registered", "user_id": user.ID})
}

// Login handles user login and session creation.
func (h *AuthHandler) Login(c echo.Context) error {
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
			CreatedAt:       entUser.CreatedAt.Format(time.RFC3339Nano), // Format here
			UpdatedAt:       entUser.UpdatedAt.Format(time.RFC3339Nano), // Format here
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
			CreatedAt: profile.CreatedAt.Format(time.RFC3339Nano), // Format here
			UpdatedAt: profile.UpdatedAt.Format(time.RFC3339Nano), // Format here
		}
	}

	// Create the final response object.
	response := dto.LoginResponse{
		Message:   "Logged in successfully!",
		User:      responseUser,
		Token:     token,
		ExpiresAt: expiry.Format("2006-01-02 15:04:05"),
	}

	// Return the JSON response.
	return c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) GetProfile(c echo.Context) error {
	// 1. Get the user ID from the context.
	// This relies on a middleware to populate the context with the user's ID.
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	// 2. Query the user and eager-load the profile using .WithProfile().
	entUser, err := h.Client.User.Query().
		Where(user.IDEQ(userID)).
		WithProfile().
		Only(c.Request().Context())
	if err != nil {
		// Differentiate between "not found" and other errors
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}
		c.Logger().Error("Database query error:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve profile")
	}

	// 3. Build the Profile DTO from the loaded edge.
	var profileResponse *dto.ProfileResponse
	if profile, err := entUser.Edges.ProfileOrErr(); err == nil && profile != nil {
		profileResponse = &dto.ProfileResponse{
			ID:        profile.ID,
			UserID:    profile.UserID, // This is now a pointer in the DTO
			Units:     profile.Units,
			Gender:    profile.Gender,
			Age:       profile.Age,
			Height:    profile.Height,
			Weight:    profile.Weight,
			CreatedAt: profile.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt: profile.UpdatedAt.Format(time.RFC3339Nano),
		}
	}

	// 4. Build the final response DTO.
	response := dto.GetProfileResponse{
		UserID:  entUser.ID,
		Email:   entUser.Email,
		Name:    entUser.Name,
		Profile: profileResponse,
	}

	// 5. Return the JSON response.
	return c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) UpdateProfile(c echo.Context) error {
	// 1. Get the authenticated user ID from the context.
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	// 2. Bind the request body to the DTO.
	var req dto.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input: "+err.Error())
	}

	// 3. Find the user's existing profile.
	// We use the edge query to find the profile associated with the user ID.
	entProfile, err := h.Client.Profile.Query().
		Where(profile.HasUserWith(user.IDEQ(userID))).
		Only(c.Request().Context())
	if err != nil {
		if ent.IsNotFound(err) {
			// If no profile exists, you might want to create one instead of returning NotFound.
			// This is a business logic decision. For now, we'll return an error.
			return echo.NewHTTPError(http.StatusNotFound, "Profile not found for this user. Please create one.")
		}
		c.Logger().Error("Database query error:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve profile for update")
	}

	// 4. Update the profile with the new data from the request.
	updatedProfile, err := entProfile.Update().
		SetUnits(req.Units).
		SetAge(req.Age).
		SetHeight(req.Height).
		SetGender(req.Gender).
		SetWeight(req.Weight).
		Save(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to update profile:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update profile")
	}

	// 5. Build the response DTO.
	// You will need to dereference the UserID pointer from the Ent entity.
	var profileUserID uuid.UUID
	if updatedProfile.UserID != nil {
		profileUserID = *updatedProfile.UserID
	}

	responseProfile := dto.ProfileResponse{
		ID:        updatedProfile.ID,
		UserID:    &profileUserID, // Assign the pointer to the DTO's pointer field
		Units:     updatedProfile.Units,
		Age:       updatedProfile.Age,
		Height:    updatedProfile.Height,
		Gender:    updatedProfile.Gender,
		Weight:    updatedProfile.Weight,
		CreatedAt: updatedProfile.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: updatedProfile.UpdatedAt.Format(time.RFC3339Nano),
	}

	response := dto.UpdateProfileResponse{
		Message: "Profile updated successfully!",
		Profile: responseProfile,
	}

	// 6. Return the JSON response.
	return c.JSON(http.StatusOK, response)
}
