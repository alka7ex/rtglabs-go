package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"rtglabs-go/dto"
	"rtglabs-go/model"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Storeegister handles user registration
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
	newUserID := uuid.New()

	// 1. Insert the new user
	insertUserQuery, insertUserArgs, err := h.sq.Insert("users").
		Columns("id", "name", "email", "password", "email_verified_at", "created_at", "updated_at").
		Values(newUserID, req.Name, req.Email, string(hashedPassword), nil, time.Now(), time.Now()).
		ToSql()
	if err != nil {
		c.Logger().Errorf("StoreRegister: Failed to build insert user query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}

	c.Logger().Infof("Generated INSERT SQL: %s", insertUserQuery)
	c.Logger().Infof("INSERT Args: %v", insertUserArgs)

	_, err = h.DB.ExecContext(ctx, insertUserQuery, insertUserArgs...)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") { // PostgreSQL specific error
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
			c.Logger().Errorf("StoreRegister: Created user not found immediately after insert (possible race condition or DB issue): %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch created user (user not found)")
		}
		c.Logger().Errorf("StoreRegister: Database query error fetching created user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch created user")
	}

	// Populate model.Profile from scanned null-aware types if a profile was found
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

	// 3. Generate token and create NEW session for the newly registered user
	token := uuid.New().String()
	expiry := time.Now().Add(7 * 24 * time.Hour) // Session expires in 7 days

	newSessionID := uuid.New()
	currentTime := time.Now() // Get current time for created_at

	insertSessionQuery, insertSessionArgs, err := h.sq.Insert("sessions").
		Columns("id", "token", "expires_at", "user_id", "created_at").
		Values(newSessionID, token, expiry, entUser.ID, currentTime).
		ToSql()

	if err != nil {
		c.Logger().Errorf("StoreRegister: Failed to build insert session query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create session (insert query build error)")
	}

	_, err = h.DB.ExecContext(ctx, insertSessionQuery, insertSessionArgs...)
	if err != nil {
		c.Logger().Errorf("StoreRegister: Failed to create session: %v", err)
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

	// Construct the RegisterResponse with token and expiry
	response := dto.RegisterResponse{
		Message:   "Registered successfully!",
		User:      responseUser,
		Token:     token,
		ExpiresAt: expiry.Format("2006-01-02 15:04:05"), // Format same as LoginResponse
	}

	return c.JSON(http.StatusCreated, response)
}
