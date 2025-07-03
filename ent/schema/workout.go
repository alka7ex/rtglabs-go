package schema

import (
	custommixin "rtglabs-go/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Workout holds the schema definition for the Workout entity.
type Workout struct {
	ent.Schema
}

func (Workout) Mixin() []ent.Mixin {
	return []ent.Mixin{
		custommixin.UUID{},
		custommixin.Timestamps{},
	}
}

func (Workout) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
	}
}

func (Workout) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("workouts").
			Unique().
			Required().
			Immutable(),
		edge.To("workout_exercises", WorkoutExercise.Type),
		edge.To("workout_logs", WorkoutLog.Type),
	}
}
