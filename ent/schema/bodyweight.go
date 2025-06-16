package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	"github.com/google/uuid"
)

// Bodyweight holds the schema definition for the Bodyweight entity.
type Bodyweight struct {
	ent.Schema
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

		field.Time("created_at").
			Default(time.Now),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),

		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

// Edges of the Bodyweight.
// func (Bodyweight) Edges() []ent.Edge {
// 	return []ent.Edge{
// 		edge.From("user", User.Type).
// 			Ref("bodyweights").
// 			Field("user_id").
// 			Unique().
// 			Required(),
// 	}
// }
