package model

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a row in the 'sessions' table.
// It maps directly to database columns using 'db' tags for SQLX.
// The relationship with User is 1:1, meaning each user can have only one such session.
type Session struct {
	ID        uuid.UUID `db:"id" json:"id"`          // Explicitly defined, not from a mixin
	UserID    uuid.UUID `db:"user_id" json:"userId"` // Foreign Key to users.id. UNIQUE and NOT NULL in DB.
	Token     string    `db:"token" json:"token"`
	ExpiresAt time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	// Note: 'updated_at' and 'deleted_at' are NOT included here,
	// as they are not defined in this specific Session schema's Fields()
	// and no Timestamps mixin is applied to Session.
}
