package store

import (
	"cmp"
	"context"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"nofx/ent"
	entposition "nofx/ent/traderposition"
)

// adaptivePriceRound rounds a price based on its magnitude to preserve meaningful precision.
// For small prices (like meme coins), it preserves more decimal places.
// It detects the number of decimal places needed from the reference price(s).
func adaptivePriceRound(price float64, referencePrices ...float64) float64 {
	if price == 0 {
		return 0
	}

	// Find the minimum magnitude among all prices (including the price itself)
	minMagnitude := math.Abs(price)
	for _, ref := range referencePrices {
		if ref > 0 && ref < minMagnitude {
			minMagnitude = ref
		}
	}

	// Determine decimal places needed based on price magnitude
	// For price 0.000000541, we need ~15 decimal places
	// For price 0.0001, we need ~8 decimal places
	// For price 1.0, we need ~4 decimal places
	var multiplier float64
	switch {
	case minMagnitude < 0.000001: // Ultra small (meme coins like CHEEMS, SHIB)
		multiplier = 1e15 // 15 decimal places
	case minMagnitude < 0.0001: // Very small (PEPE, FLOKI)
		multiplier = 1e12 // 12 decimal places
	case minMagnitude < 0.01: // Small
		multiplier = 1e10 // 10 decimal places
	case minMagnitude < 1: // Medium
		multiplier = 1e8 // 8 decimal places
	default: // Large
		multiplier = 1e6 // 6 decimal places
	}

	return math.Round(price*multiplier) / multiplier
}

// getPriceDecimalPlaces returns the number of decimal places in a price string
func getPriceDecimalPlaces(price float64) int {
	if price == 0 {
		return 0
	}
	s := strconv.FormatFloat(price, 'f', -1, 64)
	idx := strings.Index(s, ".")
	if idx == -1 {
		return 0
	}
	return len(s) - idx - 1
}

// TraderStats trading statistics metrics
type TraderStats struct {
	TotalTrades    int     `json:"total_trades"`
	WinTrades      int     `json:"win_trades"`
	LossTrades     int     `json:"loss_trades"`
	WinRate        float64 `json:"win_rate"`
	ProfitFactor   float64 `json:"profit_factor"`
	SharpeRatio    float64 `json:"sharpe_ratio"`
	TotalPnL       float64 `json:"total_pnl"`
	TotalFee       float64 `json:"total_fee"`
	AvgWin         float64 `json:"avg_win"`
	AvgLoss        float64 `json:"avg_loss"`
	MaxDrawdownPct float64 `json:"max_drawdown_pct"`
}

// TraderPosition position record
// All time fields use int64 millisecond timestamps (UTC) to avoid timezone issues
type TraderPosition struct {
	ID                 int64   `json:"id"`
	TraderID           string  `json:"trader_id"`
	ExchangeID         string  `json:"exchange_id"`
	ExchangeType       string  `json:"exchange_type"`
	ExchangePositionID string  `json:"exchange_position_id"`
	Symbol             string  `json:"symbol"`
	Side               string  `json:"side"`
	EntryQuantity      float64 `json:"entry_quantity"`
	Quantity           float64 `json:"quantity"`
	EntryPrice         float64 `json:"entry_price"`
	EntryOrderID       string  `json:"entry_order_id"`
	EntryTime          int64   `json:"entry_time"`
	ExitPrice          float64 `json:"exit_price"`
	ExitOrderID        string  `json:"exit_order_id"`
	ExitTime           int64   `json:"exit_time"`
	RealizedPnL        float64 `json:"realized_pnl"`
	Fee                float64 `json:"fee"`
	Leverage           int     `json:"leverage"`
	Status             string  `json:"status"`
	CloseReason        string  `json:"close_reason"`
	Source             string  `json:"source"`
	CreatedAt          int64   `json:"created_at"`
	UpdatedAt          int64   `json:"updated_at"`
}


