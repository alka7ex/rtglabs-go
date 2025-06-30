package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"               // This is for standard mixins like `mixin.Time`
	custommixin "rtglabs-go/ent/schema/mixin" // <-- FIX: Import your custom mixins with an alias
)

// Exercise holds the schema definition for the Exercise entity.
type Exercise struct {
	ent.Schema
}

// Mixin of the Exercise.
func (Exercise) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
		custommixin.UUID{},       // <-- FIX: Use the aliased package name
		custommixin.Timestamps{}, // <-- FIX: Use the aliased package name
	}
}

// Fields of the Exercise.
func (Exercise) Fields() []ent.Field {
	return []ent.Field{
		// FIX: The `id` field is now defined by the mixin, so you should remove it from here.
		// field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique(),
		field.String("name").Unique(),
	}
}

// Edges of the Exercise.
func (Exercise) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workout_exercises", WorkoutExercise.Type).
			Inverse(),
		edge.To("exercise_instances", ExerciseInstance.Type).
			Inverse(),
	}
}

