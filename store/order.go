package store

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"nofx/ent"
	entfill "nofx/ent/traderfill"
	entorder "nofx/ent/traderorder"
)

// TraderOrder order record
// All time fields use int64 millisecond timestamps (UTC) to avoid timezone issues
type TraderOrder struct {
	ID                int64   `json:"id"`
	TraderID          string  `json:"trader_id"`
	ExchangeID        string  `json:"exchange_id"`
	ExchangeType      string  `json:"exchange_type"`
	ExchangeOrderID   string  `json:"exchange_order_id"`
	ClientOrderID     string  `json:"client_order_id"`
	Symbol            string  `json:"symbol"`
	Side              string  `json:"side"`
	PositionSide      string  `json:"position_side"`
	Type              string  `json:"type"`
	TimeInForce       string  `json:"time_in_force"`
	Quantity          float64 `json:"quantity"`
	Price             float64 `json:"price"`
	StopPrice         float64 `json:"stop_price"`
	Status            string  `json:"status"`
	FilledQuantity    float64 `json:"filled_quantity"`
	AvgFillPrice      float64 `json:"avg_fill_price"`
	Commission        float64 `json:"commission"`
	CommissionAsset   string  `json:"commission_asset"`
	Leverage          int     `json:"leverage"`
	ReduceOnly        bool    `json:"reduce_only"`
	ClosePosition     bool    `json:"close_position"`
	WorkingType       string  `json:"working_type"`
	PriceProtect      bool    `json:"price_protect"`
	OrderAction       string  `json:"order_action"`
	RelatedPositionID int64   `json:"related_position_id"`
	CreatedAt         int64   `json:"created_at"` // Unix milliseconds UTC
	UpdatedAt         int64   `json:"updated_at"` // Unix milliseconds UTC
	FilledAt          int64   `json:"filled_at"`  // Unix milliseconds UTC
}

// fromEntTraderOrder converts ent.TraderOrder to store.TraderOrder
func fromEntTraderOrder(o *ent.TraderOrder) *TraderOrder {
	if o == nil {
		return nil
	}
	return &TraderOrder{
		ID:                o.ID,
		TraderID:          o.TraderID,
		ExchangeID:        o.ExchangeID,
		ExchangeType:      o.ExchangeType,
		ExchangeOrderID:   o.ExchangeOrderID,
		ClientOrderID:     o.ClientOrderID,
		Symbol:            o.Symbol,
		Side:              o.Side,
		PositionSide:      o.PositionSide,
		Type:              o.TypeName,
		TimeInForce:       o.TimeInForce,
		Quantity:          o.Quantity,
		Price:             o.Price,
		StopPrice:         o.StopPrice,
		Status:            o.Status,
		FilledQuantity:    o.FilledQuantity,
		AvgFillPrice:      o.AvgFillPrice,
		Commission:        o.Commission,
		CommissionAsset:   o.CommissionAsset,
		Leverage:          o.Leverage,
		ReduceOnly:        o.ReduceOnly,
		ClosePosition:     o.ClosePosition,
		WorkingType:       o.WorkingType,
		PriceProtect:      o.PriceProtect,
		OrderAction:       o.OrderAction,
		RelatedPositionID: o.RelatedPositionID,
		CreatedAt:         o.CreatedAt,
		UpdatedAt:         o.UpdatedAt,
		FilledAt:          o.FilledAt,
	}
}

// TraderFill trade record
// All time fields use int64 millisecond timestamps (UTC) to avoid timezone issues
type TraderFill struct {
	ID              int64   `json:"id"`
	TraderID        string  `json:"trader_id"`
	ExchangeID      string  `json:"exchange_id"`
	ExchangeType    string  `json:"exchange_type"`
	OrderID         int64   `json:"order_id"`
	ExchangeOrderID string  `json:"exchange_order_id"`
	ExchangeTradeID string  `json:"exchange_trade_id"`
	Symbol          string  `json:"symbol"`
	Side            string  `json:"side"`
	Price           float64 `json:"price"`
	Quantity        float64 `json:"quantity"`
	QuoteQuantity   float64 `json:"quote_quantity"`
	Commission      float64 `json:"commission"`
	CommissionAsset string  `json:"commission_asset"`
	RealizedPnL     float64 `json:"realized_pnl"`
	IsMaker         bool    `json:"is_maker"`
	CreatedAt       int64   `json:"created_at"` // Unix milliseconds UTC
}