// fromEntTraderPosition converts ent.TraderPosition to store.TraderPosition
func fromEntTraderPosition(p *ent.TraderPosition) *TraderPosition {
	if p == nil {
		return nil
	}
	pos := &TraderPosition{
		ID:                 p.ID,
		TraderID:           p.TraderID,
		ExchangeID:         p.ExchangeID,
		ExchangeType:       p.ExchangeType,
		ExchangePositionID: p.ExchangePositionID,
		Symbol:             p.Symbol,
		Side:               p.Side,
		EntryQuantity:      p.EntryQuantity,
		Quantity:           p.Quantity,
		EntryPrice:         p.EntryPrice,
		EntryOrderID:       p.EntryOrderID,
		EntryTime:          p.EntryTime,
		ExitPrice:          p.ExitPrice,
		ExitOrderID:        p.ExitOrderID,
		ExitTime:           p.ExitTime,
		RealizedPnL:        p.RealizedPnl,
		Fee:                p.Fee,
		Leverage:           p.Leverage,
		Status:             p.Status,
		CloseReason:        p.CloseReason,
		Source:             p.Source,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
	if pos.EntryQuantity == 0 {
		pos.EntryQuantity = pos.Quantity
	}
	return pos
}

// PositionStore position storage
type PositionStore struct {
	ec *ent.Client
}

// NewPositionStore creates position storage instance
func NewPositionStore() *PositionStore {
	return &PositionStore{}
}

// SetEntClient sets the ent client for testing
func (s *PositionStore) SetEntClient(ec *ent.Client) {
	s.ec = ec
}

// ClearAllPositions deletes all positions for testing
func (s *PositionStore) ClearAllPositions() error {
	if s.ec == nil {
		return fmt.Errorf("ent client not available")
	}
	ctx := context.Background()
	_, err := s.ec.TraderPosition.Delete().Exec(ctx)
	return err
}

// isPostgres checks if the database is PostgreSQL
func (s *PositionStore) isPostgres() bool {
	return false
}

// InitTables initializes position tables
func (s *PositionStore) InitTables() error {
	return nil
}

// Create creates position record
func (s *PositionStore) Create(pos *TraderPosition) error {
	pos.Status = "OPEN"
	if pos.EntryQuantity == 0 {
		pos.EntryQuantity = pos.Quantity
	}
	ctx := context.Background()
	_, err := s.ec.TraderPosition.Create().
		SetTraderID(pos.TraderID).
		SetExchangeID(pos.ExchangeID).
		SetExchangeType(pos.ExchangeType).
		SetExchangePositionID(pos.ExchangePositionID).
		SetSymbol(pos.Symbol).
		SetSide(pos.Side).
		SetEntryQuantity(pos.EntryQuantity).
		SetQuantity(pos.Quantity).
		SetEntryPrice(pos.EntryPrice).
		SetEntryOrderID(pos.EntryOrderID).
		SetEntryTime(pos.EntryTime).
		SetExitPrice(pos.ExitPrice).
		SetExitOrderID(pos.ExitOrderID).
		SetExitTime(pos.ExitTime).
		SetRealizedPnl(pos.RealizedPnL).
		SetFee(pos.Fee).
		SetLeverage(pos.Leverage).
		SetStatus(pos.Status).
		SetCloseReason(pos.CloseReason).
		SetSource(pos.Source).
		SetCreatedAt(pos.CreatedAt).
		SetUpdatedAt(pos.UpdatedAt).
		Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

// ClosePosition closes position
func (s *PositionStore) ClosePosition(id int64, exitPrice float64, exitOrderID string, realizedPnL float64, fee float64, closeReason string) error {
	nowMs := time.Now().UTC().UnixMilli()
	ctx := context.Background()
	_, err := s.ec.TraderPosition.Update().
		Where(entposition.ID(id)).
		SetExitPrice(exitPrice).
		SetExitOrderID(exitOrderID).
		SetExitTime(nowMs).
		SetRealizedPnl(realizedPnL).
		SetFee(fee).
		SetStatus("CLOSED").
		SetCloseReason(closeReason).
		SetUpdatedAt(nowMs).
		Save(ctx)
	return err
}

// UpdatePositionQuantityAndPrice updates position quantity and recalculates entry price
func (s *PositionStore) UpdatePositionQuantityAndPrice(id int64, addQty float64, addPrice float64, addFee float64) error {
	ctx := context.Background()
	pos, err := s.ec.TraderPosition.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get current position: %w", err)
	}

	currentEntryQty := pos.EntryQuantity
	if currentEntryQty == 0 {
		currentEntryQty = pos.Quantity
	}

	newQty := math.Round((pos.Quantity+addQty)*10000) / 10000
	newEntryQty := math.Round((currentEntryQty+addQty)*10000) / 10000
	newEntryPrice := (pos.EntryPrice*pos.Quantity + addPrice*addQty) / newQty
	// Use adaptive precision based on price magnitude (for meme coins with very small prices)
	newEntryPrice = adaptivePriceRound(newEntryPrice, pos.EntryPrice, addPrice)
	newFee := pos.Fee + addFee
	nowMs := time.Now().UTC().UnixMilli()

	_, err = s.ec.TraderPosition.Update().
		Where(entposition.ID(id)).
		SetQuantity(newQty).
		SetEntryQuantity(newEntryQty).
		SetEntryPrice(newEntryPrice).
		SetFee(newFee).
		SetUpdatedAt(nowMs).
		Save(ctx)
	return err
}

// ReducePositionQuantity reduces position quantity for partial close
// If quantity reaches 0 (or near 0), automatically closes the position
func (s *PositionStore) ReducePositionQuantity(id int64, reduceQty float64, exitPrice float64, addFee float64, addPnL float64) error {
	ctx := context.Background()
	pos, err := s.ec.TraderPosition.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get current position: %w", err)
	}

	newQty := math.Round((pos.Quantity-reduceQty)*10000) / 10000
	newFee := pos.Fee + addFee
	newPnL := pos.RealizedPnl + addPnL

	closedQty := pos.EntryQuantity - pos.Quantity
	newClosedQty := closedQty + reduceQty

	var newExitPrice float64
	if newClosedQty > 0 {
		newExitPrice = (pos.ExitPrice*closedQty + exitPrice*reduceQty) / newClosedQty
		// Use adaptive precision based on price magnitude (for meme coins with very small prices)
		newExitPrice = adaptivePriceRound(newExitPrice, pos.ExitPrice, exitPrice, pos.EntryPrice)
	}

	nowMs := time.Now().UTC().UnixMilli()

	// Check if position should be fully closed (quantity reduced to ~0)
	const QUANTITY_TOLERANCE = 0.0001
	if newQty <= QUANTITY_TOLERANCE {
		// Auto-close: set status to CLOSED
		_, err = s.ec.TraderPosition.Update().
			Where(entposition.ID(id)).
			SetQuantity(0).
			SetFee(newFee).
			SetExitPrice(newExitPrice).
			SetRealizedPnl(newPnL).
			SetStatus("CLOSED").
			SetExitTime(nowMs).
			SetCloseReason("sync").
			SetUpdatedAt(nowMs).
			Save(ctx)
		return err
	}

	_, err = s.ec.TraderPosition.Update().
		Where(entposition.ID(id)).
		SetQuantity(newQty).
		SetFee(newFee).
		SetExitPrice(newExitPrice).
		SetRealizedPnl(newPnL).
		SetUpdatedAt(nowMs).
		Save(ctx)
	return err
}

