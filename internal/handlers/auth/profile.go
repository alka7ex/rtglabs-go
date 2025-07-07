package handlers

import (
	"database/sql" // Added for sql.ErrNoRows and sql.Null* types
	"errors"       // Added for errors.Is
	"fmt"
	"net/http"
	"time"

	"rtglabs-go/dto"
	"rtglabs-go/model" // Import your model package

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AuthHandler should already have DB *sql.DB and sq squirrel.StatementBuilderType
// from previous refactoring steps.

// GetProfile retrieves the authenticated user's profile (maps to MVC 'Get').
func (h *AuthHandler) GetProfile(c echo.Context) error {
	// 1. Get the user ID from the context.
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	ctx := c.Request().Context()

	var entUser model.User
	var entProfile model.Profile

	// 2. Query the user and eager-load the profile using a LEFT JOIN.
	// Select all columns needed from both users and profiles tables.
	// The order here must match the order in Scan.
	// FIX: Assign 3 return values from ToSql()
	sqlQuery, args, err := h.sq.Select( // <--- CORRECTED: Capture all 3 return values
		"u.id", "u.name", "u.email", "u.email_verified_at", "u.created_at", "u.updated_at",
		"p.id", "p.user_id", "p.units", "p.age", "p.height", "p.gender", "p.weight", "p.created_at", "p.updated_at", "p.deleted_at",
	).
		From("users u").
		LeftJoin("profiles p ON u.id = p.user_id").
		Where(squirrel.Eq{"u.id": userID}).
		ToSql()

	if err != nil { // Check for error from ToSql()
		c.Logger().Errorf("GetProfile: Failed to build SQL query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve profile")
	}

	row := h.DB.QueryRowContext(ctx, sqlQuery, args...) // <--- Use sqlQuery and args

	// Variables for scanning potentially NULL profile fields from LEFT JOIN
	var (
		profileID        sql.NullString
		profileUserID    sql.NullString
		profileUnits     sql.NullInt64
		profileAge       sql.NullInt64
		profileHeight    sql.NullFloat64
		profileGender    sql.NullInt64
		profileWeight    sql.NullFloat64
		profileCreatedAt sql.NullTime
		profileUpdatedAt sql.NullTime
		profileDeletedAt sql.NullTime
	)

	err = row.Scan(
		&entUser.ID, &entUser.Name, &entUser.Email, &entUser.EmailVerifiedAt, &entUser.CreatedAt, &entUser.UpdatedAt,
		&profileID, &profileUserID, &profileUnits, &profileAge, &profileHeight, &profileGender, &profileWeight, &profileCreatedAt, &profileUpdatedAt, &profileDeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}
		c.Logger().Errorf("GetProfile: Database query error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve profile")
	}

	// 3. Populate model.Profile from scanned null-aware types if a profile was found
	// Check if profileID is valid, indicating a profile row was joined
	if profileID.Valid {
		entProfile.ID = uuid.MustParse(profileID.String)
		entProfile.UserID = uuid.MustParse(profileUserID.String)
		entProfile.Units = int(profileUnits.Int64)
		entProfile.Age = int(profileAge.Int64)
		entProfile.Height = profileHeight.Float64
		entProfile.Gender = int(profileGender.Int64)
		entProfile.Weight = profileWeight.Float64
		entProfile.CreatedAt = profileCreatedAt.Time
		entProfile.UpdatedAt = profileUpdatedAt.Time
		if profileDeletedAt.Valid {
			entProfile.DeletedAt = &profileDeletedAt.Time
		} else {
			entProfile.DeletedAt = nil
		}
	} else {
		// If no profile was found (e.g., profileID is NULL), ensure entProfile is its zero value
		entProfile = model.Profile{}
	}

	// 4. Build the Profile DTO from the loaded model.Profile.
	var profileResponse *dto.ProfileResponse
	if entProfile.ID != uuid.Nil { // Check if a profile was actually loaded
		profileResponse = &dto.ProfileResponse{
			ID:        entProfile.ID,
			UserID:    entProfile.UserID,
			Units:     entProfile.Units,
			Gender:    entProfile.Gender,
			Age:       entProfile.Age,
			Height:    entProfile.Height,
			Weight:    entProfile.Weight,
			CreatedAt: entProfile.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt: entProfile.UpdatedAt.Format(time.RFC3339Nano),
		}
	}

	// 5. Build the final response DTO.
	response := dto.GetProfileResponse{
		UserID:  entUser.ID,
		Email:   entUser.Email,
		Name:    entUser.Name,
		Profile: profileResponse,
	}

	// 6. Return the JSON response.
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

	ctx := c.Request().Context()

	var existingProfile model.Profile

	// 3. Try to find the user's existing profile
	// Select only the ID and UserID to check existence and get the profile's ID.
	selectProfileQuery, selectProfileArgs, err := h.sq.Select("id", "user_id").
		From("profiles").
		Where(squirrel.Eq{"user_id": userID}).
		ToSql()
	if err != nil { // Check for error from ToSql()
		c.Logger().Errorf("UpdateProfile: Failed to build select profile query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve profile for update")
	}

	err = h.DB.QueryRowContext(ctx, selectProfileQuery, selectProfileArgs...).
		Scan(&existingProfile.ID, &existingProfile.UserID)

	if errors.Is(err, sql.ErrNoRows) {
		// --- Create a new profile if one doesn't exist ---
		fmt.Println("Profile not found, creating a new one for user:", userID.String())

		insertProfileQuery, insertProfileArgs, err := h.sq.Insert("profiles").
			Columns("id", "user_id", "units", "age", "height", "gender", "weight"). // Assuming 'id' is explicitly set (UUID)
			Values(uuid.New(), userID, req.Units, req.Age, req.Height, req.Gender, req.Weight).
			ToSql()
		if err != nil { // Check for error from ToSql()
			c.Logger().Errorf("UpdateProfile: Failed to build create profile query: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create profile")
		}

		_, err = h.DB.ExecContext(ctx, insertProfileQuery, insertProfileArgs...)
		if err != nil {
			c.Logger().Errorf("UpdateProfile: Failed to create new profile: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create profile")
		}

		// Re-fetch the newly created profile to get all fields, including auto-generated timestamps.
		// Note: Depending on your DB and table setup, you might be able to use RETURNING clause with some drivers
		// to get the full object directly after insert. For simplicity, we'll re-fetch.
		return h.GetProfile(c) // Use GetProfile to fetch and return the newly created profile
	} else if err != nil {
		// Handle other database errors
		c.Logger().Errorf("UpdateProfile: Database query error for existing profile: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve profile for update")
	} else {
		// --- Update the existing profile if it was found ---
		fmt.Println("Existing profile found, updating it for user:", userID.String())

		updateProfileQuery, updateProfileArgs, err := h.sq.Update("profiles").
			Set("units", req.Units).
			Set("age", req.Age).
			Set("height", req.Height).
			Set("gender", req.Gender).
			Set("weight", req.Weight).
			Set("updated_at", time.Now()). // Manually set updated_at
			Where(squirrel.Eq{"id": existingProfile.ID}).
			ToSql()
		if err != nil { // Check for error from ToSql()
			c.Logger().Errorf("UpdateProfile: Failed to build update profile query: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update profile")
		}

		_, err = h.DB.ExecContext(ctx, updateProfileQuery, updateProfileArgs...)
		if err != nil {
			c.Logger().Errorf("UpdateProfile: Failed to update profile: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update profile")
		}

		// Re-fetch the updated profile to get all fields including updated timestamps.
		return h.GetProfile(c) // Use GetProfile to fetch and return the updated profile
	}
}

