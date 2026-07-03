package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type GridConfig struct {
	ent.Schema
}

func (GridConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("user_id").StructTag(`json:"user_id"`),
		field.String("trader_id").StructTag(`json:"trader_id"`),
		field.String("symbol").NotEmpty().StructTag(`json:"symbol"`),
		field.Int("grid_count").Default(10).StructTag(`json:"grid_count"`),
		field.Float("total_investment").StructTag(`json:"total_investment"`),
		field.Int("leverage").Default(5).StructTag(`json:"leverage"`),
		field.Float("upper_price").Default(0).StructTag(`json:"upper_price"`),
		field.Float("lower_price").Default(0).StructTag(`json:"lower_price"`),
		field.Bool("use_atr_bounds").Default(true).StructTag(`json:"use_atr_bounds"`),
		field.Float("atr_multiplier").Default(2.0).StructTag(`json:"atr_multiplier"`),
		field.String("distribution").Default("gaussian").StructTag(`json:"distribution"`),
		field.Float("max_drawdown_pct").Default(15.0).StructTag(`json:"max_drawdown_pct"`),
		field.Float("stop_loss_pct").Default(5.0).StructTag(`json:"stop_loss_pct"`),
		field.Float("daily_loss_limit_pct").Default(10).StructTag(`json:"daily_loss_limit_pct"`),
		field.Float("max_position_size_pct").Default(30).StructTag(`json:"max_position_size_pct"`),
		field.Int("regime_check_interval").Default(30).StructTag(`json:"regime_check_interval"`),
		field.Bool("auto_pause_on_trend").Default(true).StructTag(`json:"auto_pause_on_trend"`),
		field.Int("min_ranging_score").Default(60).StructTag(`json:"min_ranging_score"`),
		field.Int("trend_resume_threshold").Default(70).StructTag(`json:"trend_resume_threshold"`),
		field.Int("short_box_period").Default(72).StructTag(`json:"short_box_period"`),
		field.Int("mid_box_period").Default(240).StructTag(`json:"mid_box_period"`),
		field.Int("long_box_period").Default(500).StructTag(`json:"long_box_period"`),
		field.Int("narrow_regime_leverage").Default(2).StructTag(`json:"narrow_regime_leverage"`),
		field.Int("standard_regime_leverage").Default(4).StructTag(`json:"standard_regime_leverage"`),
		field.Int("wide_regime_leverage").Default(3).StructTag(`json:"wide_regime_leverage"`),
		field.Int("volatile_regime_leverage").Default(2).StructTag(`json:"volatile_regime_leverage"`),
		field.Float("narrow_regime_position_pct").Default(40).StructTag(`json:"narrow_regime_position_pct"`),
		field.Float("standard_regime_position_pct").Default(70).StructTag(`json:"standard_regime_position_pct"`),
		field.Float("wide_regime_position_pct").Default(60).StructTag(`json:"wide_regime_position_pct"`),
		field.Float("volatile_regime_position_pct").Default(40).StructTag(`json:"volatile_regime_position_pct"`),
		field.Int("order_refresh_sec").Default(300).StructTag(`json:"order_refresh_sec"`),
		field.Bool("use_maker_only").Default(true).StructTag(`json:"use_maker_only"`),
		field.Float("slippage_toler_pct").Default(0.1).StructTag(`json:"slippage_toler_pct"`),
		field.String("ai_provider").Default("deepseek").StructTag(`json:"ai_provider"`),
		field.String("ai_model").Default("deepseek-chat").StructTag(`json:"ai_model"`),
		field.Bool("is_active").Default(false).StructTag(`json:"is_active"`),
		field.Bool("enable_direction_adjust").Default(false).StructTag(`json:"enable_direction_adjust"`),
		field.Float("direction_bias_ratio").Default(0.7).StructTag(`json:"direction_bias_ratio"`),
		field.Time("created_at").Default(time.Now).Immutable().StructTag(`json:"created_at"`),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).StructTag(`json:"updated_at"`),
	}
}

func (GridConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("trader_id"),
	}
}

