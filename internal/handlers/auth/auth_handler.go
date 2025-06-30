package handlers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"rtglabs-go/ent"
	"rtglabs-go/ent/session"

	"github.com/google/uuid"
)

// AuthHandler holds the Ent client.
type AuthHandler struct {
	Client *ent.Client
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(client *ent.Client) *AuthHandler {
	return &AuthHandler{Client: client}
}

// ValidateToken checks if the token is valid and not expired.
// It returns the user ID if valid, or an error otherwise.
// This function is intended to be used by middleware.
func (h *AuthHandler) ValidateToken(token string) (uuid.UUID, error) {
	ctx := context.Background()

	session, err := h.Client.Session.
		Query().
		Where(
			session.TokenEQ(token),
			session.ExpiresAtGTE(time.Now()),
		).
		WithUser(). // <-- Eager-load the user relationship
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return uuid.Nil, errors.New("invalid or expired session token")
		}
		return uuid.Nil, fmt.Errorf("database query failed: %w", err)
	}

	// Defensive check to ensure the edge was loaded.
	if session.Edges.User == nil {
		// This can happen if the session has no associated user (e.g., a DB integrity issue).
		return uuid.Nil, errors.New("user not found for this session")
	}

	// Return the user's ID from the eager-loaded edge.
	return session.Edges.User.ID, nil
}
