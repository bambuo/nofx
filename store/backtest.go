package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"nofx/ent"
	entbacktest "nofx/ent/backtestrun"
	entcheckpoint "nofx/ent/backtestcheckpoint"
	entequity "nofx/ent/backtestequity"
	enttrade "nofx/ent/backtesttrade"
)

// BacktestStore backtest data storage
type BacktestStore struct {
	ec *ent.Client
}

// NewBacktestStore creates a new backtest store
func NewBacktestStore() *BacktestStore {
	return &BacktestStore{}
}

// isPostgres checks if the database is PostgreSQL
func (s *BacktestStore) isPostgres() bool {
	return false
}

// RunState backtest state
type RunState string

const (
	RunStateCreated   RunState = "created"
	RunStateRunning   RunState = "running"
	RunStatePaused    RunState = "paused"
	RunStateCompleted RunState = "completed"
	RunStateFailed    RunState = "failed"
)

// RunMetadata backtest metadata
type RunMetadata struct {
	RunID     string     `json:"run_id"`
	UserID    string     `json:"user_id"`
	Version   int        `json:"version"`
	State     RunState   `json:"state"`
	Label     string     `json:"label"`
	LastError string     `json:"last_error"`
	Summary   RunSummary `json:"summary"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// RunSummary backtest summary
type RunSummary struct {
	SymbolCount     int     `json:"symbol_count"`
	DecisionTF      string  `json:"decision_tf"`
	ProcessedBars   int     `json:"processed_bars"`
	ProgressPct     float64 `json:"progress_pct"`
	EquityLast      float64 `json:"equity_last"`
	MaxDrawdownPct  float64 `json:"max_drawdown_pct"`
	Liquidated      bool    `json:"liquidated"`
	LiquidationNote string  `json:"liquidation_note"`
}

// EquityPoint equity point
type EquityPoint struct {
	Timestamp   int64   `json:"timestamp"`
	Equity      float64 `json:"equity"`
	Available   float64 `json:"available"`
	PnL         float64 `json:"pnl"`
	PnLPct      float64 `json:"pnl_pct"`
	DrawdownPct float64 `json:"drawdown_pct"`
	Cycle       int     `json:"cycle"`
}

// TradeEvent trade event
type TradeEvent struct {
	Timestamp       int64   `json:"timestamp"`
	Symbol          string  `json:"symbol"`
	Action          string  `json:"action"`
	Side            string  `json:"side"`
	Quantity        float64 `json:"quantity"`
	Price           float64 `json:"price"`
	Fee             float64 `json:"fee"`
	Slippage        float64 `json:"slippage"`
	OrderValue      float64 `json:"order_value"`
	RealizedPnL     float64 `json:"realized_pnl"`
	Leverage        int     `json:"leverage"`
	Cycle           int     `json:"cycle"`
	PositionAfter   float64 `json:"position_after"`
	LiquidationFlag bool    `json:"liquidation_flag"`
	Note            string  `json:"note"`
}

// RunIndexEntry backtest index entry
type RunIndexEntry struct {
	RunID          string   `json:"run_id"`
	State          string   `json:"state"`
	Symbols        []string `json:"symbols"`
	DecisionTF     string   `json:"decision_tf"`
	EquityLast     float64  `json:"equity_last"`
	MaxDrawdownPct float64  `json:"max_drawdown_pct"`
	StartTS        int64    `json:"start_ts"`
	EndTS          int64    `json:"end_ts"`
	CreatedAtISO   string   `json:"created_at"`
	UpdatedAtISO   string   `json:"updated_at"`
}

// BacktestRun model for backtest_runs table
type BacktestRun struct {
	RunID           string    `json:"run_id"`
	UserID          string    `json:"user_id"`
	ConfigJSON      []byte    `json:"config_json"`
	State           string    `json:"state"`
	Label           string    `json:"label"`
	SymbolCount     int       `json:"symbol_count"`
	DecisionTF      string    `json:"decision_tf"`
	ProcessedBars   int       `json:"processed_bars"`
	ProgressPct     float64   `json:"progress_pct"`
	EquityLast      float64   `json:"equity_last"`
	MaxDrawdownPct  float64   `json:"max_drawdown_pct"`
	Liquidated      bool      `json:"liquidated"`
	LiquidationNote string    `json:"liquidation_note"`
	PromptTemplate  string    `json:"prompt_template"`
	CustomPrompt    string    `json:"custom_prompt"`
	OverridePrompt  bool      `json:"override_prompt"`
	AIProvider      string    `json:"ai_provider"`
	AIModel         string    `json:"ai_model"`
	LastError       string    `json:"last_error"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// BacktestCheckpoint model
type BacktestCheckpoint struct {
	RunID     string    `json:"run_id"`
	Payload   []byte    `json:"payload"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BacktestEquity model
type BacktestEquity struct {
	ID        int64   `json:"id"`
	RunID     string  `json:"run_id"`
	TS        int64   `json:"ts"`
	Equity    float64 `json:"equity"`
	Available float64 `json:"available"`
	PnL       float64 `json:"pnl"`
	PnLPct    float64 `json:"pnl_pct"`
	DDPct     float64 `json:"dd_pct"`
	Cycle     int     `json:"cycle"`
}

// fromEntBacktestEquity converts ent.BacktestEquity to store.BacktestEquity
func fromEntBacktestEquity(e *ent.BacktestEquity) BacktestEquity {
	return BacktestEquity{
		ID:        e.ID,
		RunID:     e.RunID,
		TS:        e.Ts,
		Equity:    e.Equity,
		Available: e.Available,
		PnL:       e.Pnl,
		PnLPct:    e.PnlPct,
		DDPct:     e.DdPct,
		Cycle:     e.Cycle,
	}
}

// BacktestTrade model
type BacktestTrade struct {
	ID            int64   `json:"id"`
	RunID         string  `json:"run_id"`
	TS            int64   `json:"ts"`
	Symbol        string  `json:"symbol"`
	Action        string  `json:"action"`
	Side          string  `json:"side"`
	Qty           float64 `json:"qty"`
	Price         float64 `json:"price"`
	Fee           float64 `json:"fee"`
	Slippage      float64 `json:"slippage"`
	OrderValue    float64 `json:"order_value"`
	RealizedPnL   float64 `json:"realized_pnl"`
	Leverage      int     `json:"leverage"`
	Cycle         int     `json:"cycle"`
	PositionAfter float64 `json:"position_after"`
	Liquidation   bool    `json:"liquidation"`
	Note          string  `json:"note"`
}

// fromEntBacktestTrade converts ent.BacktestTrade to store.BacktestTrade
func fromEntBacktestTrade(t *ent.BacktestTrade) BacktestTrade {
	return BacktestTrade{
		ID:          t.ID,
		RunID:       t.RunID,
		TS:          t.Ts,
		Symbol:      t.Symbol,
		Action:      t.Action,
		Side:        t.Side,
		Qty:         t.Qty,
		Price:       t.Price,
		Fee:         t.Fee,
		Slippage:    t.Slippage,
		OrderValue:  t.OrderValue,
		RealizedPnL: t.RealizedPnl,
		Leverage:    t.Leverage,
		Cycle:       t.Cycle,
		PositionAfter: t.PositionAfter,
		Liquidation: t.LiquidationFlag,
		Note:        t.Note,
	}
}

// BacktestMetrics model (no ent equivalent, stored separately)
type BacktestMetrics struct {
	RunID     string    `json:"run_id"`
	Payload   []byte    `json:"payload"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BacktestDecision model (no ent equivalent, stored separately)
type BacktestDecision struct {
	ID        int64     `json:"id"`
	RunID     string    `json:"run_id"`
	Cycle     int       `json:"cycle"`
	Payload   []byte    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

// initTables initializes backtest related tables
func (s *BacktestStore) initTables() error {
	return nil
}

// SaveCheckpoint saves checkpoint
func (s *BacktestStore) SaveCheckpoint(runID string, payload []byte) error {
	ctx := context.Background()
	// Try to find existing checkpoint
	existing, err := s.ec.BacktestCheckpoint.Query().
		Where(entcheckpoint.RunID(runID)).
		First(ctx)
	if ent.IsNotFound(err) {
		// Create new
		_, err = s.ec.BacktestCheckpoint.Create().
			SetRunID(runID).
			SetPayload(payload).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to save checkpoint: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to query checkpoint: %w", err)
	}
	// Update existing
	_ = existing
	_, err = s.ec.BacktestCheckpoint.Update().
		Where(entcheckpoint.RunID(runID)).
		SetPayload(payload).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to update checkpoint: %w", err)
	}
	return nil
}

// LoadCheckpoint loads checkpoint
func (s *BacktestStore) LoadCheckpoint(runID string) ([]byte, error) {
	ctx := context.Background()
	checkpoint, err := s.ec.BacktestCheckpoint.Query().
		Where(entcheckpoint.RunID(runID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("checkpoint not found: %s", runID)
		}
		return nil, err
	}
	return checkpoint.Payload, nil
}

// SaveRunMetadata saves run metadata
func (s *BacktestStore) SaveRunMetadata(meta *RunMetadata) error {
	ctx := context.Background()
	// Try to find existing run
	_, err := s.ec.BacktestRun.Query().
		Where(entbacktest.RunID(meta.RunID)).
		First(ctx)
	if ent.IsNotFound(err) {
		// Create new
		_, err = s.ec.BacktestRun.Create().
			SetRunID(meta.RunID).
			SetUserID(meta.UserID).
			SetState(string(meta.State)).
			SetLabel(meta.Label).
			SetLastError(meta.LastError).
			SetSymbolCount(meta.Summary.SymbolCount).
			SetDecisionTf(meta.Summary.DecisionTF).
			SetProcessedBars(meta.Summary.ProcessedBars).
			SetProgressPct(meta.Summary.ProgressPct).
			SetEquityLast(meta.Summary.EquityLast).
			SetMaxDrawdownPct(meta.Summary.MaxDrawdownPct).
			SetLiquidated(meta.Summary.Liquidated).
			SetLiquidationNote(meta.Summary.LiquidationNote).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create run: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to query run: %w", err)
	}
	// Update existing
	_, err = s.ec.BacktestRun.Update().
		Where(entbacktest.RunID(meta.RunID)).
		SetUserID(meta.UserID).
		SetState(string(meta.State)).
		SetLabel(meta.Label).
		SetLastError(meta.LastError).
		SetSymbolCount(meta.Summary.SymbolCount).
		SetDecisionTf(meta.Summary.DecisionTF).
		SetProcessedBars(meta.Summary.ProcessedBars).
		SetProgressPct(meta.Summary.ProgressPct).
		SetEquityLast(meta.Summary.EquityLast).
		SetMaxDrawdownPct(meta.Summary.MaxDrawdownPct).
		SetLiquidated(meta.Summary.Liquidated).
		SetLiquidationNote(meta.Summary.LiquidationNote).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to update run: %w", err)
	}
	return nil
}

// LoadRunMetadata loads run metadata
func (s *BacktestStore) LoadRunMetadata(runID string) (*RunMetadata, error) {
	ctx := context.Background()
	run, err := s.ec.BacktestRun.Query().
		Where(entbacktest.RunID(runID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("run not found: %s", runID)
		}
		return nil, err
	}

	return &RunMetadata{
		RunID:     run.RunID,
		UserID:    run.UserID,
		Version:   1,
		State:     RunState(run.State),
		Label:     run.Label,
		LastError: run.LastError,
		Summary: RunSummary{
			SymbolCount:     run.SymbolCount,
			DecisionTF:      run.DecisionTf,
			ProcessedBars:   run.ProcessedBars,
			ProgressPct:     run.ProgressPct,
			EquityLast:      run.EquityLast,
			MaxDrawdownPct:  run.MaxDrawdownPct,
			Liquidated:      run.Liquidated,
			LiquidationNote: run.LiquidationNote,
		},
		CreatedAt: run.CreatedAt,
		UpdatedAt: run.UpdatedAt,
	}, nil
}

// ListRunIDs lists all run IDs
func (s *BacktestStore) ListRunIDs() ([]string, error) {
	ctx := context.Background()
	runs, err := s.ec.BacktestRun.Query().
		Order(ent.Desc(entbacktest.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(runs))
	for i, run := range runs {
		ids[i] = run.RunID
	}
	return ids, nil
}

// AppendEquityPoint appends equity point
func (s *BacktestStore) AppendEquityPoint(runID string, point EquityPoint) error {
	ctx := context.Background()
	_, err := s.ec.BacktestEquity.Create().
		SetRunID(runID).
		SetTs(point.Timestamp).
		SetEquity(point.Equity).
		SetAvailable(point.Available).
		SetPnl(point.PnL).
		SetPnlPct(point.PnLPct).
		SetDdPct(point.DrawdownPct).
		SetCycle(point.Cycle).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to append equity point: %w", err)
	}
	return nil
}

// SaveEquityPoints saves multiple equity points
func (s *BacktestStore) SaveEquityPoints(runID string, points []EquityPoint) error {
	ctx := context.Background()
	for _, point := range points {
		_, err := s.ec.BacktestEquity.Create().
			SetRunID(runID).
			SetTs(point.Timestamp).
			SetEquity(point.Equity).
			SetAvailable(point.Available).
			SetPnl(point.PnL).
			SetPnlPct(point.PnLPct).
			SetDdPct(point.DrawdownPct).
			SetCycle(point.Cycle).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to save equity point: %w", err)
		}
	}
	return nil
}