type GridInstance struct {
	ent.Schema
}

func (GridInstance) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("config_id").NotEmpty().StructTag(`json:"config_id"`),
		field.String("symbol").NotEmpty().StructTag(`json:"symbol"`),
		field.String("state").NotEmpty().StructTag(`json:"state"`),
		field.Time("started_at").Optional().StructTag(`json:"started_at"`),
		field.Time("stopped_at").Optional().Nillable().StructTag(`json:"stopped_at,omitempty"`),
		field.Float("current_upper_price").Default(0).StructTag(`json:"current_upper_price"`),
		field.Float("current_lower_price").Default(0).StructTag(`json:"current_lower_price"`),
		field.Float("current_grid_spacing").Default(0).StructTag(`json:"current_grid_spacing"`),
		field.Int("active_level_count").Default(0).StructTag(`json:"active_level_count"`),
		field.String("current_regime").Default("").StructTag(`json:"current_regime"`),
		field.Int("regime_score").Default(0).StructTag(`json:"regime_score"`),
		field.Time("last_regime_check").Optional().StructTag(`json:"last_regime_check"`),
		field.Int("consecutive_trending").Default(0).StructTag(`json:"consecutive_trending"`),
		field.String("current_regime_level").Default("standard").StructTag(`json:"current_regime_level"`),
		field.Float("short_box_upper").Default(0).StructTag(`json:"short_box_upper"`),
		field.Float("short_box_lower").Default(0).StructTag(`json:"short_box_lower"`),
		field.Float("mid_box_upper").Default(0).StructTag(`json:"mid_box_upper"`),
		field.Float("mid_box_lower").Default(0).StructTag(`json:"mid_box_lower"`),
		field.Float("long_box_upper").Default(0).StructTag(`json:"long_box_upper"`),
		field.Float("long_box_lower").Default(0).StructTag(`json:"long_box_lower"`),
		field.String("breakout_level").Default("none").StructTag(`json:"breakout_level"`),
		field.String("breakout_direction").Default("").StructTag(`json:"breakout_direction"`),
		field.Int("breakout_confirm_count").Default(0).StructTag(`json:"breakout_confirm_count"`),
		field.Time("breakout_start_time").Optional().StructTag(`json:"breakout_start_time"`),
		field.Float("position_reduction_pct").Default(0).StructTag(`json:"position_reduction_pct"`),
		field.String("current_direction").Default("neutral").StructTag(`json:"current_direction"`),
		field.Time("direction_changed_at").Optional().StructTag(`json:"direction_changed_at"`),
		field.Int("direction_change_count").Default(0).StructTag(`json:"direction_change_count"`),
		field.Float("total_profit").Default(0).StructTag(`json:"total_profit"`),
		field.Float("total_fees").Default(0).StructTag(`json:"total_fees"`),
		field.Int("total_trades").Default(0).StructTag(`json:"total_trades"`),
		field.Int("winning_trades").Default(0).StructTag(`json:"winning_trades"`),
		field.Float("max_drawdown").Default(0).StructTag(`json:"max_drawdown"`),
		field.Float("current_drawdown").Default(0).StructTag(`json:"current_drawdown"`),
		field.Float("peak_equity").Default(0).StructTag(`json:"peak_equity"`),
		field.Float("daily_profit").Default(0).StructTag(`json:"daily_profit"`),
		field.Float("daily_loss").Default(0).StructTag(`json:"daily_loss"`),
		field.Time("last_daily_reset").Optional().StructTag(`json:"last_daily_reset"`),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).StructTag(`json:"updated_at"`),
	}
}

func (GridInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("config_id"),
	}
}

type GridLevel struct {
	ent.Schema
}

