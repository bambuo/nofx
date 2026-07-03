package store

import (
	"context"
	"fmt"
	"time"

	"nofx/ent"
	entequity "nofx/ent/equitysnapshot"
	enttrader "nofx/ent/trader"
)

// TraderStore trader storage
type TraderStore struct {
	ec *ent.Client
}

// NewTraderStore creates a new trader store
func NewTraderStore() *TraderStore {
	return &TraderStore{}
}

// Trader trader configuration
type Trader struct {
	ID                  string    `gorm:"primaryKey" json:"id"`
	UserID              string    `gorm:"column:user_id;not null;default:default;index" json:"user_id"`
	Name                string    `gorm:"column:name;not null" json:"name"`
	AIModelID           string    `gorm:"column:ai_model_id;not null" json:"ai_model_id"`
	ExchangeID          string    `gorm:"column:exchange_id;not null" json:"exchange_id"`
	StrategyID          string    `gorm:"column:strategy_id;default:''" json:"strategy_id"`
	InitialBalance      float64   `gorm:"column:initial_balance;not null" json:"initial_balance"`
	ScanIntervalMinutes int       `gorm:"column:scan_interval_minutes;default:3" json:"scan_interval_minutes"`
	IsRunning           bool      `gorm:"column:is_running;default:false" json:"is_running"`
	IsCrossMargin       bool      `gorm:"column:is_cross_margin;default:true" json:"is_cross_margin"`
	ShowInCompetition   bool      `gorm:"column:show_in_competition;default:true" json:"show_in_competition"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Following fields are deprecated, kept for backward compatibility, new traders should use StrategyID
	BTCETHLeverage       int    `gorm:"column:btc_eth_leverage;default:5" json:"btc_eth_leverage,omitempty"`
	AltcoinLeverage      int    `gorm:"column:altcoin_leverage;default:5" json:"altcoin_leverage,omitempty"`
	TradingSymbols       string `gorm:"column:trading_symbols;default:''" json:"trading_symbols,omitempty"`
	UseAI500             bool   `gorm:"column:use_coin_pool;default:false" json:"use_ai500,omitempty"`
	UseOITop             bool   `gorm:"column:use_oi_top;default:false" json:"use_oi_top,omitempty"`
	CustomPrompt         string `gorm:"column:custom_prompt;default:''" json:"custom_prompt,omitempty"`
	OverrideBasePrompt   bool   `gorm:"column:override_base_prompt;default:false" json:"override_base_prompt,omitempty"`
	SystemPromptTemplate string `gorm:"column:system_prompt_template;default:default" json:"system_prompt_template,omitempty"`
}

// TableName returns the table name for Trader
func (Trader) TableName() string {
	return "traders"
}

// fromEntTrader converts ent.Trader to store.Trader
func fromEntTrader(et *ent.Trader) *Trader {
	if et == nil {
		return nil
	}
	return &Trader{
		ID:                  et.ID,
		UserID:              et.UserID,
		Name:                et.Name,
		AIModelID:           et.AiModelID,
		ExchangeID:          et.ExchangeID,
		StrategyID:          et.StrategyID,
		InitialBalance:      et.InitialBalance,
		ScanIntervalMinutes: et.ScanIntervalMinutes,
		IsRunning:           et.IsRunning,
		IsCrossMargin:       et.IsCrossMargin,
		ShowInCompetition:   et.ShowInCompetition,
		CreatedAt:           et.CreatedAt,
		UpdatedAt:           et.UpdatedAt,
		BTCETHLeverage:      et.BtcEthLeverage,
		AltcoinLeverage:     et.AltcoinLeverage,
		TradingSymbols:      et.TradingSymbols,
		UseAI500:            et.UseCoinPool,
		UseOITop:            et.UseOiTop,
		CustomPrompt:        et.CustomPrompt,
		OverrideBasePrompt:  et.OverrideBasePrompt,
		SystemPromptTemplate: et.SystemPromptTemplate,
	}
}

// TraderFullConfig trader full configuration (includes AI model, exchange and strategy)
type TraderFullConfig struct {
	Trader   *Trader
	AIModel  *AIModel
	Exchange *Exchange
	Strategy *Strategy
}

func (s *TraderStore) initTables() error {
	return nil
}

// Create creates trader
func (s *TraderStore) Create(ctx context.Context, trader *Trader) error {
	_, err := s.ec.Trader.Create().
		SetID(trader.ID).
		SetUserID(trader.UserID).
		SetName(trader.Name).
		SetAiModelID(trader.AIModelID).
		SetExchangeID(trader.ExchangeID).
		SetNillableStrategyID(strPtr(trader.StrategyID)).
		SetInitialBalance(trader.InitialBalance).
		SetScanIntervalMinutes(trader.ScanIntervalMinutes).
		SetIsRunning(trader.IsRunning).
		SetIsCrossMargin(trader.IsCrossMargin).
		SetShowInCompetition(trader.ShowInCompetition).
		SetBtcEthLeverage(trader.BTCETHLeverage).
		SetAltcoinLeverage(trader.AltcoinLeverage).
		SetTradingSymbols(trader.TradingSymbols).
		SetUseCoinPool(trader.UseAI500).
		SetUseOiTop(trader.UseOITop).
		SetNillableCustomPrompt(strPtr(trader.CustomPrompt)).
		SetOverrideBasePrompt(trader.OverrideBasePrompt).
		SetNillableSystemPromptTemplate(strPtr(trader.SystemPromptTemplate)).
		Save(ctx)
	return err
}

// List gets user's trader list
func (s *TraderStore) List(ctx context.Context, userID string) ([]*Trader, error) {
	traders, err := s.ec.Trader.Query().
		Where(enttrader.UserID(userID)).
		Order(ent.Desc(enttrader.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*Trader, len(traders))
	for i, t := range traders {
		result[i] = fromEntTrader(t)
	}
	return result, nil
}

// UpdateStatus updates trader running status
func (s *TraderStore) UpdateStatus(ctx context.Context, userID, id string, isRunning bool) error {
	return s.ec.Trader.Update().
		Where(enttrader.And(enttrader.ID(id), enttrader.UserID(userID))).
		SetIsRunning(isRunning).
		Exec(ctx)
}

// UpdateShowInCompetition updates trader competition visibility
func (s *TraderStore) UpdateShowInCompetition(ctx context.Context, userID, id string, showInCompetition bool) error {
	return s.ec.Trader.Update().
		Where(enttrader.And(enttrader.ID(id), enttrader.UserID(userID))).
		SetShowInCompetition(showInCompetition).
		Exec(ctx)
}

// Update updates trader configuration
func (s *TraderStore) Update(ctx context.Context, trader *Trader) error {
	fmt.Printf("📝 TraderStore.Update: ID=%s, Name=%s, AIModelID=%s, StrategyID=%s\n",
		trader.ID, trader.Name, trader.AIModelID, trader.StrategyID)

	upd := s.ec.Trader.Update().
		Where(enttrader.And(enttrader.ID(trader.ID), enttrader.UserID(trader.UserID))).
		SetName(trader.Name).
		SetAiModelID(trader.AIModelID).
		SetExchangeID(trader.ExchangeID).
		SetStrategyID(trader.StrategyID).
		SetIsCrossMargin(trader.IsCrossMargin).
		SetShowInCompetition(trader.ShowInCompetition)

	// Only update these if > 0
	if trader.InitialBalance > 0 {
		upd = upd.SetInitialBalance(trader.InitialBalance)
	}
	if trader.ScanIntervalMinutes > 0 {
		upd = upd.SetScanIntervalMinutes(trader.ScanIntervalMinutes)
		fmt.Printf("📊 TraderStore.Update: scan_interval_minutes=%d will be saved\n", trader.ScanIntervalMinutes)
	} else {
		fmt.Printf("⚠️ TraderStore.Update: scan_interval_minutes=%d (<=0, NOT updating)\n", trader.ScanIntervalMinutes)
	}

	return upd.Exec(ctx)
}

// UpdateInitialBalance updates initial balance
func (s *TraderStore) UpdateInitialBalance(ctx context.Context, userID, id string, newBalance float64) error {
	return s.ec.Trader.Update().
		Where(enttrader.And(enttrader.ID(id), enttrader.UserID(userID))).
		SetInitialBalance(newBalance).
		Exec(ctx)
}

// UpdateCustomPrompt updates custom prompt
func (s *TraderStore) UpdateCustomPrompt(ctx context.Context, userID, id string, customPrompt string, overrideBase bool) error {
	return s.ec.Trader.Update().
		Where(enttrader.And(enttrader.ID(id), enttrader.UserID(userID))).
		SetCustomPrompt(customPrompt).
		SetOverrideBasePrompt(overrideBase).
		Exec(ctx)
}

// Delete deletes trader and associated data
func (s *TraderStore) Delete(ctx context.Context, userID, id string) error {
	// Delete associated equity snapshots first
	_, err := s.ec.EquitySnapshot.Delete().Where(entequity.TraderID(id)).Exec(ctx)
	if err != nil {
		return err
	}

	// Delete the trader
	n, err := s.ec.Trader.Delete().Where(enttrader.And(enttrader.ID(id), enttrader.UserID(userID))).Exec(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("trader not found")
	}
	return nil
}

// GetFullConfig gets trader full configuration (includes AI model, exchange and strategy)
func (s *TraderStore) GetFullConfig(ctx context.Context, userID, traderID string) (*TraderFullConfig, error) {
	trader, err := s.ec.Trader.Query().
		Where(enttrader.And(enttrader.ID(traderID), enttrader.UserID(userID))).
		WithAiModel().
		WithExchange().
		Only(ctx)
	if err != nil {
		return nil, err
	}

	full := &TraderFullConfig{
		Trader:   fromEntTrader(trader),
		AIModel:  fromEntAIModel(trader.Edges.AiModel),
		Exchange: fromEntExchange(trader.Edges.Exchange),
	}

	// Load strategy (with fallback)
	strategyStore := &StrategyStore{ec: s.ec}
	if trader.StrategyID != "" {
		full.Strategy, _ = strategyStore.Get(ctx, userID, trader.StrategyID)
	}
	if full.Strategy == nil {
		full.Strategy, _ = strategyStore.GetActive(ctx, userID)
	}

	return full, nil
}

// GetByID gets a trader by ID without requiring userID (for public APIs)
func (s *TraderStore) GetByID(ctx context.Context, traderID string) (*Trader, error) {
	t, err := s.ec.Trader.Get(ctx, traderID)
	if err != nil {
		return nil, err
	}
	return fromEntTrader(t), nil
}

// ListAll gets all traders
func (s *TraderStore) ListAll(ctx context.Context) ([]*Trader, error) {
	traders, err := s.ec.Trader.Query().Order(ent.Desc(enttrader.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*Trader, len(traders))
	for i, t := range traders {
		result[i] = fromEntTrader(t)
	}
	return result, nil
}

// ListByExchangeID gets traders that use a specific exchange
func (s *TraderStore) ListByExchangeID(ctx context.Context, userID, exchangeID string) ([]*Trader, error) {
	traders, err := s.ec.Trader.Query().
		Where(enttrader.And(enttrader.UserID(userID), enttrader.ExchangeID(exchangeID))).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*Trader, len(traders))
	for i, t := range traders {
		result[i] = fromEntTrader(t)
	}
	return result, nil
}

// ListByAIModelID gets traders that use a specific AI model
func (s *TraderStore) ListByAIModelID(ctx context.Context, userID, aiModelID string) ([]*Trader, error) {
	traders, err := s.ec.Trader.Query().
		Where(enttrader.And(enttrader.UserID(userID), enttrader.AiModelID(aiModelID))).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*Trader, len(traders))
	for i, t := range traders {
		result[i] = fromEntTrader(t)
	}
	return result, nil
}