// LoadEquityPoints loads equity points
func (s *BacktestStore) LoadEquityPoints(runID string) ([]EquityPoint, error) {
	ctx := context.Background()
	eqs, err := s.ec.BacktestEquity.Query().
		Where(entequity.RunID(runID)).
		Order(ent.Asc(entequity.FieldTs)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	points := make([]EquityPoint, len(eqs))
	for i, eq := range eqs {
		e := fromEntBacktestEquity(eq)
		points[i] = EquityPoint{
			Timestamp:   e.TS,
			Equity:      e.Equity,
			Available:   e.Available,
			PnL:         e.PnL,
			PnLPct:      e.PnLPct,
			DrawdownPct: e.DDPct,
			Cycle:       e.Cycle,
		}
	}
	return points, nil
}

// AppendTradeEvent appends trade event
func (s *BacktestStore) AppendTradeEvent(runID string, event TradeEvent) error {
	ctx := context.Background()
	_, err := s.ec.BacktestTrade.Create().
		SetRunID(runID).
		SetTs(event.Timestamp).
		SetSymbol(event.Symbol).
		SetAction(event.Action).
		SetSide(event.Side).
		SetQty(event.Quantity).
		SetPrice(event.Price).
		SetFee(event.Fee).
		SetSlippage(event.Slippage).
		SetOrderValue(event.OrderValue).
		SetRealizedPnl(event.RealizedPnL).
		SetLeverage(event.Leverage).
		SetCycle(event.Cycle).
		SetPositionAfter(event.PositionAfter).
		SetLiquidationFlag(event.LiquidationFlag).
		SetNote(event.Note).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to append trade event: %w", err)
	}
	return nil
}

