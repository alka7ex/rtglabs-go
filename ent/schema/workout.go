package schema

import (
	custommixin "rtglabs-go/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// Workout holds the schema definition for the Workout entity.
// This corresponds to your 'workouts' table.
type Workout struct {
	ent.Schema
}

// Mixin of the Workout.
// This adds common fields like 'id', 'created_at', 'updated_at', and 'deleted_at'.
func (Workout) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
		custommixin.UUID{},       // Maps to the 'id' primary key
		custommixin.Timestamps{}, // Maps to 'timestampsTz' and 'softDeletesTz'
	}
}

// Fields of the Workout.
// This defines the columns for the 'workouts' table.
func (Workout) Fields() []ent.Field {
	return []ent.Field{
		// Maps to $table->string('name').
		field.String("name"),
		// Maps to $table->foreignUuid('user_id').
		// We mark it as immutable because it's a foreign key that shouldn't change after creation.
		field.UUID("user_id", uuid.UUID{}).Immutable(),
	}
}

// Edges of the Workout.
func (Workout) Edges() []ent.Edge {
	return []ent.Edge{
		// A workout belongs to a single user (many-to-one relationship).
		// This corresponds to $table->foreignUuid('user_id').
		edge.From("user", User.Type).
			Ref("workouts"). // <-- This defines the inverse edge name
			Field("user_id").
			Unique().
			Required().Immutable(), // The foreign key is not nullable, so this is a required relationship.
	}
}

// Indexes of the Workout.
// This corresponds to $table->index('user_id').
func (Workout) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