// UpdatePositionExchangeInfo updates exchange_id and exchange_type
func (s *PositionStore) UpdatePositionExchangeInfo(id int64, exchangeID, exchangeType string) error {
	nowMs := time.Now().UTC().UnixMilli()
	ctx := context.Background()
	_, err := s.ec.TraderPosition.Update().
		Where(entposition.ID(id)).
		SetExchangeID(exchangeID).
		SetExchangeType(exchangeType).
		SetUpdatedAt(nowMs).
		Save(ctx)
	return err
}

// ClosePositionFully marks position as fully closed
// exitTimeMs is Unix milliseconds UTC
func (s *PositionStore) ClosePositionFully(id int64, exitPrice float64, exitOrderID string, exitTimeMs int64, totalRealizedPnL float64, totalFee float64, closeReason string) error {
	ctx := context.Background()
	pos, err := s.ec.TraderPosition.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get position: %w", err)
	}

	quantity := pos.Quantity
	if pos.EntryQuantity > 0 {
		quantity = pos.EntryQuantity
	}

	_, err = s.ec.TraderPosition.Update().
		Where(entposition.ID(id)).
		SetQuantity(quantity).
		SetExitPrice(exitPrice).
		SetExitOrderID(exitOrderID).
		SetExitTime(exitTimeMs).
		SetRealizedPnl(totalRealizedPnL).
		SetFee(totalFee).
		SetStatus("CLOSED").
		SetCloseReason(closeReason).
		SetUpdatedAt(time.Now().UTC().UnixMilli()).
		Save(ctx)
	return err
}

// DeleteAllOpenPositions deletes all OPEN positions for a trader
func (s *PositionStore) DeleteAllOpenPositions(traderID string) error {
	ctx := context.Background()
	_, err := s.ec.TraderPosition.Delete().
		Where(entposition.TraderID(traderID), entposition.Status("OPEN")).
		Exec(ctx)
	return err
}

// GetOpenPositions gets all open positions
func (s *PositionStore) GetOpenPositions(traderID string) ([]*TraderPosition, error) {
	positions, err := s.ec.TraderPosition.Query().
		Where(entposition.TraderID(traderID), entposition.Status("OPEN")).
		Order(ent.Desc(entposition.FieldEntryTime)).
		All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query open positions: %w", err)
	}

	result := make([]*TraderPosition, len(positions))
	for i, p := range positions {
		result[i] = fromEntTraderPosition(p)
	}
	return result, nil
}

// GetOpenPositionBySymbol gets open position for specified symbol and direction
func (s *PositionStore) GetOpenPositionBySymbol(traderID, symbol, side string) (*TraderPosition, error) {
	pos, err := s.ec.TraderPosition.Query().
		Where(
			entposition.TraderID(traderID),
			entposition.Symbol(symbol),
			entposition.Side(side),
			entposition.Status("OPEN"),
		).
		Order(ent.Desc(entposition.FieldEntryTime)).
		First(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			// Try without USDT suffix for backward compatibility
			if strings.HasSuffix(symbol, "USDT") {
				baseSymbol := strings.TrimSuffix(symbol, "USDT")
				pos, err = s.ec.TraderPosition.Query().
					Where(
						entposition.TraderID(traderID),
						entposition.Symbol(baseSymbol),
						entposition.Side(side),
						entposition.Status("OPEN"),
					).
					Order(ent.Desc(entposition.FieldEntryTime)).
					First(context.Background())
				if err != nil {
					if ent.IsNotFound(err) {
						return nil, nil
					}
					return nil, err
				}
				return fromEntTraderPosition(pos), nil
			}
			return nil, nil
		}
		return nil, err
	}
	return fromEntTraderPosition(pos), nil
}

// GetClosedPositions gets closed positions
func (s *PositionStore) GetClosedPositions(traderID string, limit int) ([]*TraderPosition, error) {
	positions, err := s.ec.TraderPosition.Query().
		Where(entposition.TraderID(traderID), entposition.Status("CLOSED")).
		Order(ent.Desc(entposition.FieldExitTime)).
		Limit(limit).
		All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query closed positions: %w", err)
	}
	result := make([]*TraderPosition, len(positions))
	for i, p := range positions {
		result[i] = fromEntTraderPosition(p)
	}
	return result, nil
}

// GetAllOpenPositions gets all traders' open positions
func (s *PositionStore) GetAllOpenPositions() ([]*TraderPosition, error) {
	positions, err := s.ec.TraderPosition.Query().
		Where(entposition.Status("OPEN")).
		Order(ent.Asc(entposition.FieldTraderID), ent.Desc(entposition.FieldEntryTime)).
		All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query all open positions: %w", err)
	}
	result := make([]*TraderPosition, len(positions))
	for i, p := range positions {
		result[i] = fromEntTraderPosition(p)
	}
	return result, nil
}

