package store

import (
	"context"
	"time"

	"nofx/ent"
	entgridconfig "nofx/ent/gridconfig"
	entgridevent "nofx/ent/gridevent"
	entgridinstance "nofx/ent/gridinstance"
	entgridlevel "nofx/ent/gridlevel"
	entregime "nofx/ent/gridregimeassessment"
)

// ==================== Grid Store Models ====================
// These models mirror the grid package types but are defined here
// to avoid import cycles between store and grid packages.

// GridConfigModel model for grid_configs table
type GridConfigModel struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TraderID  string    `json:"trader_id"`
	Symbol    string    `json:"symbol"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	GridCount       int     `json:"grid_count"`
	TotalInvestment float64 `json:"total_investment"`
	Leverage        int     `json:"leverage"`
	UpperPrice      float64 `json:"upper_price"`
	LowerPrice      float64 `json:"lower_price"`
	UseATRBounds    bool    `json:"use_atr_bounds"`
	ATRMultiplier   float64 `json:"atr_multiplier"`
	Distribution    string  `json:"distribution"`

	MaxDrawdownPct     float64 `json:"max_drawdown_pct"`
	StopLossPct        float64 `json:"stop_loss_pct"`
	DailyLossLimitPct  float64 `json:"daily_loss_limit_pct"`
	MaxPositionSizePct float64 `json:"max_position_size_pct"`

	RegimeCheckInterval  int  `json:"regime_check_interval"`
	AutoPauseOnTrend     bool `json:"auto_pause_on_trend"`
	MinRangingScore      int  `json:"min_ranging_score"`
	TrendResumeThreshold int  `json:"trend_resume_threshold"`

	// Box indicator periods (1h candles)
	ShortBoxPeriod int `json:"short_box_period"`
	MidBoxPeriod   int `json:"mid_box_period"`
	LongBoxPeriod  int `json:"long_box_period"`

	// Effective leverage limits by regime level
	NarrowRegimeLeverage   int `json:"narrow_regime_leverage"`
	StandardRegimeLeverage int `json:"standard_regime_leverage"`
	WideRegimeLeverage     int `json:"wide_regime_leverage"`
	VolatileRegimeLeverage int `json:"volatile_regime_leverage"`

	// Position limits by regime level (percentage of total investment)
	NarrowRegimePositionPct   float64 `json:"narrow_regime_position_pct"`
	StandardRegimePositionPct float64 `json:"standard_regime_position_pct"`
	WideRegimePositionPct     float64 `json:"wide_regime_position_pct"`
	VolatileRegimePositionPct float64 `json:"volatile_regime_position_pct"`

	OrderRefreshSec  int     `json:"order_refresh_sec"`
	UseMakerOnly     bool    `json:"use_maker_only"`
	SlippageTolerPct float64 `json:"slippage_toler_pct"`

	AIProvider string `json:"ai_provider"`
	AIModel    string `json:"ai_model"`
	IsActive   bool   `json:"is_active"`

	// Direction adjustment settings
	EnableDirectionAdjust bool    `json:"enable_direction_adjust"`
	DirectionBiasRatio    float64 `json:"direction_bias_ratio"`
}

// GridInstanceModel model for grid_instances table
type GridInstanceModel struct {
	ID        string     `json:"id"`
	ConfigID  string     `json:"config_id"`
	Symbol    string     `json:"symbol"`
	State     string     `json:"state"`
	StartedAt time.Time  `json:"started_at"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`

	CurrentUpperPrice   float64   `json:"current_upper_price"`
	CurrentLowerPrice   float64   `json:"current_lower_price"`
	CurrentGridSpacing  float64   `json:"current_grid_spacing"`
	ActiveLevelCount    int       `json:"active_level_count"`
	CurrentRegime       string    `json:"current_regime"`
	RegimeScore         int       `json:"regime_score"`
	LastRegimeCheck     time.Time `json:"last_regime_check"`
	ConsecutiveTrending int       `json:"consecutive_trending"`

	// Current regime level (narrow/standard/wide/volatile/trending)
	CurrentRegimeLevel string `json:"current_regime_level"`

	// Box state
	ShortBoxUpper float64 `json:"short_box_upper"`
	ShortBoxLower float64 `json:"short_box_lower"`
	MidBoxUpper   float64 `json:"mid_box_upper"`
	MidBoxLower   float64 `json:"mid_box_lower"`
	LongBoxUpper  float64 `json:"long_box_upper"`
	LongBoxLower  float64 `json:"long_box_lower"`

	// Breakout state
	BreakoutLevel        string    `json:"breakout_level"`
	BreakoutDirection    string    `json:"breakout_direction"`
	BreakoutConfirmCount int       `json:"breakout_confirm_count"`
	BreakoutStartTime    time.Time `json:"breakout_start_time"`

	// Position adjustment due to breakout
	PositionReductionPct float64 `json:"position_reduction_pct"`

	// Grid direction adjustment state
	CurrentDirection       string    `json:"current_direction"`
	DirectionChangedAt     time.Time `json:"direction_changed_at"`
	DirectionChangeCount   int       `json:"direction_change_count"`

	TotalProfit     float64   `json:"total_profit"`
	TotalFees       float64   `json:"total_fees"`
	TotalTrades     int       `json:"total_trades"`
	WinningTrades   int       `json:"winning_trades"`
	MaxDrawdown     float64   `json:"max_drawdown"`
	CurrentDrawdown float64   `json:"current_drawdown"`
	PeakEquity      float64   `json:"peak_equity"`
	DailyProfit     float64   `json:"daily_profit"`
	DailyLoss       float64   `json:"daily_loss"`
	LastDailyReset  time.Time `json:"last_daily_reset"`
}

// GridLevelModel model for grid_levels table
type GridLevelModel struct {
	ID               string     `json:"id"`
	InstanceID       string     `json:"instance_id"`
	LevelIndex       int        `json:"level_index"`
	Price            float64    `json:"price"`
	State            string     `json:"state"`
	Side             string     `json:"side"`
	OrderID          string     `json:"order_id,omitempty"`
	OrderPrice       float64    `json:"order_price,omitempty"`
	OrderQuantity    float64    `json:"order_quantity,omitempty"`
	OrderCreatedAt   *time.Time `json:"order_created_at,omitempty"`
	PositionSize     float64    `json:"position_size,omitempty"`
	PositionEntry    float64    `json:"position_entry,omitempty"`
	PositionOpenAt   *time.Time `json:"position_open_at,omitempty"`
	AllocationWeight float64    `json:"allocation_weight"`
	AllocatedUSD     float64    `json:"allocated_usd"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// GridEventModel model for grid_events table
