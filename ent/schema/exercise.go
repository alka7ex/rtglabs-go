package schema

import (
	custommixin "rtglabs-go/ent/schema/mixin" // Using your custom mixin

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Exercise holds the schema definition for the Exercise entity.
// This corresponds to your 'exercises' table.
type Exercise struct {
	ent.Schema
}

// Mixin of the Exercise.
// This adds common fields like 'id', 'created_at', 'updated_at', and 'deleted_at'.
func (Exercise) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},             // Maps to 'created_at' and 'updated_at'
		custommixin.UUID{},       // Maps to the 'id' field
		custommixin.Timestamps{}, // Maps to 'created_at', 'updated_at', and 'soft_deletes'
	}
}

// Fields of the Exercise.
// This defines the columns for the 'exercises' table.
func (Exercise) Fields() []ent.Field {
	return []ent.Field{
		// The 'id' field is handled by the mixin.
		// Maps to $table->string('name'). We'll make it unique for master data.
		field.String("name").Unique(),
	}
}

// Edges of the Exercise.
// As requested, no relationships are defined for now.
func (Exercise) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("exercise_instances", ExerciseInstance.Type),
		edge.To("workout_exercises", WorkoutExercise.Type),
	}
}
