package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid" // Import uuid package
)

// PasswordResetToken holds the schema definition for the PasswordResetToken entity.
type PrivateToken struct {
	ent.Schema
}

// Fields of the PasswordResetToken.
func (PrivateToken) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New), // Correct placement of DefaultFunc
		field.String("token").
			Unique().
			NotEmpty(),
		field.String("type").
			NotEmpty(),
		field.Time("expires_at"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the PasswordResetToken.
func (PrivateToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("private_token").
			Unique().
			Required(),
	}
}

