package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Trader struct {
	ent.Schema
}

func (Trader) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("user_id").NotEmpty().Default("default").StructTag(`json:"user_id"`),
		field.String("name").NotEmpty().StructTag(`json:"name"`),
		field.String("ai_model_id").StructTag(`json:"ai_model_id"`),
		field.String("exchange_id").StructTag(`json:"exchange_id"`),
		field.String("strategy_id").Optional().Default("").StructTag(`json:"strategy_id"`),
		field.Float("initial_balance").Default(0).StructTag(`json:"initial_balance"`),
		field.Int("scan_interval_minutes").Default(3).StructTag(`json:"scan_interval_minutes"`),
		field.Bool("is_running").Default(false).StructTag(`json:"is_running"`),
		field.Bool("is_cross_margin").Default(true).StructTag(`json:"is_cross_margin"`),
		field.Bool("show_in_competition").Default(true).StructTag(`json:"show_in_competition"`),
		field.Int("btc_eth_leverage").Optional().Default(5).StructTag(`json:"btc_eth_leverage,omitempty"`),
		field.Int("altcoin_leverage").Optional().Default(5).StructTag(`json:"altcoin_leverage,omitempty"`),
		field.String("trading_symbols").Optional().Default("").StructTag(`json:"trading_symbols,omitempty"`),
		field.Bool("use_coin_pool").Optional().Default(false).StructTag(`json:"use_ai500,omitempty"`),
		field.Bool("use_oi_top").Optional().Default(false).StructTag(`json:"use_oi_top,omitempty"`),
		field.String("custom_prompt").Optional().Default("").StructTag(`json:"custom_prompt,omitempty"`),
		field.Bool("override_base_prompt").Optional().Default(false).StructTag(`json:"override_base_prompt,omitempty"`),
		field.String("system_prompt_template").Optional().Default("default").StructTag(`json:"system_prompt_template,omitempty"`),
		field.Time("created_at").Default(time.Now).Immutable().StructTag(`json:"created_at"`),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).StructTag(`json:"updated_at"`),
	}
}

func (Trader) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("ai_model", AIModel.Type).Field("ai_model_id").Unique().Required().StructTag(`json:"ai_model"`),
		edge.To("exchange", Exchange.Type).Field("exchange_id").Unique().Required().StructTag(`json:"exchange"`),
	}
}

func (Trader) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
