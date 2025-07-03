package handlers

import (
	"net/http"
	"time"

	"rtglabs-go/dto"
	"rtglabs-go/ent/user"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Register handles user registration
func (h *AuthHandler) StoreRegister(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password")
	}

	newUser, err := h.Client.User.Create().
		SetName(req.Name).
		SetEmail(req.Email).
		SetPassword(string(hashed)).
		Save(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusConflict, "Email already registered")
	}

	// Optional: eager-load the profile (will be nil for new users)
	entUser, err := h.Client.User.Query().
		Where(user.IDEQ(newUser.ID)).
		WithProfile().
		Only(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch created user")
	}

	// Prepare response DTO
	response := dto.BaseUserResponse{
		ID:              entUser.ID,
		Name:            entUser.Name,
		Email:           entUser.Email,
		EmailVerifiedAt: entUser.EmailVerifiedAt,
		CreatedAt:       entUser.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:       entUser.UpdatedAt.Format(time.RFC3339Nano),
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"message": "Registered successfully!",
		"user":    response,
	})
}