type GridEventModel struct {
	ID          string    `json:"id"`
	InstanceID  string    `json:"instance_id"`
	LevelID     string    `json:"level_id,omitempty"`
	EventType   string    `json:"event_type"`
	EventTime   time.Time `json:"event_time"`
	Price       float64   `json:"price,omitempty"`
	Quantity    float64   `json:"quantity,omitempty"`
	Side        string    `json:"side,omitempty"`
	PnL         float64   `json:"pnl,omitempty"`
	Fee         float64   `json:"fee,omitempty"`
	Message     string    `json:"message,omitempty"`
	OldRegime   string    `json:"old_regime,omitempty"`
	NewRegime   string    `json:"new_regime,omitempty"`
	TriggerType string    `json:"trigger_type,omitempty"`
	RawData     string    `json:"raw_data,omitempty"`
}

// GridRegimeAssessmentModel model for grid_regime_assessments table
type GridRegimeAssessmentModel struct {
	ID              string    `json:"id"`
	InstanceID      string    `json:"instance_id"`
	AssessedAt      time.Time `json:"assessed_at"`
	Regime          string    `json:"regime"`
	Score           int       `json:"score"`
	Confidence      float64   `json:"confidence"`
	BollingerSignal int       `json:"bollinger_signal"`
	EMASignal       int       `json:"ema_signal"`
	MACDSignal      int       `json:"macd_signal"`
	VolumeSignal    int       `json:"volume_signal"`
	OISignal        int       `json:"oi_signal"`
	FundingSignal   int       `json:"funding_signal"`
	CandleSignal    int       `json:"candle_signal"`
	ATR14           float64   `json:"atr14"`
	BollingerWidth  float64   `json:"bollinger_width"`
	EMADistance     float64   `json:"ema_distance"`
	CurrentPrice    float64   `json:"current_price"`
	AIReasoning     string    `json:"ai_reasoning"`
}

// fromEntGridConfigModel converts ent.GridConfig to store.GridConfigModel
func fromEntGridConfigModel(c *ent.GridConfig) GridConfigModel {
	return GridConfigModel{
		ID:                      c.ID,
		UserID:                  c.UserID,
		TraderID:                c.TraderID,
		Symbol:                  c.Symbol,
		CreatedAt:               c.CreatedAt,
		UpdatedAt:               c.UpdatedAt,
		GridCount:               c.GridCount,
		TotalInvestment:         c.TotalInvestment,
		Leverage:                c.Leverage,
		UpperPrice:              c.UpperPrice,
		LowerPrice:              c.LowerPrice,
		UseATRBounds:            c.UseAtrBounds,
		ATRMultiplier:           c.AtrMultiplier,
		Distribution:            c.Distribution,
		MaxDrawdownPct:          c.MaxDrawdownPct,
		StopLossPct:             c.StopLossPct,
		DailyLossLimitPct:       c.DailyLossLimitPct,
		MaxPositionSizePct:      c.MaxPositionSizePct,
		RegimeCheckInterval:     c.RegimeCheckInterval,
		AutoPauseOnTrend:        c.AutoPauseOnTrend,
		MinRangingScore:         c.MinRangingScore,
		TrendResumeThreshold:    c.TrendResumeThreshold,
		ShortBoxPeriod:          c.ShortBoxPeriod,
		MidBoxPeriod:            c.MidBoxPeriod,
		LongBoxPeriod:           c.LongBoxPeriod,
		NarrowRegimeLeverage:    c.NarrowRegimeLeverage,
		StandardRegimeLeverage:  c.StandardRegimeLeverage,
		WideRegimeLeverage:      c.WideRegimeLeverage,
		VolatileRegimeLeverage:  c.VolatileRegimeLeverage,
		NarrowRegimePositionPct: c.NarrowRegimePositionPct,
		StandardRegimePositionPct: c.StandardRegimePositionPct,
		WideRegimePositionPct:   c.WideRegimePositionPct,
		VolatileRegimePositionPct: c.VolatileRegimePositionPct,
		OrderRefreshSec:         c.OrderRefreshSec,
		UseMakerOnly:            c.UseMakerOnly,
		SlippageTolerPct:        c.SlippageTolerPct,
		AIProvider:              c.AiProvider,
		AIModel:                 c.AiModel,
		IsActive:                c.IsActive,
		EnableDirectionAdjust:   c.EnableDirectionAdjust,
		DirectionBiasRatio:      c.DirectionBiasRatio,
	}
}