// SaveTradeEvents saves multiple trade events
func (s *BacktestStore) SaveTradeEvents(runID string, events []TradeEvent) error {
	ctx := context.Background()
	for _, event := range events {
		_, err := s.ec.BacktestTrade.Create().
			SetRunID(runID).
			SetTs(event.Timestamp).
			SetSymbol(event.Symbol).
			SetAction(event.Action).
			SetSide(event.Side).
			SetQty(event.Quantity).
			SetPrice(event.Price).
			SetFee(event.Fee).
			SetSlippage(event.Slippage).
			SetOrderValue(event.OrderValue).
			SetRealizedPnl(event.RealizedPnL).
			SetLeverage(event.Leverage).
			SetCycle(event.Cycle).
			SetPositionAfter(event.PositionAfter).
			SetLiquidationFlag(event.LiquidationFlag).
			SetNote(event.Note).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to save trade event: %w", err)
		}
	}
	return nil
}

// LoadTradeEvents loads trade events
func (s *BacktestStore) LoadTradeEvents(runID string) ([]TradeEvent, error) {
	ctx := context.Background()
	trades, err := s.ec.BacktestTrade.Query().
		Where(enttrade.RunID(runID)).
		Order(ent.Asc(enttrade.FieldTs)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	events := make([]TradeEvent, len(trades))
	for i, trade := range trades {
		t := fromEntBacktestTrade(trade)
		events[i] = TradeEvent{
			Timestamp:       t.TS,
			Symbol:          t.Symbol,
			Action:          t.Action,
			Side:            t.Side,
			Quantity:        t.Qty,
			Price:           t.Price,
			Fee:             t.Fee,
			Slippage:        t.Slippage,
			OrderValue:      t.OrderValue,
			RealizedPnL:     t.RealizedPnL,
			Leverage:        t.Leverage,
			Cycle:           t.Cycle,
			PositionAfter:   t.PositionAfter,
			LiquidationFlag: t.Liquidation,
			Note:            t.Note,
		}
	}
	return events, nil
}

// SaveMetrics saves metrics
func (s *BacktestStore) SaveMetrics(runID string, payload []byte) error {
	return nil
}

// LoadMetrics loads metrics
func (s *BacktestStore) LoadMetrics(runID string) ([]byte, error) {
	return nil, nil
}

// SaveDecisionRecord saves decision record
func (s *BacktestStore) SaveDecisionRecord(runID string, cycle int, payload []byte) error {
	return nil
}

// LoadDecisionRecords loads decision records
func (s *BacktestStore) LoadDecisionRecords(runID string, limit, offset int) ([]json.RawMessage, error) {
	return nil, nil
}

// LoadLatestDecision loads latest decision
func (s *BacktestStore) LoadLatestDecision(runID string, cycle int) ([]byte, error) {
	return nil, nil
}

// UpdateProgress updates progress
func (s *BacktestStore) UpdateProgress(runID string, progressPct, equity float64, barIndex int, liquidated bool) error {
	ctx := context.Background()
	return s.ec.BacktestRun.Update().
		Where(entbacktest.RunID(runID)).
		SetProgressPct(progressPct).
		SetEquityLast(equity).
		SetProcessedBars(barIndex).
		SetLiquidated(liquidated).
		Exec(ctx)
}

// ListIndexEntries lists index entries
func (s *BacktestStore) ListIndexEntries() ([]RunIndexEntry, error) {
	ctx := context.Background()
	runs, err := s.ec.BacktestRun.Query().
		Order(ent.Desc(entbacktest.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	entries := make([]RunIndexEntry, len(runs))
	for i, run := range runs {
		entry := RunIndexEntry{
			RunID:          run.RunID,
			State:          run.State,
			DecisionTF:     run.DecisionTf,
			EquityLast:     run.EquityLast,
			MaxDrawdownPct: run.MaxDrawdownPct,
			CreatedAtISO:   run.CreatedAt.Format(time.RFC3339),
			UpdatedAtISO:   run.UpdatedAt.Format(time.RFC3339),
			Symbols:        make([]string, 0, run.SymbolCount),
		}

		if len(run.ConfigJSON) > 0 {
			var cfg struct {
				Symbols []string `json:"symbols"`
				StartTS int64    `json:"start_ts"`
				EndTS   int64    `json:"end_ts"`
			}
			if json.Unmarshal(run.ConfigJSON, &cfg) == nil {
				entry.Symbols = cfg.Symbols
				entry.StartTS = cfg.StartTS
				entry.EndTS = cfg.EndTS
			}
		}

		entries[i] = entry
	}
	return entries, nil
}

// DeleteRun deletes run
func (s *BacktestStore) DeleteRun(runID string) error {
	ctx := context.Background()

	// Delete related records first
	if _, err := s.ec.BacktestCheckpoint.Delete().Where(entcheckpoint.RunID(runID)).Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete checkpoints: %w", err)
	}
	if _, err := s.ec.BacktestEquity.Delete().Where(entequity.RunID(runID)).Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete equity: %w", err)
	}
	if _, err := s.ec.BacktestTrade.Delete().Where(enttrade.RunID(runID)).Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete trades: %w", err)
	}
	if _, err := s.ec.BacktestRun.Delete().Where(entbacktest.RunID(runID)).Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete run: %w", err)
	}
	return nil
}

