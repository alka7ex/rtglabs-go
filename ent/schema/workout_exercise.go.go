package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
	custommixin "rtglabs-go/ent/schema/mixin" // <-- FIX: Import your custom mixins with an alias
)

// WorkoutExercise holds the schema definition for the WorkoutExercise entity.
// This is a pivot table mapping workouts to exercises.
type WorkoutExercise struct {
	ent.Schema
}

// Mixin of the WorkoutExercise.
func (WorkoutExercise) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
		custommixin.UUID{},
		custommixin.Timestamps{}, // Our custom timestamps/soft-delete mixin
	}
}

// Fields of the WorkoutExercise.
func (WorkoutExercise) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique(),
		// Foreign key fields for the edges.
		field.UUID("workout_id", uuid.UUID{}).Immutable(),                     // Maps to $table->foreignUuid('workout_id')
		field.UUID("exercise_id", uuid.UUID{}).Immutable(),                    // Maps to $table->foreignUuid('exercise_id')
		field.UUID("exercise_instance_id", uuid.UUID{}).Nillable().Optional(), // Maps to $table->foreignUuid('exercise_instance_id')->nullable()
		field.Uint("order").Nillable().Optional(),                             // Maps to $table->unsignedInteger('order')->nullable()
		field.Uint("sets").Nillable().Optional(),                              // Maps to $table->unsignedInteger('sets')->nullable()
		field.Float("weight").Nillable().Optional(),                           // Maps to $table->decimal('weight', 8, 2)->nullable()
		field.Uint("reps").Nillable().Optional(),                              // Maps to $table->unsignedInteger('reps')->nullable()
	}
}

// Edges of the WorkoutExercise.
func (WorkoutExercise) Edges() []ent.Edge {
	return []ent.Edge{
		// A workout_exercise belongs to one workout (many-to-one).
		edge.From("workout", Workout.Type).
			Ref("workout_exercises").
			Field("workout_id").
			Unique().
			Required(),
		// A workout_exercise belongs to one exercise (many-to-one).
		edge.From("exercise", Exercise.Type).
			Ref("workout_exercises").
			Field("exercise_id").
			Unique().
			Required(),
		// A workout_exercise can have one exercise instance (optional many-to-one).
		edge.From("exercise_instance", ExerciseInstance.Type).
			Ref("workout_exercises").
			Field("exercise_instance_id").
			Unique().
			Nillable().
			Optional(),
	}
}

// Indexes of the WorkoutExercise.
func (WorkoutExercise) Indexes() []ent.Index {
	return []ent.Index{
		// Maps to $table->unique(['workout_id', 'exercise_id']);
		index.Fields("workout_id", "exercise_id").Unique(),
	}
}