func (GridLevel) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("instance_id").NotEmpty().StructTag(`json:"instance_id"`),
		field.Int("level_index").StructTag(`json:"level_index"`),
		field.Float("price").StructTag(`json:"price"`),
		field.String("state").NotEmpty().StructTag(`json:"state"`),
		field.String("side").Optional().StructTag(`json:"side"`),
		field.String("order_id").Optional().StructTag(`json:"order_id,omitempty"`),
		field.Float("order_price").Optional().StructTag(`json:"order_price,omitempty"`),
		field.Float("order_quantity").Optional().StructTag(`json:"order_quantity,omitempty"`),
		field.Time("order_created_at").Optional().Nillable().StructTag(`json:"order_created_at,omitempty"`),
		field.Float("position_size").Optional().StructTag(`json:"position_size,omitempty"`),
		field.Float("position_entry").Optional().StructTag(`json:"position_entry,omitempty"`),
		field.Time("position_open_at").Optional().Nillable().StructTag(`json:"position_open_at,omitempty"`),
		field.Float("allocation_weight").Default(0).StructTag(`json:"allocation_weight"`),
		field.Float("allocated_usd").Default(0).StructTag(`json:"allocated_usd"`),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).StructTag(`json:"updated_at"`),
	}
}

func (GridLevel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("instance_id"),
	}
}

type GridEvent struct {
	ent.Schema
}

func (GridEvent) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("instance_id").NotEmpty().StructTag(`json:"instance_id"`),
		field.String("level_id").Optional().StructTag(`json:"level_id,omitempty"`),
		field.String("event_type").NotEmpty().StructTag(`json:"event_type"`),
		field.Time("event_time").Default(time.Now).Immutable().StructTag(`json:"event_time"`),
		field.Float("price").Optional().StructTag(`json:"price,omitempty"`),
		field.Float("quantity").Optional().StructTag(`json:"quantity,omitempty"`),
		field.String("side").Optional().StructTag(`json:"side,omitempty"`),
		field.Float("pnl").Optional().StructTag(`json:"pnl,omitempty"`),
		field.Float("fee").Optional().StructTag(`json:"fee,omitempty"`),
		field.String("message").Optional().StructTag(`json:"message,omitempty"`),
		field.String("old_regime").Optional().StructTag(`json:"old_regime,omitempty"`),
		field.String("new_regime").Optional().StructTag(`json:"new_regime,omitempty"`),
		field.String("trigger_type").Optional().StructTag(`json:"trigger_type,omitempty"`),
		field.Text("raw_data").Optional().StructTag(`json:"raw_data,omitempty"`),
	}
}

func (GridEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("instance_id"),
		index.Fields("level_id"),
	}
}

type GridRegimeAssessment struct {
	ent.Schema
}

func (GridRegimeAssessment) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().StorageKey("id").StructTag(`json:"id"`),
		field.String("instance_id").NotEmpty().StructTag(`json:"instance_id"`),
		field.Time("assessed_at").Default(time.Now).Immutable().StructTag(`json:"assessed_at"`),
		field.String("regime").NotEmpty().StructTag(`json:"regime"`),
		field.Int("score").StructTag(`json:"score"`),
		field.Float("confidence").Default(0).StructTag(`json:"confidence"`),
		field.Int("bollinger_signal").Default(0).StructTag(`json:"bollinger_signal"`),
		field.Int("ema_signal").Default(0).StructTag(`json:"ema_signal"`),
		field.Int("macd_signal").Default(0).StructTag(`json:"macd_signal"`),
		field.Int("volume_signal").Default(0).StructTag(`json:"volume_signal"`),
		field.Int("oi_signal").Default(0).StructTag(`json:"oi_signal"`),
		field.Int("funding_signal").Default(0).StructTag(`json:"funding_signal"`),
		field.Int("candle_signal").Default(0).StructTag(`json:"candle_signal"`),
		field.Float("atr14").Default(0).StructTag(`json:"atr14"`),
		field.Float("bollinger_width").Default(0).StructTag(`json:"bollinger_width"`),
		field.Float("ema_distance").Default(0).StructTag(`json:"ema_distance"`),
		field.Float("current_price").Default(0).StructTag(`json:"current_price"`),
		field.Text("ai_reasoning").Optional().StructTag(`json:"ai_reasoning"`),
	}
}

func (GridRegimeAssessment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("instance_id"),
	}
}
