package handlers

import (
	"database/sql" // Added for sql.ErrNoRows and sql.Null* types
	"errors"       // Added for errors.Is
	"net/http"
	"strings"
	"time"

	"rtglabs-go/dto"
	"rtglabs-go/model" // Import your model package

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	// Removed: "rtglabs-go/ent", "rtglabs-go/ent/user"
)

// StoreRegister handles user registration
// Assumes AuthHandler has DB *sql.DB and sq squirrel.StatementBuilderType
func (h *AuthHandler) StoreRegister(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Logger().Errorf("StoreRegister: Failed to hash password: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password")
	}

	ctx := c.Request().Context()
	newUserID := uuid.New() // Generate a UUID for the new user

	// 1. Insert the new user
	insertUserQuery, insertUserArgs, err := h.sq.Insert("users"). // Assuming table name is 'users'
									Columns("id", "name", "email", "password", "created_at", "updated_at").
									Values(newUserID, req.Name, req.Email, string(hashedPassword), time.Now(), time.Now()).
									ToSql()
	if err != nil {
		c.Logger().Errorf("StoreRegister: Failed to build insert user query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}

	_, err = h.DB.ExecContext(ctx, insertUserQuery, insertUserArgs...)
	if err != nil {
		// Handle unique constraint violation for email
		// The exact error check depends on your database driver.
		// For PostgreSQL, you might check for pq.Error and its Code.
		// For MySQL, you might check for mysql.MySQLError and its Number.
		// A common approach is to check the error message string or use specific driver packages.
		// This is a generic check; you might need to make it more specific.
		if errors.Is(err, sql.ErrTxDone) { // Example for some database errors after a transaction
			c.Logger().Errorf("StoreRegister: Transaction error: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user due to transaction error")
		} else if err.Error() == "UNIQUE constraint failed: users.email" { // Example for SQLite
			return echo.NewHTTPError(http.StatusConflict, "Email already registered")
		} else if strings.Contains(err.Error(), "duplicate key value violates unique constraint") { // Example for PostgreSQL
			return echo.NewHTTPError(http.StatusConflict, "Email already registered")
		} else if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "for key 'users.email'") { // Example for MySQL
			return echo.NewHTTPError(http.StatusConflict, "Email already registered")
		}
		c.Logger().Errorf("StoreRegister: Failed to create new user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}

	var entUser model.User
	var entProfile model.Profile

	// 2. Query the newly created user and eager-load the profile using a LEFT JOIN.
	// Select all columns needed from both users and profiles tables.
	// The order here must match the order in Scan.
	fetchUserQuery, fetchUserArgs, err := h.sq.Select(
		"u.id", "u.name", "u.email", "u.email_verified_at", "u.created_at", "u.updated_at",
		"p.id", "p.user_id", "p.units", "p.age", "p.height", "p.gender", "p.weight", "p.created_at", "p.updated_at", "p.deleted_at",
	).
		From("users u").
		LeftJoin("profiles p ON u.id = p.user_id").
		Where(squirrel.Eq{"u.id": newUserID}). // Query by the ID of the newly created user
		ToSql()
	if err != nil {
		c.Logger().Errorf("StoreRegister: Failed to build fetch user query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch created user")
	}

	row := h.DB.QueryRowContext(ctx, fetchUserQuery, fetchUserArgs...)

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
			// This case should ideally not happen if the insert succeeded,
			// but it's good for defensive programming.
			c.Logger().Errorf("StoreRegister: Created user not found immediately after insert (possible race condition or DB issue): %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch created user (user not found)")
		}
		c.Logger().Errorf("StoreRegister: Database query error fetching created user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch created user")
	}

	// Populate model.Profile from scanned null-aware types if a profile was found
	// For new users, this will likely be empty as no profile is created during initial registration.
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
		entProfile = model.Profile{} // Ensure it's zeroed if no profile found
	}

	// Prepare response DTO
	responseUser := dto.UserWithProfileResponse{ // Use UserWithProfileResponse to match StoreLogin
		BaseUserResponse: dto.BaseUserResponse{
			ID:              entUser.ID,
			Name:            entUser.Name,
			Email:           entUser.Email,
			EmailVerifiedAt: entUser.EmailVerifiedAt,
			CreatedAt:       entUser.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:       entUser.UpdatedAt.Format(time.RFC3339Nano),
		},
	}

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

	return c.JSON(http.StatusCreated, echo.Map{
		"message": "Registered successfully!",
		"user":    responseUser, // Return the full user response with (empty) profile
	})
}
