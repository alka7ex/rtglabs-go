package schema

import (
	custommixin "rtglabs-go/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// ExerciseInstance holds the schema definition for the ExerciseInstance entity.
// This corresponds to your 'exercise_instances' table.
type ExerciseInstance struct {
	ent.Schema
}

// Mixin of the ExerciseInstance.
// This adds common fields like 'id', 'created_at', 'updated_at', and 'deleted_at'.
func (ExerciseInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
		custommixin.UUID{},       // Maps to the 'id' primary key
		custommixin.Timestamps{}, // Maps to 'timestampsTz' and 'softDeletesTz'
	}
}

// Fields of the ExerciseInstance.
// This defines the columns for the 'exercise_instances' table.
func (ExerciseInstance) Fields() []ent.Field {
	return []ent.Field{
		// Maps to foreignUuid('workout_log_id')->nullable().
		field.UUID("workout_log_id", uuid.UUID{}).Nillable().Optional(),
		// Maps to foreignUuid('exercise_id'). We'll mark it immutable for consistency.
		field.UUID("exercise_id", uuid.UUID{}).Immutable(),
	}
}

// Edges of the ExerciseInstance.
// As requested, no relationships are defined for now.
func (ExerciseInstance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workout_exercises", WorkoutExercise.Type),
	}
}

// Indexes of the ExerciseInstance.
func (ExerciseInstance) Indexes() []ent.Index {
	return []ent.Index{
		// Maps to $table->index('workout_log_id').
		index.Fields("workout_log_id"),
		// Maps to $table->index('exercise_id').
		index.Fields("exercise_id"),
	}
}
