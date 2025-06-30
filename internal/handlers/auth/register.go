package handlers

import (
	"net/http"

	"rtglabs-go/dto"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Register handles user registration (maps to MVC 'Store' or 'Post').
func (h *AuthHandler) StoreRegister(c echo.Context) error {
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
