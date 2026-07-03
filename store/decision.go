package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"nofx/ent"
	entdecision "nofx/ent/decisionrecord"
	entposition "nofx/ent/traderposition"
)

// DecisionStore decision log storage
type DecisionStore struct {
	ec *ent.Client
}

// DecisionRecordDB internal GORM model for decision_records table
type DecisionRecordDB struct {
	ID                  int64     `json:"id"`
	TraderID            string    `json:"trader_id"`
	CycleNumber         int       `json:"cycle_number"`
	Timestamp           time.Time `json:"timestamp"`
	SystemPrompt        string    `json:"system_prompt"`
	InputPrompt         string    `json:"input_prompt"`
	CoTTrace            string    `json:"cot_trace"`
	DecisionJSON        string    `json:"decision_json"`
	RawResponse         string    `json:"raw_response"`
	CandidateCoins      string    `json:"candidate_coins"`
	ExecutionLog        string    `json:"execution_log"`
	Decisions           string    `json:"decisions"`
	Success             bool      `json:"success"`
	ErrorMessage        string    `json:"error_message"`
	AIRequestDurationMs int64     `json:"ai_request_duration_ms"`
	CreatedAt           time.Time `json:"created_at"`
}

// DecisionRecord decision record (external API struct)
type DecisionRecord struct {
	ID                  int64              `json:"id"`
	TraderID            string             `json:"trader_id"`
	CycleNumber         int                `json:"cycle_number"`
	Timestamp           time.Time          `json:"timestamp"`
	SystemPrompt        string             `json:"system_prompt"`
	InputPrompt         string             `json:"input_prompt"`
	CoTTrace            string             `json:"cot_trace"`
	DecisionJSON        string             `json:"decision_json"`
	RawResponse         string             `json:"raw_response"` // Raw AI response for debugging
	CandidateCoins      []string           `json:"candidate_coins"`
	ExecutionLog        []string           `json:"execution_log"`
	Success             bool               `json:"success"`
	ErrorMessage        string             `json:"error_message"`
	AIRequestDurationMs int64              `json:"ai_request_duration_ms"`
	AccountState        AccountSnapshot    `json:"account_state"`
	Positions           []PositionSnapshot `json:"positions"`
	Decisions           []DecisionAction   `json:"decisions"`
}

// AccountSnapshot account state snapshot
type AccountSnapshot struct {
	TotalBalance          float64 `json:"total_balance"`
	AvailableBalance      float64 `json:"available_balance"`
	TotalUnrealizedProfit float64 `json:"total_unrealized_profit"`
	PositionCount         int     `json:"position_count"`
	MarginUsedPct         float64 `json:"margin_used_pct"`
	InitialBalance        float64 `json:"initial_balance"`
}

// PositionSnapshot position snapshot
type PositionSnapshot struct {
	Symbol           string  `json:"symbol"`
	Side             string  `json:"side"`
	PositionAmt      float64 `json:"position_amt"`
	EntryPrice       float64 `json:"entry_price"`
	MarkPrice        float64 `json:"mark_price"`
	UnrealizedProfit float64 `json:"unrealized_profit"`
	Leverage         float64 `json:"leverage"`
	LiquidationPrice float64 `json:"liquidation_price"`
}

// DecisionAction decision action
type DecisionAction struct {
	Action     string    `json:"action"`
	Symbol     string    `json:"symbol"`
	Quantity   float64   `json:"quantity"`
	Leverage   int       `json:"leverage"`
	Price      float64   `json:"price"`
	StopLoss   float64   `json:"stop_loss,omitempty"`   // Stop loss price
	TakeProfit float64   `json:"take_profit,omitempty"` // Take profit price
	Confidence int       `json:"confidence,omitempty"`  // AI confidence (0-100)
	Reasoning  string    `json:"reasoning,omitempty"`   // Brief reasoning
	OrderID    int64     `json:"order_id"`
	Timestamp  time.Time `json:"timestamp"`
	Success    bool      `json:"success"`
	Error      string    `json:"error"`
}

// Statistics statistics information
type Statistics struct {
	TotalCycles         int `json:"total_cycles"`
	SuccessfulCycles    int `json:"successful_cycles"`
	FailedCycles        int `json:"failed_cycles"`
	TotalOpenPositions  int `json:"total_open_positions"`
	TotalClosePositions int `json:"total_close_positions"`
}

