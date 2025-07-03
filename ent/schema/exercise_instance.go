package schema

import (
	custommixin "rtglabs-go/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
)

// ExerciseInstance holds the schema definition for the ExerciseInstance entity.
type ExerciseInstance struct {
	ent.Schema
}

func (ExerciseInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		custommixin.UUID{},
		custommixin.Timestamps{},
	}
}

func (ExerciseInstance) Fields() []ent.Field {
	return []ent.Field{}
}

func (ExerciseInstance) Edges() []ent.Edge {
	return []ent.Edge{
		// edge.From("workout_log", WorkoutLog.Type).
		// 	Ref("exercise_instances").
		// 	Unique().
		// 	Required(), // if not nullable
		//
		edge.From("exercise", Exercise.Type).
			Ref("exercise_instances").
			Unique().
			Required(), // remove this if exercise_id is nullable

		edge.To("workout_exercises", WorkoutExercise.Type),
	}
}