// GetPositionStats gets position statistics
func (s *PositionStore) GetPositionStats(traderID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	positions, err := s.ec.TraderPosition.Query().
		Where(entposition.TraderID(traderID), entposition.Status("CLOSED")).
		All(context.Background())
	if err != nil {
		return nil, err
	}

	total := len(positions)
	wins := 0
	var totalPnL, totalFee float64
	for _, pos := range positions {
		if pos.RealizedPnl > 0 {
			wins++
		}
		totalPnL += pos.RealizedPnl
		totalFee += pos.Fee
	}

	stats["total_trades"] = total
	stats["win_trades"] = wins
	stats["total_pnl"] = totalPnL
	stats["total_fee"] = totalFee
	if total > 0 {
		stats["win_rate"] = float64(wins) / float64(total) * 100
	} else {
		stats["win_rate"] = 0.0
	}

	return stats, nil
}

// GetFullStats gets complete trading statistics
func (s *PositionStore) GetFullStats(traderID string) (*TraderStats, error) {
	stats := &TraderStats{}

	positions, err := s.ec.TraderPosition.Query().
		Where(entposition.TraderID(traderID), entposition.Status("CLOSED")).
		Order(ent.Asc(entposition.FieldExitTime)).
		All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query position statistics: %w", err)
	}
	if len(positions) == 0 {
		return stats, nil
	}

	var pnls []float64
	var totalWin, totalLoss float64

	for _, pos := range positions {
		stats.TotalTrades++
		stats.TotalPnL += pos.RealizedPnl
		stats.TotalFee += pos.Fee
		pnls = append(pnls, pos.RealizedPnl)

		if pos.RealizedPnl > 0 {
			stats.WinTrades++
			totalWin += pos.RealizedPnl
		} else if pos.RealizedPnl < 0 {
			stats.LossTrades++
			totalLoss += -pos.RealizedPnl
		}
	}

	if stats.TotalTrades > 0 {
		stats.WinRate = float64(stats.WinTrades) / float64(stats.TotalTrades) * 100
	}
	if totalLoss > 0 {
		stats.ProfitFactor = totalWin / totalLoss
	}
	if stats.WinTrades > 0 {
		stats.AvgWin = totalWin / float64(stats.WinTrades)
	}
	if stats.LossTrades > 0 {
		stats.AvgLoss = totalLoss / float64(stats.LossTrades)
	}
	if len(pnls) > 1 {
		stats.SharpeRatio = calculateSharpeRatioFromPnls(pnls)
	}
	if len(pnls) > 0 {
		stats.MaxDrawdownPct = calculateMaxDrawdownFromPnls(pnls)
	}

	return stats, nil
}

// RecentTrade recent trade record
type RecentTrade struct {
	Symbol       string  `json:"symbol"`
	Side         string  `json:"side"`
	EntryPrice   float64 `json:"entry_price"`
	ExitPrice    float64 `json:"exit_price"`
	RealizedPnL  float64 `json:"realized_pnl"`
	PnLPct       float64 `json:"pnl_pct"`
	EntryTime    int64   `json:"entry_time"`
	ExitTime     int64   `json:"exit_time"`
	HoldDuration string  `json:"hold_duration"`
}

// GetRecentTrades gets recent closed trades
func (s *PositionStore) GetRecentTrades(traderID string, limit int) ([]RecentTrade, error) {
	positions, err := s.ec.TraderPosition.Query().
		Where(entposition.TraderID(traderID), entposition.Status("CLOSED")).
		Order(ent.Desc(entposition.FieldExitTime)).
		Limit(limit).
		All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query recent trades: %w", err)
	}

	var trades []RecentTrade
	for _, pos := range positions {
		t := RecentTrade{
			Symbol:      pos.Symbol,
			Side:        strings.ToLower(pos.Side),
			EntryPrice:  pos.EntryPrice,
			ExitPrice:   pos.ExitPrice,
			RealizedPnL: pos.RealizedPnl,
			EntryTime:   pos.EntryTime / 1000, // Convert ms to seconds for API compatibility
		}

		if pos.ExitTime > 0 {
			t.ExitTime = pos.ExitTime / 1000 // Convert ms to seconds
			durationMs := pos.ExitTime - pos.EntryTime
			t.HoldDuration = formatDurationMs(durationMs)
		}

		if pos.EntryPrice > 0 {
			if t.Side == "long" {
				t.PnLPct = (pos.ExitPrice - pos.EntryPrice) / pos.EntryPrice * 100 * float64(pos.Leverage)
			} else {
				t.PnLPct = (pos.EntryPrice - pos.ExitPrice) / pos.EntryPrice * 100 * float64(pos.Leverage)
			}
		}

		trades = append(trades, t)
	}

	return trades, nil
}

// formatDuration formats a duration
func formatDuration(d time.Duration) string {
	return formatDurationMs(d.Milliseconds())
}

// formatDurationMs formats a duration in milliseconds
func formatDurationMs(ms int64) string {
	seconds := ms / 1000
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24

	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	if hours < 24 {
		remainingMins := minutes % 60
		if remainingMins == 0 {
			return fmt.Sprintf("%dh", hours)
		}
		return fmt.Sprintf("%dh%dm", hours, remainingMins)
	}
	remainingHours := hours % 24
	if remainingHours == 0 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dd%dh", days, remainingHours)
}

// calculateSharpeRatioFromPnls calculates Sharpe ratio
func calculateSharpeRatioFromPnls(pnls []float64) float64 {
	if len(pnls) < 2 {
		return 0
	}

	var sum float64
	for _, pnl := range pnls {
		sum += pnl
	}
	mean := sum / float64(len(pnls))

	var variance float64
	for _, pnl := range pnls {
		variance += (pnl - mean) * (pnl - mean)
	}
	stdDev := math.Sqrt(variance / float64(len(pnls)-1))

	if stdDev == 0 {
		return 0
	}

	return mean / stdDev
}

