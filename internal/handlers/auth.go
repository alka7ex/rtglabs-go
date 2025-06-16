package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"rtglabs-go/dto"
	"rtglabs-go/ent"
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

func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	user, err := h.Client.User.Query().Where(user.EmailEQ(req.Email)).Only(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	token := uuid.New().String()
	expiry := time.Now().Add(7 * 24 * time.Hour) // 7 days

	_, err = h.Client.Session.Create().
		SetToken(token).
		SetExpiresAt(expiry).
		SetUser(user).
		Save(c.Request().Context())

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create session")
	}

	return c.JSON(http.StatusOK, echo.Map{"token": token})
}