// fromEntGridInstanceModel converts ent.GridInstance to store.GridInstanceModel
func fromEntGridInstanceModel(i *ent.GridInstance) GridInstanceModel {
	return GridInstanceModel{
		ID:                   i.ID,
		ConfigID:             i.ConfigID,
		Symbol:               i.Symbol,
		State:                i.State,
		StartedAt:            i.StartedAt,
		StoppedAt:            i.StoppedAt,
		UpdatedAt:            i.UpdatedAt,
		CurrentUpperPrice:    i.CurrentUpperPrice,
		CurrentLowerPrice:    i.CurrentLowerPrice,
		CurrentGridSpacing:   i.CurrentGridSpacing,
		ActiveLevelCount:     i.ActiveLevelCount,
		CurrentRegime:        i.CurrentRegime,
		RegimeScore:          i.RegimeScore,
		LastRegimeCheck:      i.LastRegimeCheck,
		ConsecutiveTrending:  i.ConsecutiveTrending,
		CurrentRegimeLevel:   i.CurrentRegimeLevel,
		ShortBoxUpper:        i.ShortBoxUpper,
		ShortBoxLower:        i.ShortBoxLower,
		MidBoxUpper:          i.MidBoxUpper,
		MidBoxLower:          i.MidBoxLower,
		LongBoxUpper:         i.LongBoxUpper,
		LongBoxLower:         i.LongBoxLower,
		BreakoutLevel:        i.BreakoutLevel,
		BreakoutDirection:    i.BreakoutDirection,
		BreakoutConfirmCount: i.BreakoutConfirmCount,
		BreakoutStartTime:    i.BreakoutStartTime,
		PositionReductionPct: i.PositionReductionPct,
		CurrentDirection:     i.CurrentDirection,
		DirectionChangedAt:   i.DirectionChangedAt,
		DirectionChangeCount: i.DirectionChangeCount,
		TotalProfit:          i.TotalProfit,
		TotalFees:            i.TotalFees,
		TotalTrades:          i.TotalTrades,
		WinningTrades:        i.WinningTrades,
		MaxDrawdown:          i.MaxDrawdown,
		CurrentDrawdown:      i.CurrentDrawdown,
		PeakEquity:           i.PeakEquity,
		DailyProfit:          i.DailyProfit,
		DailyLoss:            i.DailyLoss,
		LastDailyReset:       i.LastDailyReset,
	}
}

// fromEntGridLevelModel converts ent.GridLevel to store.GridLevelModel
func fromEntGridLevelModel(l *ent.GridLevel) GridLevelModel {
	return GridLevelModel{
		ID:               l.ID,
		InstanceID:       l.InstanceID,
		LevelIndex:       l.LevelIndex,
		Price:            l.Price,
		State:            l.State,
		Side:             l.Side,
		OrderID:          l.OrderID,
		OrderPrice:       l.OrderPrice,
		OrderQuantity:    l.OrderQuantity,
		OrderCreatedAt:   l.OrderCreatedAt,
		PositionSize:     l.PositionSize,
		PositionEntry:    l.PositionEntry,
		PositionOpenAt:   l.PositionOpenAt,
		AllocationWeight: l.AllocationWeight,
		AllocatedUSD:     l.AllocatedUsd,
		UpdatedAt:        l.UpdatedAt,
	}
}

// fromEntGridEventModel converts ent.GridEvent to store.GridEventModel
func fromEntGridEventModel(e *ent.GridEvent) GridEventModel {
	return GridEventModel{
		ID:          e.ID,
		InstanceID:  e.InstanceID,
		LevelID:     e.LevelID,
		EventType:   e.EventType,
		EventTime:   e.EventTime,
		Price:       e.Price,
		Quantity:    e.Quantity,
		Side:        e.Side,
		PnL:         e.Pnl,
		Fee:         e.Fee,
		Message:     e.Message,
		OldRegime:   e.OldRegime,
		NewRegime:   e.NewRegime,
		TriggerType: e.TriggerType,
		RawData:     e.RawData,
	}
}

// fromEntGridRegimeAssessmentModel converts ent.GridRegimeAssessment to store.GridRegimeAssessmentModel
func fromEntGridRegimeAssessmentModel(a *ent.GridRegimeAssessment) GridRegimeAssessmentModel {
	return GridRegimeAssessmentModel{
		ID:              a.ID,
		InstanceID:      a.InstanceID,
		AssessedAt:      a.AssessedAt,
		Regime:          a.Regime,
		Score:           a.Score,
		Confidence:      a.Confidence,
		BollingerSignal: a.BollingerSignal,
		EMASignal:       a.EmaSignal,
		MACDSignal:      a.MacdSignal,
		VolumeSignal:    a.VolumeSignal,
		OISignal:        a.OiSignal,
		FundingSignal:   a.FundingSignal,
		CandleSignal:    a.CandleSignal,
		ATR14:           a.Atr14,
		BollingerWidth:  a.BollingerWidth,
		EMADistance:     a.EmaDistance,
		CurrentPrice:    a.CurrentPrice,
		AIReasoning:     a.AiReasoning,
	}
}

// ==================== Grid Store ====================

// GridStore provides database operations for grid trading
type GridStore struct {
	ec *ent.Client
}

// NewGridStore creates a new grid store
func NewGridStore() *GridStore {
	return &GridStore{}
}

// InitTables initializes grid-related tables
func (s *GridStore) InitTables() error {
	return nil
}

// ==================== Config Operations ====================

