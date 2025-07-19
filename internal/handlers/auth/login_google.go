package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"rtglabs-go/dto"
	"rtglabs-go/model"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/idtoken"
)

// StoreGoogleLogin handles Google Sign-In authentication

func (h *AuthHandler) StoreGoogleLogin(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		IDToken string `json:"id_token" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// 1. Verify ID token
	payload, err := idtoken.Validate(ctx, req.IDToken, h.GoogleClientID)
	if err != nil {
		c.Logger().Errorf("❌ Failed to verify Google ID token: %v", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid ID token")
	}

	sub := payload.Subject
	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	verified, _ := payload.Claims["email_verified"].(bool)

	var user model.User

	// 2. Check if user exists by google_id
	err = h.DB.QueryRowContext(ctx, `
		SELECT id, name, email, email_verified_at, created_at, updated_at
		FROM users WHERE google_id = $1 LIMIT 1
	`, sub).Scan(
		&user.ID, &user.Name, &user.Email, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// 3. Not found by google_id — check by email
		err = h.DB.QueryRowContext(ctx, `
			SELECT id, name, email, email_verified_at, created_at, updated_at
			FROM users WHERE email = $1 LIMIT 1
		`, email).Scan(
			&user.ID, &user.Name, &user.Email, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			// 4. New user
			now := time.Now()
			user = model.User{
				ID:        uuid.New(),
				Name:      name,
				Email:     email,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if verified {
				user.EmailVerifiedAt = &now
			}

			_, err := h.sq.Insert("users").
				Columns("id", "name", "email", "password", "email_verified_at", "created_at", "updated_at", "google_id").
				Values(user.ID, user.Name, user.Email, nil, user.EmailVerifiedAt, user.CreatedAt, user.UpdatedAt, sub).
				RunWith(h.DB).
				ExecContext(ctx)
			if err != nil {
				c.Logger().Errorf("❌ Failed to insert new Google user: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
			}
		} else if err != nil {
			c.Logger().Errorf("❌ Failed to lookup user by email: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check email")
		} else {
			// 5. Found user by email — update google_id
			_, err := h.sq.Update("users").
				Set("google_id", sub).
				Set("updated_at", time.Now()).
				Where("id = ?", user.ID).
				RunWith(h.DB).
				ExecContext(ctx)
			if err != nil {
				c.Logger().Errorf("❌ Failed to link google_id to existing user: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to link Google account")
			}
		}
	} else if err != nil {
		c.Logger().Errorf("❌ Failed to lookup user by Google ID: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error during Google login")
	}

	// 6. Create session
	sessionID := uuid.New()
	token := uuid.New().String()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	createdAt := time.Now()

	_, err = h.sq.Insert("sessions").
		Columns("id", "token", "expires_at", "user_id", "created_at", "provider", "google_id").
		Values(sessionID, token, expiresAt, user.ID, createdAt, "google", sub).
		RunWith(h.DB).
		ExecContext(ctx)
	if err != nil {
		c.Logger().Errorf("❌ Failed to create session: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create session")
	}

	return c.JSON(http.StatusOK, dto.LoginResponse{
		Message: "Logged in successfully via Google!",
		User: dto.UserWithProfileResponse{
			BaseUserResponse: dto.BaseUserResponse{
				ID:              user.ID,
				Name:            user.Name,
				Email:           user.Email,
				EmailVerifiedAt: user.EmailVerifiedAt,
				CreatedAt:       user.CreatedAt.Format(time.RFC3339Nano),
				UpdatedAt:       user.UpdatedAt.Format(time.RFC3339Nano),
			},
		},
		Token:     token,
		ExpiresAt: expiresAt.Format("2006-01-02 15:04:05"),
	})
}