// NewDecisionStore creates a new DecisionStore
func NewDecisionStore() *DecisionStore {
	return &DecisionStore{}
}

// initTables initializes AI decision log tables
func (s *DecisionStore) initTables() error {
	return nil
}

// toRecord converts DB model to API struct
func (db *DecisionRecordDB) toRecord() *DecisionRecord {
	record := &DecisionRecord{
		ID:                  db.ID,
		TraderID:            db.TraderID,
		CycleNumber:         db.CycleNumber,
		Timestamp:           db.Timestamp,
		SystemPrompt:        db.SystemPrompt,
		InputPrompt:         db.InputPrompt,
		CoTTrace:            db.CoTTrace,
		DecisionJSON:        db.DecisionJSON,
		RawResponse:         db.RawResponse,
		Success:             db.Success,
		ErrorMessage:        db.ErrorMessage,
		AIRequestDurationMs: db.AIRequestDurationMs,
	}
	json.Unmarshal([]byte(db.CandidateCoins), &record.CandidateCoins)
	json.Unmarshal([]byte(db.ExecutionLog), &record.ExecutionLog)
	json.Unmarshal([]byte(db.Decisions), &record.Decisions)
	return record
}

// fromEntDecisionRecord converts ent.DecisionRecord to store.DecisionRecordDB
func fromEntDecisionRecord(ed *ent.DecisionRecord) *DecisionRecordDB {
	if ed == nil {
		return nil
	}
	return &DecisionRecordDB{
		ID:                  ed.ID,
		TraderID:            ed.TraderID,
		CycleNumber:         ed.CycleNumber,
		Timestamp:           ed.Timestamp,
		SystemPrompt:        ed.SystemPrompt,
		InputPrompt:         ed.InputPrompt,
		CoTTrace:            ed.CotTrace,
		DecisionJSON:        ed.DecisionJSON,
		RawResponse:         ed.RawResponse,
		CandidateCoins:      ed.CandidateCoins,
		ExecutionLog:        ed.ExecutionLog,
		Decisions:           ed.Decisions,
		Success:             ed.Success,
		ErrorMessage:        ed.ErrorMessage,
		AIRequestDurationMs: ed.AiRequestDurationMs,
		CreatedAt:           ed.CreatedAt,
	}
}

// LogDecision logs decision
func (s *DecisionStore) LogDecision(record *DecisionRecord) error {
	ctx := context.Background()
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now().UTC()
	} else {
		record.Timestamp = record.Timestamp.UTC()
	}

	candidateCoinsJSON, _ := json.Marshal(record.CandidateCoins)
	executionLogJSON, _ := json.Marshal(record.ExecutionLog)
	decisionsJSON, _ := json.Marshal(record.Decisions)

	created, err := s.ec.DecisionRecord.Create().
		SetTraderID(record.TraderID).
		SetCycleNumber(record.CycleNumber).
		SetTimestamp(record.Timestamp).
		SetSystemPrompt(record.SystemPrompt).
		SetInputPrompt(record.InputPrompt).
		SetCotTrace(record.CoTTrace).
		SetDecisionJSON(record.DecisionJSON).
		SetRawResponse(record.RawResponse).
		SetCandidateCoins(string(candidateCoinsJSON)).
		SetExecutionLog(string(executionLogJSON)).
		SetDecisions(string(decisionsJSON)).
		SetSuccess(record.Success).
		SetErrorMessage(record.ErrorMessage).
		SetAiRequestDurationMs(record.AIRequestDurationMs).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert decision record: %w", err)
	}
	record.ID = created.ID
	return nil
}

// GetLatestRecords gets the latest N records for specified trader (sorted by time in ascending order: old to new)
func (s *DecisionStore) GetLatestRecords(traderID string, n int) ([]*DecisionRecord, error) {
	ctx := context.Background()
	dbRecords, err := s.ec.DecisionRecord.Query().
		Where(entdecision.TraderID(traderID)).
		Order(ent.Desc(entdecision.FieldTimestamp)).
		Limit(n).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query decision records: %w", err)
	}
	records := make([]*DecisionRecord, len(dbRecords))
	for i, db := range dbRecords {
		records[len(dbRecords)-1-i] = fromEntDecisionRecord(db).toRecord()
	}
	return records, nil
}

