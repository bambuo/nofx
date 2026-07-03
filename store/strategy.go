package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"nofx/ent"
	entstrategy "nofx/ent/strategy"
)

// StrategyStore strategy storage
type StrategyStore struct {
	ec *ent.Client
}

// Strategy strategy configuration
type Strategy struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	IsActive      bool      `json:"is_active"`
	IsDefault     bool      `json:"is_default"`
	IsPublic      bool      `json:"is_public"`
	ConfigVisible bool      `json:"config_visible"`
	Config        string    `json:"config"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// fromEntStrategy converts ent.Strategy to store.Strategy
func fromEntStrategy(es *ent.Strategy) *Strategy {
	if es == nil {
		return nil
	}
	return &Strategy{
		ID:            es.ID,
		UserID:        es.UserID,
		Name:          es.Name,
		Description:   es.Description,
		IsActive:      es.IsActive,
		IsDefault:     es.IsDefault,
		IsPublic:      es.IsPublic,
		ConfigVisible: es.ConfigVisible,
		Config:        es.Config,
		CreatedAt:     es.CreatedAt,
		UpdatedAt:     es.UpdatedAt,
	}
}

// StrategyConfig strategy configuration details (JSON structure)
type StrategyConfig struct {
	// Strategy type: "ai_trading" (default) or "grid_trading"
	StrategyType string `json:"strategy_type,omitempty"`

	// language setting: "zh" for Chinese, "en" for English
	// This determines the language used for data formatting and prompt generation
	Language string `json:"language,omitempty"`
	// coin source configuration
	CoinSource CoinSourceConfig `json:"coin_source"`
	// quantitative data configuration
	Indicators IndicatorConfig `json:"indicators"`
	// custom prompt (appended at the end)
	CustomPrompt string `json:"custom_prompt,omitempty"`
	// risk control configuration
	RiskControl RiskControlConfig `json:"risk_control"`
	// editable sections of System Prompt
	PromptSections PromptSectionsConfig `json:"prompt_sections,omitempty"`

	// Grid trading configuration (only used when StrategyType == "grid_trading")
	GridConfig *GridStrategyConfig `json:"grid_config,omitempty"`
}

// GridStrategyConfig grid trading specific configuration
type GridStrategyConfig struct {
	// Trading pair (e.g., "BTCUSDT")
	Symbol string `json:"symbol"`
	// Number of grid levels (5-50)
	GridCount int `json:"grid_count"`
	// Total investment in USDT
	TotalInvestment float64 `json:"total_investment"`
	// Leverage (1-20)
	Leverage int `json:"leverage"`
	// Upper price boundary (0 = auto-calculate from ATR)
	UpperPrice float64 `json:"upper_price"`
	// Lower price boundary (0 = auto-calculate from ATR)
	LowerPrice float64 `json:"lower_price"`
	// Use ATR to auto-calculate bounds
	UseATRBounds bool `json:"use_atr_bounds"`
	// ATR multiplier for bound calculation (default 2.0)
	ATRMultiplier float64 `json:"atr_multiplier"`
	// Position distribution: "uniform" | "gaussian" | "pyramid"
	Distribution string `json:"distribution"`
	// Maximum drawdown percentage before emergency exit
	MaxDrawdownPct float64 `json:"max_drawdown_pct"`
	// Stop loss percentage per position
	StopLossPct float64 `json:"stop_loss_pct"`
	// Daily loss limit percentage
	DailyLossLimitPct float64 `json:"daily_loss_limit_pct"`
	// Use maker-only orders for lower fees
	UseMakerOnly bool `json:"use_maker_only"`
	// Enable automatic grid direction adjustment based on box breakouts
	EnableDirectionAdjust bool `json:"enable_direction_adjust"`
	// Direction bias ratio for long_bias/short_bias modes (default 0.7 = 70%/30%)
	DirectionBiasRatio float64 `json:"direction_bias_ratio"`
}

// PromptSectionsConfig editable sections of System Prompt
type PromptSectionsConfig struct {
	// role definition (title + description)
	RoleDefinition string `json:"role_definition,omitempty"`
	// trading frequency awareness
	TradingFrequency string `json:"trading_frequency,omitempty"`
	// entry standards
	EntryStandards string `json:"entry_standards,omitempty"`
	// decision process
	DecisionProcess string `json:"decision_process,omitempty"`
}

// CoinSourceConfig coin source configuration
type CoinSourceConfig struct {
	// source type: "static" | "ai500" | "oi_top" | "oi_low" | "mixed"
	SourceType string `json:"source_type"`
	// static coin list (used when source_type = "static")
	StaticCoins []string `json:"static_coins,omitempty"`
	// excluded coins list (filtered out from all sources)
	ExcludedCoins []string `json:"excluded_coins,omitempty"`
	// whether to use AI500 coin pool
	UseAI500 bool `json:"use_ai500"`
	// AI500 coin pool maximum count
	AI500Limit int `json:"ai500_limit,omitempty"`
	// whether to use OI Top (持仓增加榜，适合做多)
	UseOITop bool `json:"use_oi_top"`
	// OI Top maximum count
	OITopLimit int `json:"oi_top_limit,omitempty"`
	// whether to use OI Low (持仓减少榜，适合做空)
	UseOILow bool `json:"use_oi_low"`
	// OILow maximum count
	OILowLimit int `json:"oi_low_limit,omitempty"`
}

// IndicatorConfig indicator configuration
type IndicatorConfig struct {
	// K-line configuration
	Klines KlineConfig `json:"klines"`
	// raw kline data (OHLCV) - always enabled, required for AI analysis
	EnableRawKlines bool `json:"enable_raw_klines"`
	// technical indicator switches
	EnableEMA         bool `json:"enable_ema"`
	EnableMACD        bool `json:"enable_macd"`
	EnableRSI         bool `json:"enable_rsi"`
	EnableATR         bool `json:"enable_atr"`
	EnableBOLL        bool `json:"enable_boll"`         // Bollinger Bands
	EnableVolume      bool `json:"enable_volume"`
	EnableOI          bool `json:"enable_oi"`           // open interest
	EnableFundingRate bool `json:"enable_funding_rate"` // funding rate
	// EMA period configuration
	EMAPeriods []int `json:"ema_periods,omitempty"` // default [20, 50]
	// RSI period configuration
	RSIPeriods []int `json:"rsi_periods,omitempty"` // default [7, 14]
	// ATR period configuration
	ATRPeriods []int `json:"atr_periods,omitempty"` // default [14]
	// BOLL period configuration (period, standard deviation multiplier is fixed at 2)
	BOLLPeriods []int `json:"boll_periods,omitempty"` // default [20] - can select multiple timeframes
	// external data sources
	ExternalDataSources []ExternalDataSource `json:"external_data_sources,omitempty"`

	// ========== NofxOS Unified API Configuration ==========
	// Unified API Key for all NofxOS data sources
	NofxOSAPIKey string `json:"nofxos_api_key,omitempty"`

	// quantitative data sources (capital flow, position changes, price changes)
	EnableQuantData    bool `json:"enable_quant_data"`    // whether to enable quantitative data
	EnableQuantOI      bool `json:"enable_quant_oi"`      // whether to show OI data
	EnableQuantNetflow bool `json:"enable_quant_netflow"` // whether to show Netflow data

	// OI ranking data (market-wide open interest increase/decrease rankings)
	EnableOIRanking   bool   `json:"enable_oi_ranking"`             // whether to enable OI ranking data
	OIRankingDuration string `json:"oi_ranking_duration,omitempty"` // duration: 1h, 4h, 24h
	OIRankingLimit    int    `json:"oi_ranking_limit,omitempty"`    // number of entries (default 10)

	// NetFlow ranking data (market-wide fund flow rankings - institution/personal)
	EnableNetFlowRanking   bool   `json:"enable_netflow_ranking"`             // whether to enable NetFlow ranking data
	NetFlowRankingDuration string `json:"netflow_ranking_duration,omitempty"` // duration: 1h, 4h, 24h
	NetFlowRankingLimit    int    `json:"netflow_ranking_limit,omitempty"`    // number of entries (default 10)

	// Price ranking data (market-wide gainers/losers)
	EnablePriceRanking   bool   `json:"enable_price_ranking"`             // whether to enable price ranking data
	PriceRankingDuration string `json:"price_ranking_duration,omitempty"` // durations: "1h" or "1h,4h,24h"
	PriceRankingLimit    int    `json:"price_ranking_limit,omitempty"`    // number of entries per ranking (default 10)
}

// KlineConfig K-line configuration
type KlineConfig struct {
	// primary timeframe: "1m", "3m", "5m", "15m", "1h", "4h"
	PrimaryTimeframe string `json:"primary_timeframe"`
	// primary timeframe K-line count
	PrimaryCount int `json:"primary_count"`
	// longer timeframe
	LongerTimeframe string `json:"longer_timeframe,omitempty"`
	// longer timeframe K-line count
	LongerCount int `json:"longer_count,omitempty"`
	// whether to enable multi-timeframe analysis
	EnableMultiTimeframe bool `json:"enable_multi_timeframe"`
	// selected timeframe list (new: supports multi-timeframe selection)
	SelectedTimeframes []string `json:"selected_timeframes,omitempty"`
}

// ExternalDataSource external data source configuration
type ExternalDataSource struct {
	Name        string            `json:"name"`         // data source name
	Type        string            `json:"type"`         // type: "api" | "webhook"
	URL         string            `json:"url"`          // API URL
	Method      string            `json:"method"`       // HTTP method
	Headers     map[string]string `json:"headers,omitempty"`
	DataPath    string            `json:"data_path,omitempty"`    // JSON data path
	RefreshSecs int               `json:"refresh_secs,omitempty"` // refresh interval (seconds)
}

// RiskControlConfig risk control configuration
type RiskControlConfig struct {
	// Max number of coins held simultaneously (CODE ENFORCED)
	MaxPositions int `json:"max_positions"`

	// BTC/ETH exchange leverage for opening positions (AI guided)
	BTCETHMaxLeverage int `json:"btc_eth_max_leverage"`
	// Altcoin exchange leverage for opening positions (AI guided)
	AltcoinMaxLeverage int `json:"altcoin_max_leverage"`

	// BTC/ETH single position max value = equity x this ratio (CODE ENFORCED, default: 5)
	BTCETHMaxPositionValueRatio float64 `json:"btc_eth_max_position_value_ratio"`
	// Altcoin single position max value = equity x this ratio (CODE ENFORCED, default: 1)
	AltcoinMaxPositionValueRatio float64 `json:"altcoin_max_position_value_ratio"`

	// Max margin utilization (e.g. 0.9 = 90%) (CODE ENFORCED)
	MaxMarginUsage float64 `json:"max_margin_usage"`
	// Min position size in USDT (CODE ENFORCED)
	MinPositionSize float64 `json:"min_position_size"`

	// Min take_profit / stop_loss ratio (AI guided)
	MinRiskRewardRatio float64 `json:"min_risk_reward_ratio"`
	// Min AI confidence to open position (AI guided)
	MinConfidence int `json:"min_confidence"`
}

// NewStrategyStore creates a new StrategyStore
func NewStrategyStore() *StrategyStore {
	return &StrategyStore{}
}

func (s *StrategyStore) initTables() error {
	return nil
}

func (s *StrategyStore) initDefaultData() error {
	return nil
}

// GetDefaultStrategyConfig returns the default strategy configuration for the given language
func GetDefaultStrategyConfig(lang string) StrategyConfig {
	// Normalize language to "zh" or "en"
	normalizedLang := "zh"
	if lang == "en" {
		normalizedLang = "en"
	}

	config := StrategyConfig{
		Language: normalizedLang,
		CoinSource: CoinSourceConfig{
			SourceType: "ai500",
			UseAI500:   true,
			AI500Limit: 10,
			UseOITop:   false,
			OITopLimit: 10,
			UseOILow:   false,
			OILowLimit: 10,
		},
		Indicators: IndicatorConfig{
			Klines: KlineConfig{
				PrimaryTimeframe:     "5m",
				PrimaryCount:         30,
				LongerTimeframe:      "4h",
				LongerCount:          10,
				EnableMultiTimeframe: true,
				SelectedTimeframes:   []string{"5m", "15m", "1h", "4h"},
			},
			EnableRawKlines:   true,
			EnableEMA:         false,
			EnableMACD:        false,
			EnableRSI:         false,
			EnableATR:         false,
			EnableBOLL:        false,
			EnableVolume:      true,
			EnableOI:          true,
			EnableFundingRate: true,
			EMAPeriods:        []int{20, 50},
			RSIPeriods:        []int{7, 14},
			ATRPeriods:        []int{14},
			BOLLPeriods:       []int{20},
			// NofxOS unified API key
			NofxOSAPIKey: "cm_568c67eae410d912c54c",
			// Quant data
			EnableQuantData:    true,
			EnableQuantOI:      true,
			EnableQuantNetflow: true,
			// OI ranking data
			EnableOIRanking:   true,
			OIRankingDuration: "1h",
			OIRankingLimit:    10,
			// NetFlow ranking data
			EnableNetFlowRanking:   true,
			NetFlowRankingDuration: "1h",
			NetFlowRankingLimit:    10,
			// Price ranking data
			EnablePriceRanking:   true,
			PriceRankingDuration: "1h,4h,24h",
			PriceRankingLimit:    10,
		},
		RiskControl: RiskControlConfig{
			MaxPositions:                    3,
			BTCETHMaxLeverage:               5,
			AltcoinMaxLeverage:              5,
			BTCETHMaxPositionValueRatio:     5.0,
			AltcoinMaxPositionValueRatio:    1.0,
			MaxMarginUsage:                  0.9,
			MinPositionSize:                 12,
			MinRiskRewardRatio:              3.0,
			MinConfidence:                   75,
		},
	}

	if lang == "zh" {
		config.PromptSections = PromptSectionsConfig{
			RoleDefinition: `# 你是一个专业的加密货币交易AI