// fromEntTraderFill converts ent.TraderFill to store.TraderFill
func fromEntTraderFill(f *ent.TraderFill) *TraderFill {
	if f == nil {
		return nil
	}
	return &TraderFill{
		ID:              f.ID,
		TraderID:        f.TraderID,
		ExchangeID:      f.ExchangeID,
		ExchangeType:    f.ExchangeType,
		OrderID:         f.OrderID,
		ExchangeOrderID: f.ExchangeOrderID,
		ExchangeTradeID: f.ExchangeTradeID,
		Symbol:          f.Symbol,
		Side:            f.Side,
		Price:           f.Price,
		Quantity:        f.Quantity,
		QuoteQuantity:   f.QuoteQuantity,
		Commission:      f.Commission,
		CommissionAsset: f.CommissionAsset,
		RealizedPnL:     f.RealizedPnl,
		IsMaker:         f.IsMaker,
		CreatedAt:       f.CreatedAt,
	}
}

// OrderStore order storage
type OrderStore struct {
	ec *ent.Client
}

// NewOrderStore creates order storage instance
func NewOrderStore() *OrderStore {
	return &OrderStore{}
}

// InitTables initializes order tables
func (s *OrderStore) InitTables() error {
	return nil
}

// CreateOrder creates order record
func (s *OrderStore) CreateOrder(order *TraderOrder) error {
	// Check if order already exists
	existing, err := s.GetOrderByExchangeID(order.ExchangeID, order.ExchangeOrderID)
	if err != nil {
		return fmt.Errorf("failed to check existing order: %w", err)
	}
	if existing != nil {
		order.ID = existing.ID
		order.CreatedAt = existing.CreatedAt
		order.UpdatedAt = existing.UpdatedAt
		return nil
	}

	ctx := context.Background()
	created, err := s.ec.TraderOrder.Create().
		SetTraderID(order.TraderID).
		SetExchangeID(order.ExchangeID).
		SetExchangeType(order.ExchangeType).
		SetExchangeOrderID(order.ExchangeOrderID).
		SetClientOrderID(order.ClientOrderID).
		SetSymbol(order.Symbol).
		SetSide(order.Side).
		SetPositionSide(order.PositionSide).
		SetTypeName(order.Type).
		SetTimeInForce(order.TimeInForce).
		SetQuantity(order.Quantity).
		SetPrice(order.Price).
		SetStopPrice(order.StopPrice).
		SetStatus(order.Status).
		SetFilledQuantity(order.FilledQuantity).
		SetAvgFillPrice(order.AvgFillPrice).
		SetCommission(order.Commission).
		SetCommissionAsset(order.CommissionAsset).
		SetLeverage(order.Leverage).
		SetReduceOnly(order.ReduceOnly).
		SetClosePosition(order.ClosePosition).
		SetWorkingType(order.WorkingType).
		SetPriceProtect(order.PriceProtect).
		SetOrderAction(order.OrderAction).
		SetRelatedPositionID(order.RelatedPositionID).
		SetCreatedAt(order.CreatedAt).
		SetUpdatedAt(order.UpdatedAt).
		SetFilledAt(order.FilledAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	order.ID = created.ID
	return nil
}

// UpdateOrderStatus updates order status
func (s *OrderStore) UpdateOrderStatus(id int64, status string, filledQty, avgPrice, commission float64) error {
	now := time.Now().UTC().UnixMilli()
	update := s.ec.TraderOrder.Update().
		Where(entorder.ID(id)).
		SetStatus(status).
		SetFilledQuantity(filledQty).
		SetAvgFillPrice(avgPrice).
		SetCommission(commission).
		SetUpdatedAt(now)

	if status == "FILLED" {
		update.SetFilledAt(now)
	}

	_, err := update.Save(context.Background())
	return err
}

// CreateFill creates fill record
func (s *OrderStore) CreateFill(fill *TraderFill) error {
	// Check if fill already exists
	existing, err := s.GetFillByExchangeTradeID(fill.ExchangeID, fill.ExchangeTradeID)
	if err != nil {
		return fmt.Errorf("failed to check existing fill: %w", err)
	}
	if existing != nil {
		fill.ID = existing.ID
		fill.CreatedAt = existing.CreatedAt
		return nil
	}

	ctx := context.Background()
	created, err := s.ec.TraderFill.Create().
		SetTraderID(fill.TraderID).
		SetExchangeID(fill.ExchangeID).
		SetExchangeType(fill.ExchangeType).
		SetOrderID(fill.OrderID).
		SetExchangeOrderID(fill.ExchangeOrderID).
		SetExchangeTradeID(fill.ExchangeTradeID).
		SetSymbol(fill.Symbol).
		SetSide(fill.Side).
		SetPrice(fill.Price).
		SetQuantity(fill.Quantity).
		SetQuoteQuantity(fill.QuoteQuantity).
		SetCommission(fill.Commission).
		SetCommissionAsset(fill.CommissionAsset).
		SetRealizedPnl(fill.RealizedPnL).
		SetIsMaker(fill.IsMaker).
		SetCreatedAt(fill.CreatedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create fill: %w", err)
	}
	fill.ID = created.ID
	return nil
}

// GetFillByExchangeTradeID gets fill by exchange trade ID
func (s *OrderStore) GetFillByExchangeTradeID(exchangeID, exchangeTradeID string) (*TraderFill, error) {
	fill, err := s.ec.TraderFill.Query().
		Where(entfill.ExchangeID(exchangeID), entfill.ExchangeTradeID(exchangeTradeID)).
		Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get fill: %w", err)
	}
	return fromEntTraderFill(fill), nil
}

