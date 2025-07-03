package schema

import (
	custommixin "rtglabs-go/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Profile holds the schema definition for the Profile entity.
type Profile struct {
	ent.Schema
}

func (Profile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		custommixin.UUID{},       // Adds 'id' primary key
		custommixin.Timestamps{}, // Adds timestamps & soft deletes
	}
}

func (Profile) Fields() []ent.Field {
	return []ent.Field{
		field.Int("units"),
		field.Int("age"),
		field.Float("height").
			SchemaType(map[string]string{
				"mysql":    "decimal(10, 2)",
				"postgres": "numeric(10, 2)",
				"sqlite":   "numeric",
			}),
		field.Int("gender"),
		field.Float("weight").
			SchemaType(map[string]string{
				"mysql":    "decimal(10, 2)",
				"postgres": "numeric(10, 2)",
				"sqlite":   "numeric",
			}),
	}
}

func (Profile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("profile").
			Unique(),
	}
}

