package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"rtglabs-go/dto"
	"rtglabs-go/model"

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

	var entUser model.User
	var entProfile model.Profile

	// 1. Query User and their Profile using a LEFT JOIN
	userQuery := h.sq.Select(
		"u.id", "u.name", "u.email", "u.password", "u.email_verified_at", "u.created_at", "u.updated_at",
		"p.id", "p.user_id", "p.units", "p.age", "p.height", "p.gender", "p.weight", "p.created_at", "p.updated_at", "p.deleted_at",
	).
		From("users u").
		LeftJoin("profiles p ON u.id = p.user_id").
		Where(squirrel.Eq{"u.email": req.Email}).
		Limit(1)

	sqlQuery, args, err := userQuery.ToSql()
	if err != nil {
		c.Logger().Errorf("StoreLogin: Failed to build SQL query for user and profile: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to authenticate")
	}

	row := h.DB.QueryRowContext(ctx, sqlQuery, args...)

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
		&entUser.ID, &entUser.Name, &entUser.Email, &entUser.Password, &entUser.EmailVerifiedAt, &entUser.CreatedAt, &entUser.UpdatedAt,
		&profileID, &profileUserID, &profileUnits, &profileAge, &profileHeight, &profileGender, &profileWeight, &profileCreatedAt, &profileUpdatedAt, &profileDeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
		}
		c.Logger().Errorf("StoreLogin: Database query error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to authenticate")
	}

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
		entProfile = model.Profile{}
	}

	// 2. Validate password
	if err := bcrypt.CompareHashAndPassword([]byte(entUser.Password), []byte(req.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
	}

	// Delete existing session for this user_id
	deleteSessionQuery, deleteSessionArgs, err := h.sq.Delete("sessions").
		Where(squirrel.Eq{"user_id": entUser.ID}).
		ToSql()
	if err != nil {
		c.Logger().Errorf("StoreLogin: Failed to build delete session query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to manage session (delete query build error)")
	}

	_, err = h.DB.ExecContext(ctx, deleteSessionQuery, deleteSessionArgs...)
	if err != nil {
		c.Logger().Warnf("StoreLogin: Failed to delete old session for user %s: %v", entUser.ID, err)
	}

	// 3. Generate token and create NEW session
	token := uuid.New().String()
	expiry := time.Now().Add(7 * 24 * time.Hour)

	newSessionID := uuid.New()
	currentTime := time.Now() // Get current time for created_at

	// --- FIX START: REMOVED "updated_at" from Columns and Values ---
	insertSessionQuery, insertSessionArgs, err := h.sq.Insert("sessions").
		Columns("id", "token", "expires_at", "user_id", "created_at"). // Removed "updated_at"
		Values(newSessionID, token, expiry, entUser.ID, currentTime).  // Removed currentTime for "updated_at"
		ToSql()
	// --- FIX END ---

	if err != nil {
		c.Logger().Errorf("StoreLogin: Failed to build insert session query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create session (insert query build error)")
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

	response := dto.LoginResponse{
		Message:   "Logged in successfully!",
		User:      responseUser,
		Token:     token,
		ExpiresAt: expiry.Format("2006-01-02 15:04:05"),
	}

	return c.JSON(http.StatusOK, response)
}