// GetOrderByExchangeID gets order by exchange order ID
func (s *OrderStore) GetOrderByExchangeID(exchangeID, exchangeOrderID string) (*TraderOrder, error) {
	order, err := s.ec.TraderOrder.Query().
		Where(entorder.ExchangeID(exchangeID), entorder.ExchangeOrderID(exchangeOrderID)).
		Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return fromEntTraderOrder(order), nil
}

// GetTraderOrders gets trader's order list
func (s *OrderStore) GetTraderOrders(traderID string, limit int) ([]*TraderOrder, error) {
	orders, err := s.ec.TraderOrder.Query().
		Where(entorder.TraderID(traderID)).
		Order(ent.Desc(entorder.FieldCreatedAt)).
		Limit(limit).
		All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	result := make([]*TraderOrder, len(orders))
	for i, o := range orders {
		result[i] = fromEntTraderOrder(o)
	}
	return result, nil
}

// GetTraderOrdersFiltered gets trader's order list with optional symbol and status filters
func (s *OrderStore) GetTraderOrdersFiltered(traderID string, symbol string, status string, limit int) ([]*TraderOrder, error) {
	query := s.ec.TraderOrder.Query().Where(entorder.TraderID(traderID))
	if symbol != "" {
		query = query.Where(entorder.Symbol(symbol))
	}
	if status != "" {
		query = query.Where(entorder.Status(status))
	}
	orders, err := query.Order(ent.Desc(entorder.FieldCreatedAt)).Limit(limit).All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	result := make([]*TraderOrder, len(orders))
	for i, o := range orders {
		result[i] = fromEntTraderOrder(o)
	}
	return result, nil
}

// GetOrderFills gets order's fill records
func (s *OrderStore) GetOrderFills(orderID int64) ([]*TraderFill, error) {
	fills, err := s.ec.TraderFill.Query().
		Where(entfill.OrderID(orderID)).
		Order(ent.Asc(entfill.FieldCreatedAt)).
		All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query fills: %w", err)
	}
	result := make([]*TraderFill, len(fills))
	for i, f := range fills {
		result[i] = fromEntTraderFill(f)
	}
	return result, nil
}

// GetTraderOrderStats gets trader's order statistics
func (s *OrderStore) GetTraderOrderStats(traderID string) (map[string]interface{}, error) {
	orders, err := s.ec.TraderOrder.Query().
		Where(entorder.TraderID(traderID)).
		All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get order stats: %w", err)
	}

	var totalOrders, filledOrders, canceledOrders int
	var totalCommission, totalVolume float64
	for _, o := range orders {
		totalOrders++
		totalCommission += o.Commission
		switch o.Status {
		case "FILLED":
			filledOrders++
			totalVolume += o.FilledQuantity * o.AvgFillPrice
		case "CANCELED":
			canceledOrders++
		}
	}

	return map[string]interface{}{
		"total_orders":     totalOrders,
		"filled_orders":    filledOrders,
		"canceled_orders":  canceledOrders,
		"total_commission": totalCommission,
		"total_volume":     totalVolume,
	}, nil
}