// calculateMaxDrawdownFromPnls calculates maximum drawdown
func calculateMaxDrawdownFromPnls(pnls []float64) float64 {
	if len(pnls) == 0 {
		return 0
	}

	const startingEquity = 10000.0
	equity := startingEquity
	peak := startingEquity
	var maxDD float64

	for _, pnl := range pnls {
		equity += pnl
		peak = max(peak, equity)
		if peak > 0 {
			dd := (peak - equity) / peak * 100
			maxDD = max(maxDD, dd)
		}
	}

	return maxDD
}

// SymbolStats per-symbol trading statistics
type SymbolStats struct {
	Symbol      string  `json:"symbol"`
	TotalTrades int     `json:"total_trades"`
	WinTrades   int     `json:"win_trades"`
	WinRate     float64 `json:"win_rate"`
	TotalPnL    float64 `json:"total_pnl"`
	AvgPnL      float64 `json:"avg_pnl"`
	AvgHoldMins float64 `json:"avg_hold_mins"`
}

// GetSymbolStats gets per-symbol trading statistics
func (s *PositionStore) GetSymbolStats(traderID string, limit int) ([]SymbolStats, error) {
	positions, err := s.ec.TraderPosition.Query().
		Where(entposition.TraderID(traderID), entposition.Status("CLOSED")).
		All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query symbol stats: %w", err)
	}

	// Group by symbol
	symbolMap := make(map[string]*SymbolStats)
	symbolHoldMins := make(map[string][]float64)

	for _, pos := range positions {
		if _, ok := symbolMap[pos.Symbol]; !ok {
			symbolMap[pos.Symbol] = &SymbolStats{Symbol: pos.Symbol}
			symbolHoldMins[pos.Symbol] = []float64{}
		}
		s := symbolMap[pos.Symbol]
		s.TotalTrades++
		s.TotalPnL += pos.RealizedPnl
		if pos.RealizedPnl > 0 {
			s.WinTrades++
		}

		if pos.ExitTime > 0 {
			holdMins := float64(pos.ExitTime-pos.EntryTime) / 60000.0 // ms to minutes
			symbolHoldMins[pos.Symbol] = append(symbolHoldMins[pos.Symbol], holdMins)
		}
	}

	var stats []SymbolStats
	for symbol, s := range symbolMap {
		if s.TotalTrades > 0 {
			s.WinRate = float64(s.WinTrades) / float64(s.TotalTrades) * 100
			s.AvgPnL = s.TotalPnL / float64(s.TotalTrades)
		}
		if len(symbolHoldMins[symbol]) > 0 {
			var totalMins float64
			for _, m := range symbolHoldMins[symbol] {
				totalMins += m
			}
			s.AvgHoldMins = totalMins / float64(len(symbolHoldMins[symbol]))
		}
		stats = append(stats, *s)
	}

	// Sort by TotalPnL descending and limit
	slices.SortStableFunc(stats, func(a, b SymbolStats) int {
		return -cmp.Compare(a.TotalPnL, b.TotalPnL)
	})

	if limit > 0 && len(stats) > limit {
		stats = stats[:limit]
	}

	return stats, nil
}

// HoldingTimeStats holding duration analysis
type HoldingTimeStats struct {
	Range      string  `json:"range"`
	TradeCount int     `json:"trade_count"`
	WinRate    float64 `json:"win_rate"`
	AvgPnL     float64 `json:"avg_pnl"`
}

// GetHoldingTimeStats analyzes performance by holding duration
func (s *PositionStore) GetHoldingTimeStats(traderID string) ([]HoldingTimeStats, error) {
	positions, err := s.ec.TraderPosition.Query().
		Where(entposition.TraderID(traderID), entposition.Status("CLOSED"), entposition.ExitTimeGT(0)).
		All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query holding time stats: %w", err)
	}

	rangeStats := map[string]*struct {
		count   int
		wins    int
		totalPnL float64
	}{
		"<1h":   {},
		"1-4h":  {},
		"4-24h": {},
		">24h":  {},
	}

	for _, pos := range positions {
		if pos.ExitTime == 0 {
			continue
		}
		holdHours := float64(pos.ExitTime-pos.EntryTime) / 3600000.0 // ms to hours

		var rangeKey string
		switch {
		case holdHours < 1:
			rangeKey = "<1h"
		case holdHours < 4:
			rangeKey = "1-4h"
		case holdHours < 24:
			rangeKey = "4-24h"
		default:
			rangeKey = ">24h"
		}

		r := rangeStats[rangeKey]
		r.count++
		r.totalPnL += pos.RealizedPnl
		if pos.RealizedPnl > 0 {
			r.wins++
		}
	}

	var stats []HoldingTimeStats
	for _, rangeKey := range []string{"<1h", "1-4h", "4-24h", ">24h"} {
		r := rangeStats[rangeKey]
		if r.count > 0 {
			stats = append(stats, HoldingTimeStats{
				Range:      rangeKey,
				TradeCount: r.count,
				WinRate:    float64(r.wins) / float64(r.count) * 100,
				AvgPnL:     r.totalPnL / float64(r.count),
			})
		}
	}

	return stats, nil
}

// DirectionStats long/short performance comparison
type DirectionStats struct {
	Side       string  `json:"side"`
	TradeCount int     `json:"trade_count"`
	WinRate    float64 `json:"win_rate"`
	TotalPnL   float64 `json:"total_pnl"`
	AvgPnL     float64 `json:"avg_pnl"`
}

