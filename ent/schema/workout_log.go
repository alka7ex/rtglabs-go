package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
	customixin "rtglabs-go/ent/schema/mixin"
)

// WorkoutLog holds the schema definition for the WorkoutLog entity.
type WorkoutLog struct {
	ent.Schema
}

// Mixin of the WorkoutLog.
func (WorkoutLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
		customixin.UUID{},
		customixin.Timestamps{}, // Our custom mixin for created_at, updated_at, and soft deletes
	}
}

// Fields of the WorkoutLog.
func (WorkoutLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique(),
		field.UUID("workout_id", uuid.UUID{}).Nillable().Optional(), // Maps to nullable foreignUuid('workout_id')
		field.UUID("user_id", uuid.UUID{}).Immutable(),              // Maps to foreignUuid('user_id')
		field.Time("started_at").Nillable().Optional(),
		field.Time("finished_at").Nillable().Optional(),
		field.Int("status").Default(0), // Maps to integer('status')->default(0)
		field.Uint("total_active_duration_seconds").Default(0),
		field.Uint("total_pause_duration_seconds").Default(0),
	}
}

// Edges of the WorkoutLog.
func (WorkoutLog) Edges() []ent.Edge {
	return []ent.Edge{
		// A workout log can optionally be based on a workout template.
		edge.From("workout", Workout.Type).
			Ref("workout_logs").
			Field("workout_id"),
		// A workout log belongs to a user.
		edge.From("user", User.Type).
			Ref("workout_logs").
			Field("user_id").
			Unique().
			Required(),
		// A workout log has many exercise instances and sets.
		edge.To("exercise_instances", ExerciseInstance.Type),
		edge.To("exercise_sets", ExerciseSet.Type),
	}
}

// Indexes of the WorkoutLog.
func (WorkoutLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workout_id"),
		index.Fields("user_id"),
		index.Fields("status"),
	}
}
