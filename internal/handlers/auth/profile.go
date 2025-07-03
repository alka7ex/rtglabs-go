package handlers

import (
	"fmt"
	"net/http"
	"time"

	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"rtglabs-go/ent/profile"
	"rtglabs-go/ent/user"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// GetProfile retrieves the authenticated user's profile (maps to MVC 'Get').
func (h *AuthHandler) GetProfile(c echo.Context) error {
	// 1. Get the user ID from the context.
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	// 2. Query the user and eager-load the profile.
	entUser, err := h.Client.User.Query().
		Where(user.IDEQ(userID)).
		WithProfile().
		Only(c.Request().Context())
	if err != nil {
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
			UserID:    profile.Edges.User.ID,
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

// UpdateProfile updates or creates the authenticated user's profile (maps to MVC 'Update').
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

	var updatedProfile *ent.Profile
	var err error

	// 3. Try to find the user's existing profile.
	entProfile, err := h.Client.Profile.Query().
		Where(profile.HasUserWith(user.IDEQ(userID))).
		Only(c.Request().Context())

	if ent.IsNotFound(err) {
		// --- Create a new profile if one doesn't exist ---
		fmt.Println("Profile not found, creating a new one for user:", userID.String())
		updatedProfile, err = h.Client.Profile.Create().
			SetUserID(userID).
			SetUnits(req.Units).
			SetAge(req.Age).
			SetHeight(req.Height).
			SetGender(req.Gender).
			SetWeight(req.Weight).
			Save(c.Request().Context())
		if err != nil {
			c.Logger().Error("Failed to create new profile:", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create profile")
		}
	} else if err != nil {
		// Handle other database errors
		c.Logger().Error("Database query error:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve profile for update")
	} else {
		// --- Update the existing profile if it was found ---
		fmt.Println("Existing profile found, updating it for user:", userID.String())
		updatedProfile, err = entProfile.Update().
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
	}

	// 4. Build the response DTO from the updated/created profile.
	responseProfile := dto.ProfileResponse{
		ID:        updatedProfile.ID,
		UserID:    updatedProfile.Edges.User.ID,
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

	// 5. Return the JSON response.
	return c.JSON(http.StatusOK, response)
}