// GetDirectionStats analyzes long vs short performance
func (s *PositionStore) GetDirectionStats(traderID string) ([]DirectionStats, error) {
	positions, err := s.ec.TraderPosition.Query().
		Where(entposition.TraderID(traderID), entposition.Status("CLOSED")).
		All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query direction stats: %w", err)
	}

	sideStats := make(map[string]*DirectionStats)
	for _, pos := range positions {
		if _, ok := sideStats[pos.Side]; !ok {
			sideStats[pos.Side] = &DirectionStats{Side: pos.Side}
		}
		s := sideStats[pos.Side]
		s.TradeCount++
		s.TotalPnL += pos.RealizedPnl
		if pos.RealizedPnl > 0 {
			s.WinRate++
		}
	}

	var stats []DirectionStats
	for _, s := range sideStats {
		if s.TradeCount > 0 {
			s.AvgPnL = s.TotalPnL / float64(s.TradeCount)
			s.WinRate = s.WinRate / float64(s.TradeCount) * 100
		}
		stats = append(stats, *s)
	}

	return stats, nil
}

// HistorySummary comprehensive trading history for AI context
type HistorySummary struct {
	TotalTrades    int     `json:"total_trades"`
	WinRate        float64 `json:"win_rate"`
	TotalPnL       float64 `json:"total_pnl"`
	AvgTradeReturn float64 `json:"avg_trade_return"`

	BestSymbols  []SymbolStats `json:"best_symbols"`
	WorstSymbols []SymbolStats `json:"worst_symbols"`

	LongWinRate  float64 `json:"long_win_rate"`
	ShortWinRate float64 `json:"short_win_rate"`
	LongPnL      float64 `json:"long_pnl"`
	ShortPnL     float64 `json:"short_pnl"`

	AvgHoldingMins float64 `json:"avg_holding_mins"`
	BestHoldRange  string  `json:"best_hold_range"`

	RecentWinRate float64 `json:"recent_win_rate"`
	RecentPnL     float64 `json:"recent_pnl"`

	CurrentStreak int `json:"current_streak"`
	MaxWinStreak  int `json:"max_win_streak"`
	MaxLoseStreak int `json:"max_lose_streak"`
}

// GetHistorySummary generates comprehensive AI context summary
func (s *PositionStore) GetHistorySummary(traderID string) (*HistorySummary, error) {
	summary := &HistorySummary{}

	fullStats, err := s.GetFullStats(traderID)
	if err != nil {
		return nil, err
	}
	summary.TotalTrades = fullStats.TotalTrades
	summary.WinRate = fullStats.WinRate
	summary.TotalPnL = fullStats.TotalPnL
	if fullStats.TotalTrades > 0 {
		summary.AvgTradeReturn = fullStats.TotalPnL / float64(fullStats.TotalTrades)
	}

	symbolStats, _ := s.GetSymbolStats(traderID, 20)
	if len(symbolStats) > 0 {
		for i := 0; i < len(symbolStats) && i < 3; i++ {
			if symbolStats[i].TotalPnL > 0 {
				summary.BestSymbols = append(summary.BestSymbols, symbolStats[i])
			}
		}
		for i := len(symbolStats) - 1; i >= 0 && len(summary.WorstSymbols) < 3; i-- {
			if symbolStats[i].TotalPnL < 0 {
				summary.WorstSymbols = append(summary.WorstSymbols, symbolStats[i])
			}
		}
	}

	dirStats, _ := s.GetDirectionStats(traderID)
	for _, d := range dirStats {
		if d.Side == "LONG" {
			summary.LongWinRate = d.WinRate
			summary.LongPnL = d.TotalPnL
		} else if d.Side == "SHORT" {
			summary.ShortWinRate = d.WinRate
			summary.ShortPnL = d.TotalPnL
		}
	}

	holdStats, _ := s.GetHoldingTimeStats(traderID)
	var bestHoldWinRate float64
	for _, h := range holdStats {
		if h.WinRate > bestHoldWinRate && h.TradeCount >= 3 {
			bestHoldWinRate = h.WinRate
			summary.BestHoldRange = h.Range
		}
	}

	// Calculate average holding time
	avgHoldingPositions, err := s.ec.TraderPosition.Query().
		Where(entposition.TraderID(traderID), entposition.Status("CLOSED"), entposition.ExitTimeGT(0)).
		All(context.Background())
	if err == nil && len(avgHoldingPositions) > 0 {
		var totalMins float64
		for _, pos := range avgHoldingPositions {
			if pos.ExitTime > 0 {
				totalMins += float64(pos.ExitTime-pos.EntryTime) / 60000.0 // ms to minutes
			}
		}
		summary.AvgHoldingMins = totalMins / float64(len(avgHoldingPositions))
	}

	// Recent 20 trades
	recentEnts, err := s.ec.TraderPosition.Query().
		Where(entposition.TraderID(traderID), entposition.Status("CLOSED")).
		Order(ent.Desc(entposition.FieldExitTime)).
		Limit(20).
		All(context.Background())
	if err == nil {
		for _, pos := range recentEnts {
			summary.RecentPnL += pos.RealizedPnl
			if pos.RealizedPnl > 0 {
				summary.RecentWinRate++
			}
		}
		if len(recentEnts) > 0 {
			summary.RecentWinRate = summary.RecentWinRate / float64(len(recentEnts)) * 100
		}
	}

	// Calculate streaks
	s.calculateStreaks(traderID, summary)

	return summary, nil
}

