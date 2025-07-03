package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	custommixin "rtglabs-go/ent/schema/mixin"
)

// ExerciseSet holds the schema definition for the ExerciseSet entity.
type ExerciseSet struct {
	ent.Schema
}

func (ExerciseSet) Mixin() []ent.Mixin {
	return []ent.Mixin{
		custommixin.UUID{},
		custommixin.Timestamps{}, // timestampsTz + softDeletesTz
	}
}

func (ExerciseSet) Fields() []ent.Field {
	return []ent.Field{
		field.Float("weight").
			SchemaType(map[string]string{"postgres": "decimal(8,2)"}).
			Optional().
			Nillable(),
		field.Int("reps").Optional().Nillable(),
		field.Int("set_number"),
		field.Time("finished_at").Optional().Nillable(),
		field.Int("status").Default(0),
	}
}

func (ExerciseSet) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workout_log", WorkoutLog.Type).
			Ref("exercise_sets").
			Unique().
			Required().
			Annotations(
				entsql.OnDelete(entsql.Cascade),
			),

		edge.From("exercise", Exercise.Type).
			Ref("exercise_sets").
			Unique().
			Required().
			Annotations(
				entsql.OnDelete(entsql.Cascade),
			),

		edge.From("exercise_instance", ExerciseInstance.Type).
			Ref("exercise_sets").
			Unique().
			Annotations(
				entsql.OnDelete(entsql.Cascade),
			),
	}
}

func (ExerciseSet) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
	}
}