// GetAllLatestRecords gets the latest N records for all traders
func (s *DecisionStore) GetAllLatestRecords(n int) ([]*DecisionRecord, error) {
	ctx := context.Background()
	dbRecords, err := s.ec.DecisionRecord.Query().
		Order(ent.Desc(entdecision.FieldTimestamp)).
		Limit(n).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query decision records: %w", err)
	}
	records := make([]*DecisionRecord, len(dbRecords))
	for i, db := range dbRecords {
		records[len(dbRecords)-1-i] = fromEntDecisionRecord(db).toRecord()
	}
	return records, nil
}

// GetRecordsByDate gets all records for a specified trader on a specified date
func (s *DecisionStore) GetRecordsByDate(traderID string, date time.Time) ([]*DecisionRecord, error) {
	ctx := context.Background()
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)
	dbRecords, err := s.ec.DecisionRecord.Query().
		Where(
			entdecision.TraderID(traderID),
			entdecision.TimestampGTE(startOfDay),
			entdecision.TimestampLT(endOfDay),
		).
		Order(ent.Asc(entdecision.FieldTimestamp)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query decision records: %w", err)
	}
	records := make([]*DecisionRecord, len(dbRecords))
	for i, db := range dbRecords {
		records[i] = fromEntDecisionRecord(db).toRecord()
	}
	return records, nil
}

// CleanOldRecords cleans old records from N days ago
func (s *DecisionStore) CleanOldRecords(traderID string, days int) (int64, error) {
	ctx := context.Background()
	cutoffTime := time.Now().AddDate(0, 0, -days)
	n, err := s.ec.DecisionRecord.Delete().
		Where(entdecision.TraderID(traderID), entdecision.TimestampLT(cutoffTime)).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to clean old records: %w", err)
	}
	return int64(n), nil
}

// GetStatistics gets statistics information for specified trader
func (s *DecisionStore) GetStatistics(traderID string) (*Statistics, error) {
	ctx := context.Background()
	stats := &Statistics{}

	totalCount, err := s.ec.DecisionRecord.Query().Where(entdecision.TraderID(traderID)).Count(ctx)
	if err != nil {
		return nil, err
	}
	successCount, err := s.ec.DecisionRecord.Query().Where(entdecision.TraderID(traderID), entdecision.Success(true)).Count(ctx)
	if err != nil {
		return nil, err
	}

	stats.TotalCycles = totalCount
	stats.SuccessfulCycles = successCount
	stats.FailedCycles = totalCount - successCount

	// Count from trader_positions table
	openCount, err := s.ec.TraderPosition.Query().Where(entposition.TraderID(traderID)).Count(ctx)
	if err == nil {
		stats.TotalOpenPositions = openCount
	}
	closedCount, err := s.ec.TraderPosition.Query().Where(entposition.TraderID(traderID), entposition.Status("CLOSED")).Count(ctx)
	if err == nil {
		stats.TotalClosePositions = closedCount
	}

	return stats, nil
}

// GetAllStatistics gets statistics information for all traders
func (s *DecisionStore) GetAllStatistics() (*Statistics, error) {
	ctx := context.Background()
	stats := &Statistics{}

	totalCount, err := s.ec.DecisionRecord.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	successCount, err := s.ec.DecisionRecord.Query().Where(entdecision.Success(true)).Count(ctx)
	if err != nil {
		return nil, err
	}

	stats.TotalCycles = totalCount
	stats.SuccessfulCycles = successCount
	stats.FailedCycles = totalCount - successCount

	// Count from trader_positions table
	openCount, err := s.ec.TraderPosition.Query().Count(ctx)
	if err == nil {
		stats.TotalOpenPositions = openCount
	}
	closedCount, err := s.ec.TraderPosition.Query().Where(entposition.Status("CLOSED")).Count(ctx)
	if err == nil {
		stats.TotalClosePositions = closedCount
	}

	return stats, nil
}

// GetLastCycleNumber gets the last cycle number for specified trader
func (s *DecisionStore) GetLastCycleNumber(traderID string) (int, error) {
	ctx := context.Background()
	record, err := s.ec.DecisionRecord.Query().
		Where(entdecision.TraderID(traderID)).
		Order(ent.Desc(entdecision.FieldCycleNumber)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	return record.CycleNumber, nil
}
