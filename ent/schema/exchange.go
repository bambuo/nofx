package schema

import (
	"nofx/crypto"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Exchange struct {
	ent.Schema
}

func (Exchange) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("exchange_type").NotEmpty().Default("").StructTag(`json:"exchange_type"`),
		field.String("account_name").NotEmpty().Default("").StructTag(`json:"account_name"`),
		field.String("user_id").NotEmpty().Default("default").StructTag(`json:"user_id"`),
		field.String("name").NotEmpty().StructTag(`json:"name"`),
		field.String("type").NotEmpty().StructTag(`json:"type"`),
		field.Bool("enabled").Default(false).StructTag(`json:"enabled"`),
		field.String("api_key").
			GoType(crypto.EncryptedString("")).
			Optional().
			Sensitive(),
		field.String("secret_key").
			GoType(crypto.EncryptedString("")).
			Optional().
			Sensitive(),
		field.String("passphrase").
			GoType(crypto.EncryptedString("")).
			Optional().
			Sensitive(),
		field.Bool("testnet").Default(false).StructTag(`json:"testnet"`),
		field.String("hyperliquid_wallet_addr").Optional().Default("").StructTag(`json:"hyperliquidWalletAddr,omitempty"`),
		field.String("aster_user").Optional().Default("").StructTag(`json:"asterUser,omitempty"`),
		field.String("aster_signer").Optional().Default("").StructTag(`json:"asterSigner,omitempty"`),
		field.String("aster_private_key").
			GoType(crypto.EncryptedString("")).
			Optional().
			Sensitive(),
		field.String("lighter_wallet_addr").Optional().Default("").StructTag(`json:"lighterWalletAddr,omitempty"`),
		field.String("lighter_private_key").
			GoType(crypto.EncryptedString("")).
			Optional().
			Sensitive(),
		field.String("lighter_api_key_private_key").
			GoType(crypto.EncryptedString("")).
			Optional().
			Sensitive(),
		field.Int("lighter_api_key_index").Default(0).StructTag(`json:"lighterAPIKeyIndex"`),
		field.Time("created_at").Default(time.Now).Immutable().StructTag(`json:"created_at"`),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).StructTag(`json:"updated_at"`),
	}
}

func (Exchange) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
