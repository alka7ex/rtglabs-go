package handlers

import (
	"net/http"
	"time"

	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"rtglabs-go/ent/privatetoken"
	"rtglabs-go/ent/user"
	mail "rtglabs-go/provider" // Import email sender

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// ForgotPasswordHandler holds dependencies for password reset operations.
type ForgotPasswordHandler struct {
	Client      *ent.Client
	EmailSender mail.EmailSender
	AppBaseURL  string
}

// NewForgotPasswordHandler creates a new ForgotPasswordHandler instance.
func NewForgotPasswordHandler(client *ent.Client, emailSender mail.EmailSender, appBaseURL string) *ForgotPasswordHandler {
	return &ForgotPasswordHandler{
		Client:      client,
		EmailSender: emailSender,
		AppBaseURL:  appBaseURL,
	}
}

// ForgotPassword handles sending a password reset email
func (h *ForgotPasswordHandler) ForgotPassword(c echo.Context) error { // Receiver is now ForgotPasswordHandler
	var req dto.ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	entUser, err := h.Client.User.Query().
		Where(user.EmailEQ(req.Email)).
		Only(c.Request().Context())
	if err != nil {
		if ent.IsNotFound(err) {
			return c.JSON(http.StatusOK, map[string]string{"message": "If an account with that email exists, a password reset link has been sent."})
		}
		c.Logger().Error("Database query error:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process request")
	}

	// Invalidate any existing password reset tokens for this user
	_, err = h.Client.PrivateToken.Update().
		Where(
			privatetoken.HasUserWith(user.IDEQ(entUser.ID)),
			privatetoken.TypeEQ(dto.TokenTypePasswordReset),
		).
		SetExpiresAt(time.Now().Add(-1 * time.Hour)).
		Save(c.Request().Context())
	if err != nil && !ent.IsNotFound(err) {
		c.Logger().Error("Failed to invalidate old password reset tokens:", err)
	}

	// Generate a new token
	token := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err = h.Client.PrivateToken.Create().
		SetToken(token).
		SetType(dto.TokenTypePasswordReset).
		SetExpiresAt(expiresAt).
		SetUser(entUser).
		Save(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to create password reset token:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create password reset token")
	}

	resetLink := h.AppBaseURL + "/reset-password?token=" + token // Access from the new struct's field

	err = h.EmailSender.SendPasswordResetEmail(entUser.Email, resetLink) // Access from the new struct's field
	if err != nil {
		c.Logger().Error("Failed to send password reset email:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to send password reset email, but reset token was created.")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "If an account with that email exists, a password reset link has been sent."})
}

// ResetPassword handles resetting the user's password using a token
func (h *ForgotPasswordHandler) ResetPassword(c echo.Context) error { // Receiver is now ForgotPasswordHandler
	var req dto.ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	resetToken, err := h.Client.PrivateToken.Query().
		Where(
			privatetoken.TokenEQ(req.Token),
			privatetoken.TypeEQ(dto.TokenTypePasswordReset),
			privatetoken.ExpiresAtGT(time.Now()),
		).
		WithUser().
		Only(c.Request().Context())
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid or expired password reset token.")
		}
		c.Logger().Error("Database query error:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process request")
	}

	if req.NewPassword != req.ConfirmNewPassword {
		return echo.NewHTTPError(http.StatusBadRequest, "New password and confirmation do not match.")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.Logger().Error("Failed to hash password:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reset password")
	}

	_, err = resetToken.Edges.User.Update().
		SetPassword(string(hashedPassword)).
		Save(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to update user password:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reset password")
	}

	err = h.Client.PrivateToken.DeleteOne(resetToken).Exec(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to delete used password reset token:", err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Your password has been reset successfully."})
}
