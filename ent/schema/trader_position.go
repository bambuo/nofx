package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type TraderPosition struct {
	ent.Schema
}

func (TraderPosition) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("trader_id").NotEmpty().StructTag(`json:"trader_id"`),
		field.String("exchange_id").Default("").StructTag(`json:"exchange_id"`),
		field.String("exchange_type").Default("").StructTag(`json:"exchange_type"`),
		field.String("exchange_position_id").Default("").StructTag(`json:"exchange_position_id"`),
		field.String("symbol").NotEmpty().StructTag(`json:"symbol"`),
		field.String("side").NotEmpty().StructTag(`json:"side"`),
		field.Float("entry_quantity").Default(0).StructTag(`json:"entry_quantity"`),
		field.Float("quantity").Default(0).StructTag(`json:"quantity"`),
		field.Float("entry_price").Default(0).StructTag(`json:"entry_price"`),
		field.String("entry_order_id").Default("").StructTag(`json:"entry_order_id"`),
		field.Int64("entry_time").Default(0).StructTag(`json:"entry_time"`),
		field.Float("exit_price").Default(0).StructTag(`json:"exit_price"`),
		field.String("exit_order_id").Default("").StructTag(`json:"exit_order_id"`),
		field.Int64("exit_time").Default(0).StructTag(`json:"exit_time"`),
		field.Float("realized_pnl").Default(0).StructTag(`json:"realized_pnl"`),
		field.Float("fee").Default(0).StructTag(`json:"fee"`),
		field.Int("leverage").Default(1).StructTag(`json:"leverage"`),
		field.String("status").Default("OPEN").StructTag(`json:"status"`),
		field.String("close_reason").Default("").StructTag(`json:"close_reason"`),
		field.String("source").Default("system").StructTag(`json:"source"`),
		field.Int64("created_at").Default(0).StructTag(`json:"created_at"`),
		field.Int64("updated_at").Default(0).StructTag(`json:"updated_at"`),
	}
}

func (TraderPosition) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("trader_id"),
		index.Fields("exchange_id"),
		index.Fields("entry_time"),
		index.Fields("exit_time"),
		index.Fields("status"),
	}
}
