package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"rtglabs-go/ent"
	"rtglabs-go/ent/user"
)

type AuthHandler struct {
	Client *ent.Client
}

func NewAuthHandler(client *ent.Client) *AuthHandler {
	return &AuthHandler{Client: client}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
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
	var req LoginRequest
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
