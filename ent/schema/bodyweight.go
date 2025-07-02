package schema

import (
	custommixin "rtglabs-go/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/google/uuid"
)

// Bodyweight holds the schema definition for the Bodyweight entity.
type Bodyweight struct {
	ent.Schema
}

func (Bodyweight) Mixin() []ent.Mixin {
	return []ent.Mixin{
		custommixin.UUID{},       // Maps to the 'id' primary key
		custommixin.Timestamps{}, // Maps to 'timestampsTz' and 'softDeletesTz'
	}
}

// Fields of the Bodyweight.
func (Bodyweight) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Unique(),
		field.UUID("user_id", uuid.UUID{}),
		field.Float("weight").
			Positive(),
		field.String("unit").
			NotEmpty(),
	}
}

// Edges of the Bodyweight.
func (Bodyweight) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("bodyweights").
			Field("user_id").
			Unique().
			Required(),
	}
}
