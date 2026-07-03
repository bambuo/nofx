package schema

import (
	"nofx/crypto"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type AIModel struct {
	ent.Schema
}

func (AIModel) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("user_id").NotEmpty().Default("default").StructTag(`json:"user_id"`),
		field.String("name").NotEmpty().StructTag(`json:"name"`),
		field.String("provider").NotEmpty().StructTag(`json:"provider"`),
		field.Bool("enabled").Default(false).StructTag(`json:"enabled"`),
		field.String("api_key").
			GoType(crypto.EncryptedString("")).
			Optional().
			Sensitive(),
		field.String("custom_api_url").Optional().Default("").StructTag(`json:"customApiUrl,omitempty"`),
		field.String("custom_model_name").Optional().Default("").StructTag(`json:"customModelName,omitempty"`),
		field.Time("created_at").Default(time.Now).Immutable().StructTag(`json:"created_at"`),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).StructTag(`json:"updated_at"`),
	}
}

func (AIModel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