// CleanupDuplicateOrders cleans up duplicate order records
func (s *OrderStore) CleanupDuplicateOrders() (int, error) {
	orders, err := s.ec.TraderOrder.Query().
		Order(ent.Asc(entorder.FieldID)).
		All(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to query orders: %w", err)
	}

	seen := make(map[string]bool)
	var toDelete []int64
	for _, o := range orders {
		key := o.ExchangeID + "|" + o.ExchangeOrderID
		if seen[key] {
			toDelete = append(toDelete, o.ID)
		}
		seen[key] = true
	}

	if len(toDelete) == 0 {
		return 0, nil
	}

	deleted, err := s.ec.TraderOrder.Delete().
		Where(entorder.IDIn(toDelete...)).
		Exec(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to delete duplicate orders: %w", err)
	}
	return deleted, nil
}

// CleanupDuplicateFills cleans up duplicate fill records
func (s *OrderStore) CleanupDuplicateFills() (int, error) {
	fills, err := s.ec.TraderFill.Query().
		Order(ent.Asc(entfill.FieldID)).
		All(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to query fills: %w", err)
	}

	seen := make(map[string]bool)
	var toDelete []int64
	for _, f := range fills {
		key := f.ExchangeID + "|" + f.ExchangeTradeID
		if seen[key] {
			toDelete = append(toDelete, f.ID)
		}
		seen[key] = true
	}

	if len(toDelete) == 0 {
		return 0, nil
	}

	deleted, err := s.ec.TraderFill.Delete().
		Where(entfill.IDIn(toDelete...)).
		Exec(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to delete duplicate fills: %w", err)
	}
	return deleted, nil
}

// GetDuplicateOrdersCount gets duplicate orders count
func (s *OrderStore) GetDuplicateOrdersCount() (int, error) {
	orders, err := s.ec.TraderOrder.Query().
		All(context.Background())
	if err != nil {
		return 0, err
	}

	seen := make(map[string]bool)
	duplicates := 0
	for _, o := range orders {
		key := o.ExchangeID + "|" + o.ExchangeOrderID
		if seen[key] {
			duplicates++
		}
		seen[key] = true
	}
	return duplicates, nil
}

// GetDuplicateFillsCount gets duplicate fills count
func (s *OrderStore) GetDuplicateFillsCount() (int, error) {
	fills, err := s.ec.TraderFill.Query().
		All(context.Background())
	if err != nil {
		return 0, err
	}

	seen := make(map[string]bool)
	duplicates := 0
	for _, f := range fills {
		key := f.ExchangeID + "|" + f.ExchangeTradeID
		if seen[key] {
			duplicates++
		}
		seen[key] = true
	}
	return duplicates, nil
}

// GetMaxTradeIDsByExchange returns max trade ID for each symbol for a given exchange
func (s *OrderStore) GetMaxTradeIDsByExchange(exchangeID string) (map[string]int64, error) {
	fills, err := s.ec.TraderFill.Query().
		Where(entfill.ExchangeID(exchangeID), entfill.ExchangeTradeIDNEQ("")).
		All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query trade IDs: %w", err)
	}

	// Find max trade ID per symbol in Go (handles 64-bit integers properly)
	result := make(map[string]int64)
	for _, f := range fills {
		tradeID, err := strconv.ParseInt(f.ExchangeTradeID, 10, 64)
		if err != nil {
			continue // Skip non-numeric trade IDs
		}
		if tradeID > result[f.Symbol] {
			result[f.Symbol] = tradeID
		}
	}
	return result, nil
}

// GetLastFillTimeByExchange returns the most recent fill time (Unix ms) for a given exchange
// Used to recover sync state after service restart
func (s *OrderStore) GetLastFillTimeByExchange(exchangeID string) (int64, error) {
	fill, err := s.ec.TraderFill.Query().
		Where(entfill.ExchangeID(exchangeID)).
		Order(ent.Desc(entfill.FieldCreatedAt)).
		First(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	return fill.CreatedAt, nil
}

// GetRecentFillSymbolsByExchange returns distinct symbols with fills since given time (Unix ms)
func (s *OrderStore) GetRecentFillSymbolsByExchange(exchangeID string, sinceMs int64) ([]string, error) {
	fills, err := s.ec.TraderFill.Query().
		Where(entfill.ExchangeID(exchangeID), entfill.CreatedAtGTE(sinceMs)).
		All(context.Background())
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var symbols []string
	for _, f := range fills {
		if !seen[f.Symbol] {
			symbols = append(symbols, f.Symbol)
			seen[f.Symbol] = true
		}
	}
	return symbols, nil
}
