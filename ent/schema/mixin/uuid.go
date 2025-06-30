package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// UUID implements an ent.Mixin for UUID primary keys.
// It defines a UUID `id` field with a default value of a new UUID.
type UUID struct {
	mixin.Schema
}

// Fields of the UUID mixin.
func (UUID) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Unique(),
	}
}
