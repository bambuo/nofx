package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("email").Unique().NotEmpty().StructTag(`json:"email"`),
		field.String("password_hash").Sensitive(),
		field.String("otp_secret").Optional().Sensitive(),
		field.Bool("otp_verified").Default(false).StructTag(`json:"otp_verified"`),
		field.Time("created_at").Default(time.Now).Immutable().StructTag(`json:"created_at"`),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).StructTag(`json:"updated_at"`),
	}
}
