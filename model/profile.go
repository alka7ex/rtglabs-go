package model

import (
	"time"

	"github.com/google/uuid"
)

// Profile represents a row in the 'profiles' table.
// It maps directly to database columns using 'db' tags for SQLX.
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
	DeletedAt *time.Time `db:"deleted_at" json:"deletedAt,omitempty"` // From custommixin.Timestamps (for soft deletes), nullable
}

// NOTE ON THE 1:1 RELATIONSHIP (user -> profile):
// In a SQLX/Squirrel setup, you don't embed the `User` struct directly into `Profile`.
// Instead, the `UserID` field explicitly holds the foreign key.
//
// To fetch a User with their Profile, you'd typically perform a JOIN:
//
// SELECT
//     u.id AS user_id, u.name AS user_name, u.email AS user_email, -- etc.
//     p.id AS profile_id, p.units AS profile_units, p.age AS profile_age -- etc.
// FROM
//     users u
// JOIN
//     profiles p ON u.id = p.user_id
// WHERE
//     u.id = 'some-uuid';
//
// You would then scan the results into separate `model.User` and `model.Profile` structs,
// or into a custom struct designed to hold the joined data.
