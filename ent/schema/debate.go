package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type DebateSession struct {
	ent.Schema
}

func (DebateSession) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("user_id").NotEmpty().StructTag(`json:"user_id"`),
		field.String("name").NotEmpty().StructTag(`json:"name"`),
		field.String("strategy_id").NotEmpty().StructTag(`json:"strategy_id"`),
		field.String("status").Default("pending").StructTag(`json:"status"`),
		field.String("symbol").Default("").StructTag(`json:"symbol"`),
		field.Int("max_rounds").Default(3).StructTag(`json:"max_rounds"`),
		field.Int("current_round").Default(0).StructTag(`json:"current_round"`),
		field.Int("interval_minutes").Default(5).StructTag(`json:"interval_minutes"`),
		field.String("prompt_variant").Default("balanced").StructTag(`json:"prompt_variant"`),
		field.Text("final_decision").Optional().StructTag(`json:"final_decision,omitempty"`),
		field.Bool("auto_execute").Default(false).StructTag(`json:"auto_execute"`),
		field.String("trader_id").Optional().Default("").StructTag(`json:"trader_id,omitempty"`),
		field.Bool("enable_oi_ranking").Default(false).StructTag(`json:"enable_oi_ranking"`),
		field.Int("oi_ranking_limit").Default(10).StructTag(`json:"oi_ranking_limit"`),
		field.String("oi_duration").Default("1h").StructTag(`json:"oi_duration"`),
		field.Time("created_at").Default(time.Now).Immutable().StructTag(`json:"created_at"`),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).StructTag(`json:"updated_at"`),
	}
}

func (DebateSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("status"),
	}
}

type DebateParticipant struct {
	ent.Schema
}

func (DebateParticipant) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("session_id").NotEmpty().StructTag(`json:"session_id"`),
		field.String("ai_model_id").NotEmpty().StructTag(`json:"ai_model_id"`),
		field.String("ai_model_name").NotEmpty().StructTag(`json:"ai_model_name"`),
		field.String("provider").NotEmpty().StructTag(`json:"provider"`),
		field.String("personality").NotEmpty().StructTag(`json:"personality"`),
		field.String("color").NotEmpty().StructTag(`json:"color"`),
		field.Int("speak_order").Default(0).StructTag(`json:"speak_order"`),
		field.Time("created_at").Default(time.Now).Immutable().StructTag(`json:"created_at"`),
	}
}

func (DebateParticipant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_id"),
	}
}

type DebateMessage struct {
	ent.Schema
}

func (DebateMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("session_id").NotEmpty().StructTag(`json:"session_id"`),
		field.Int("round").StructTag(`json:"round"`),
		field.String("ai_model_id").NotEmpty().StructTag(`json:"ai_model_id"`),
		field.String("ai_model_name").NotEmpty().StructTag(`json:"ai_model_name"`),
		field.String("provider").NotEmpty().StructTag(`json:"provider"`),
		field.String("personality").NotEmpty().StructTag(`json:"personality"`),
		field.String("message_type").NotEmpty().StructTag(`json:"message_type"`),
		field.Text("content").StructTag(`json:"content"`),
		field.Text("decision_raw").Optional().StructTag(`json:"-"`),
		field.Int("confidence").Default(0).StructTag(`json:"confidence"`),
		field.Time("created_at").Default(time.Now).Immutable().StructTag(`json:"created_at"`),
	}
}

func (DebateMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_id"),
	}
}

type DebateVote struct {
	ent.Schema
}

func (DebateVote) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("session_id").NotEmpty().StructTag(`json:"session_id"`),
		field.String("ai_model_id").NotEmpty().StructTag(`json:"ai_model_id"`),
		field.String("ai_model_name").NotEmpty().StructTag(`json:"ai_model_name"`),
		field.String("action").NotEmpty().StructTag(`json:"action"`),
		field.String("symbol").NotEmpty().StructTag(`json:"symbol"`),
		field.Int("confidence").Default(0).StructTag(`json:"confidence"`),
		field.Int("leverage").Default(5).StructTag(`json:"leverage"`),
		field.Float("position_pct").Default(0.2).StructTag(`json:"position_pct"`),
		field.Float("stop_loss_pct").Default(0.03).StructTag(`json:"stop_loss_pct"`),
		field.Float("take_profit_pct").Default(0.06).StructTag(`json:"take_profit_pct"`),
		field.Text("reasoning").Optional().StructTag(`json:"reasoning"`),
		field.Time("created_at").Default(time.Now).Immutable().StructTag(`json:"created_at"`),
	}
}

func (DebateVote) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_id"),
	}
}
