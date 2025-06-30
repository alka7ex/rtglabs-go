package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
	customixin "rtglabs-go/ent/schema/mixin"
)

// Workout holds the schema definition for the Workout entity.
type Workout struct {
	ent.Schema
}

// Mixin of the Workout.
func (Workout) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
		customixin.UUID{},
		customixin.Timestamps{}, // Our custom timestamps/soft-delete mixin
	}
}

// Fields of the Workout.
func (Workout) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique(), // Maps to $table->uuid('id')->primary();
		field.String("name"), // Maps to $table->string('name');
		// The user_id is defined as a field to make it visible and indexed.
		field.UUID("user_id", uuid.UUID{}).Immutable(), // Maps to $table->foreignUuid('user_id')
	}
}

// Edges of the Workout.
func (Workout) Edges() []ent.Edge {
	return []ent.Edge{
		// A workout belongs to a single user (many-to-one relationship).
		edge.From("user", User.Type).
			Ref("workouts").
			Field("user_id"). // The foreign key field
			Unique().         // A workout can only belong to one user
			Required(),       // A workout must have a user
		// A workout has many workout_exercises (one-to-many relationship).
		edge.To("workout_exercises", WorkoutExercise.Type).
			Annotations(
				// This will ensure that when a workout is deleted, its exercises are also deleted (cascade).
				edge.Annotation{
					edge.OnDelete: edge.Cascade,
				},
			),
	}
}