// calculateStreaks calculates win/loss streaks
func (s *PositionStore) calculateStreaks(traderID string, summary *HistorySummary) {
	positions, err := s.ec.TraderPosition.Query().
		Where(entposition.TraderID(traderID), entposition.Status("CLOSED")).
		Order(ent.Desc(entposition.FieldExitTime)).
		All(context.Background())
	if err != nil || len(positions) == 0 {
		return
	}

	var currentStreak, maxWin, maxLose int
	var prevWin *bool
	isFirst := true

	for _, pos := range positions {
		isWin := pos.RealizedPnl > 0

		if isFirst {
			if isWin {
				currentStreak = 1
			} else {
				currentStreak = -1
			}
			isFirst = false
		}

		if prevWin == nil {
			prevWin = &isWin
		} else if *prevWin == isWin {
			if isWin {
				currentStreak++
				maxWin = max(maxWin, currentStreak)
			} else {
				currentStreak--
				maxLose = max(maxLose, -currentStreak)
			}
		} else {
			if isWin {
				currentStreak = 1
			} else {
				currentStreak = -1
			}
			*prevWin = isWin
		}
	}

	summary.CurrentStreak = currentStreak
	summary.MaxWinStreak = maxWin
	summary.MaxLoseStreak = maxLose
}

// ExistsWithExchangePositionID checks if a position exists
func (s *PositionStore) ExistsWithExchangePositionID(exchangeID, exchangePositionID string) (bool, error) {
	if exchangePositionID == "" {
		return false, nil
	}
	exists, err := s.ec.TraderPosition.Query().
		Where(entposition.ExchangeID(exchangeID), entposition.ExchangePositionID(exchangePositionID)).
		Exist(context.Background())
	if err != nil {
		return false, fmt.Errorf("failed to check position existence: %w", err)
	}
	return exists, nil
}

// GetOpenPositionByExchangePositionID gets an OPEN position by exchange_position_id
func (s *PositionStore) GetOpenPositionByExchangePositionID(exchangeID, exchangePositionID string) (*TraderPosition, error) {
	if exchangePositionID == "" {
		return nil, nil
	}
	pos, err := s.ec.TraderPosition.Query().
		Where(
			entposition.ExchangeID(exchangeID),
			entposition.ExchangePositionID(exchangePositionID),
			entposition.Status("OPEN"),
		).
		First(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return fromEntTraderPosition(pos), nil
}

// ClosedPnLRecord represents a closed position record from exchange
// All time fields use int64 millisecond timestamps (UTC)
type ClosedPnLRecord struct {
	Symbol      string
	Side        string
	EntryPrice  float64
	ExitPrice   float64
	Quantity    float64
	RealizedPnL float64
	Fee         float64
	Leverage    int
	EntryTime   int64 // Unix milliseconds UTC
	ExitTime    int64 // Unix milliseconds UTC
	OrderID     string
	CloseType   string
	ExchangeID  string
}

// CreateFromClosedPnL creates a closed position record from exchange data
func (s *PositionStore) CreateFromClosedPnL(traderID, exchangeID, exchangeType string, record *ClosedPnLRecord) (bool, error) {
	if record.Symbol == "" {
		return false, nil
	}

	side := strings.ToUpper(record.Side)
	if side == "LONG" || side == "BUY" {
		side = "LONG"
	} else if side == "SHORT" || side == "SELL" {
		side = "SHORT"
	} else {
		return false, nil
	}

	if record.Quantity <= 0 || record.ExitPrice <= 0 || record.EntryPrice <= 0 {
		return false, nil
	}

	exchangePositionID := record.ExchangeID
	if exchangePositionID == "" {
		exchangePositionID = fmt.Sprintf("%s_%s_%d_%.8f", record.Symbol, side, record.ExitTime, record.RealizedPnL)
	}

	exists, err := s.ExistsWithExchangePositionID(exchangeID, exchangePositionID)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	exitTimeMs := record.ExitTime
	entryTimeMs := record.EntryTime

	// Validate timestamps (must be after year 2000 = ~946684800000 ms)
	minValidTime := int64(946684800000) // 2000-01-01 UTC in milliseconds
	if exitTimeMs < minValidTime {
		return false, nil
	}
	if entryTimeMs < minValidTime {
		entryTimeMs = exitTimeMs
	}
	if entryTimeMs > exitTimeMs {
		entryTimeMs = min(entryTimeMs, exitTimeMs)
	}

	nowMs := time.Now().UTC().UnixMilli()
	pos := &TraderPosition{
		TraderID:           traderID,
		ExchangeID:         exchangeID,
		ExchangeType:       exchangeType,
		ExchangePositionID: exchangePositionID,
		Symbol:             record.Symbol,
		Side:               side,
		Quantity:           record.Quantity,
		EntryQuantity:      record.Quantity,
		EntryPrice:         record.EntryPrice,
		EntryTime:          entryTimeMs,
		ExitPrice:          record.ExitPrice,
		ExitOrderID:        record.OrderID,
		ExitTime:           exitTimeMs,
		RealizedPnL:        record.RealizedPnL,
		Fee:                record.Fee,
		Leverage:           record.Leverage,
		Status:             "CLOSED",
		CloseReason:        record.CloseType,
		Source:             "sync",
		CreatedAt:          nowMs,
		UpdatedAt:          nowMs,
	}

	ctx := context.Background()
	_, err = s.ec.TraderPosition.Create().
		SetTraderID(pos.TraderID).
		SetExchangeID(pos.ExchangeID).
		SetExchangeType(pos.ExchangeType).
		SetExchangePositionID(pos.ExchangePositionID).
		SetSymbol(pos.Symbol).
		SetSide(pos.Side).
		SetEntryQuantity(pos.EntryQuantity).
		SetQuantity(pos.Quantity).
		SetEntryPrice(pos.EntryPrice).
		SetEntryOrderID(pos.EntryOrderID).
		SetEntryTime(pos.EntryTime).
		SetExitPrice(pos.ExitPrice).
		SetExitOrderID(pos.ExitOrderID).
		SetExitTime(pos.ExitTime).
		SetRealizedPnl(pos.RealizedPnL).
		SetFee(pos.Fee).
		SetLeverage(pos.Leverage).
		SetStatus(pos.Status).
		SetCloseReason(pos.CloseReason).
		SetSource(pos.Source).
		SetCreatedAt(pos.CreatedAt).
		SetUpdatedAt(pos.UpdatedAt).
		Save(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "unique constraint") {
			return false, nil
		}
		return false, fmt.Errorf("failed to create position from closed PnL: %w", err)
	}

	return true, nil
}

