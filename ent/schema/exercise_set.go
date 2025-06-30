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

// ExerciseSet holds the schema definition for the ExerciseSet entity.
type ExerciseSet struct {
	ent.Schema
}

// Mixin of the ExerciseSet.
func (ExerciseSet) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
		customixin.UUID{},
		customixin.Timestamps{}, // Our custom mixin for created_at, updated_at, and soft deletes
	}
}

// Fields of the ExerciseSet.
func (ExerciseSet) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique(),
		field.UUID("workout_log_id", uuid.UUID{}).Immutable(),                 // Maps to foreignUuid('workout_log_id')
		field.UUID("exercise_id", uuid.UUID{}).Immutable(),                    // Maps to foreignUuid('exercise_id')
		field.UUID("exercise_instance_id", uuid.UUID{}).Nillable().Optional(), // Maps to nullable foreignUuid
		field.Float("weight").Nillable().Optional(),                           // Maps to decimal('weight')
		field.Int("reps").Nillable().Optional(),                               // Maps to integer('reps')
		field.Int("set_number"),                                               // Maps to integer('set_number')
		field.Time("finished_at").Nillable().Optional(),
		field.Int("status").Default(0), // Maps to integer('status')->default(0)
	}
}

// Edges of the ExerciseSet.
func (ExerciseSet) Edges() []ent.Edge {
	return []ent.Edge{
		// An exercise set belongs to a workout log.
		edge.From("workout_log", WorkoutLog.Type).
			Ref("exercise_sets").
			Field("workout_log_id").
			Unique().
			Required().
			Annotations(
				edge.Annotation{
					edge.OnDelete: edge.Cascade,
				},
			),
		// An exercise set belongs to a master exercise.
		edge.From("exercise", Exercise.Type).
			Ref("exercise_sets").
			Field("exercise_id").
			Unique().
			Required().
			Annotations(
				edge.Annotation{
					edge.OnDelete: edge.Cascade,
				},
			),
		// An exercise set can optionally be linked to an exercise instance.
		edge.From("exercise_instance", ExerciseInstance.Type).
			Ref("exercise_sets").
			Field("exercise_instance_id").
			Unique().
			Nillable().
			Optional().
			Annotations(
				edge.Annotation{
					edge.OnDelete: edge.Cascade,
				},
			),
	}
}

// Indexes of the ExerciseSet.
func (ExerciseSet) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workout_log_id"),
		index.Fields("exercise_id"),
		index.Fields("exercise_instance_id"),
		index.Fields("status"),
	}
}