你的任务是根据提供的市场数据做出交易决策。你是一个经验丰富的量化交易员，擅长技术分析和风险管理。`,
			TradingFrequency: `# ⏱️ 交易频率意识

- 优秀交易员：每天2-4笔 ≈ 每小时0.1-0.2笔
- 每小时超过2笔 = 过度交易
- 单笔持仓时间 ≥ 30-60分钟
如果你发现自己每个周期都在交易 → 标准太低；如果持仓不到30分钟就平仓 → 太冲动。`,
			EntryStandards: `# 🎯 入场标准（严格）

只在多个信号共振时入场。自由使用任何有效的分析方法，避免单一指标、信号矛盾、横盘震荡、或平仓后立即重新开仓等低质量行为。`,
			DecisionProcess: `# 📋 决策流程

1. 检查持仓 → 是否止盈/止损
2. 扫描候选币种 + 多时间框架 → 是否存在强信号
3. 先写思维链，再输出结构化JSON`,
		}
	} else {
		config.PromptSections = PromptSectionsConfig{
			RoleDefinition: `# You are a professional cryptocurrency trading AI

Your task is to make trading decisions based on the provided market data. You are an experienced quantitative trader skilled in technical analysis and risk management.`,
			TradingFrequency: `# ⏱️ Trading Frequency Awareness

