package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Strategy struct {
	ent.Schema
}

func (Strategy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("user_id").NotEmpty().Default("").StructTag(`json:"user_id"`),
		field.String("name").NotEmpty().StructTag(`json:"name"`),
		field.String("description").Optional().Default("").StructTag(`json:"description"`),
		field.Bool("is_active").Default(false).StructTag(`json:"is_active"`),
		field.Bool("is_default").Default(false).StructTag(`json:"is_default"`),
		field.Bool("is_public").Default(false).StructTag(`json:"is_public"`),
		field.Bool("config_visible").Default(true).StructTag(`json:"config_visible"`),
		field.Text("config").Default("{}").StructTag(`json:"config"`),
		field.Time("created_at").Default(time.Now).Immutable().StructTag(`json:"created_at"`),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).StructTag(`json:"updated_at"`),
	}
}

func (Strategy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("is_active"),
		index.Fields("is_public"),
	}
}
