package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type BacktestRun struct {
	ent.Schema
}

func (BacktestRun) Fields() []ent.Field {
	return []ent.Field{
		field.String("run_id").Unique().StorageKey("run_id").StructTag(`json:"run_id"`),
		field.String("user_id").Default("").StructTag(`json:"user_id"`),
		field.Bytes("config_json").Optional().StructTag(`json:"config_json,omitempty"`),
		field.String("state").Default("created").StructTag(`json:"state"`),
		field.String("label").Default("").StructTag(`json:"label"`),
		field.Int("symbol_count").Default(0).StructTag(`json:"symbol_count"`),
		field.String("decision_tf").Default("").StructTag(`json:"decision_tf"`),
		field.Int("processed_bars").Default(0).StructTag(`json:"processed_bars"`),
		field.Float("progress_pct").Default(0).StructTag(`json:"progress_pct"`),
		field.Float("equity_last").Default(0).StructTag(`json:"equity_last"`),
		field.Float("max_drawdown_pct").Default(0).StructTag(`json:"max_drawdown_pct"`),
		field.Bool("liquidated").Default(false).StructTag(`json:"liquidated"`),
		field.String("liquidation_note").Default("").StructTag(`json:"liquidation_note"`),
		field.String("prompt_template").Default("").StructTag(`json:"prompt_template"`),
		field.String("custom_prompt").Default("").StructTag(`json:"custom_prompt"`),
		field.Bool("override_prompt").Default(false).StructTag(`json:"override_prompt"`),
		field.String("ai_provider").Default("").StructTag(`json:"ai_provider"`),
		field.String("ai_model").Default("").StructTag(`json:"ai_model"`),
		field.String("last_error").Default("").StructTag(`json:"last_error"`),
		field.Time("created_at").Default(time.Now).Immutable().StructTag(`json:"created_at"`),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).StructTag(`json:"updated_at"`),
	}
}

func (BacktestRun) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}

type BacktestCheckpoint struct {
	ent.Schema
}

func (BacktestCheckpoint) Fields() []ent.Field {
	return []ent.Field{
		field.String("run_id").Unique().StorageKey("run_id").StructTag(`json:"run_id"`),
		field.Bytes("payload").NotEmpty().StructTag(`json:"payload"`),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).StructTag(`json:"updated_at"`),
	}
}

type BacktestEquity struct {
	ent.Schema
}

func (BacktestEquity) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("run_id").NotEmpty().StructTag(`json:"run_id"`),
		field.Int64("ts").StructTag(`json:"ts"`),
		field.Float("equity").StructTag(`json:"equity"`),
		field.Float("available").StructTag(`json:"available"`),
		field.Float("pnl").StructTag(`json:"pnl"`),
		field.Float("pnl_pct").StructTag(`json:"pnl_pct"`),
		field.Float("dd_pct").StructTag(`json:"dd_pct"`),
		field.Int("cycle").StructTag(`json:"cycle"`),
	}
}

func (BacktestEquity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("run_id", "ts"),
	}
}

type BacktestTrade struct {
	ent.Schema
}

func (BacktestTrade) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("run_id").NotEmpty().StructTag(`json:"run_id"`),
		field.Int64("ts").StructTag(`json:"ts"`),
		field.String("symbol").NotEmpty().StructTag(`json:"symbol"`),
		field.String("action").NotEmpty().StructTag(`json:"action"`),
		field.String("side").Default("").StructTag(`json:"side"`),
		field.Float("qty").Default(0).StructTag(`json:"qty"`),
		field.Float("price").Default(0).StructTag(`json:"price"`),
		field.Float("fee").Default(0).StructTag(`json:"fee"`),
		field.Float("slippage").Default(0).StructTag(`json:"slippage"`),
		field.Float("order_value").Default(0).StructTag(`json:"order_value"`),
		field.Float("realized_pnl").Default(0).StructTag(`json:"realized_pnl"`),
		field.Int("leverage").Default(1).StructTag(`json:"leverage"`),
		field.Int("cycle").Default(0).StructTag(`json:"cycle"`),
		field.Float("position_after").Default(0).StructTag(`json:"position_after"`),
		field.Bool("liquidation_flag").Default(false).StructTag(`json:"liquidation_flag"`),
		field.String("note").Default("").StructTag(`json:"note"`),
	}
}

func (BacktestTrade) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("run_id", "ts"),
	}
}
