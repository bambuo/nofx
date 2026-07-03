package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type TraderOrder struct {
	ent.Schema
}

func (TraderOrder) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("trader_id").NotEmpty().StructTag(`json:"trader_id"`),
		field.String("exchange_id").Default("").StructTag(`json:"exchange_id"`),
		field.String("exchange_type").Default("").StructTag(`json:"exchange_type"`),
		field.String("exchange_order_id").NotEmpty().StructTag(`json:"exchange_order_id"`),
		field.String("client_order_id").Default("").StructTag(`json:"client_order_id"`),
		field.String("symbol").NotEmpty().StructTag(`json:"symbol"`),
		field.String("side").NotEmpty().StructTag(`json:"side"`),
		field.String("position_side").Default("").StructTag(`json:"position_side"`),
		field.String("type_name").NotEmpty().StorageKey("type").StructTag(`json:"type"`),
		field.String("time_in_force").Default("GTC").StructTag(`json:"time_in_force"`),
		field.Float("quantity").Default(0).StructTag(`json:"quantity"`),
		field.Float("price").Default(0).StructTag(`json:"price"`),
		field.Float("stop_price").Default(0).StructTag(`json:"stop_price"`),
		field.String("status").Default("NEW").StructTag(`json:"status"`),
		field.Float("filled_quantity").Default(0).StructTag(`json:"filled_quantity"`),
		field.Float("avg_fill_price").Default(0).StructTag(`json:"avg_fill_price"`),
		field.Float("commission").Default(0).StructTag(`json:"commission"`),
		field.String("commission_asset").Default("USDT").StructTag(`json:"commission_asset"`),
		field.Int("leverage").Default(1).StructTag(`json:"leverage"`),
		field.Bool("reduce_only").Default(false).StructTag(`json:"reduce_only"`),
		field.Bool("close_position").Default(false).StructTag(`json:"close_position"`),
		field.String("working_type").Default("CONTRACT_PRICE").StructTag(`json:"working_type"`),
		field.Bool("price_protect").Default(false).StructTag(`json:"price_protect"`),
		field.String("order_action").Default("").StructTag(`json:"order_action"`),
		field.Int64("related_position_id").Default(0).StructTag(`json:"related_position_id"`),
		field.Int64("created_at").Default(0).StructTag(`json:"created_at"`),
		field.Int64("updated_at").Default(0).StructTag(`json:"updated_at"`),
		field.Int64("filled_at").Default(0).StructTag(`json:"filled_at"`),
	}
}

func (TraderOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("trader_id"),
		index.Fields("symbol"),
		index.Fields("status"),
		index.Fields("exchange_order_id").Unique(),
	}
}

type TraderFill struct {
	ent.Schema
}

func (TraderFill) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("trader_id").NotEmpty().StructTag(`json:"trader_id"`),
		field.String("exchange_id").Default("").StructTag(`json:"exchange_id"`),
		field.String("exchange_type").Default("").StructTag(`json:"exchange_type"`),
		field.Int64("order_id").Default(0).StructTag(`json:"order_id"`),
		field.String("exchange_order_id").NotEmpty().StructTag(`json:"exchange_order_id"`),
		field.String("exchange_trade_id").NotEmpty().StructTag(`json:"exchange_trade_id"`),
		field.String("symbol").NotEmpty().StructTag(`json:"symbol"`),
		field.String("side").NotEmpty().StructTag(`json:"side"`),
		field.Float("price").Default(0).StructTag(`json:"price"`),
		field.Float("quantity").Default(0).StructTag(`json:"quantity"`),
		field.Float("quote_quantity").Default(0).StructTag(`json:"quote_quantity"`),
		field.Float("commission").Default(0).StructTag(`json:"commission"`),
		field.String("commission_asset").Default("").StructTag(`json:"commission_asset"`),
		field.Float("realized_pnl").Default(0).StructTag(`json:"realized_pnl"`),
		field.Bool("is_maker").Default(false).StructTag(`json:"is_maker"`),
		field.Int64("created_at").Default(0).StructTag(`json:"created_at"`),
	}
}

func (TraderFill) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("trader_id"),
		index.Fields("order_id"),
		index.Fields("exchange_trade_id").Unique(),
	}
}
