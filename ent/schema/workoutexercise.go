package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// WorkoutExercise holds the schema definition for the WorkoutExercise entity.
type WorkoutExercise struct {
	ent.Schema
}

// Fields of the WorkoutExercise.
func (WorkoutExercise) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Unique(),

		field.UUID("workout_id", uuid.UUID{}),
		field.UUID("exercise_id", uuid.UUID{}),
		field.UUID("exercise_instance_id", uuid.UUID{}).
			Optional().
			Nillable(),

		field.Int("order"),
		field.Int("sets"),
		field.Int("reps"),
		field.Float("weight"),

		field.Time("created_at").
			Default(time.Now),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),

		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

// Edges of the WorkoutExercise.
// func (WorkoutExercise) Edges() []ent.Edge {
// 	return []ent.Edge{
// 		edge.From("workout", Workout.Type).
// 			Ref("workout_exercises").
// 			Field("workout_id").
// 			Required().
// 			Unique(),
//
// 		edge.From("exercise", Exercise.Type).
// 			Ref("workout_exercises").
// 			Field("exercise_id").
// 			Required().
// 			Unique(),
// 	}
// }