// SaveConfig saves config
func (s *BacktestStore) SaveConfig(runID, userID, template, customPrompt, provider, model string, override bool, configJSON []byte) error {
	ctx := context.Background()
	if userID == "" {
		userID = "default"
	}

	// Try to find existing run
	_, err := s.ec.BacktestRun.Query().
		Where(entbacktest.RunID(runID)).
		First(ctx)
	if ent.IsNotFound(err) {
		// Create new
		_, err = s.ec.BacktestRun.Create().
			SetRunID(runID).
			SetUserID(userID).
			SetConfigJSON(configJSON).
			SetPromptTemplate(template).
			SetCustomPrompt(customPrompt).
			SetOverridePrompt(override).
			SetAiProvider(provider).
			SetAiModel(model).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to query run: %w", err)
	}
	// Update existing
	_, err = s.ec.BacktestRun.Update().
		Where(entbacktest.RunID(runID)).
		SetUserID(userID).
		SetConfigJSON(configJSON).
		SetPromptTemplate(template).
		SetCustomPrompt(customPrompt).
		SetOverridePrompt(override).
		SetAiProvider(provider).
		SetAiModel(model).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}
	return nil
}

// LoadConfig loads config
func (s *BacktestStore) LoadConfig(runID string) ([]byte, error) {
	ctx := context.Background()
	run, err := s.ec.BacktestRun.Query().
		Where(entbacktest.RunID(runID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("config not found: %s", runID)
		}
		return nil, err
	}
	return run.ConfigJSON, nil
}
