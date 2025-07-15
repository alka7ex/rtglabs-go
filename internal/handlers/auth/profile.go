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
	var entProfile model.Profile // This will hold the data from the 'profiles' table

	// --- Query User and Profile Data with LEFT JOIN ---
	// Select all columns needed from both users and profiles tables.
	// Note: We are NOT selecting 'p.weight' here as it will come from bodyweights.
	sqlQuery, args, err := h.sq.Select(
		"u.id", "u.name", "u.email", "u.email_verified_at", "u.created_at", "u.updated_at",
		"p.id", "p.user_id", "p.units", "p.age", "p.height", "p.gender", "p.created_at", "p.updated_at", "p.deleted_at", // Exclude p.weight
	).
		From("users u").
		LeftJoin("profiles p ON u.id = p.user_id").
		Where(squirrel.Eq{"u.id": userID}).
		ToSql()

	if err != nil {
		c.Logger().Errorf("GetProfile: Failed to build user/profile query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve user/profile data")
	}

	row := h.DB.QueryRowContext(ctx, sqlQuery, args...)

	// Variables for scanning potentially NULL profile fields from LEFT JOIN
	var (
		profileID        sql.NullString
		profileUserID    sql.NullString
		profileUnits     sql.NullInt64
		profileAge       sql.NullInt64
		profileHeight    sql.NullFloat64
		profileGender    sql.NullInt64
		profileCreatedAt sql.NullTime
		profileUpdatedAt sql.NullTime
		profileDeletedAt sql.NullTime
	)

	err = row.Scan(
		&entUser.ID, &entUser.Name, &entUser.Email, &entUser.EmailVerifiedAt, &entUser.CreatedAt, &entUser.UpdatedAt,
		&profileID, &profileUserID, &profileUnits, &profileAge, &profileHeight, &profileGender, &profileCreatedAt, &profileUpdatedAt, &profileDeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}
		c.Logger().Errorf("GetProfile: Database query error for user/profile: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve user/profile data")
	}

	// Populate model.Profile from scanned null-aware types if a profile was found
	if profileID.Valid {
		entProfile.ID = uuid.MustParse(profileID.String)
		entProfile.UserID = uuid.MustParse(profileUserID.String)
		entProfile.Units = int(profileUnits.Int64)
		entProfile.Age = int(profileAge.Int64)
		entProfile.Height = profileHeight.Float64
		entProfile.Gender = int(profileGender.Int64)
		entProfile.CreatedAt = profileCreatedAt.Time
		entProfile.UpdatedAt = profileUpdatedAt.Time
		if profileDeletedAt.Valid {
			entProfile.DeletedAt = &profileDeletedAt.Time
		} else {
			entProfile.DeletedAt = nil
		}
	} else {
		// If no profile was found, ensure entProfile is its zero value
		entProfile = model.Profile{}
	}

	// --- Query the Latest Bodyweight ---
	var latestBodyweightValue sql.NullFloat64 // This will hold the weight value, can be NULL
	var latestBodyweightUnit sql.NullString   // Optional: if you also want the unit

	latestBWQuery, latestBWArgs, err := h.sq.Select(
		"weight", "unit",
	).
		From("bodyweights").
		Where(squirrel.And{
			squirrel.Eq{"user_id": userID},
			squirrel.Expr("deleted_at IS NULL"),
		}).
		OrderBy("created_at DESC").
		Limit(1).
		ToSql()

	if err != nil {
		c.Logger().Errorf("GetProfile: Failed to build latest bodyweight query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve latest bodyweight")
	}

	bwRow := h.DB.QueryRowContext(ctx, latestBWQuery, latestBWArgs...)

	err = bwRow.Scan(&latestBodyweightValue, &latestBodyweightUnit)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.Logger().Errorf("GetProfile: Database query error for latest bodyweight: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve latest bodyweight")
	}

	// --- Build the Profile DTO ---
	var profileResponse *dto.ProfileResponse
	if entProfile.ID != uuid.Nil || latestBodyweightValue.Valid {
		// Create a profile response. If entProfile is nil, it means no profile table entry,
		// but we still want to potentially show the weight.
		profileResponse = &dto.ProfileResponse{}

		// Populate fields from the 'profiles' table (if a profile existed)
		if entProfile.ID != uuid.Nil {
			profileResponse.ID = entProfile.ID
			profileResponse.UserID = entProfile.UserID
			profileResponse.Units = entProfile.Units
			profileResponse.Gender = entProfile.Gender
			profileResponse.Age = entProfile.Age
			profileResponse.Height = entProfile.Height
			profileResponse.CreatedAt = entProfile.CreatedAt.Format(time.RFC3339Nano)
			profileResponse.UpdatedAt = entProfile.UpdatedAt.Format(time.RFC3339Nano)
			// Note: entProfile.Weight is not set because it's sourced from bodyweights
		}

		// Always populate Weight from the latest bodyweight if available
		if latestBodyweightValue.Valid {
			profileResponse.Weight = latestBodyweightValue.Float64
			// You might also want to add a Unit field to ProfileResponse if it's relevant here
			// profileResponse.WeightUnit = latestBodyweightUnit.String // Example if you add WeightUnit to ProfileResponse DTO
		} else {
			// If no latest bodyweight, set weight to nil/zero value
			profileResponse.Weight = 0.0 // Or nil if Weight is a pointer in DTO
		}
	}

	// 5. Build the final response DTO.
	response := dto.GetProfileResponse{
		UserID:  entUser.ID,
		Email:   entUser.Email,
		Name:    entUser.Name,
		Profile: profileResponse, // This field now combines profile and latest weight
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
	tx, err := h.DB.BeginTx(ctx, nil) // Start a transaction for atomicity
	if err != nil {
		c.Logger().Errorf("UpdateProfile: Failed to begin transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update profile")
	}
	defer tx.Rollback() // Rollback on error unless committed

	var existingProfile model.Profile

	// 3. Try to find the user's existing profile
	selectProfileQuery, selectProfileArgs, err := h.sq.Select("id", "user_id").
		From("profiles").
		Where(squirrel.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		c.Logger().Errorf("UpdateProfile: Failed to build select profile query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve profile for update")
	}

	err = tx.QueryRowContext(ctx, selectProfileQuery, selectProfileArgs...).
		Scan(&existingProfile.ID, &existingProfile.UserID)

	if errors.Is(err, sql.ErrNoRows) {
		// --- Create a new profile if one doesn't exist ---
		fmt.Println("Profile not found, creating a new one for user:", userID.String())

		// Note: 'Weight' is NOT included in the INSERT for 'profiles' table
		insertProfileQuery, insertProfileArgs, err := h.sq.Insert("profiles").
			Columns("id", "user_id", "units", "age", "height", "gender").
			Values(uuid.New(), userID, req.Units, req.Age, req.Height, req.Gender).
			ToSql()
		if err != nil {
			c.Logger().Errorf("UpdateProfile: Failed to build create profile query: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create profile")
		}

		_, err = tx.ExecContext(ctx, insertProfileQuery, insertProfileArgs...)
		if err != nil {
			c.Logger().Errorf("UpdateProfile: Failed to create new profile: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create profile")
		}
	} else if err != nil {
		// Handle other database errors
		c.Logger().Errorf("UpdateProfile: Database query error for existing profile: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve profile for update")
	} else {
		// --- Update the existing profile if it was found ---
		fmt.Println("Existing profile found, updating it for user:", userID.String())

		// Note: 'Weight' is NOT included in the UPDATE for 'profiles' table
		updateProfileQuery, updateProfileArgs, err := h.sq.Update("profiles").
			Set("units", req.Units).
			Set("age", req.Age).
			Set("height", req.Height).
			Set("gender", req.Gender).
			Set("updated_at", time.Now()). // Manually set updated_at
			Where(squirrel.Eq{"id": existingProfile.ID}).
			ToSql()
		if err != nil {
			c.Logger().Errorf("UpdateProfile: Failed to build update profile query: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update profile")
		}

		_, err = tx.ExecContext(ctx, updateProfileQuery, updateProfileArgs...)
		if err != nil {
			c.Logger().Errorf("UpdateProfile: Failed to update profile: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update profile")
		}
	}

	// --- Handle Bodyweight Insertion (always create a new record if weight is provided) ---
	// It's good practice to ensure weight is a positive value if applicable for your domain.
	if req.Weight > 0 { // Only insert if weight is provided and valid (e.g., > 0)
		insertBodyweightQuery, insertBodyweightArgs, err := h.sq.Insert("bodyweights").
			Columns("id", "user_id", "weight", "unit").   // Assuming a default unit or you'll get it from request
			Values(uuid.New(), userID, req.Weight, "kg"). // TODO: Determine default unit or pass from req
			ToSql()
		if err != nil {
			c.Logger().Errorf("UpdateProfile: Failed to build insert bodyweight query: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to record bodyweight")
		}

		_, err = tx.ExecContext(ctx, insertBodyweightQuery, insertBodyweightArgs...)
		if err != nil {
			c.Logger().Errorf("UpdateProfile: Failed to insert new bodyweight: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to record bodyweight")
		}
	}

	// Commit the transaction if all operations were successful
	if err = tx.Commit(); err != nil {
		c.Logger().Errorf("UpdateProfile: Failed to commit transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to complete profile update")
	}

	// Re-fetch and return the updated profile (which will now correctly include the new latest weight)
	return h.GetProfile(c)
}
