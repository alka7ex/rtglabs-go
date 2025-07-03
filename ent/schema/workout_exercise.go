package schema

import (
	custommixin "rtglabs-go/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// WorkoutExercise holds the schema definition for the WorkoutExercise entity.
type WorkoutExercise struct {
	ent.Schema
}

func (WorkoutExercise) Mixin() []ent.Mixin {
	return []ent.Mixin{
		custommixin.UUID{},
		custommixin.Timestamps{},
	}
}

func (WorkoutExercise) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("order").Optional().Nillable(),
		field.Uint("sets").Optional().Nillable(),
		field.Float("weight").Optional().Nillable(),
		field.Uint("reps").Optional().Nillable(),
	}
}

func (WorkoutExercise) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workout", Workout.Type).
			Ref("workout_exercises").
			Unique().
			Required().
			Immutable(),

		edge.From("exercise", Exercise.Type).
			Ref("workout_exercises").
			Unique().
			Required().
			Immutable(),

		edge.From("exercise_instance", ExerciseInstance.Type).
			Ref("workout_exercises").
			Unique().
			Immutable(), // ❓ Remove `.Required()` if nullable
	}
}

