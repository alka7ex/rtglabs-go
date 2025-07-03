package schema

import (
	custommixin "rtglabs-go/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Bodyweight holds the schema definition for the Bodyweight entity.
type Bodyweight struct {
	ent.Schema
}

func (Bodyweight) Mixin() []ent.Mixin {
	return []ent.Mixin{
		custommixin.UUID{},       // Provides the 'id' field
		custommixin.Timestamps{}, // Provides created_at, updated_at, deleted_at
	}
}

func (Bodyweight) Fields() []ent.Field {
	return []ent.Field{
		field.Float("weight").
			Positive(),
		field.String("unit").
			NotEmpty(),
	}
}

func (Bodyweight) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("bodyweights").
			Unique().
			Required(),
	}
}

