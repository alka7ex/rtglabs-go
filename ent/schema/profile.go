package schema

import (
	custommixin "rtglabs-go/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Profile holds the schema definition for the Profile entity.
type Profile struct {
	ent.Schema
}

func (Profile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		custommixin.UUID{},       // Maps to the 'id' primary key
		custommixin.Timestamps{}, // Maps to 'timestampsTz' and 'softDeletesTz'
	}
}

// Fields of the Profile.
func (Profile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.Int("units"),
		field.Int("age"),
		field.Float("height").
			// Specify the precision and scale for the decimal type in the database.
			// This corresponds to 'decimal:2' in Laravel.
			SchemaType(map[string]string{
				"mysql":    "decimal(10, 2)",
				"postgres": "numeric(10, 2)",
				"sqlite":   "numeric",
			}),
		field.Int("gender"),
		field.Float("weight").
			// Specify the precision and scale for the decimal type in the database.
			// This corresponds to 'decimal:2' in Laravel.
			SchemaType(map[string]string{
				"mysql":    "decimal(10, 2)",
				"postgres": "numeric(10, 2)",
				"sqlite":   "numeric",
			}),
		field.UUID("user_id", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

// Edges of the Profile.
func (Profile) Edges() []ent.Edge {
	return []ent.Edge{
		// Defines a one-to-one or many-to-one relationship with the User entity.
		// The `Unique()` method makes it a one-to-one relationship.
		// The `Field("user_id")` connects the edge to the foreign key field.
		edge.From("user", User.Type).
			Ref("profile").
			Field("user_id").
			Unique(),
	}
}

// Hooks of the Profile.
func (Profile) Hooks() []ent.Hook {
	return nil
}

// Indexes of the Profile.
func (Profile) Indexes() []ent.Index {
	return nil
}
