package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Unique(),

		field.String("name").
			NotEmpty(),

		field.String("email").
			Unique().
			NotEmpty(),

		field.String("password").
			Sensitive().
			NotEmpty(),

		field.Time("email_verified_at").
			Optional().
			Nillable(),
	}
}

// Edges of the User.
// func (User) Edges() []ent.Edge {
// 	return []ent.Edge{
// 		edge.To("workouts", Workout.Type),
// 		edge.To("bodyweights", Bodyweight.Type),
// 	}
// }
//
