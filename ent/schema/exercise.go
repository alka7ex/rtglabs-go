package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Exercise holds the schema definition for the Exercise entity.
type Exercise struct {
	ent.Schema
}

// Fields of the Exercise.
func (Exercise) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Unique(),

		field.String("name").
			NotEmpty(),

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

// Edges of the Exercise.
func (Exercise) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workout_exercises", WorkoutExercise.Type).
			Ref("exercise"),

		edge.From("workout", Workout.Type).
			Ref("workout_exercises").
			Field("workout_id").
			Required().
			Unique(),
	}
}
