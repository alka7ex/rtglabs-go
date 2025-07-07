package model

import (
	"time"

	"github.com/google/uuid"
)

// PrivateToken represents a row in the 'private_tokens' table.
// It maps directly to database columns using 'db' tags for SQLX.
//
// As per the Ent schema's `Unique()` edge.From, each user can have only one PrivateToken.
type PrivateToken struct {
	ID        uuid.UUID `db:"id" json:"id"`          // Explicitly defined with default uuid.New
	UserID    uuid.UUID `db:"user_id" json:"userId"` // Foreign Key to users.id, UNIQUE and NOT NULL
	Token     string    `db:"token" json:"token"`    // Unique and NotEmpty
	Type      string    `db:"type" json:"type"`      // NotEmpty (CAUTION: 'type' is a SQL keyword, often quoted or renamed to 'token_type' in DDL)
	ExpiresAt time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"` // Default time.Now, Immutable (application logic)
	// Note: 'updated_at' and 'deleted_at' are NOT included here,
	// as they are not defined in this specific PrivateToken schema's Fields()
	// and no Timestamps mixin is applied.
}

// NOTE ON EDGES (Relationships):
// The `UserID` field is the foreign key connecting to the `users` table.
// To fetch the associated `User` for a `PrivateToken`, you would perform a SQL JOIN.
