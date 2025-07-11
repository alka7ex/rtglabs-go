package handlers

import (
	"database/sql" // <--- NEW: Import for standard SQL DB
	"errors"
	"log"
	"net/http"
	"time"

	"rtglabs-go/dto"
	mail "rtglabs-go/provider" // Import email sender

	"github.com/Masterminds/squirrel" // <--- NEW: Import squirrel
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	// REMOVE: "rtglabs-go/ent", "rtglabs-go/ent/privatetoken", "rtglabs-go/ent/user"
)

type PrivateToken struct {
	ID        uuid.UUID // Assuming an auto-incrementing ID for private_tokens
	Token     string
	Type      string
	UserID    uuid.UUID // Foreign key to users.id
	ExpiresAt time.Time
}

// ForgotPasswordHandler holds dependencies for password reset operations.
type ForgotPasswordHandler struct {
	DB          *sql.DB // <--- CHANGED: From *ent.Client to *sql.DB
	EmailSender mail.EmailSender
	AppBaseURL  string
	sq          squirrel.StatementBuilderType // Add squirrel builder
}

// NewForgotPasswordHandler creates a new ForgotPasswordHandler instance.
// It now accepts *sql.DB.
func NewForgotPasswordHandler(db *sql.DB, emailSender mail.EmailSender, appBaseURL string) *ForgotPasswordHandler {
	// Initialize squirrel with the appropriate placeholder format for your DB
	// squirrel.Question for MySQL/SQLite, squirrel.Dollar for PostgreSQL
	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	return &ForgotPasswordHandler{
		DB:          db,
		EmailSender: emailSender,
		AppBaseURL:  appBaseURL,
		sq:          sq, // Assign the squirrel builder
	}
}

// ForgotPassword handles sending a password reset email
func (h *ForgotPasswordHandler) ForgotPassword(c echo.Context) error {
	var req dto.ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	ctx := c.Request().Context()

	// 1. Find the user by email
	var entUser User // Use our custom User struct
	query, args, err := h.sq.Select("id", "email").From("users").Where(squirrel.Eq{"email": req.Email}).ToSql()
	if err != nil {
		c.Logger().Errorf("ForgotPassword: Failed to build user query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process request")
	}

	row := h.DB.QueryRowContext(ctx, query, args...)
	err = row.Scan(&entUser.ID, &entUser.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Always return a generic message for security reasons
			return c.JSON(http.StatusOK, map[string]string{"message": "If an account with that email exists, a password reset link has been sent."})
		}
		c.Logger().Errorf("ForgotPassword: Database query error for user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process request")
	}

	// 2. Invalidate any existing password reset tokens for this user
	// This will update all existing tokens for the user to expire immediately.
	updateQuery, updateArgs, err := h.sq.Update("private_tokens"). // Assuming table name is 'private_tokens'
									Set("expires_at", time.Now().Add(-1*time.Hour)).
									Where(squirrel.Eq{
			"user_id": entUser.ID,
			"type":    dto.TokenTypePasswordReset,
		}).
		ToSql()
	if err != nil {
		c.Logger().Errorf("ForgotPassword: Failed to build invalidate token query: %v", err)
		// Don't return an error here, try to continue as it's not critical
	} else {
		_, err = h.DB.ExecContext(ctx, updateQuery, updateArgs...)
		if err != nil && !errors.Is(err, sql.ErrNoRows) { // Check for actual errors, not just no rows affected
			c.Logger().Errorf("ForgotPassword: Failed to invalidate old password reset tokens: %v", err)
		}
	}

	// 3. Generate and create a new token
	token := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Hour)

	insertQuery, insertArgs, err := h.sq.Insert("private_tokens").
		Columns("token", "type", "user_id", "expires_at").
		Values(token, dto.TokenTypePasswordReset, entUser.ID, expiresAt).
		ToSql()
	if err != nil {
		c.Logger().Errorf("ForgotPassword: Failed to build create token query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create password reset token")
	}

	_, err = h.DB.ExecContext(ctx, insertQuery, insertArgs...)
	if err != nil {
		c.Logger().Errorf("ForgotPassword: Failed to create password reset token: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create password reset token")
	}

	// 4. Send email
	resetLink := h.AppBaseURL + "/reset-password?token=" + token

	// resetLink := "https://tunnel.rtglabs.net" + "/reset-password?token=" + token
	err = h.EmailSender.SendPasswordResetEmail(entUser.Email, resetLink)
	if err != nil {
		c.Logger().Errorf("ForgotPassword: Failed to send password reset email: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to send password reset email, but reset token was created.")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "If an account with that email exists, a password reset link has been sent."})
}

