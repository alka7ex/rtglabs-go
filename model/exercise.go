package model

import (
	"time"

	"github.com/google/uuid"
)

// Bodyweight represents a row in the 'bodyweights' table.
// It maps directly to database columns using 'db' tags for SQLX.
//
// IMPORTANT NOTE: As per your Ent schema, the `Unique()` constraint on the 'user' edge
// implies a 1:1 relationship from User to Bodyweight. This is highly unusual for a
// bodyweight *log*, as a user typically logs many bodyweights over time.
// Your database DDL (below) will reflect this UNIQUE constraint on the foreign key column.
type Exercise struct {
	ID        uuid.UUID  `db:"id" json:"id"` // From custommixin.UUID
	Name      string     `db:"name" json:"name"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`           // From custommixin.Timestamps
	UpdatedAt time.Time  `db:"updated_at" json:"updatedAt"`           // From custommixin.Timestamps
	DeletedAt *time.Time `db:"deleted_at" json:"deletedAt,omitempty"` // For soft deletes, nullable
}

// NOTE ON EDGES (Relationships):
// The `UserID` field is the foreign key connecting to the `users` table.
// To fetch the associated `User` for a `Bodyweight`, you would perform a SQL JOIN.
