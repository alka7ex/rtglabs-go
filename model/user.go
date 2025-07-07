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
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"email_verified_at,omitempty"` // Nullable field, uses pointer
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`                         // From custommixin.Timestamps
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`                         // From custommixin.Timestamps
	DeletedAt       *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`               // From custommixin.Timestamps (for soft deletes), nullable
}

// NOTE ON EDGES (Relationships):
// The 'Edges' defined in your Ent schema (e.g., to Bodyweight, Session, Profile)
// are NOT represented as fields within this 'User' struct directly.
//
// Instead:
// - Foreign keys exist on the *other* tables (e.g., 'bodyweights' table would have a 'user_id' column).
// - To fetch related data (e.g., all bodyweights for a user), you would write separate SQL queries
//   (often using JOINs or separate selects) and map them into separate structs.
//   For example:
//   func GetUserWithBodyweights(db *sqlx.DB, userID uuid.UUID) (*model.User, []model.Bodyweight, error) {
//       // ... fetch user
//       // ... fetch bodyweights WHERE user_id = userID
//   }

// You would define other model structs similarly:
// type Bodyweight struct {
//     ID        uuid.UUID  `db:"id" json:"id"`
//     UserID    uuid.UUID  `db:"user_id" json:"userId"` // Foreign key
//     WeightKg  float64    `db:"weight_kg" json:"weightKg"`
//     LoggedAt  time.Time  `db:"logged_at" json:"loggedAt"`
//     CreatedAt time.Time  `db:"created_at" json:"createdAt"`
//     UpdatedAt time.Time  `db:"updated_at" json:"updatedAt"`
// }

// And so on for Session, Profile, Workout, WorkoutLog, PrivateToken.