// ResetPassword handles resetting the user's password using a token
func (h *ForgotPasswordHandler) ResetPassword(c echo.Context) error {
	var req dto.ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[RESET_PASSWORD_ERROR] Failed to bind request: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	ctx := c.Request().Context()

	if req.NewPassword != req.ConfirmNewPassword {
		log.Printf("[RESET_PASSWORD_WARN] New password and confirmation do not match for token: %s", req.Token)
		return echo.NewHTTPError(http.StatusBadRequest, "New password and confirmation do not match.")
	}

	// 1. Find and validate the private token
	var privateToken PrivateToken
	// Correctly use squirrel's WHERE clause with time.Now()
	// Squirrel will handle inserting time.Now() as a parameter ($1)
	query, args, err := h.sq.Select("id", "user_id").
		From("private_tokens").
		Where(squirrel.Eq{
			"token": req.Token,
			"type":  dto.TokenTypePasswordReset,
		}).
		Where("expires_at > ?", time.Now()). // <--- This line is still problematic if the '?' is used literally by squirrel
		// CORRECT WAY TO DO THIS WITH SQUIRREL FOR POSTGRES (using `Gt`)
		// Where(squirrel.Gt{"expires_at": time.Now()}).
		// Or, if you prefer the raw string, explicitly use $1:
		// Where("expires_at > $1", time.Now()).
		ToSql()

	// Let's rewrite the above to be safer and more idiomatic Squirrel for PostgreSQL
	// We'll use squirrel.Gt (Greater Than) for the time comparison.
	// This lets squirrel manage the placeholder correctly.
	queryBuilder := h.sq.Select("id", "user_id").
		From("private_tokens").
		Where(squirrel.Eq{
			"token": req.Token,
			"type":  dto.TokenTypePasswordReset,
		}).
		Where(squirrel.Gt{"expires_at": time.Now()}) // <--- This is the correct way for time comparison with Squirrel

	query, args, err = queryBuilder.ToSql() // Rebuild query and args
	if err != nil {
		log.Printf("[RESET_PASSWORD_ERROR] Failed to build token query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process request")
	}
	log.Printf("[RESET_PASSWORD_DEBUG] Executing token query: SQL='%s', Args='%v'", query, args)

	row := h.DB.QueryRowContext(ctx, query, args...)
	err = row.Scan(&privateToken.ID, &privateToken.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[RESET_PASSWORD_INFO] Invalid or expired token received: %s", req.Token)
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid or expired password reset token.")
		}
		log.Printf("[RESET_PASSWORD_ERROR] Database query error for token '%s': %v", req.Token, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process request")
	}
	log.Printf("[RESET_PASSWORD_DEBUG] Valid token found for UserID: %s, TokenID: %d", privateToken.UserID, privateToken.ID)

	// 2. Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[RESET_PASSWORD_ERROR] Failed to hash password: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reset password")
	}
	log.Printf("[RESET_PASSWORD_DEBUG] Password hashed successfully.")

	// 3. Update the user's password
	updateUserQuery, updateUserArgs, err := h.sq.Update("users").
		Set("password", string(hashedPassword)).
		Where(squirrel.Eq{"id": privateToken.UserID}).
		ToSql()
	if err != nil {
		log.Printf("[RESET_PASSWORD_ERROR] Failed to build update user password query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reset password")
	}
	log.Printf("[RESET_PASSWORD_DEBUG] Executing user password update query: SQL='%s', Args='%v'", updateUserQuery, updateUserArgs)

	_, err = h.DB.ExecContext(ctx, updateUserQuery, updateUserArgs...)
	if err != nil {
		log.Printf("[RESET_PASSWORD_ERROR] Failed to update user password for ID '%s': %v", privateToken.UserID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reset password")
	}
	log.Printf("[RESET_PASSWORD_INFO] User password updated successfully for ID: %s", privateToken.UserID)

	// 4. Delete the used private token
	deleteTokenQuery, deleteTokenArgs, err := h.sq.Delete("private_tokens").
		Where(squirrel.Eq{"id": privateToken.ID}).
		ToSql()
	if err != nil {
		log.Printf("[RESET_PASSWORD_ERROR] Failed to build delete token query: %v", err)
		// Don't return an error here, just log it as the password reset was successful
	} else {
		_, err = h.DB.ExecContext(ctx, deleteTokenQuery, deleteTokenArgs...)
		if err != nil {
			log.Printf("[RESET_PASSWORD_ERROR] Failed to delete used password reset token ID %d: %v", privateToken.ID, err)
		} else {
			log.Printf("[RESET_PASSWORD_INFO] Used password reset token ID %d deleted successfully.", privateToken.ID)
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Your password has been reset successfully."})
}
