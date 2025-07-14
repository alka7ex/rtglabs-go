package model

import (
	"time"

	"github.com/google/uuid"
)

// Profile represents a row in the 'profiles' table.
type Profile struct {
	ID        uuid.UUID  `db:"id" json:"id"`          // From custommixin.UUID
	UserID    uuid.UUID  `db:"user_id" json:"userId"` // Foreign Key to users.id. This must be UNIQUE in the DB.
	Units     int        `db:"units" json:"units"`
	Age       int        `db:"age" json:"age"`
	Height    float64    `db:"height" json:"height"` // float64 is suitable for DECIMAL/NUMERIC types
	Gender    int        `db:"gender" json:"gender"`
	Weight    float64    `db:"weight" json:"weight"`                  // float64 is suitable for DECIMAL/NUMERIC types
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`           // From custommixin.Timestamps
	UpdatedAt time.Time  `db:"updated_at" json:"updatedAt"`           // From custommixin.Timestamps
	DeletedAt *time.Time `db:"deleted_at" json:"deletedAt"` // From custommixin.Timestamps (for soft deletes), nullable
}