- Excellent trader: 2-4 trades per day ≈ 0.1-0.2 trades per hour
- >2 trades per hour = overtrading
- Single position holding time ≥ 30-60 minutes
If you find yourself trading every cycle → standards are too low; if closing positions in <30 minutes → too impulsive.`,
			EntryStandards: `# 🎯 Entry Standards (Strict)

Only enter positions when multiple signals resonate. Freely use any effective analysis methods, avoid low-quality behaviors such as single indicators, contradictory signals, sideways oscillation, or immediately restarting after closing positions.`,
			DecisionProcess: `# 📋 Decision Process

1. Check positions → whether to take profit/stop loss
2. Scan candidate coins + multi-timeframe → whether strong signals exist
3. Write chain of thought first, then output structured JSON`,
		}
	}

	return config
}

// Create creates a strategy using ent
func (s *StrategyStore) Create(ctx context.Context, strategy *Strategy) error {
	if s.ec == nil {
		return fmt.Errorf("ent client not available")
	}
	_, err := s.ec.Strategy.Create().
		SetID(strategy.ID).
		SetUserID(strategy.UserID).
		SetName(strategy.Name).
		SetNillableDescription(strPtr(strategy.Description)).
		SetIsActive(strategy.IsActive).
		SetIsDefault(strategy.IsDefault).
		SetIsPublic(strategy.IsPublic).
		SetConfigVisible(strategy.ConfigVisible).
		SetConfig(strategy.Config).
		Save(ctx)
	return err
}

// Update updates a strategy using ent
func (s *StrategyStore) Update(ctx context.Context, strategy *Strategy) error {
	if s.ec == nil {
		return fmt.Errorf("ent client not available")
	}
	return s.ec.Strategy.Update().
		Where(entstrategy.And(entstrategy.ID(strategy.ID), entstrategy.UserID(strategy.UserID))).
		SetName(strategy.Name).
		SetNillableDescription(strPtr(strategy.Description)).
		SetConfig(strategy.Config).
		SetIsPublic(strategy.IsPublic).
		SetConfigVisible(strategy.ConfigVisible).
		Exec(ctx)
}

// Delete deletes a strategy using ent
func (s *StrategyStore) Delete(ctx context.Context, userID, id string) error {
	if s.ec == nil {
		return fmt.Errorf("ent client not available")
	}
	st, err := s.ec.Strategy.Query().Where(entstrategy.ID(id)).First(ctx)
	if err == nil && st.IsDefault {
		return fmt.Errorf("cannot delete system default strategy")
	}
	n, err := s.ec.Strategy.Delete().Where(entstrategy.And(entstrategy.ID(id), entstrategy.UserID(userID))).Exec(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("strategy not found")
	}
	return nil
}

// List gets user's strategy list using ent
func (s *StrategyStore) List(ctx context.Context, userID string) ([]*Strategy, error) {
	if s.ec == nil {
		return nil, fmt.Errorf("ent client not available")
	}
	strategies, err := s.ec.Strategy.Query().
		Where(entstrategy.Or(entstrategy.UserID(userID), entstrategy.IsDefault(true))).
		Order(ent.Desc(entstrategy.FieldIsDefault), ent.Desc(entstrategy.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*Strategy, len(strategies))
	for i, s := range strategies {
		result[i] = fromEntStrategy(s)
	}
	return result, nil
}

// ListPublic gets all public strategies using ent
func (s *StrategyStore) ListPublic(ctx context.Context) ([]*Strategy, error) {
	if s.ec == nil {
		return nil, fmt.Errorf("ent client not available")
	}
	strategies, err := s.ec.Strategy.Query().
		Where(entstrategy.IsPublic(true)).
		Order(ent.Desc(entstrategy.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*Strategy, len(strategies))
	for i, s := range strategies {
		result[i] = fromEntStrategy(s)
	}
	return result, nil
}

// Get gets a single strategy using ent
func (s *StrategyStore) Get(ctx context.Context, userID, id string) (*Strategy, error) {
	if s.ec == nil {
		return nil, fmt.Errorf("ent client not available")
	}
	st, err := s.ec.Strategy.Query().
		Where(entstrategy.And(entstrategy.ID(id), entstrategy.Or(entstrategy.UserID(userID), entstrategy.IsDefault(true)))).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return fromEntStrategy(st), nil
}

// GetActive gets user's currently active strategy using ent
func (s *StrategyStore) GetActive(ctx context.Context, userID string) (*Strategy, error) {
	if s.ec == nil {
		return nil, fmt.Errorf("ent client not available")
	}
	st, err := s.ec.Strategy.Query().
		Where(entstrategy.And(entstrategy.UserID(userID), entstrategy.IsActive(true))).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}
	if err == nil {
		return fromEntStrategy(st), nil
	}
	return s.GetDefault(ctx)
}

// GetDefault gets system default strategy using ent
func (s *StrategyStore) GetDefault(ctx context.Context) (*Strategy, error) {
	if s.ec == nil {
		return nil, fmt.Errorf("ent client not available")
	}
	st, err := s.ec.Strategy.Query().Where(entstrategy.IsDefault(true)).First(ctx)
	if err != nil {
		return nil, err
	}
	return fromEntStrategy(st), nil
}

// SetActive sets active strategy using ent (deactivates others first)
func (s *StrategyStore) SetActive(ctx context.Context, userID, strategyID string) error {
	if s.ec == nil {
		return fmt.Errorf("ent client not available")
	}
	tx, err := s.ec.Tx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := tx.Strategy.Update().
		Where(entstrategy.UserID(userID)).
		SetIsActive(false).
		Exec(ctx); err != nil {
		return err
	}
	if err := tx.Strategy.Update().
		Where(entstrategy.And(entstrategy.ID(strategyID), entstrategy.Or(entstrategy.UserID(userID), entstrategy.IsDefault(true)))).
		SetIsActive(true).
		Exec(ctx); err != nil {
		return err
	}
	return tx.Commit()
}

// Duplicate duplicates a strategy (used to create custom strategy based on default strategy)
func (s *StrategyStore) Duplicate(ctx context.Context, userID, sourceID, newID, newName string) error {
	// get source strategy
	source, err := s.Get(ctx, userID, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get source strategy: %w", err)
	}

	// create new strategy
	newStrategy := &Strategy{
		ID:          newID,
		UserID:      userID,
		Name:        newName,
		Description: "Created based on [" + source.Name + "]",
		IsActive:    false,
		IsDefault:   false,
		Config:      source.Config,
	}

	return s.Create(ctx, newStrategy)
}

// ParseConfig parse strategy configuration JSON
func (s *Strategy) ParseConfig() (*StrategyConfig, error) {
	var config StrategyConfig
	if err := json.Unmarshal([]byte(s.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to parse strategy configuration: %w", err)
	}
	return &config, nil
}

// SetConfig set strategy configuration
func (s *Strategy) SetConfig(config *StrategyConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize strategy configuration: %w", err)
	}
	s.Config = string(data)
	return nil
}
