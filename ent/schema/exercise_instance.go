package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
	customixin "rtglabs-go/ent/schema/mixin"
)

// ExerciseInstance holds the schema definition for the ExerciseInstance entity.
type ExerciseInstance struct {
	ent.Schema
}

// Mixin of the ExerciseInstance.
func (ExerciseInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
		customixin.UUID{},
		customixin.Timestamps{}, // Our custom mixin for created_at, updated_at, and soft deletes
	}
}

// Fields of the ExerciseInstance.
func (ExerciseInstance) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique(), // Maps to $table->uuid('id')->primary()
		// Foreign key field for the nullable relationship with WorkoutLog.
		field.UUID("workout_log_id", uuid.UUID{}).
			Nillable().  // Maps to ->nullable()
			Optional().  // Field is optional in the creation
			Immutable(), // Make the foreign key immutable after creation
		// Foreign key field for the required relationship with Exercise.
		field.UUID("exercise_id", uuid.UUID{}).Immutable(), // Maps to $table->foreignUuid('exercise_id')
	}
}

// Edges of the ExerciseInstance.
func (ExerciseInstance) Edges() []ent.Edge {
	return []ent.Edge{
		// A many-to-one edge to WorkoutLog. It is nullable/optional.
		edge.From("workout_log", WorkoutLog.Type).
			Ref("exercise_instances").
			Field("workout_log_id").
			Unique(),
		// A many-to-one edge to Exercise. It is required.
		edge.From("exercise", Exercise.Type).
			Ref("exercise_instances").
			Field("exercise_id").
			Unique(). // A single instance belongs to one master exercise
			Required(),
	}
}
