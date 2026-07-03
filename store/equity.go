package store

import (
	"context"
	"time"

	"nofx/ent"
	entequity "nofx/ent/equitysnapshot"
)

// EquityStore account equity storage (for plotting return curves)
type EquityStore struct {
	ec *ent.Client
}

// EquitySnapshot equity snapshot
type EquitySnapshot struct {
	ID            int64     `json:"id"`
	TraderID      string    `json:"trader_id"`
	Timestamp     time.Time `json:"timestamp"`
	TotalEquity   float64   `json:"total_equity"`
	Balance       float64   `json:"balance"`
	UnrealizedPnL float64   `json:"unrealized_pnl"`
	PositionCount int       `json:"position_count"`
	MarginUsedPct float64   `json:"margin_used_pct"`
	CreatedAt     time.Time `json:"created_at"`
}

// NewEquityStore creates a new EquityStore
func NewEquityStore() *EquityStore {
	return &EquityStore{}
}

// initTables initializes equity tables
func (s *EquityStore) initTables() error {
	return nil
}

// fromEntEquitySnapshot converts ent.EquitySnapshot to store.EquitySnapshot
func fromEntEquitySnapshot(es *ent.EquitySnapshot) *EquitySnapshot {
	if es == nil {
		return nil
	}
	return &EquitySnapshot{
		ID:            es.ID,
		TraderID:      es.TraderID,
		Timestamp:     es.Timestamp,
		TotalEquity:   es.TotalEquity,
		Balance:       es.Balance,
		UnrealizedPnL: es.UnrealizedPnl,
		PositionCount: es.PositionCount,
		MarginUsedPct: es.MarginUsedPct,
		CreatedAt:     es.CreatedAt,
	}
}

// Save saves equity snapshot
func (s *EquityStore) Save(snapshot *EquitySnapshot) error {
	ctx := context.Background()
	if snapshot.Timestamp.IsZero() {
		snapshot.Timestamp = time.Now().UTC()
	} else {
		snapshot.Timestamp = snapshot.Timestamp.UTC()
	}
	_, err := s.ec.EquitySnapshot.Create().
		SetTraderID(snapshot.TraderID).
		SetTimestamp(snapshot.Timestamp).
		SetTotalEquity(snapshot.TotalEquity).
		SetBalance(snapshot.Balance).
		SetUnrealizedPnl(snapshot.UnrealizedPnL).
		SetPositionCount(snapshot.PositionCount).
		SetMarginUsedPct(snapshot.MarginUsedPct).
		Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

// GetLatest gets the latest N equity records for specified trader (sorted in ascending chronological order: old to new)
func (s *EquityStore) GetLatest(traderID string, limit int) ([]*EquitySnapshot, error) {
	ctx := context.Background()
	snapshots, err := s.ec.EquitySnapshot.Query().
		Where(entequity.TraderID(traderID)).
		Order(ent.Desc(entequity.FieldTimestamp)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	// Reverse to sort from old to new (suitable for plotting curves)
	result := make([]*EquitySnapshot, len(snapshots))
	for i, s := range snapshots {
		result[len(snapshots)-1-i] = fromEntEquitySnapshot(s)
	}
	return result, nil
}

// GetByTimeRange gets equity records within specified time range
func (s *EquityStore) GetByTimeRange(traderID string, start, end time.Time) ([]*EquitySnapshot, error) {
	ctx := context.Background()
	snapshots, err := s.ec.EquitySnapshot.Query().
		Where(
			entequity.TraderID(traderID),
			entequity.TimestampGTE(start),
			entequity.TimestampLTE(end),
		).
		Order(ent.Asc(entequity.FieldTimestamp)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*EquitySnapshot, len(snapshots))
	for i, s := range snapshots {
		result[i] = fromEntEquitySnapshot(s)
	}
	return result, nil
}

// GetAllTradersLatest gets latest equity for all traders (for leaderboards)
func (s *EquityStore) GetAllTradersLatest() (map[string]*EquitySnapshot, error) {
	ctx := context.Background()
	snapshots, err := s.ec.EquitySnapshot.Query().
		Order(ent.Desc(entequity.FieldTimestamp)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Deduplicate by TraderID, keeping only the latest (first after DESC order)
	seen := make(map[string]bool)
	result := make(map[string]*EquitySnapshot)
	for _, snap := range snapshots {
		if !seen[snap.TraderID] {
			seen[snap.TraderID] = true
			result[snap.TraderID] = fromEntEquitySnapshot(snap)
		}
	}
	return result, nil
}

// CleanOldRecords cleans old records from N days ago
func (s *EquityStore) CleanOldRecords(traderID string, days int) (int64, error) {
	ctx := context.Background()
	cutoffTime := time.Now().AddDate(0, 0, -days)
	n, err := s.ec.EquitySnapshot.Delete().
		Where(entequity.TraderID(traderID), entequity.TimestampLT(cutoffTime)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return int64(n), nil
}

// GetCount gets record count for specified trader
func (s *EquityStore) GetCount(traderID string) (int, error) {
	ctx := context.Background()
	count, err := s.ec.EquitySnapshot.Query().
		Where(entequity.TraderID(traderID)).
		Count(ctx)
	return count, err
}

// MigrateFromDecision migrates data from old decision_account_snapshots table
// Migration is no longer needed as ent manages the schema
func (s *EquityStore) MigrateFromDecision() (int64, error) {
	return 0, nil
}
