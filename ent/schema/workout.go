package schema

import (
	"time"

	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Workout holds the schema definition for the Workout entity.
type Workout struct {
	ent.Schema
}

// Fields of the Workout.
func (Workout) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Unique(),

		field.UUID("user_id", uuid.UUID{}),

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

// Edges of the Workout.
func (Workout) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("workouts").
			Field("user_id").
			Required().
			Unique(),

		edge.To("workout_exercises", WorkoutExercise.Type).
			Ref("workouts"),
	}
}

