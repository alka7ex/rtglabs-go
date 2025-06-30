package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
	custommixin "rtglabs-go/ent/schema/mixin"
)

// WorkoutExercise holds the schema definition for the WorkoutExercise entity.
// This corresponds to your 'workout_exercises' pivot table.
type WorkoutExercise struct {
	ent.Schema
}

// Mixin of the WorkoutExercise.
// This adds common fields like 'id', 'created_at', 'updated_at', and 'deleted_at'.
func (WorkoutExercise) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
		custommixin.UUID{},       // Maps to the 'id' primary key
		custommixin.Timestamps{}, // Maps to 'timestampsTz' and 'softDeletesTz'
	}
}

// Fields of the WorkoutExercise.
// This defines the columns for the 'workout_exercises' table.
func (WorkoutExercise) Fields() []ent.Field {
	return []ent.Field{
		// Maps to foreignUuid('workout_id'). We'll mark it immutable.
		field.UUID("workout_id", uuid.UUID{}).Immutable(),
		// Maps to foreignUuid('exercise_id'). We'll mark it immutable.
		field.UUID("exercise_id", uuid.UUID{}).Immutable(),
		// Maps to foreignUuid('exercise_instance_id')->nullable().
		field.UUID("exercise_instance_id", uuid.UUID{}).Nillable().Optional(),
		// Maps to unsignedInteger('order')->nullable().
		field.Uint("order").Nillable().Optional(),
		// Maps to unsignedInteger('sets')->nullable().
		field.Uint("sets").Nillable().Optional(),
		// Maps to decimal('weight', 8, 2)->nullable().
		field.Float("weight").Nillable().Optional(),
		// Maps to unsignedInteger('reps')->nullable().
		field.Uint("reps").Nillable().Optional(),
	}
}

// Edges of the WorkoutExercise.
// As requested, no relationships are defined for now.
func (WorkoutExercise) Edges() []ent.Edge {
	return []ent.Edge{
		// No edges for now. We will add the relations to 'workout', 'exercise', and 'exercise_instance' later.
	}
}

// Indexes of the WorkoutExercise.
func (WorkoutExercise) Indexes() []ent.Index {
	return []ent.Index{
		// Maps to $table->unique(['workout_id', 'exercise_id']).
		index.Fields("workout_id", "exercise_id").Unique(),
		// Maps to $table->index('workout_id').
		index.Fields("workout_id"),
		// Maps to $table->index('exercise_id').
		index.Fields("exercise_id"),
		// Maps to $table->index('exercise_instance_id').
		index.Fields("exercise_instance_id"),
	}
}
