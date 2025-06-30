package mixin

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Timestamps implements a mixin for soft-deletable entities with timestamps.
type Timestamps struct {
	mixin.Schema
}

// Fields of the Timestamps mixin.
func (Timestamps) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Immutable().
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now), // Automatically updates on save
		field.Time("deleted_at").
			Nillable(). // Nullable to allow for non-deleted state
			Optional(), // Optional in the creation
	}
}
