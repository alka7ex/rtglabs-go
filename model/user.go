package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a row in the 'users' table.
// It maps directly to database columns using 'db' tags for SQLX.
// 'json' tags are included for potential API serialization.
type User struct {
	ID              uuid.UUID  `db:"id" json:"id"` // From custommixin.UUID
	Name            string     `db:"name" json:"name"`
	Email           string     `db:"email" json:"email"`
	Password        string     `db:"password" json:"-"`                                    // 'json:"-"' omits it from JSON serialization for security
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"email_verified_at"` // Nullable field, uses pointer
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`                         // From custommixin.Timestamps
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
	// From custommixin.Timestamps
}
