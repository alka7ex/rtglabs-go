package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	custommixin "rtglabs-go/ent/schema/mixin"
)

// WorkoutLog holds the schema definition for the WorkoutLog entity.
type WorkoutLog struct {
	ent.Schema
}

func (WorkoutLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		custommixin.UUID{},
		custommixin.Timestamps{}, // timestampsTz + softDeletesTz
	}
}

func (WorkoutLog) Fields() []ent.Field {
	return []ent.Field{
		field.Time("started_at").Optional().Nillable(),
		field.Time("finished_at").Optional().Nillable(),
		field.Int("status").Default(0),
		field.Uint("total_active_duration_seconds").Default(0),
		field.Uint("total_pause_duration_seconds").Default(0),
	}
}

func (WorkoutLog) Edges() []ent.Edge {
	return []ent.Edge{
		// user_id
		edge.From("user", User.Type).
			Ref("workout_logs").
			Unique().
			Required().
			Annotations(
				entsql.OnDelete(entsql.Cascade),
			),

		// workout_id (nullable)
		edge.From("workout", Workout.Type).
			Ref("workout_logs").
			Unique().
			Annotations(
				entsql.OnDelete(entsql.Cascade),
			),

		// has many exercise_sets
		edge.To("exercise_sets", ExerciseSet.Type).
			Annotations(
				entsql.OnDelete(entsql.Cascade),
			),
		edge.To("exercise_instances", ExerciseInstance.Type),
	}
}

func (WorkoutLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
	}
}

