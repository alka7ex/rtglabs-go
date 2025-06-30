package schema

import (
	custommixin "rtglabs-go/ent/schema/mixin"

	"entgo.io/ent"
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
// As requested, no relationships are defined for now.
func (Workout) Edges() []ent.Edge {
	return []ent.Edge{
		// No edges for now. We will add the relation to 'user' later.
	}
}

// Indexes of the Workout.
// This corresponds to $table->index('user_id').
func (Workout) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
