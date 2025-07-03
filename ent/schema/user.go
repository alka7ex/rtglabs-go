package schema

import (
	custommixin "rtglabs-go/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		custommixin.UUID{},       // Provides the 'id' field (uuid.UUID)
		custommixin.Timestamps{}, // Provides created_at, updated_at, deleted_at
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
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
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bodyweights", Bodyweight.Type),
		edge.To("sessions", Session.Type),
		edge.To("profile", Profile.Type).Unique(),
		edge.To("workouts", Workout.Type),
		edge.To("workout_logs", WorkoutLog.Type),
	}
}