// SaveGridConfig saves or updates a grid configuration
func (s *GridStore) SaveGridConfig(config *GridConfigModel) error {
	ctx := context.Background()
	config.UpdatedAt = time.Now()
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now()
	}

	exists, err := s.ec.GridConfig.Query().Where(entgridconfig.ID(config.ID)).Exist(ctx)
	if err != nil {
		return err
	}

	if exists {
		_, err = s.ec.GridConfig.UpdateOneID(config.ID).
			SetUserID(config.UserID).
			SetTraderID(config.TraderID).
			SetSymbol(config.Symbol).
			SetGridCount(config.GridCount).
			SetTotalInvestment(config.TotalInvestment).
			SetLeverage(config.Leverage).
			SetUpperPrice(config.UpperPrice).
			SetLowerPrice(config.LowerPrice).
			SetUseAtrBounds(config.UseATRBounds).
			SetAtrMultiplier(config.ATRMultiplier).
			SetDistribution(config.Distribution).
			SetMaxDrawdownPct(config.MaxDrawdownPct).
			SetStopLossPct(config.StopLossPct).
			SetDailyLossLimitPct(config.DailyLossLimitPct).
			SetMaxPositionSizePct(config.MaxPositionSizePct).
			SetRegimeCheckInterval(config.RegimeCheckInterval).
			SetAutoPauseOnTrend(config.AutoPauseOnTrend).
			SetMinRangingScore(config.MinRangingScore).
			SetTrendResumeThreshold(config.TrendResumeThreshold).
			SetShortBoxPeriod(config.ShortBoxPeriod).
			SetMidBoxPeriod(config.MidBoxPeriod).
			SetLongBoxPeriod(config.LongBoxPeriod).
			SetNarrowRegimeLeverage(config.NarrowRegimeLeverage).
			SetStandardRegimeLeverage(config.StandardRegimeLeverage).
			SetWideRegimeLeverage(config.WideRegimeLeverage).
			SetVolatileRegimeLeverage(config.VolatileRegimeLeverage).
			SetNarrowRegimePositionPct(config.NarrowRegimePositionPct).
			SetStandardRegimePositionPct(config.StandardRegimePositionPct).
			SetWideRegimePositionPct(config.WideRegimePositionPct).
			SetVolatileRegimePositionPct(config.VolatileRegimePositionPct).
			SetOrderRefreshSec(config.OrderRefreshSec).
			SetUseMakerOnly(config.UseMakerOnly).
			SetSlippageTolerPct(config.SlippageTolerPct).
			SetAiProvider(config.AIProvider).
			SetAiModel(config.AIModel).
			SetIsActive(config.IsActive).
			SetEnableDirectionAdjust(config.EnableDirectionAdjust).
			SetDirectionBiasRatio(config.DirectionBiasRatio).
			Save(ctx)
		return err
	}

	_, err = s.ec.GridConfig.Create().
		SetID(config.ID).
		SetUserID(config.UserID).
		SetTraderID(config.TraderID).
		SetSymbol(config.Symbol).
		SetGridCount(config.GridCount).
		SetTotalInvestment(config.TotalInvestment).
		SetLeverage(config.Leverage).
		SetUpperPrice(config.UpperPrice).
		SetLowerPrice(config.LowerPrice).
		SetUseAtrBounds(config.UseATRBounds).
		SetAtrMultiplier(config.ATRMultiplier).
		SetDistribution(config.Distribution).
		SetMaxDrawdownPct(config.MaxDrawdownPct).
		SetStopLossPct(config.StopLossPct).
		SetDailyLossLimitPct(config.DailyLossLimitPct).
		SetMaxPositionSizePct(config.MaxPositionSizePct).
		SetRegimeCheckInterval(config.RegimeCheckInterval).
		SetAutoPauseOnTrend(config.AutoPauseOnTrend).
		SetMinRangingScore(config.MinRangingScore).
		SetTrendResumeThreshold(config.TrendResumeThreshold).
		SetShortBoxPeriod(config.ShortBoxPeriod).
		SetMidBoxPeriod(config.MidBoxPeriod).
		SetLongBoxPeriod(config.LongBoxPeriod).
		SetNarrowRegimeLeverage(config.NarrowRegimeLeverage).
		SetStandardRegimeLeverage(config.StandardRegimeLeverage).
		SetWideRegimeLeverage(config.WideRegimeLeverage).
		SetVolatileRegimeLeverage(config.VolatileRegimeLeverage).
		SetNarrowRegimePositionPct(config.NarrowRegimePositionPct).
		SetStandardRegimePositionPct(config.StandardRegimePositionPct).
		SetWideRegimePositionPct(config.WideRegimePositionPct).
		SetVolatileRegimePositionPct(config.VolatileRegimePositionPct).
		SetOrderRefreshSec(config.OrderRefreshSec).
		SetUseMakerOnly(config.UseMakerOnly).
		SetSlippageTolerPct(config.SlippageTolerPct).
		SetAiProvider(config.AIProvider).
		SetAiModel(config.AIModel).
		SetIsActive(config.IsActive).
		SetEnableDirectionAdjust(config.EnableDirectionAdjust).
		SetDirectionBiasRatio(config.DirectionBiasRatio).
		Save(ctx)
	return err
}

// LoadGridConfig loads a grid configuration by ID
func (s *GridStore) LoadGridConfig(id string) (*GridConfigModel, error) {
	ctx := context.Background()
	config, err := s.ec.GridConfig.Query().
		Where(entgridconfig.ID(id)).
		First(ctx)
	if err != nil {
		return nil, err
	}
	result := fromEntGridConfigModel(config)
	return &result, nil
}

// LoadGridConfigByTrader loads a grid configuration by trader ID
func (s *GridStore) LoadGridConfigByTrader(traderID string) (*GridConfigModel, error) {
	ctx := context.Background()
	config, err := s.ec.GridConfig.Query().
		Where(entgridconfig.TraderID(traderID), entgridconfig.IsActive(true)).
		First(ctx)
	if err != nil {
		return nil, err
	}
	result := fromEntGridConfigModel(config)
	return &result, nil
}