// GetLastClosedPositionTime gets the most recent exit time (Unix ms)
func (s *PositionStore) GetLastClosedPositionTime(traderID string) (int64, error) {
	pos, err := s.ec.TraderPosition.Query().
		Where(
			entposition.TraderID(traderID),
			entposition.Status("CLOSED"),
			entposition.ExitTimeGT(0),
		).
		Order(ent.Desc(entposition.FieldExitTime)).
		First(context.Background())
	if err != nil {
		if ent.IsNotFound(err) || pos == nil {
			return time.Now().UTC().Add(-30 * 24 * time.Hour).UnixMilli(), nil
		}
		return 0, fmt.Errorf("failed to get last closed position time: %w", err)
	}
	return pos.ExitTime, nil
}

// CreateOpenPosition creates an open position
func (s *PositionStore) CreateOpenPosition(pos *TraderPosition) error {
	if pos.ExchangePositionID != "" && pos.ExchangeID != "" {
		existingPos, err := s.GetOpenPositionByExchangePositionID(pos.ExchangeID, pos.ExchangePositionID)
		if err != nil {
			return err
		}
		if existingPos != nil {
			return s.UpdatePositionQuantityAndPrice(existingPos.ID, pos.Quantity, pos.EntryPrice, pos.Fee)
		}
		exists, err := s.ExistsWithExchangePositionID(pos.ExchangeID, pos.ExchangePositionID)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
	}

	if pos.Status == "" {
		pos.Status = "OPEN"
	}
	if pos.Source == "" {
		pos.Source = "system"
	}
	if pos.EntryQuantity == 0 {
		pos.EntryQuantity = pos.Quantity
	}

	ctx := context.Background()
	_, err := s.ec.TraderPosition.Create().
		SetTraderID(pos.TraderID).
		SetExchangeID(pos.ExchangeID).
		SetExchangeType(pos.ExchangeType).
		SetExchangePositionID(pos.ExchangePositionID).
		SetSymbol(pos.Symbol).
		SetSide(pos.Side).
		SetEntryQuantity(pos.EntryQuantity).
		SetQuantity(pos.Quantity).
		SetEntryPrice(pos.EntryPrice).
		SetEntryOrderID(pos.EntryOrderID).
		SetEntryTime(pos.EntryTime).
		SetExitPrice(pos.ExitPrice).
		SetExitOrderID(pos.ExitOrderID).
		SetExitTime(pos.ExitTime).
		SetRealizedPnl(pos.RealizedPnL).
		SetFee(pos.Fee).
		SetLeverage(pos.Leverage).
		SetStatus(pos.Status).
		SetCloseReason(pos.CloseReason).
		SetSource(pos.Source).
		SetCreatedAt(pos.CreatedAt).
		SetUpdatedAt(pos.UpdatedAt).
		Save(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "unique constraint") {
			existingPos, findErr := s.GetOpenPositionByExchangePositionID(pos.ExchangeID, pos.ExchangePositionID)
			if findErr != nil {
				return findErr
			}
			if existingPos != nil {
				return s.UpdatePositionQuantityAndPrice(existingPos.ID, pos.Quantity, pos.EntryPrice, pos.Fee)
			}
			return nil
		}
		return fmt.Errorf("failed to create open position: %w", err)
	}

	return nil
}

// ClosePositionWithAccurateData closes a position with accurate data from exchange
// exitTimeMs is Unix milliseconds UTC
func (s *PositionStore) ClosePositionWithAccurateData(id int64, exitPrice float64, exitOrderID string, exitTimeMs int64, realizedPnL float64, fee float64, closeReason string) error {
	ctx := context.Background()
	_, err := s.ec.TraderPosition.Update().
		Where(entposition.ID(id)).
		SetExitPrice(exitPrice).
		SetExitOrderID(exitOrderID).
		SetExitTime(exitTimeMs).
		SetRealizedPnl(realizedPnL).
		SetFee(fee).
		SetStatus("CLOSED").
		SetCloseReason(closeReason).
		SetUpdatedAt(time.Now().UTC().UnixMilli()).
		Save(ctx)
	return err
}

// SyncClosedPositions syncs closed positions from exchange
func (s *PositionStore) SyncClosedPositions(traderID, exchangeID, exchangeType string, records []ClosedPnLRecord) (int, int, error) {
	created, skipped := 0, 0
	for _, record := range records {
		rec := record
		wasCreated, err := s.CreateFromClosedPnL(traderID, exchangeID, exchangeType, &rec)
		if err != nil {
			return created, skipped, fmt.Errorf("failed to sync position: %w", err)
		}
		if wasCreated {
			created++
		} else {
			skipped++
		}
	}
	return created, skipped, nil
}
