package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"rtglabs-go/dto"
	"rtglabs-go/model" // <--- NEW: Import your model package

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// StoreLogin handles user login and session creation
// Assumes AuthHandler has DB *sql.DB and sq squirrel.StatementBuilderType
func (h *AuthHandler) StoreLogin(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	ctx := c.Request().Context()

	var entUser model.User       // <--- Use your model.User struct
	var entProfile model.Profile // <--- Use your model.Profile struct

	// 1. Query User and their Profile using a LEFT JOIN
	// The column names in Select MUST match your database schema.
	// The order here must match the order in Scan.
	// Note: We select all columns for both user and profile
	// If a column is nullable, it must be selected with the correct type.
	userQuery := h.sq.Select(
		"u.id", "u.name", "u.email", "u.password", "u.email_verified_at", "u.created_at", "u.updated_at",
		"p.id", "p.user_id", "p.units", "p.age", "p.height", "p.gender", "p.weight", "p.created_at", "p.updated_at", "p.deleted_at", // Include all profile fields
	).
		From("users u").                            // Assuming 'users' is your table name
		LeftJoin("profiles p ON u.id = p.user_id"). // Assuming 'profiles' is your table name
		Where(squirrel.Eq{"u.email": req.Email}).
		Limit(1) // Ensure only one user is fetched

	sqlQuery, args, err := userQuery.ToSql()
	if err != nil {
		c.Logger().Errorf("StoreLogin: Failed to build SQL query for user and profile: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to authenticate")
	}

	row := h.DB.QueryRowContext(ctx, sqlQuery, args...)

	// Scan results directly into the model structs where possible.
	// For columns from the LEFT JOINed table that might be NULL (e.g., if no profile exists),
	// you still need to use sql.Null* or pointers.
	// We'll scan into temporary variables for the profile part and then
	// manually assign to entProfile if the profile was found.
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
		&entUser.ID, &entUser.Name, &entUser.Email, &entUser.Password, &entUser.EmailVerifiedAt, &entUser.CreatedAt, &entUser.UpdatedAt, // Directly scan into User
		&profileID, &profileUserID, &profileUnits, &profileAge, &profileHeight, &profileGender, &profileWeight, &profileCreatedAt, &profileUpdatedAt, &profileDeletedAt, // Scan into null-aware types for Profile
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
		}
		c.Logger().Errorf("StoreLogin: Database query error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to authenticate")
	}

	// Manually populate entProfile if a profile was found (i.e., profileID is valid)
	// uuid.Nil is the zero value for uuid.UUID, which implies no profile was found if the ID is not valid.
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
		// If no profile was found, reset entProfile to its zero value
		entProfile = model.Profile{}
	}

	// 2. Validate password
	if err := bcrypt.CompareHashAndPassword([]byte(entUser.Password), []byte(req.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
	}

	// 3. Generate token and create session
	token := uuid.New().String()
	expiry := time.Now().Add(7 * 24 * time.Hour)

	insertSessionQuery, insertSessionArgs, err := h.sq.Insert("sessions"). // Assuming 'sessions' is your table name
										Columns("token", "expires_at", "user_id").
										Values(token, expiry, entUser.ID).
										ToSql()
	if err != nil {
		c.Logger().Errorf("StoreLogin: Failed to build insert session query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create session")
	}

	_, err = h.DB.ExecContext(ctx, insertSessionQuery, insertSessionArgs...)
	if err != nil {
		c.Logger().Errorf("StoreLogin: Failed to create session: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create session")
	}

	// 4. Prepare response
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

	// Populate profile if loaded (check if entProfile.ID is not the zero UUID)
	if entProfile.ID != uuid.Nil {
		responseUser.Profile = &dto.ProfileResponse{
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

	// Final response
	response := dto.LoginResponse{
		Message:   "Logged in successfully!",
		User:      responseUser,
		Token:     token,
		ExpiresAt: expiry.Format("2006-01-02 15:04:05"),
	}

	return c.JSON(http.StatusOK, response)
}