// ListGridConfigs lists all grid configurations for a user
func (s *GridStore) ListGridConfigs(userID string) ([]GridConfigModel, error) {
	ctx := context.Background()
	configs, err := s.ec.GridConfig.Query().
		Where(entgridconfig.UserID(userID)).
		Order(ent.Desc(entgridconfig.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]GridConfigModel, len(configs))
	for i, c := range configs {
		result[i] = fromEntGridConfigModel(c)
	}
	return result, nil
}

// DeleteGridConfig deletes a grid configuration and all related data
func (s *GridStore) DeleteGridConfig(id string) error {
	ctx := context.Background()
	tx, err := s.ec.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	instances, err := tx.GridInstance.Query().
		Where(entgridinstance.ConfigID(id)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, instance := range instances {
		if _, err := tx.GridLevel.Delete().
			Where(entgridlevel.InstanceID(instance.ID)).
			Exec(ctx); err != nil {
			tx.Rollback()
			return err
		}
		if _, err := tx.GridEvent.Delete().
			Where(entgridevent.InstanceID(instance.ID)).
			Exec(ctx); err != nil {
			tx.Rollback()
			return err
		}
		if _, err := tx.GridRegimeAssessment.Delete().
			Where(entregime.InstanceID(instance.ID)).
			Exec(ctx); err != nil {
			tx.Rollback()
			return err
		}
	}

	if _, err := tx.GridInstance.Delete().
		Where(entgridinstance.ConfigID(id)).
		Exec(ctx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.GridConfig.DeleteOneID(id).Exec(ctx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// ==================== Instance Operations ====================

// SaveGridInstance saves or updates a grid instance
func (s *GridStore) SaveGridInstance(instance *GridInstanceModel) error {
	ctx := context.Background()
	instance.UpdatedAt = time.Now()

	exists, err := s.ec.GridInstance.Query().Where(entgridinstance.ID(instance.ID)).Exist(ctx)
	if err != nil {
		return err
	}

	if exists {
		_, err = s.ec.GridInstance.UpdateOneID(instance.ID).
			SetConfigID(instance.ConfigID).
			SetSymbol(instance.Symbol).
			SetState(instance.State).
			SetStartedAt(instance.StartedAt).
			SetNillableStoppedAt(instance.StoppedAt).
			SetCurrentUpperPrice(instance.CurrentUpperPrice).
			SetCurrentLowerPrice(instance.CurrentLowerPrice).
			SetCurrentGridSpacing(instance.CurrentGridSpacing).
			SetActiveLevelCount(instance.ActiveLevelCount).
			SetCurrentRegime(instance.CurrentRegime).
			SetRegimeScore(instance.RegimeScore).
			SetLastRegimeCheck(instance.LastRegimeCheck).
			SetConsecutiveTrending(instance.ConsecutiveTrending).
			SetCurrentRegimeLevel(instance.CurrentRegimeLevel).
			SetShortBoxUpper(instance.ShortBoxUpper).
			SetShortBoxLower(instance.ShortBoxLower).
			SetMidBoxUpper(instance.MidBoxUpper).
			SetMidBoxLower(instance.MidBoxLower).
			SetLongBoxUpper(instance.LongBoxUpper).
			SetLongBoxLower(instance.LongBoxLower).
			SetBreakoutLevel(instance.BreakoutLevel).
			SetBreakoutDirection(instance.BreakoutDirection).
			SetBreakoutConfirmCount(instance.BreakoutConfirmCount).
			SetBreakoutStartTime(instance.BreakoutStartTime).
			SetPositionReductionPct(instance.PositionReductionPct).
			SetCurrentDirection(instance.CurrentDirection).
			SetDirectionChangedAt(instance.DirectionChangedAt).
			SetDirectionChangeCount(instance.DirectionChangeCount).
			SetTotalProfit(instance.TotalProfit).
			SetTotalFees(instance.TotalFees).
			SetTotalTrades(instance.TotalTrades).
			SetWinningTrades(instance.WinningTrades).
			SetMaxDrawdown(instance.MaxDrawdown).
			SetCurrentDrawdown(instance.CurrentDrawdown).
			SetPeakEquity(instance.PeakEquity).
			SetDailyProfit(instance.DailyProfit).
			SetDailyLoss(instance.DailyLoss).
			SetLastDailyReset(instance.LastDailyReset).
			Save(ctx)
		return err
	}

	_, err = s.ec.GridInstance.Create().
		SetID(instance.ID).
		SetConfigID(instance.ConfigID).
		SetSymbol(instance.Symbol).
		SetState(instance.State).
		SetStartedAt(instance.StartedAt).
		SetNillableStoppedAt(instance.StoppedAt).
		SetCurrentUpperPrice(instance.CurrentUpperPrice).
		SetCurrentLowerPrice(instance.CurrentLowerPrice).
		SetCurrentGridSpacing(instance.CurrentGridSpacing).
		SetActiveLevelCount(instance.ActiveLevelCount).
		SetCurrentRegime(instance.CurrentRegime).
		SetRegimeScore(instance.RegimeScore).
		SetLastRegimeCheck(instance.LastRegimeCheck).
		SetConsecutiveTrending(instance.ConsecutiveTrending).
		SetCurrentRegimeLevel(instance.CurrentRegimeLevel).
		SetShortBoxUpper(instance.ShortBoxUpper).
		SetShortBoxLower(instance.ShortBoxLower).
		SetMidBoxUpper(instance.MidBoxUpper).
		SetMidBoxLower(instance.MidBoxLower).
		SetLongBoxUpper(instance.LongBoxUpper).
		SetLongBoxLower(instance.LongBoxLower).
		SetBreakoutLevel(instance.BreakoutLevel).
		SetBreakoutDirection(instance.BreakoutDirection).
		SetBreakoutConfirmCount(instance.BreakoutConfirmCount).
		SetBreakoutStartTime(instance.BreakoutStartTime).
		SetPositionReductionPct(instance.PositionReductionPct).
		SetCurrentDirection(instance.CurrentDirection).
		SetDirectionChangedAt(instance.DirectionChangedAt).
		SetDirectionChangeCount(instance.DirectionChangeCount).
		SetTotalProfit(instance.TotalProfit).
		SetTotalFees(instance.TotalFees).
		SetTotalTrades(instance.TotalTrades).
		SetWinningTrades(instance.WinningTrades).
		SetMaxDrawdown(instance.MaxDrawdown).
		SetCurrentDrawdown(instance.CurrentDrawdown).
		SetPeakEquity(instance.PeakEquity).
		SetDailyProfit(instance.DailyProfit).
		SetDailyLoss(instance.DailyLoss).
		SetLastDailyReset(instance.LastDailyReset).
		Save(ctx)
	return err
}

// LoadGridInstance loads a grid instance by config ID
func (s *GridStore) LoadGridInstance(configID string) (*GridInstanceModel, error) {
	ctx := context.Background()
	instance, err := s.ec.GridInstance.Query().
		Where(entgridinstance.ConfigID(configID)).
		Order(ent.Desc(entgridinstance.FieldStartedAt)).
		First(ctx)
	if err != nil {
		return nil, err
	}
	result := fromEntGridInstanceModel(instance)
	return &result, nil
}

// LoadGridInstanceByID loads a grid instance by ID
func (s *GridStore) LoadGridInstanceByID(id string) (*GridInstanceModel, error) {
	ctx := context.Background()
	instance, err := s.ec.GridInstance.Query().
		Where(entgridinstance.ID(id)).
		First(ctx)
	if err != nil {
		return nil, err
	}
	result := fromEntGridInstanceModel(instance)
	return &result, nil
}

// ListGridInstances lists all instances for a config
func (s *GridStore) ListGridInstances(configID string) ([]GridInstanceModel, error) {
	ctx := context.Background()
	instances, err := s.ec.GridInstance.Query().
		Where(entgridinstance.ConfigID(configID)).
		Order(ent.Desc(entgridinstance.FieldStartedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]GridInstanceModel, len(instances))
	for i, inst := range instances {
		result[i] = fromEntGridInstanceModel(inst)
	}
	return result, nil
}

// ==================== Level Operations ====================

// SaveGridLevel saves or updates a grid level
func (s *GridStore) SaveGridLevel(level *GridLevelModel) error {
	ctx := context.Background()
	level.UpdatedAt = time.Now()

	exists, err := s.ec.GridLevel.Query().Where(entgridlevel.ID(level.ID)).Exist(ctx)
	if err != nil {
		return err
	}

	if exists {
		_, err = s.ec.GridLevel.UpdateOneID(level.ID).
			SetInstanceID(level.InstanceID).
			SetLevelIndex(level.LevelIndex).
			SetPrice(level.Price).
			SetState(level.State).
			SetSide(level.Side).
			SetOrderID(level.OrderID).
			SetOrderPrice(level.OrderPrice).
			SetOrderQuantity(level.OrderQuantity).
			SetNillableOrderCreatedAt(level.OrderCreatedAt).
			SetPositionSize(level.PositionSize).
			SetPositionEntry(level.PositionEntry).
			SetNillablePositionOpenAt(level.PositionOpenAt).
			SetAllocationWeight(level.AllocationWeight).
			SetAllocatedUsd(level.AllocatedUSD).
			Save(ctx)
		return err
	}

	_, err = s.ec.GridLevel.Create().
		SetID(level.ID).
		SetInstanceID(level.InstanceID).
		SetLevelIndex(level.LevelIndex).
		SetPrice(level.Price).
		SetState(level.State).
		SetSide(level.Side).
		SetOrderID(level.OrderID).
		SetOrderPrice(level.OrderPrice).
		SetOrderQuantity(level.OrderQuantity).
		SetNillableOrderCreatedAt(level.OrderCreatedAt).
		SetPositionSize(level.PositionSize).
		SetPositionEntry(level.PositionEntry).
		SetNillablePositionOpenAt(level.PositionOpenAt).
		SetAllocationWeight(level.AllocationWeight).
		SetAllocatedUsd(level.AllocatedUSD).
		Save(ctx)
	return err
}

// SaveGridLevels saves multiple grid levels
func (s *GridStore) SaveGridLevels(levels []GridLevelModel) error {
	if len(levels) == 0 {
		return nil
	}
	ctx := context.Background()
	now := time.Now()
	for i := range levels {
		levels[i].UpdatedAt = now
		exists, err := s.ec.GridLevel.Query().Where(entgridlevel.ID(levels[i].ID)).Exist(ctx)
		if err != nil {
			return err
		}
		if exists {
			_, err = s.ec.GridLevel.UpdateOneID(levels[i].ID).
				SetInstanceID(levels[i].InstanceID).
				SetLevelIndex(levels[i].LevelIndex).
				SetPrice(levels[i].Price).
				SetState(levels[i].State).
				SetSide(levels[i].Side).
				SetOrderID(levels[i].OrderID).
				SetOrderPrice(levels[i].OrderPrice).
				SetOrderQuantity(levels[i].OrderQuantity).
				SetNillableOrderCreatedAt(levels[i].OrderCreatedAt).
				SetPositionSize(levels[i].PositionSize).
				SetPositionEntry(levels[i].PositionEntry).
				SetNillablePositionOpenAt(levels[i].PositionOpenAt).
				SetAllocationWeight(levels[i].AllocationWeight).
				SetAllocatedUsd(levels[i].AllocatedUSD).
				Save(ctx)
			if err != nil {
				return err
			}
		} else {
			_, err = s.ec.GridLevel.Create().
				SetID(levels[i].ID).
				SetInstanceID(levels[i].InstanceID).
				SetLevelIndex(levels[i].LevelIndex).
				SetPrice(levels[i].Price).
				SetState(levels[i].State).
				SetSide(levels[i].Side).
				SetOrderID(levels[i].OrderID).
				SetOrderPrice(levels[i].OrderPrice).
				SetOrderQuantity(levels[i].OrderQuantity).
				SetNillableOrderCreatedAt(levels[i].OrderCreatedAt).
				SetPositionSize(levels[i].PositionSize).
				SetPositionEntry(levels[i].PositionEntry).
				SetNillablePositionOpenAt(levels[i].PositionOpenAt).
				SetAllocationWeight(levels[i].AllocationWeight).
				SetAllocatedUsd(levels[i].AllocatedUSD).
				Save(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// LoadGridLevels loads all levels for an instance
func (s *GridStore) LoadGridLevels(instanceID string) ([]GridLevelModel, error) {
	ctx := context.Background()
	levels, err := s.ec.GridLevel.Query().
		Where(entgridlevel.InstanceID(instanceID)).
		Order(ent.Asc(entgridlevel.FieldLevelIndex)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]GridLevelModel, len(levels))
	for i, l := range levels {
		result[i] = fromEntGridLevelModel(l)
	}
	return result, nil
}

// DeleteGridLevels deletes all levels for an instance
func (s *GridStore) DeleteGridLevels(instanceID string) error {
	ctx := context.Background()
	_, err := s.ec.GridLevel.Delete().
		Where(entgridlevel.InstanceID(instanceID)).
		Exec(ctx)
	return err
}

// ==================== Event Operations ====================

// SaveGridEvent saves a grid event
func (s *GridStore) SaveGridEvent(event *GridEventModel) error {
	ctx := context.Background()
	if event.EventTime.IsZero() {
		event.EventTime = time.Now()
	}
	_, err := s.ec.GridEvent.Create().
		SetID(event.ID).
		SetInstanceID(event.InstanceID).
		SetLevelID(event.LevelID).
		SetEventType(event.EventType).
		SetEventTime(event.EventTime).
		SetPrice(event.Price).
		SetQuantity(event.Quantity).
		SetSide(event.Side).
		SetPnl(event.PnL).
		SetFee(event.Fee).
		SetMessage(event.Message).
		SetOldRegime(event.OldRegime).
		SetNewRegime(event.NewRegime).
		SetTriggerType(event.TriggerType).
		SetRawData(event.RawData).
		Save(ctx)
	return err
}

// LoadRecentGridEvents loads recent events for an instance
func (s *GridStore) LoadRecentGridEvents(instanceID string, limit int) ([]GridEventModel, error) {
	ctx := context.Background()
	query := s.ec.GridEvent.Query().
		Where(entgridevent.InstanceID(instanceID)).
		Order(ent.Desc(entgridevent.FieldEventTime))
	if limit > 0 {
		query = query.Limit(limit)
	}
	events, err := query.All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]GridEventModel, len(events))
	for i, e := range events {
		result[i] = fromEntGridEventModel(e)
	}
	return result, nil
}

// LoadGridEventsByType loads events of a specific type
func (s *GridStore) LoadGridEventsByType(instanceID, eventType string, limit int) ([]GridEventModel, error) {
	ctx := context.Background()
	query := s.ec.GridEvent.Query().
		Where(entgridevent.InstanceID(instanceID), entgridevent.EventType(eventType)).
		Order(ent.Desc(entgridevent.FieldEventTime))
	if limit > 0 {
		query = query.Limit(limit)
	}
	events, err := query.All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]GridEventModel, len(events))
	for i, e := range events {
		result[i] = fromEntGridEventModel(e)
	}
	return result, nil
}

// CountGridEvents counts events for an instance
func (s *GridStore) CountGridEvents(instanceID string) (int64, error) {
	ctx := context.Background()
	count, err := s.ec.GridEvent.Query().
		Where(entgridevent.InstanceID(instanceID)).
		Count(ctx)
	return int64(count), err
}

// ==================== Regime Assessment Operations ====================

// SaveGridRegimeAssessment saves a regime assessment
func (s *GridStore) SaveGridRegimeAssessment(assessment *GridRegimeAssessmentModel) error {
	ctx := context.Background()
	if assessment.AssessedAt.IsZero() {
		assessment.AssessedAt = time.Now()
	}
	_, err := s.ec.GridRegimeAssessment.Create().
		SetID(assessment.ID).
		SetInstanceID(assessment.InstanceID).
		SetAssessedAt(assessment.AssessedAt).
		SetRegime(assessment.Regime).
		SetScore(assessment.Score).
		SetConfidence(assessment.Confidence).
		SetBollingerSignal(assessment.BollingerSignal).
		SetEmaSignal(assessment.EMASignal).
		SetMacdSignal(assessment.MACDSignal).
		SetVolumeSignal(assessment.VolumeSignal).
		SetOiSignal(assessment.OISignal).
		SetFundingSignal(assessment.FundingSignal).
		SetCandleSignal(assessment.CandleSignal).
		SetAtr14(assessment.ATR14).
		SetBollingerWidth(assessment.BollingerWidth).
		SetEmaDistance(assessment.EMADistance).
		SetCurrentPrice(assessment.CurrentPrice).
		SetAiReasoning(assessment.AIReasoning).
		Save(ctx)
	return err
}

// LoadLatestGridRegime loads the latest regime assessment
func (s *GridStore) LoadLatestGridRegime(instanceID string) (*GridRegimeAssessmentModel, error) {
	ctx := context.Background()
	assessment, err := s.ec.GridRegimeAssessment.Query().
		Where(entregime.InstanceID(instanceID)).
		Order(ent.Desc(entregime.FieldAssessedAt)).
		First(ctx)
	if err != nil {
		return nil, err
	}
	result := fromEntGridRegimeAssessmentModel(assessment)
	return &result, nil
}

// LoadGridRegimeHistory loads regime assessment history
func (s *GridStore) LoadGridRegimeHistory(instanceID string, limit int) ([]GridRegimeAssessmentModel, error) {
	ctx := context.Background()
	query := s.ec.GridRegimeAssessment.Query().
		Where(entregime.InstanceID(instanceID)).
		Order(ent.Desc(entregime.FieldAssessedAt))
	if limit > 0 {
		query = query.Limit(limit)
	}
	assessments, err := query.All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]GridRegimeAssessmentModel, len(assessments))
	for i, a := range assessments {
		result[i] = fromEntGridRegimeAssessmentModel(a)
	}
	return result, nil
}

// ==================== Statistics Operations ====================

// GetGridInstanceStatistics returns statistics for an instance
func (s *GridStore) GetGridInstanceStatistics(instanceID string) (map[string]interface{}, error) {
	ctx := context.Background()
	instance, err := s.ec.GridInstance.Query().
		Where(entgridinstance.ID(instanceID)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	// Count events by type
	events, err := s.ec.GridEvent.Query().
		Where(entgridevent.InstanceID(instanceID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	eventCountMap := make(map[string]int64)
	for _, e := range events {
		eventCountMap[e.EventType]++
	}

	// Get latest regime
	latestRegime, err := s.ec.GridRegimeAssessment.Query().
		Where(entregime.InstanceID(instanceID)).
		Order(ent.Desc(entregime.FieldAssessedAt)).
		First(ctx)
	latestRegimeScore := 0
	if err == nil {
		latestRegimeScore = latestRegime.Score
	}

	winRate := 0.0
	if instance.TotalTrades > 0 {
		winRate = float64(instance.WinningTrades) / float64(instance.TotalTrades) * 100
	}

	return map[string]interface{}{
		"instance_id":         instance.ID,
		"state":               instance.State,
		"started_at":          instance.StartedAt,
		"stopped_at":          instance.StoppedAt,
		"total_profit":        instance.TotalProfit,
		"total_fees":          instance.TotalFees,
		"total_trades":        instance.TotalTrades,
		"winning_trades":      instance.WinningTrades,
		"win_rate":            winRate,
		"max_drawdown":        instance.MaxDrawdown,
		"current_drawdown":    instance.CurrentDrawdown,
		"peak_equity":         instance.PeakEquity,
		"active_level_count":  instance.ActiveLevelCount,
		"current_regime":      instance.CurrentRegime,
		"regime_score":        instance.RegimeScore,
		"event_counts":        eventCountMap,
		"latest_regime_score": latestRegimeScore,
	}, nil
}

// GetGridPerformanceMetrics returns performance metrics for a time period
func (s *GridStore) GetGridPerformanceMetrics(instanceID string, from, to time.Time) (map[string]interface{}, error) {
	ctx := context.Background()

	// Count trades in period
	tradeCount, _ := s.ec.GridEvent.Query().
		Where(
			entgridevent.InstanceID(instanceID),
			entgridevent.EventType("order_filled"),
			entgridevent.EventTimeGTE(from),
			entgridevent.EventTimeLTE(to),
		).
		Count(ctx)

	buyCount, _ := s.ec.GridEvent.Query().
		Where(
			entgridevent.InstanceID(instanceID),
			entgridevent.EventType("order_filled"),
			entgridevent.Side("buy"),
			entgridevent.EventTimeGTE(from),
			entgridevent.EventTimeLTE(to),
		).
		Count(ctx)

	sellCount, _ := s.ec.GridEvent.Query().
		Where(
			entgridevent.InstanceID(instanceID),
			entgridevent.EventType("order_filled"),
			entgridevent.Side("sell"),
			entgridevent.EventTimeGTE(from),
			entgridevent.EventTimeLTE(to),
		).
		Count(ctx)

	// Sum profit/loss from events in period
	pnlEvents, err := s.ec.GridEvent.Query().
		Where(
			entgridevent.InstanceID(instanceID),
			entgridevent.EventTimeGTE(from),
			entgridevent.EventTimeLTE(to),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	var totalPnL, totalFee float64
	for _, e := range pnlEvents {
		totalPnL += e.Pnl
		totalFee += e.Fee
	}

	// Count regime changes
	regimeChanges, _ := s.ec.GridEvent.Query().
		Where(
			entgridevent.InstanceID(instanceID),
			entgridevent.EventType("regime_change"),
			entgridevent.EventTimeGTE(from),
			entgridevent.EventTimeLTE(to),
		).
		Count(ctx)

	return map[string]interface{}{
		"period_start":   from,
		"period_end":     to,
		"total_fills":    int64(tradeCount),
		"buy_fills":      int64(buyCount),
		"sell_fills":     int64(sellCount),
		"total_pnl":      totalPnL,
		"total_fees":     totalFee,
		"net_pnl":        totalPnL - totalFee,
		"regime_changes": int64(regimeChanges),
	}, nil
}
