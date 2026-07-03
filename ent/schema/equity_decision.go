package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type EquitySnapshot struct {
	ent.Schema
}

func (EquitySnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("trader_id").NotEmpty().StructTag(`json:"trader_id"`),
		field.Time("timestamp").StructTag(`json:"timestamp"`),
		field.Float("total_equity").Default(0).StructTag(`json:"total_equity"`),
		field.Float("balance").Default(0).StructTag(`json:"balance"`),
		field.Float("unrealized_pnl").Default(0).StructTag(`json:"unrealized_pnl"`),
		field.Int("position_count").Default(0).StructTag(`json:"position_count"`),
		field.Float("margin_used_pct").Default(0).StructTag(`json:"margin_used_pct"`),
		field.Time("created_at").Default(time.Now).Immutable().StructTag(`json:"created_at"`),
	}
}

func (EquitySnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("trader_id", "timestamp"),
		index.Fields("timestamp"),
	}
}

type DecisionRecord struct {
	ent.Schema
}

func (DecisionRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("trader_id").NotEmpty().StructTag(`json:"trader_id"`),
		field.Int("cycle_number").Default(0).StructTag(`json:"cycle_number"`),
		field.Time("timestamp").StructTag(`json:"timestamp"`),
		field.Text("system_prompt").Default("").StructTag(`json:"system_prompt"`),
		field.Text("input_prompt").Default("").StructTag(`json:"input_prompt"`),
		field.Text("cot_trace").Default("").StructTag(`json:"cot_trace"`),
		field.Text("decision_json").Default("").StructTag(`json:"decision_json"`),
		field.Text("raw_response").Default("").StructTag(`json:"raw_response"`),
		field.Text("candidate_coins").Default("").StructTag(`json:"candidate_coins"`),
		field.Text("execution_log").Default("").StructTag(`json:"execution_log"`),
		field.Text("decisions").Default("[]").StructTag(`json:"decisions"`),
		field.Bool("success").Default(false).StructTag(`json:"success"`),
		field.Text("error_message").Default("").StructTag(`json:"error_message"`),
		field.Int64("ai_request_duration_ms").Default(0).StructTag(`json:"ai_request_duration_ms"`),
		field.Time("created_at").Default(time.Now).Immutable().StructTag(`json:"created_at"`),
	}
}

func (DecisionRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("trader_id", "timestamp"),
		index.Fields("timestamp"),
	}
}
