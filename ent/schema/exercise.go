package schema

import (
	custommixin "rtglabs-go/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Exercise holds the schema definition for the Exercise entity.
type Exercise struct {
	ent.Schema
}

func (Exercise) Mixin() []ent.Mixin {
	return []ent.Mixin{
		custommixin.UUID{},
		custommixin.Timestamps{},
	}
}

func (Exercise) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique(),
	}
}

func (Exercise) Edges() []ent.Edge {
	return []ent.Edge{
		// These define forward relations.
		// Ent will create reverse foreign keys in the other tables.
		edge.To("exercise_instances", ExerciseInstance.Type),
		edge.To("workout_exercises", WorkoutExercise.Type),
	}
}
