package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Location holds the schema definition for the Location entity.
type Location struct {
	ent.Schema
}

// Fields of the Location.
func (Location) Fields() []ent.Field {
	return []ent.Field{
		field.String("chrononym").
			NotEmpty(),
		field.String("definition").
			NotEmpty(),
		field.String("context").
			NotEmpty(),
		field.String("district").
			NotEmpty(),
		field.String("selsovet").
			NotEmpty(),
		field.Float("latitude").
			Optional(),
		field.Float("longitude").
			Optional(),
		field.String("comment").
			Optional(),
		field.String("year").
			NotEmpty(),
		field.String("district_ss").
			NotEmpty(),
	}
}

// Edges of the Location.
func (Location) Edges() []ent.Edge {
	return nil
}
