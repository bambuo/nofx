package kernel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nofx/logger"
	"nofx/market"
	"nofx/mcp"
	"nofx/provider/binance"
	"nofx/provider/nofxos"
	"nofx/security"
	"nofx/store"
	"regexp"
	"strings"
	"time"
)

// ============================================================================
// Pre-compiled regular expressions (performance optimization)
// ============================================================================

var (
	// Safe regex: precisely match ```json code blocks
	reJSONFence      = regexp.MustCompile(`(?is)` + "```json\\s*(\\[\\s*\\{.*?\\}\\s*\\])\\s*```")
	reJSONArray      = regexp.MustCompile(`(?is)\[\s*\{.*?\}\s*\]`)
	reArrayHead      = regexp.MustCompile(`^\[\s*\{`)
	reArrayOpenSpace = regexp.MustCompile(`^\[\s+\{`)
	reInvisibleRunes = regexp.MustCompile("[\u200B\u200C\u200D\uFEFF]")

	// XML tag extraction (supports any characters in reasoning chain)
	reReasoningTag = regexp.MustCompile(`(?s)<reasoning>(.*?)</reasoning>`)
	reDecisionTag  = regexp.MustCompile(`(?s)<decision>(.*?)</decision>`)
)

const (
	decisionRequestTemperature = 0.2
	decisionRequestTopP        = 0.2
	decisionRequestMaxTokens   = 1800
	defaultMinRiskRewardRatio  = 3.0
	coinProviderNofxOS         = "nofxos"
	coinProviderBinance        = "binance"
	marketDataQuarantineTTL    = 15 * time.Minute
)

// ============================================================================
// Type Definitions
// ============================================================================

// PositionInfo position information
type PositionInfo struct {
	Symbol           string  `json:"symbol"`
	Side             string  `json:"side"` // "long" or "short"
	EntryPrice       float64 `json:"entry_price"`
	MarkPrice        float64 `json:"mark_price"`
	Quantity         float64 `json:"quantity"`
	Leverage         int     `json:"leverage"`
	UnrealizedPnL    float64 `json:"unrealized_pnl"`
	UnrealizedPnLPct float64 `json:"unrealized_pnl_pct"`
	PeakPnLPct       float64 `json:"peak_pnl_pct"` // Historical peak profit percentage
	LiquidationPrice float64 `json:"liquidation_price"`
	MarginUsed       float64 `json:"margin_used"`
	UpdateTime       int64   `json:"update_time"` // Position update timestamp (milliseconds)
}

// AccountInfo account information
type AccountInfo struct {
	TotalEquity      float64 `json:"total_equity"`      // Account equity
	AvailableBalance float64 `json:"available_balance"` // Available balance
	UnrealizedPnL    float64 `json:"unrealized_pnl"`    // Unrealized profit/loss
	TotalPnL         float64 `json:"total_pnl"`         // Total profit/loss
	TotalPnLPct      float64 `json:"total_pnl_pct"`     // Total profit/loss percentage
	MarginUsed       float64 `json:"margin_used"`       // Used margin
	MarginUsedPct    float64 `json:"margin_used_pct"`   // Margin usage rate
	PositionCount    int     `json:"position_count"`    // Number of positions
}

// CandidateCoin candidate coin (from coin pool)
type CandidateCoin struct {
	Symbol  string   `json:"symbol"`
	Sources []string `json:"sources"` // Sources: "ai500", "oi_top", "binance_score", etc.
}

// DataQualitySummary describes which symbols were actually analyzable this round.
type DataQualitySummary struct {
	CandidateCount     int                    `json:"candidate_count"`
	AnalyzableCount    int                    `json:"analyzable_count"`
	MarketDataFailures []MarketDataFailure    `json:"market_data_failures,omitempty"`
	LiquidityFiltered  []LiquidityFilterIssue `json:"liquidity_filtered,omitempty"`
}

// MarketDataFailure records symbols that could not be analyzed due to missing market data.
type MarketDataFailure struct {
	Symbol string `json:"symbol"`
	Reason string `json:"reason"`
}

// LiquidityFilterIssue records symbols excluded by the OI liquidity gate.
type LiquidityFilterIssue struct {
	Symbol            string  `json:"symbol"`
	OIValueMillions   float64 `json:"oi_value_millions"`
	ThresholdMillions float64 `json:"threshold_millions"`
}

// OITopData open interest increase data (for AI decision reference)
type OITopData struct {
	Rank              int     // OI increase ranking
	OIDeltaPercent    float64 // Open interest change percentage (1 hour)
	OIDeltaValue      float64 // Open interest change value
	PriceDeltaPercent float64 // Price change percentage
}

// TradingStats trading statistics (for AI input)
type TradingStats struct {
	TotalTrades    int     `json:"total_trades"`     // Total number of trades (closed)
	WinRate        float64 `json:"win_rate"`         // Win rate (%)
	ProfitFactor   float64 `json:"profit_factor"`    // Profit factor
	SharpeRatio    float64 `json:"sharpe_ratio"`     // Sharpe ratio
	TotalPnL       float64 `json:"total_pnl"`        // Total profit/loss
	AvgWin         float64 `json:"avg_win"`          // Average win
	AvgLoss        float64 `json:"avg_loss"`         // Average loss
	MaxDrawdownPct float64 `json:"max_drawdown_pct"` // Maximum drawdown (%)
}

// RecentOrder recently completed order (for AI input)
type RecentOrder struct {
	Symbol       string  `json:"symbol"`        // Trading pair
	Side         string  `json:"side"`          // long/short
	EntryPrice   float64 `json:"entry_price"`   // Entry price
	ExitPrice    float64 `json:"exit_price"`    // Exit price
	RealizedPnL  float64 `json:"realized_pnl"`  // Realized profit/loss
	PnLPct       float64 `json:"pnl_pct"`       // Profit/loss percentage
	EntryTime    string  `json:"entry_time"`    // Entry time
	ExitTime     string  `json:"exit_time"`     // Exit time
	HoldDuration string  `json:"hold_duration"` // Hold duration, e.g. "2h30m"
}

// Context trading context (complete information passed to AI)
type Context struct {
	CurrentTime        string                             `json:"current_time"`
	RuntimeMinutes     int                                `json:"runtime_minutes"`
	CallCount          int                                `json:"call_count"`
	Account            AccountInfo                        `json:"account"`
	Positions          []PositionInfo                     `json:"positions"`
	CandidateCoins     []CandidateCoin                    `json:"candidate_coins"`
	PromptVariant      string                             `json:"prompt_variant,omitempty"`
	TradingStats       *TradingStats                      `json:"trading_stats,omitempty"`
	RecentOrders       []RecentOrder                      `json:"recent_orders,omitempty"`
	DataQuality        *DataQualitySummary                `json:"data_quality,omitempty"`
	MarketDataMap      map[string]*market.Data            `json:"-"`
	MultiTFMarket      map[string]map[string]*market.Data `json:"-"`
	OITopDataMap       map[string]*OITopData              `json:"-"`
	QuantDataMap       map[string]*QuantData              `json:"-"`
	OIRankingData      *nofxos.OIRankingData              `json:"-"` // Market-wide OI ranking data
	NetFlowRankingData *nofxos.NetFlowRankingData         `json:"-"` // Market-wide fund flow ranking data
	PriceRankingData   *nofxos.PriceRankingData           `json:"-"` // Market-wide price gainers/losers
	BTCETHLeverage     int                                `json:"-"`
	AltcoinLeverage    int                                `json:"-"`
	Timeframes         []string                           `json:"-"`
}

// Decision AI trading decision
type Decision struct {
	Symbol string `json:"symbol"`
	Action string `json:"action"` // Standard: "open_long", "open_short", "close_long", "close_short", "hold", "wait"
	// Grid actions: "place_buy_limit", "place_sell_limit", "cancel_order", "cancel_all_orders", "pause_grid", "resume_grid", "adjust_grid"

	// Opening position parameters
	Leverage        int     `json:"leverage,omitempty"`
	PositionSizeUSD float64 `json:"position_size_usd,omitempty"`
	StopLoss        float64 `json:"stop_loss,omitempty"`
	TakeProfit      float64 `json:"take_profit,omitempty"`

	// Grid trading parameters
	Price      float64 `json:"price,omitempty"`       // Limit order price (for grid)
	Quantity   float64 `json:"quantity,omitempty"`    // Order quantity (for grid)
	LevelIndex int     `json:"level_index,omitempty"` // Grid level index
	OrderID    string  `json:"order_id,omitempty"`    // Order ID (for cancel)

	// Common parameters
	Confidence int     `json:"confidence,omitempty"` // Confidence level (0-100)
	RiskUSD    float64 `json:"risk_usd,omitempty"`   // Maximum USD risk
	Reasoning  string  `json:"reasoning"`
}

// FullDecision AI's complete decision (including chain of thought)
type FullDecision struct {
	SystemPrompt        string     `json:"system_prompt"`
	UserPrompt          string     `json:"user_prompt"`
	CoTTrace            string     `json:"cot_trace"`
	Decisions           []Decision `json:"decisions"`
	RawResponse         string     `json:"raw_response"`
	Timestamp           time.Time  `json:"timestamp"`
	AIRequestDurationMs int64      `json:"ai_request_duration_ms,omitempty"`
}

// QuantData quantitative data structure (fund flow, position changes, price changes)
type QuantData struct {
	Symbol      string             `json:"symbol"`
	Price       float64            `json:"price"`
	Netflow     *NetflowData       `json:"netflow,omitempty"`
	OI          map[string]*OIData `json:"oi,omitempty"`
	PriceChange map[string]float64 `json:"price_change,omitempty"`
}

type NetflowData struct {
	Institution *FlowTypeData `json:"institution,omitempty"`
	Personal    *FlowTypeData `json:"personal,omitempty"`
}

type FlowTypeData struct {
	Future map[string]float64 `json:"future,omitempty"`
	Spot   map[string]float64 `json:"spot,omitempty"`
}

type OIData struct {
	CurrentOI float64                 `json:"current_oi"`
	Delta     map[string]*OIDeltaData `json:"delta,omitempty"`
}

type OIDeltaData struct {
	OIDelta        float64 `json:"oi_delta"`
	OIDeltaValue   float64 `json:"oi_delta_value"`
	OIDeltaPercent float64 `json:"oi_delta_percent"`
}

// ============================================================================
// StrategyEngine - Core Strategy Execution Engine
// ============================================================================

// StrategyEngine strategy execution engine
type StrategyEngine struct {
	config             *store.StrategyConfig
	nofxosClient       *nofxos.Client
	binanceFutures     *binance.FuturesClient
	quarantinedSymbols map[string]time.Time
}

// NewStrategyEngine creates strategy execution engine
func NewStrategyEngine(config *store.StrategyConfig) *StrategyEngine {
	// Create NofxOS client with API key from config
	apiKey := config.Indicators.NofxOSAPIKey
	if apiKey == "" {
		apiKey = nofxos.DefaultAuthKey
	}
	client := nofxos.NewClient(nofxos.DefaultBaseURL, apiKey)

	return &StrategyEngine{
		config:             config,
		nofxosClient:       client,
		binanceFutures:     binance.NewFuturesClient(binance.DefaultFuturesBaseURL),
		quarantinedSymbols: make(map[string]time.Time),
	}
}

// GetRiskControlConfig gets risk control configuration
func (e *StrategyEngine) GetRiskControlConfig() store.RiskControlConfig {
	return e.config.RiskControl
}

// GetLanguage returns the language from config or falls back to auto-detection
func (e *StrategyEngine) GetLanguage() Language {
	switch e.config.Language {
	case "zh":
		return LangChinese
	case "en":
		return LangEnglish
	default:
		// Fall back to auto-detection from prompt content for backward compatibility
		return detectLanguage(e.config.PromptSections.RoleDefinition)
	}
}

// GetConfig gets complete strategy configuration
func (e *StrategyEngine) GetConfig() *store.StrategyConfig {
	return e.config
}

// ============================================================================
// Entry Functions - Main API
// ============================================================================

// GetFullDecision gets AI's complete trading decision (batch analysis of all coins and positions)
// Uses default strategy configuration - for production use GetFullDecisionWithStrategy with explicit config
func GetFullDecision(ctx *Context, mcpClient mcp.AIClient) (*FullDecision, error) {
	defaultConfig := store.GetDefaultStrategyConfig("en")
	engine := NewStrategyEngine(&defaultConfig)
	return GetFullDecisionWithStrategy(ctx, mcpClient, engine, "")
}

// GetFullDecisionWithStrategy uses StrategyEngine to get AI decision (unified prompt generation)
func GetFullDecisionWithStrategy(ctx *Context, mcpClient mcp.AIClient, engine *StrategyEngine, variant string) (*FullDecision, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if engine == nil {
		defaultConfig := store.GetDefaultStrategyConfig("en")
		engine = NewStrategyEngine(&defaultConfig)
	}

	// 1. Fetch market data using strategy config
	if len(ctx.MarketDataMap) == 0 {
		if err := fetchMarketDataWithStrategy(ctx, engine); err != nil {
			return nil, fmt.Errorf("failed to fetch market data: %w", err)
		}
	}

	// Ensure OITopDataMap is initialized
	if ctx.OITopDataMap == nil {
		ctx.OITopDataMap = make(map[string]*OITopData)
		oiPositions, err := engine.nofxosClient.GetOITopPositions()
		if err == nil {
			for _, pos := range oiPositions {
				ctx.OITopDataMap[pos.Symbol] = &OITopData{
					Rank:              pos.Rank,
					OIDeltaPercent:    pos.OIDeltaPercent,
					OIDeltaValue:      pos.OIDeltaValue,
					PriceDeltaPercent: pos.PriceDeltaPercent,
				}
			}
		}
	}

	// 2. Build System Prompt using strategy engine
	riskConfig := engine.GetRiskControlConfig()
	systemPrompt := engine.BuildSystemPrompt(ctx.Account.TotalEquity, variant)

	// 3. Build User Prompt using strategy engine
	userPrompt := engine.BuildUserPrompt(ctx)

	// 4. Call AI API
	aiCallStart := time.Now()
	aiRequest, err := buildDecisionRequest(systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to build AI request: %w", err)
	}
	aiResponse, err := mcpClient.CallWithRequest(aiRequest)
	aiCallDuration := time.Since(aiCallStart)
	if err != nil {
		return nil, fmt.Errorf("AI API call failed: %w", err)
	}

	// 5. Parse AI response
	decision, err := parseFullDecisionResponse(
		aiResponse,
		ctx.Account.TotalEquity,
		riskConfig.BTCETHMaxLeverage,
		riskConfig.AltcoinMaxLeverage,
		riskConfig.BTCETHMaxPositionValueRatio,
		riskConfig.AltcoinMaxPositionValueRatio,
		riskConfig.MinRiskRewardRatio,
	)

	if decision != nil {
		decision.Timestamp = time.Now()
		decision.SystemPrompt = systemPrompt
		decision.UserPrompt = userPrompt
		decision.AIRequestDurationMs = aiCallDuration.Milliseconds()
		decision.RawResponse = aiResponse
	}

	if err != nil {
		return decision, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return decision, nil
}

func buildDecisionRequest(systemPrompt, userPrompt string) (*mcp.Request, error) {
	return mcp.NewRequestBuilder().
		WithSystemPrompt(systemPrompt).
		WithUserPrompt(userPrompt).
		WithTemperature(decisionRequestTemperature).
		WithTopP(decisionRequestTopP).
		WithMaxTokens(decisionRequestMaxTokens).
		Build()
}

// ============================================================================
// Market Data Fetching
// ============================================================================

// fetchMarketDataWithStrategy fetches market data using strategy config (multiple timeframes)
func fetchMarketDataWithStrategy(ctx *Context, engine *StrategyEngine) error {
	config := engine.GetConfig()
	ctx.MarketDataMap = make(map[string]*market.Data)
	ctx.DataQuality = &DataQualitySummary{
		CandidateCount: len(ctx.CandidateCoins),
	}

	timeframes := config.Indicators.Klines.SelectedTimeframes
	primaryTimeframe := config.Indicators.Klines.PrimaryTimeframe
	klineCount := config.Indicators.Klines.PrimaryCount

	// Compatible with old configuration
	if len(timeframes) == 0 {
		if primaryTimeframe != "" {
			timeframes = append(timeframes, primaryTimeframe)
		} else {
			timeframes = append(timeframes, "3m")
		}
		if config.Indicators.Klines.LongerTimeframe != "" {
			timeframes = append(timeframes, config.Indicators.Klines.LongerTimeframe)
		}
	}
	if primaryTimeframe == "" {
		primaryTimeframe = timeframes[0]
	}
	if klineCount <= 0 {
		klineCount = 30
	}

	logger.Infof("📊 Strategy timeframes: %v, Primary: %s, Kline count: %d", timeframes, primaryTimeframe, klineCount)

	// 1. First fetch data for position coins (must fetch)
	for _, pos := range ctx.Positions {
		data, err := market.GetWithTimeframes(pos.Symbol, timeframes, primaryTimeframe, klineCount)
		if err != nil {
			logger.Infof("⚠️  Failed to fetch market data for position %s: %v", pos.Symbol, err)
			ctx.DataQuality.MarketDataFailures = append(ctx.DataQuality.MarketDataFailures, MarketDataFailure{
				Symbol: pos.Symbol,
				Reason: err.Error(),
			})
			continue
		}
		ctx.MarketDataMap[pos.Symbol] = data
	}

	// 2. Fetch data for all candidate coins
	positionSymbols := make(map[string]bool)
	for _, pos := range ctx.Positions {
		positionSymbols[pos.Symbol] = true
	}

	const minOIThresholdMillions = 15.0 // 15M USD minimum open interest value

	for _, coin := range ctx.CandidateCoins {
		if _, exists := ctx.MarketDataMap[coin.Symbol]; exists {
			continue
		}

		if engine.isSymbolQuarantined(coin.Symbol) {
			reason := fmt.Sprintf("temporarily quarantined after recent market data failure (%s)", marketDataQuarantineTTL)
			logger.Infof("⏳ %s %s, skipping coin", coin.Symbol, reason)
			ctx.DataQuality.MarketDataFailures = append(ctx.DataQuality.MarketDataFailures, MarketDataFailure{
				Symbol: coin.Symbol,
				Reason: reason,
			})
			continue
		}

		data, err := market.GetWithTimeframes(coin.Symbol, timeframes, primaryTimeframe, klineCount)
		if err != nil {
			logger.Infof("⚠️  Failed to fetch market data for %s: %v", coin.Symbol, err)
			if shouldQuarantineMarketDataError(err) {
				engine.quarantineSymbol(coin.Symbol)
			}
			ctx.DataQuality.MarketDataFailures = append(ctx.DataQuality.MarketDataFailures, MarketDataFailure{
				Symbol: coin.Symbol,
				Reason: err.Error(),
			})
			continue
		}

		// Liquidity filter (skip for xyz dex assets - they don't have OI data from Binance)
		isExistingPosition := positionSymbols[coin.Symbol]
		isXyzAsset := market.IsXyzDexAsset(coin.Symbol)
		if !isExistingPosition && !isXyzAsset && data.OpenInterest != nil && data.CurrentPrice > 0 {
			oiValue := data.OpenInterest.Latest * data.CurrentPrice
			oiValueInMillions := oiValue / 1_000_000
			if oiValueInMillions < minOIThresholdMillions {
				logger.Infof("⚠️  %s OI value too low (%.2fM USD < %.1fM), skipping coin",
					coin.Symbol, oiValueInMillions, minOIThresholdMillions)
				ctx.DataQuality.LiquidityFiltered = append(ctx.DataQuality.LiquidityFiltered, LiquidityFilterIssue{
					Symbol:            coin.Symbol,
					OIValueMillions:   oiValueInMillions,
					ThresholdMillions: minOIThresholdMillions,
				})
				continue
			}
		}

		ctx.MarketDataMap[coin.Symbol] = data
	}

	ctx.DataQuality.AnalyzableCount = len(ctx.MarketDataMap)
	logger.Infof("📊 Successfully fetched multi-timeframe market data for %d coins", len(ctx.MarketDataMap))
	return nil
}

func (e *StrategyEngine) isSymbolQuarantined(symbol string) bool {
	if len(e.quarantinedSymbols) == 0 {
		return false
	}

	until, ok := e.quarantinedSymbols[symbol]
	if !ok {
		return false
	}
	if time.Now().Before(until) {
		return true
	}
	delete(e.quarantinedSymbols, symbol)
	return false
}

func (e *StrategyEngine) quarantineSymbol(symbol string) {
	if e.quarantinedSymbols == nil {
		e.quarantinedSymbols = make(map[string]time.Time)
	}
	e.quarantinedSymbols[symbol] = time.Now().Add(marketDataQuarantineTTL)
}

func shouldQuarantineMarketDataError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "stale") ||
		strings.Contains(msg, "k-line data is empty") ||
		strings.Contains(msg, "fallback returned empty")
}

// ============================================================================
// Candidate Coins
// ============================================================================

// GetCandidateCoins gets candidate coins based on strategy configuration
func (e *StrategyEngine) GetCandidateCoins() ([]CandidateCoin, error) {
	var candidates []CandidateCoin
	symbolSources := make(map[string][]string)

	coinSource := e.config.CoinSource

	switch coinSource.SourceType {
	case "static":
		for _, symbol := range coinSource.StaticCoins {
			symbol = market.Normalize(symbol)
			candidates = append(candidates, CandidateCoin{
				Symbol:  symbol,
				Sources: []string{"static"},
			})
		}

		return e.filterExcludedCoins(candidates), nil

	case "ai500":
		// 检查 use_ai500 标志，如果为 false 则回退到静态币种
		if !coinSource.UseAI500 {
			logger.Infof("⚠️  source_type is 'ai500' but use_ai500 is false, falling back to static coins")
			for _, symbol := range coinSource.StaticCoins {
				symbol = market.Normalize(symbol)
				candidates = append(candidates, CandidateCoin{
					Symbol:  symbol,
					Sources: []string{"static"},
				})
			}
			return e.filterExcludedCoins(candidates), nil
		}
		coins, err := e.getAI500Coins(coinSource.AI500Limit)
		if err != nil {
			return nil, err
		}
		// 空列表是正常情况，直接返回
		return e.filterExcludedCoins(coins), nil

	case "oi_top":
		// 检查 use_oi_top 标志，如果为 false 则回退到静态币种
		if !coinSource.UseOITop {
			logger.Infof("⚠️  source_type is 'oi_top' but use_oi_top is false, falling back to static coins")
			for _, symbol := range coinSource.StaticCoins {
				symbol = market.Normalize(symbol)
				candidates = append(candidates, CandidateCoin{
					Symbol:  symbol,
					Sources: []string{"static"},
				})
			}
			return e.filterExcludedCoins(candidates), nil
		}
		coins, err := e.getOITopCoins(coinSource.OITopLimit)
		if err != nil {
			return nil, err
		}
		// 空列表是正常情况，直接返回
		return e.filterExcludedCoins(coins), nil

	case "oi_low":
		// 持仓减少，适合做空
		if !coinSource.UseOILow {
			logger.Infof("⚠️  source_type is 'oi_low' but use_oi_low is false, falling back to static coins")
			for _, symbol := range coinSource.StaticCoins {
				symbol = market.Normalize(symbol)
				candidates = append(candidates, CandidateCoin{
					Symbol:  symbol,
					Sources: []string{"static"},
				})
			}
			return e.filterExcludedCoins(candidates), nil
		}
		coins, err := e.getOILowCoins(coinSource.OILowLimit)
		if err != nil {
			return nil, err
		}
		// 空列表是正常情况，直接返回
		return e.filterExcludedCoins(coins), nil

	case "mixed":
		if coinSource.UseAI500 {
			poolCoins, err := e.getAI500Coins(coinSource.AI500Limit)
			if err != nil {
				logger.Infof("⚠️  Failed to get score top coins: %v", err)
			} else {
				for _, coin := range poolCoins {
					symbolSources[coin.Symbol] = append(symbolSources[coin.Symbol], coin.Sources...)
				}
			}
		}

		if coinSource.UseOITop {
			oiCoins, err := e.getOITopCoins(coinSource.OITopLimit)
			if err != nil {
				logger.Infof("⚠️  Failed to get OI increase coins: %v", err)
			} else {
				for _, coin := range oiCoins {
					symbolSources[coin.Symbol] = append(symbolSources[coin.Symbol], coin.Sources...)
				}
			}
		}

		if coinSource.UseOILow {
			oiLowCoins, err := e.getOILowCoins(coinSource.OILowLimit)
			if err != nil {
				logger.Infof("⚠️  Failed to get OI decrease coins: %v", err)
			} else {
				for _, coin := range oiLowCoins {
					symbolSources[coin.Symbol] = append(symbolSources[coin.Symbol], coin.Sources...)
				}
			}
		}

		for _, symbol := range coinSource.StaticCoins {
			symbol = market.Normalize(symbol)
			if _, exists := symbolSources[symbol]; !exists {
				symbolSources[symbol] = []string{"static"}
			} else {
				symbolSources[symbol] = append(symbolSources[symbol], "static")
			}
		}

		for symbol, sources := range symbolSources {
			candidates = append(candidates, CandidateCoin{
				Symbol:  symbol,
				Sources: sources,
			})
		}
		return e.filterExcludedCoins(candidates), nil

	default:
		return nil, fmt.Errorf("unknown coin source type: %s", coinSource.SourceType)
	}
}

// filterExcludedCoins removes excluded coins from the candidates list
func (e *StrategyEngine) filterExcludedCoins(candidates []CandidateCoin) []CandidateCoin {
	if len(e.config.CoinSource.ExcludedCoins) == 0 {
		return candidates
	}

	// Build excluded set for O(1) lookup
	excluded := make(map[string]bool)
	for _, coin := range e.config.CoinSource.ExcludedCoins {
		normalized := market.Normalize(coin)
		excluded[normalized] = true
	}

	// Filter out excluded coins
	filtered := make([]CandidateCoin, 0, len(candidates))
	for _, c := range candidates {
		if !excluded[c.Symbol] {
			filtered = append(filtered, c)
		} else {
			logger.Infof("🚫 Excluded coin: %s", c.Symbol)
		}
	}

	return filtered
}

func (e *StrategyEngine) getAI500Coins(limit int) ([]CandidateCoin, error) {
	if limit <= 0 {
		limit = 30
	}

	return e.getScoreCoinsFromProviders(limit)
}

func (e *StrategyEngine) getOITopCoins(limit int) ([]CandidateCoin, error) {
	if limit <= 0 {
		limit = 10
	}

	return e.getOIIncreaseCoinsFromProviders(limit)
}

func (e *StrategyEngine) getOILowCoins(limit int) ([]CandidateCoin, error) {
	if limit <= 0 {
		limit = 10
	}

	return e.getOIDecreaseCoinsFromProviders(limit)
}

func (e *StrategyEngine) getScoreCoinsFromProviders(limit int) ([]CandidateCoin, error) {
	var lastErr error
	for _, provider := range e.coinSourceProviders() {
		switch provider {
		case coinProviderNofxOS:
			symbols, err := e.nofxosClient.GetTopRatedCoins(limit)
			if err != nil {
				lastErr = err
				logger.Infof("⚠️  NofxOS score top source unavailable: %v", err)
				continue
			}
			if len(symbols) == 0 {
				logger.Infof("ℹ️  NofxOS score top source returned empty list")
				continue
			}
			return candidateCoinsFromSymbols(symbols, "ai500"), nil
		case coinProviderBinance:
			coins, err := e.getBinanceScoreCoins(limit)
			if err != nil {
				lastErr = err
				logger.Infof("⚠️  Binance score source unavailable: %v", err)
				continue
			}
			if len(coins) == 0 {
				logger.Infof("ℹ️  Binance score source returned empty list")
				continue
			}
			return coins, nil
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return []CandidateCoin{}, nil
}

func (e *StrategyEngine) getOIIncreaseCoinsFromProviders(limit int) ([]CandidateCoin, error) {
	var lastErr error
	for _, provider := range e.coinSourceProviders() {
		switch provider {
		case coinProviderNofxOS:
			positions, err := e.nofxosClient.GetOITopPositions()
			if err != nil {
				lastErr = err
				logger.Infof("⚠️  NofxOS OI top source unavailable: %v", err)
				continue
			}
			if len(positions) == 0 {
				logger.Infof("ℹ️  NofxOS OI top source returned empty list")
				continue
			}
			return candidateCoinsFromOIPositions(positions, limit, "oi_top"), nil
		case coinProviderBinance:
			coins, err := e.getBinanceOITopCoins(limit)
			if err != nil {
				lastErr = err
				logger.Infof("⚠️  Binance OI increase source unavailable: %v", err)
				continue
			}
			if len(coins) == 0 {
				logger.Infof("ℹ️  Binance OI increase source returned empty list")
				continue
			}
			return coins, nil
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return []CandidateCoin{}, nil
}

func (e *StrategyEngine) getOIDecreaseCoinsFromProviders(limit int) ([]CandidateCoin, error) {
	var lastErr error
	for _, provider := range e.coinSourceProviders() {
		switch provider {
		case coinProviderNofxOS:
			positions, err := e.nofxosClient.GetOILowPositions()
			if err != nil {
				lastErr = err
				logger.Infof("⚠️  NofxOS OI low source unavailable: %v", err)
				continue
			}
			if len(positions) == 0 {
				logger.Infof("ℹ️  NofxOS OI low source returned empty list")
				continue
			}
			return candidateCoinsFromOIPositions(positions, limit, "oi_low"), nil
		case coinProviderBinance:
			coins, err := e.getBinanceOILowCoins(limit)
			if err != nil {
				lastErr = err
				logger.Infof("⚠️  Binance OI decrease source unavailable: %v", err)
				continue
			}
			if len(coins) == 0 {
				logger.Infof("ℹ️  Binance OI decrease source returned empty list")
				continue
			}
			return coins, nil
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return []CandidateCoin{}, nil
}

func candidateCoinsFromOIPositions(positions []nofxos.OIPosition, limit int, source string) []CandidateCoin {
	var candidates []CandidateCoin
	for i, pos := range positions {
		if i >= limit {
			break
		}
		symbol := market.Normalize(pos.Symbol)
		candidates = append(candidates, CandidateCoin{
			Symbol:  symbol,
			Sources: []string{source},
		})
	}
	return candidates
}

func (e *StrategyEngine) coinSourceProviders() []string {
	configured := e.config.CoinSource.Providers
	if len(configured) == 0 {
		configured = []string{coinProviderBinance}
	}

	providers := make([]string, 0, len(configured))
	seen := make(map[string]bool, len(configured))
	for _, provider := range configured {
		provider = strings.ToLower(strings.TrimSpace(provider))
		if seen[provider] {
			continue
		}
		switch provider {
		case coinProviderNofxOS:
			if !e.hasCustomNofxOSAPIKey() {
				logger.Infof("ℹ️  NofxOS coin provider skipped because no custom NofxOS API key is configured")
				continue
			}
			seen[provider] = true
			providers = append(providers, provider)
		case coinProviderBinance:
			seen[provider] = true
			providers = append(providers, provider)
		default:
			logger.Infof("⚠️  Unknown coin source provider ignored: %s", provider)
		}
	}
	if len(providers) == 0 {
		return []string{coinProviderBinance}
	}
	return providers
}

func (e *StrategyEngine) hasCustomNofxOSAPIKey() bool {
	apiKey := strings.TrimSpace(e.config.Indicators.NofxOSAPIKey)
	return apiKey != "" && apiKey != nofxos.DefaultAuthKey
}

func (e *StrategyEngine) getBinanceScoreCoins(limit int) ([]CandidateCoin, error) {
	symbols, err := e.binanceFutures.GetTopScoredSymbols(limit)
	if err != nil {
		return nil, err
	}
	return candidateCoinsFromSymbols(symbols, "binance_score"), nil
}

func (e *StrategyEngine) getBinanceOITopCoins(limit int) ([]CandidateCoin, error) {
	symbols, err := e.binanceFutures.GetOIIncreasingSymbols(limit)
	if err != nil {
		return nil, err
	}
	return candidateCoinsFromSymbols(symbols, "binance_oi_top"), nil
}

func (e *StrategyEngine) getBinanceOILowCoins(limit int) ([]CandidateCoin, error) {
	symbols, err := e.binanceFutures.GetOIDecreasingSymbols(limit)
	if err != nil {
		return nil, err
	}
	return candidateCoinsFromSymbols(symbols, "binance_oi_low"), nil
}

func candidateCoinsFromSymbols(symbols []string, source string) []CandidateCoin {
	candidates := make([]CandidateCoin, 0, len(symbols))
	for _, symbol := range symbols {
		candidates = append(candidates, CandidateCoin{
			Symbol:  market.Normalize(symbol),
			Sources: []string{source},
		})
	}
	return candidates
}

// ============================================================================
// External & Quant Data
// ============================================================================

// FetchMarketData fetches market data based on strategy configuration
func (e *StrategyEngine) FetchMarketData(symbol string) (*market.Data, error) {
	return market.Get(symbol)
}

// FetchExternalData fetches external data sources
func (e *StrategyEngine) FetchExternalData() (map[string]interface{}, error) {
	externalData := make(map[string]interface{})

	for _, source := range e.config.Indicators.ExternalDataSources {
		data, err := e.fetchSingleExternalSource(source)
		if err != nil {
			logger.Infof("⚠️  Failed to fetch external data source [%s]: %v", source.Name, err)
			continue
		}
		externalData[source.Name] = data
	}

	return externalData, nil
}

func (e *StrategyEngine) fetchSingleExternalSource(source store.ExternalDataSource) (interface{}, error) {
	// SSRF Protection: Validate URL before making request
	if err := security.ValidateURL(source.URL); err != nil {
		return nil, fmt.Errorf("external source URL validation failed: %w", err)
	}

	timeout := time.Duration(source.RefreshSecs) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Use SSRF-safe HTTP client
	client := security.SafeHTTPClient(timeout)

	req, err := http.NewRequest(source.Method, source.URL, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range source.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if source.DataPath != "" {
		result = extractJSONPath(result, source.DataPath)
	}

	return result, nil
}

func extractJSONPath(data interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}

	return current
}

// FetchQuantData fetches quantitative data for a single coin
func (e *StrategyEngine) FetchQuantData(symbol string) (*QuantData, error) {
	if !e.config.Indicators.EnableQuantData {
		return nil, nil
	}

	// Use nofxos client with unified API key
	include := "oi,price"
	if e.config.Indicators.EnableQuantNetflow {
		include = "netflow,oi,price"
	}

	nofxosData, err := e.nofxosClient.GetCoinData(symbol, include)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quant data: %w", err)
	}

	if nofxosData == nil {
		return nil, nil
	}

	// Convert nofxos.QuantData to kernel.QuantData
	quantData := &QuantData{
		Symbol:      nofxosData.Symbol,
		Price:       nofxosData.Price,
		PriceChange: nofxosData.PriceChange,
	}

	// Convert OI data
	if nofxosData.OI != nil {
		quantData.OI = make(map[string]*OIData)
		for exchange, oiData := range nofxosData.OI {
			if oiData != nil {
				kData := &OIData{
					CurrentOI: oiData.CurrentOI,
				}
				if oiData.Delta != nil {
					kData.Delta = make(map[string]*OIDeltaData)
					for dur, delta := range oiData.Delta {
						if delta != nil {
							kData.Delta[dur] = &OIDeltaData{
								OIDelta:        delta.OIDelta,
								OIDeltaValue:   delta.OIDeltaValue,
								OIDeltaPercent: delta.OIDeltaPercent,
							}
						}
					}
				}
				quantData.OI[exchange] = kData
			}
		}
	}

	// Convert Netflow data
	if nofxosData.Netflow != nil {
		quantData.Netflow = &NetflowData{}
		if nofxosData.Netflow.Institution != nil {
			quantData.Netflow.Institution = &FlowTypeData{
				Future: nofxosData.Netflow.Institution.Future,
				Spot:   nofxosData.Netflow.Institution.Spot,
			}
		}
		if nofxosData.Netflow.Personal != nil {
			quantData.Netflow.Personal = &FlowTypeData{
				Future: nofxosData.Netflow.Personal.Future,
				Spot:   nofxosData.Netflow.Personal.Spot,
			}
		}
	}

	return quantData, nil
}

// FetchQuantDataBatch batch fetches quantitative data
func (e *StrategyEngine) FetchQuantDataBatch(symbols []string) map[string]*QuantData {
	result := make(map[string]*QuantData)

	if !e.config.Indicators.EnableQuantData {
		return result
	}

	for _, symbol := range symbols {
		data, err := e.FetchQuantData(symbol)
		if err != nil {
			logger.Infof("⚠️  Failed to fetch quantitative data for %s: %v", symbol, err)
			continue
		}
		if data != nil {
			result[symbol] = data
		}
	}

	return result
}

// FetchOIRankingData fetches market-wide OI ranking data
func (e *StrategyEngine) FetchOIRankingData() *nofxos.OIRankingData {
	indicators := e.config.Indicators
	if !indicators.EnableOIRanking {
		return nil
	}

	duration := indicators.OIRankingDuration
	if duration == "" {
		duration = "1h"
	}

	limit := indicators.OIRankingLimit
	if limit <= 0 {
		limit = 10
	}

	logger.Infof("📊 Fetching OI ranking data (duration: %s, limit: %d)", duration, limit)

	data, err := e.nofxosClient.GetOIRanking(duration, limit)
	if err != nil {
		logger.Warnf("⚠️  Failed to fetch OI ranking data: %v", err)
		return nil
	}

	logger.Infof("✓ OI ranking data ready: %d top, %d low positions",
		len(data.TopPositions), len(data.LowPositions))

	return data
}

// FetchNetFlowRankingData fetches market-wide NetFlow ranking data
func (e *StrategyEngine) FetchNetFlowRankingData() *nofxos.NetFlowRankingData {
	indicators := e.config.Indicators
	if !indicators.EnableNetFlowRanking {
		return nil
	}

	duration := indicators.NetFlowRankingDuration
	if duration == "" {
		duration = "1h"
	}

	limit := indicators.NetFlowRankingLimit
	if limit <= 0 {
		limit = 10
	}

	logger.Infof("💰 Fetching NetFlow ranking data (duration: %s, limit: %d)", duration, limit)

	data, err := e.nofxosClient.GetNetFlowRanking(duration, limit)
	if err != nil {
		logger.Warnf("⚠️  Failed to fetch NetFlow ranking data: %v", err)
		return nil
	}

	logger.Infof("✓ NetFlow ranking data ready: inst_in=%d, inst_out=%d, retail_in=%d, retail_out=%d",
		len(data.InstitutionFutureTop), len(data.InstitutionFutureLow),
		len(data.PersonalFutureTop), len(data.PersonalFutureLow))

	return data
}

// FetchPriceRankingData fetches market-wide price ranking data (gainers/losers)
func (e *StrategyEngine) FetchPriceRankingData() *nofxos.PriceRankingData {
	indicators := e.config.Indicators
	if !indicators.EnablePriceRanking {
		return nil
	}

	durations := indicators.PriceRankingDuration
	if durations == "" {
		durations = "1h"
	}

	limit := indicators.PriceRankingLimit
	if limit <= 0 {
		limit = 10
	}

	logger.Infof("📈 Fetching Price ranking data (durations: %s, limit: %d)", durations, limit)

	data, err := e.nofxosClient.GetPriceRanking(durations, limit)
	if err != nil {
		logger.Warnf("⚠️  Failed to fetch Price ranking data: %v", err)
		return nil
	}

	logger.Infof("✓ Price ranking data ready for %d durations", len(data.Durations))

	return data
}

// ============================================================================
// Prompt Building - System Prompt
// ============================================================================

// BuildSystemPrompt builds System Prompt according to strategy configuration
func (e *StrategyEngine) BuildSystemPrompt(accountEquity float64, variant string) string {
	var sb strings.Builder
	riskControl := e.config.RiskControl
	promptSections := e.config.PromptSections

	// 0. Data Dictionary & Schema (ensure AI understands all fields)
	lang := e.GetLanguage()
	schemaPrompt := GetSchemaPrompt(lang)
	sb.WriteString(schemaPrompt)
	sb.WriteString("\n\n")
	sb.WriteString("---\n\n")

	// 1. Role definition (editable)
	if promptSections.RoleDefinition != "" {
		sb.WriteString(promptSections.RoleDefinition)
		sb.WriteString("\n\n")
	} else if lang == LangChinese {
		sb.WriteString("# 你是一名专业的加密货币交易 AI\n\n")
		sb.WriteString("你的任务是基于提供的市场数据做出交易决策。\n\n")
	} else {
		sb.WriteString("# You are a professional cryptocurrency trading AI\n\n")
		sb.WriteString("Your task is to make trading decisions based on provided market data.\n\n")
	}

	// 2. Trading mode variant
	switch strings.ToLower(strings.TrimSpace(variant)) {
	case "aggressive":
		if lang == LangChinese {
			sb.WriteString("## 模式: 激进\n- 优先捕捉趋势突破，当信心度 ≥ 70 时可分批建仓\n- 允许较高仓位，但必须严格设置止损并说明风险收益比\n\n")
		} else {
			sb.WriteString("## Mode: Aggressive\n- Prioritize capturing trend breakouts, can build positions in batches when confidence ≥ 70\n- Allow higher positions, but must strictly set stop-loss and explain risk-reward ratio\n\n")
		}
	case "conservative":
		if lang == LangChinese {
			sb.WriteString("## 模式: 保守\n- 仅在多信号共振时开仓\n- 以保本为首要目标，连续亏损后必须暂停多个周期\n\n")
		} else {
			sb.WriteString("## Mode: Conservative\n- Only open positions when multiple signals resonate\n- Prioritize cash preservation, must pause for multiple periods after consecutive losses\n\n")
		}
	case "scalping":
		if lang == LangChinese {
			sb.WriteString("## 模式: 短线\n- 关注短期动量，较小盈利目标但需要快速行动\n- 如果价格在两个 K 线内未按预期方向移动，立即减仓或止损\n\n")
		} else {
			sb.WriteString("## Mode: Scalping\n- Focus on short-term momentum, smaller profit targets but require quick action\n- If price doesn't move as expected within two bars, immediately reduce position or stop-loss\n\n")
		}
	}

	// 3. Hard constraints (risk control)
	btcEthPosValueRatio := riskControl.BTCETHMaxPositionValueRatio
	if btcEthPosValueRatio <= 0 {
		btcEthPosValueRatio = 5.0
	}
	altcoinPosValueRatio := riskControl.AltcoinMaxPositionValueRatio
	if altcoinPosValueRatio <= 0 {
		altcoinPosValueRatio = 1.0
	}
	minRiskRewardRatio := riskControl.MinRiskRewardRatio
	if minRiskRewardRatio <= 0 {
		minRiskRewardRatio = defaultMinRiskRewardRatio
	}

	if lang == LangChinese {
		sb.WriteString("# 硬约束（风险控制）\n\n")
		sb.WriteString("## 代码强制执行（后端校验，无法绕过）：\n")
		sb.WriteString(fmt.Sprintf("- 最大持仓数量: %d 个币种\n", riskControl.MaxPositions))
		sb.WriteString(fmt.Sprintf("- 仓位价值限制(山寨币): 最大 %.0f USDT (= 权益 %.0f × %.1f倍)\n",
			accountEquity*altcoinPosValueRatio, accountEquity, altcoinPosValueRatio))
		sb.WriteString(fmt.Sprintf("- 仓位价值限制(BTC/ETH): 最大 %.0f USDT (= 权益 %.0f × %.1f倍)\n",
			accountEquity*btcEthPosValueRatio, accountEquity, btcEthPosValueRatio))
		sb.WriteString(fmt.Sprintf("- 最大保证金使用率: ≤%.0f%%\n", riskControl.MaxMarginUsage*100))
		sb.WriteString(fmt.Sprintf("- 最小开仓金额: ≥%.0f USDT\n\n", riskControl.MinPositionSize))

		sb.WriteString("## AI 建议（推荐遵守）：\n")
		sb.WriteString(fmt.Sprintf("- 交易杠杆: 山寨币最大 %dx | BTC/ETH 最大 %dx\n",
			riskControl.AltcoinMaxLeverage, riskControl.BTCETHMaxLeverage))
		sb.WriteString(fmt.Sprintf("- 风险收益比: ≥1:%.1f (止盈 / 止损)\n", minRiskRewardRatio))
		sb.WriteString(fmt.Sprintf("- 最小信心度: ≥%d 才能开仓\n\n", riskControl.MinConfidence))

		sb.WriteString("## 仓位大小指引\n")
		sb.WriteString("根据你的信心度和上述仓位价值限制计算 `position_size_usd`：\n")
		sb.WriteString("- 高信心度(≥85): 使用仓位限制的 80-100%%\n")
		sb.WriteString("- 中信心度(70-84): 使用仓位限制的 50-80%%\n")
		sb.WriteString("- 低信心度(60-69): 使用仓位限制的 30-50%%\n")
		sb.WriteString(fmt.Sprintf("- 示例: 权益 %.0f USDT, BTC/ETH 倍数 %.1f倍, 最大仓位 %.0f USDT\n",
			accountEquity, btcEthPosValueRatio, accountEquity*btcEthPosValueRatio))
		sb.WriteString("- **禁止** 直接使用 available_balance 作为 position_size_usd，请使用仓位价值限制！\n\n")
	} else {
		sb.WriteString("# Hard Constraints (Risk Control)\n\n")
		sb.WriteString("## CODE ENFORCED (Backend validation, cannot be bypassed):\n")
		sb.WriteString(fmt.Sprintf("- Max Positions: %d coins simultaneously\n", riskControl.MaxPositions))
		sb.WriteString(fmt.Sprintf("- Position Value Limit (Altcoins): max %.0f USDT (= equity %.0f × %.1fx)\n",
			accountEquity*altcoinPosValueRatio, accountEquity, altcoinPosValueRatio))
		sb.WriteString(fmt.Sprintf("- Position Value Limit (BTC/ETH): max %.0f USDT (= equity %.0f × %.1fx)\n",
			accountEquity*btcEthPosValueRatio, accountEquity, btcEthPosValueRatio))
		sb.WriteString(fmt.Sprintf("- Max Margin Usage: ≤%.0f%%\n", riskControl.MaxMarginUsage*100))
		sb.WriteString(fmt.Sprintf("- Min Position Size: ≥%.0f USDT\n\n", riskControl.MinPositionSize))

		sb.WriteString("## AI GUIDED (Recommended, you should follow):\n")
		sb.WriteString(fmt.Sprintf("- Trading Leverage: Altcoins max %dx | BTC/ETH max %dx\n",
			riskControl.AltcoinMaxLeverage, riskControl.BTCETHMaxLeverage))
		sb.WriteString(fmt.Sprintf("- Risk-Reward Ratio: ≥1:%.1f (take_profit / stop_loss)\n", minRiskRewardRatio))
		sb.WriteString(fmt.Sprintf("- Min Confidence: ≥%d to open position\n\n", riskControl.MinConfidence))

		// Position sizing guidance
		sb.WriteString("## Position Sizing Guidance\n")
		sb.WriteString("Calculate `position_size_usd` based on your confidence and the Position Value Limits above:\n")
		sb.WriteString("- High confidence (≥85): Use 80-100%% of max position value limit\n")
		sb.WriteString("- Medium confidence (70-84): Use 50-80%% of max position value limit\n")
		sb.WriteString("- Low confidence (60-69): Use 30-50%% of max position value limit\n")
		sb.WriteString(fmt.Sprintf("- Example: With equity %.0f and BTC/ETH ratio %.1fx, max is %.0f USDT\n",
			accountEquity, btcEthPosValueRatio, accountEquity*btcEthPosValueRatio))
		sb.WriteString("- **DO NOT** just use available_balance as position_size_usd. Use the Position Value Limits!\n\n")
	}

	// 4. Trading frequency (editable)
	if promptSections.TradingFrequency != "" {
		sb.WriteString(promptSections.TradingFrequency)
		sb.WriteString("\n\n")
	} else if lang == LangChinese {
		sb.WriteString("# ⏱️ 交易频率意识\n\n")
		sb.WriteString("- 优秀交易员: 2-4 笔/天 ≈ 0.1-0.2 笔/小时\n")
		sb.WriteString("- >2 笔/小时 = 过度交易\n")
		sb.WriteString("- 单笔持仓时间 ≥ 30-60 分钟\n")
		sb.WriteString("如果每个周期都在交易，说明标准太低；如果持仓 < 30 分钟就平仓，说明过于急躁。\n\n")
	} else {
		sb.WriteString("# ⏱️ Trading Frequency Awareness\n\n")
		sb.WriteString("- Excellent traders: 2-4 trades/day ≈ 0.1-0.2 trades/hour\n")
		sb.WriteString("- >2 trades/hour = Overtrading\n")
		sb.WriteString("- Single position hold time ≥ 30-60 minutes\n")
		sb.WriteString("If you find yourself trading every period → standards too low; if closing positions < 30 minutes → too impatient.\n\n")
	}

	// 5. Entry standards (editable)
	if promptSections.EntryStandards != "" {
		sb.WriteString(promptSections.EntryStandards)
		if lang == LangChinese {
			sb.WriteString("\n\n你拥有以下指标数据：\n")
		} else {
			sb.WriteString("\n\nYou have the following indicator data:\n")
		}
		e.writeAvailableIndicators(&sb, lang)
		if lang == LangChinese {
			sb.WriteString(fmt.Sprintf("\n**信心度 ≥ %d** 才能开仓。\n\n", riskControl.MinConfidence))
		} else {
			sb.WriteString(fmt.Sprintf("\n**Confidence ≥ %d** required to open positions.\n\n", riskControl.MinConfidence))
		}
	} else if lang == LangChinese {
		sb.WriteString("# 🎯 入场标准（严格）\n\n")
		sb.WriteString("仅在多个信号共振时开仓。你拥有：\n")
		e.writeAvailableIndicators(&sb, lang)
		sb.WriteString(fmt.Sprintf("\n可以使用任何有效分析方法，但 **信心度 ≥ %d** 才能开仓；避免单一指标、信号矛盾、横盘震荡、平仓后立即反向开仓等低质量行为。\n\n", riskControl.MinConfidence))
	} else {
		sb.WriteString("# 🎯 Entry Standards (Strict)\n\n")
		sb.WriteString("Only open positions when multiple signals resonate. You have:\n")
		e.writeAvailableIndicators(&sb, lang)
		sb.WriteString(fmt.Sprintf("\nFeel free to use any effective analysis method, but **confidence ≥ %d** required to open positions; avoid low-quality behaviors such as single indicators, contradictory signals, sideways consolidation, reopening immediately after closing, etc.\n\n", riskControl.MinConfidence))
	}

	// 6. Decision process (editable)
	if promptSections.DecisionProcess != "" {
		sb.WriteString(promptSections.DecisionProcess)
		sb.WriteString("\n\n")
	} else if lang == LangChinese {
		sb.WriteString("# 📋 决策流程\n\n")
		sb.WriteString("1. 检查持仓 → 是否应该止盈/止损\n")
		sb.WriteString("2. 扫描候选币种 + 多时间框架 → 是否有强信号\n")
		sb.WriteString("3. 输出简短、可审计的理由摘要，再输出结构化 JSON\n\n")
	} else {
		sb.WriteString("# 📋 Decision Process\n\n")
		sb.WriteString("1. Check positions → Should we take profit/stop-loss\n")
		sb.WriteString("2. Scan candidate coins + multi-timeframe → Are there strong signals\n")
		sb.WriteString("3. Output a brief auditable rationale summary, then structured JSON\n\n")
	}

	// 7. Custom Prompt
	if e.config.CustomPrompt != "" {
		if lang == LangChinese {
			sb.WriteString("# 📌 个性化交易策略\n\n")
		} else {
			sb.WriteString("# 📌 Personalized Trading Strategy\n\n")
		}
		sb.WriteString(e.config.CustomPrompt)
		sb.WriteString("\n\n")
		if lang == LangChinese {
			sb.WriteString("注意：上述个性化策略是对基本规则的补充，不能违反后续最终硬约束和风险控制原则。\n\n")
		} else {
			sb.WriteString("Note: The personalized strategy above supplements the base rules and cannot override the final hard constraints or risk controls below.\n\n")
		}
	}

	// 8. Output format
	if lang == LangChinese {
		sb.WriteString("# 输出格式（严格遵循）\n\n")
		sb.WriteString("**必须使用 XML 标签 <reasoning> 和 <decision> 分隔理由摘要和决策 JSON，避免解析错误**\n\n")
		sb.WriteString("## 格式要求\n\n")
		sb.WriteString("<reasoning>\n")
		sb.WriteString("- 用 3-6 条短句概括关键证据、风险点和无效条件\n")
		sb.WriteString("- 不要输出完整隐藏推理链，不要在此处放 JSON\n")
		sb.WriteString("</reasoning>\n\n")
		sb.WriteString("<decision>\n")
		sb.WriteString("第二步：JSON 决策数组\n\n")
	} else {
		sb.WriteString("# Output Format (Strictly Follow)\n\n")
		sb.WriteString("**Must use XML tags <reasoning> and <decision> to separate the rationale summary and decision JSON, avoiding parsing errors**\n\n")
		sb.WriteString("## Format Requirements\n\n")
		sb.WriteString("<reasoning>\n")
		sb.WriteString("- Summarize key evidence, risks, and invalidation conditions in 3-6 short bullets\n")
		sb.WriteString("- Do not output hidden chain-of-thought and do not place JSON here\n")
		sb.WriteString("</reasoning>\n\n")
		sb.WriteString("<decision>\n")
		sb.WriteString("Step 2: JSON decision array\n\n")
	}
	sb.WriteString("```json\n[\n")
	// Use the actual configured position value ratio for BTC/ETH in the example
	examplePositionSize := accountEquity * btcEthPosValueRatio
	sb.WriteString(fmt.Sprintf("  {\"symbol\": \"BTCUSDT\", \"action\": \"open_short\", \"leverage\": %d, \"position_size_usd\": %.0f, \"stop_loss\": 97000, \"take_profit\": 91000, \"confidence\": 85, \"risk_usd\": 300},\n",
		riskControl.BTCETHMaxLeverage, examplePositionSize))
	sb.WriteString("  {\"symbol\": \"ETHUSDT\", \"action\": \"close_long\"}\n")
	sb.WriteString("]\n```\n")
	sb.WriteString("</decision>\n\n")
	if lang == LangChinese {
		sb.WriteString("## 字段说明\n\n")
		sb.WriteString("- `action`: open_long | open_short | close_long | close_short | hold | wait\n")
		sb.WriteString(fmt.Sprintf("- `confidence`: 0-100（建议开仓 ≥ %d）\n", riskControl.MinConfidence))
		sb.WriteString("- 开仓时必须包含: leverage, position_size_usd, stop_loss, take_profit, confidence, risk_usd\n")
		sb.WriteString("- **重要**: 所有数值必须是计算后的数字，不能是公式/表达式（例如使用 `27.76` 而非 `3000 * 0.01`）\n\n")
	} else {
		sb.WriteString("## Field Description\n\n")
		sb.WriteString("- `action`: open_long | open_short | close_long | close_short | hold | wait\n")
		sb.WriteString(fmt.Sprintf("- `confidence`: 0-100 (opening recommended ≥ %d)\n", riskControl.MinConfidence))
		sb.WriteString("- Required when opening: leverage, position_size_usd, stop_loss, take_profit, confidence, risk_usd\n")
		sb.WriteString("- **IMPORTANT**: All numeric values must be calculated numbers, NOT formulas/expressions (e.g., use `27.76` not `3000 * 0.01`)\n\n")
	}

	if lang == LangChinese {
		sb.WriteString("## 最终硬约束\n\n")
		sb.WriteString("- 个性化策略、模式偏好和市场观点都不能覆盖风险控制、输出格式和后端校验规则。\n")
		sb.WriteString(fmt.Sprintf("- 风险收益比必须 ≥1:%.1f，开仓信心度建议 ≥%d。\n", minRiskRewardRatio, riskControl.MinConfidence))
	} else {
		sb.WriteString("## Final Hard Constraints\n\n")
		sb.WriteString("- Personalized strategy, mode preference, and market opinion cannot override risk controls, output format, or backend validation rules.\n")
		sb.WriteString(fmt.Sprintf("- Risk-reward ratio must be ≥1:%.1f and opening confidence should be ≥%d.\n", minRiskRewardRatio, riskControl.MinConfidence))
	}

	return sb.String()
}

func (e *StrategyEngine) writeAvailableIndicators(sb *strings.Builder, lang Language) {
	indicators := e.config.Indicators
	kline := indicators.Klines

	if lang == LangChinese {
		sb.WriteString(fmt.Sprintf("- %s 价格序列", kline.PrimaryTimeframe))
		if kline.EnableMultiTimeframe {
			sb.WriteString(fmt.Sprintf(" + %s K线序列\n", kline.LongerTimeframe))
		} else {
			sb.WriteString("\n")
		}

		if indicators.EnableEMA {
			sb.WriteString("- EMA 指标")
			if len(indicators.EMAPeriods) > 0 {
				sb.WriteString(fmt.Sprintf(" (周期: %v)", indicators.EMAPeriods))
			}
			sb.WriteString("\n")
		}

		if indicators.EnableMACD {
			sb.WriteString("- MACD 指标\n")
		}

		if indicators.EnableRSI {
			sb.WriteString("- RSI 指标")
			if len(indicators.RSIPeriods) > 0 {
				sb.WriteString(fmt.Sprintf(" (周期: %v)", indicators.RSIPeriods))
			}
			sb.WriteString("\n")
		}

		if indicators.EnableATR {
			sb.WriteString("- ATR 指标")
			if len(indicators.ATRPeriods) > 0 {
				sb.WriteString(fmt.Sprintf(" (周期: %v)", indicators.ATRPeriods))
			}
			sb.WriteString("\n")
		}

		if indicators.EnableBOLL {
			sb.WriteString("- 布林带 (BOLL) - 上轨/中轨/下轨")
			if len(indicators.BOLLPeriods) > 0 {
				sb.WriteString(fmt.Sprintf(" (周期: %v)", indicators.BOLLPeriods))
			}
			sb.WriteString("\n")
		}

		if indicators.EnableVolume {
			sb.WriteString("- 成交量数据\n")
		}

		if indicators.EnableOI {
			sb.WriteString("- 持仓量 (OI) 数据\n")
		}

		if indicators.EnableFundingRate {
			sb.WriteString("- 资金费率\n")
		}

		if len(e.config.CoinSource.StaticCoins) > 0 || e.config.CoinSource.UseAI500 || e.config.CoinSource.UseOITop {
			sb.WriteString("- 综合评分榜 / OI 持仓增加过滤标签（如有）\n")
		}

		if indicators.EnableQuantData {
			sb.WriteString("- 量化数据（机构/散户资金流向、仓位变化、多周期价格变化）\n")
		}
	} else {
		sb.WriteString(fmt.Sprintf("- %s price series", kline.PrimaryTimeframe))
		if kline.EnableMultiTimeframe {
			sb.WriteString(fmt.Sprintf(" + %s K-line series\n", kline.LongerTimeframe))
		} else {
			sb.WriteString("\n")
		}

		if indicators.EnableEMA {
			sb.WriteString("- EMA indicators")
			if len(indicators.EMAPeriods) > 0 {
				sb.WriteString(fmt.Sprintf(" (periods: %v)", indicators.EMAPeriods))
			}
			sb.WriteString("\n")
		}

		if indicators.EnableMACD {
			sb.WriteString("- MACD indicators\n")
		}

		if indicators.EnableRSI {
			sb.WriteString("- RSI indicators")
			if len(indicators.RSIPeriods) > 0 {
				sb.WriteString(fmt.Sprintf(" (periods: %v)", indicators.RSIPeriods))
			}
			sb.WriteString("\n")
		}

		if indicators.EnableATR {
			sb.WriteString("- ATR indicators")
			if len(indicators.ATRPeriods) > 0 {
				sb.WriteString(fmt.Sprintf(" (periods: %v)", indicators.ATRPeriods))
			}
			sb.WriteString("\n")
		}

		if indicators.EnableBOLL {
			sb.WriteString("- Bollinger Bands (BOLL) - Upper/Middle/Lower bands")
			if len(indicators.BOLLPeriods) > 0 {
				sb.WriteString(fmt.Sprintf(" (periods: %v)", indicators.BOLLPeriods))
			}
			sb.WriteString("\n")
		}

		if indicators.EnableVolume {
			sb.WriteString("- Volume data\n")
		}

		if indicators.EnableOI {
			sb.WriteString("- Open Interest (OI) data\n")
		}

		if indicators.EnableFundingRate {
			sb.WriteString("- Funding rate\n")
		}

		if len(e.config.CoinSource.StaticCoins) > 0 || e.config.CoinSource.UseAI500 || e.config.CoinSource.UseOITop {
			sb.WriteString("- Score Top / OI Increase filter tags (if available)\n")
		}

		if indicators.EnableQuantData {
			sb.WriteString("- Quantitative data (institutional/retail fund flow, position changes, multi-period price changes)\n")
		}
	}
}

// ============================================================================
// Prompt Building - User Prompt
// ============================================================================

// BuildUserPrompt builds User Prompt based on strategy configuration
func (e *StrategyEngine) BuildUserPrompt(ctx *Context) string {
	var sb strings.Builder

	// System status
	lang := e.GetLanguage()
	if lang == LangChinese {
		sb.WriteString(fmt.Sprintf("时间: %s | 周期: #%d | 运行时长: %d 分钟\n\n",
			ctx.CurrentTime, ctx.CallCount, ctx.RuntimeMinutes))
	} else {
		sb.WriteString(fmt.Sprintf("Time: %s | Period: #%d | Runtime: %d minutes\n\n",
			ctx.CurrentTime, ctx.CallCount, ctx.RuntimeMinutes))
	}

	// BTC market
	if btcData, hasBTC := ctx.MarketDataMap["BTCUSDT"]; hasBTC {
		if lang == LangChinese {
			sb.WriteString(fmt.Sprintf("BTC: %.2f (1小时: %+.2f%%, 4小时: %+.2f%%) | MACD: %.4f | RSI: %.2f\n\n",
				btcData.CurrentPrice, btcData.PriceChange1h, btcData.PriceChange4h,
				btcData.CurrentMACD, btcData.CurrentRSI7))
		} else {
			sb.WriteString(fmt.Sprintf("BTC: %.2f (1h: %+.2f%%, 4h: %+.2f%%) | MACD: %.4f | RSI: %.2f\n\n",
				btcData.CurrentPrice, btcData.PriceChange1h, btcData.PriceChange4h,
				btcData.CurrentMACD, btcData.CurrentRSI7))
		}
	}

	// Account information
	if lang == LangChinese {
		sb.WriteString(fmt.Sprintf("账户: 权益 %.2f | 余额 %.2f (%.1f%%) | 盈亏 %+.2f%% | 保证金 %.1f%% | 持仓 %d\n\n",
			ctx.Account.TotalEquity,
			ctx.Account.AvailableBalance,
			(ctx.Account.AvailableBalance/ctx.Account.TotalEquity)*100,
			ctx.Account.TotalPnLPct,
			ctx.Account.MarginUsedPct,
			ctx.Account.PositionCount))
	} else {
		sb.WriteString(fmt.Sprintf("Account: Equity %.2f | Balance %.2f (%.1f%%) | PnL %+.2f%% | Margin %.1f%% | Positions %d\n\n",
			ctx.Account.TotalEquity,
			ctx.Account.AvailableBalance,
			(ctx.Account.AvailableBalance/ctx.Account.TotalEquity)*100,
			ctx.Account.TotalPnLPct,
			ctx.Account.MarginUsedPct,
			ctx.Account.PositionCount))
	}

	// Recently completed orders (placed before positions to ensure visibility)
	if len(ctx.RecentOrders) > 0 {
		if lang == LangChinese {
			sb.WriteString("## 最近完成的交易\n")
		} else {
			sb.WriteString("## Recent Completed Trades\n")
		}
		for i, order := range ctx.RecentOrders {
			var resultStr string
			var entryLabel string
			var exitLabel string
			if lang == LangChinese {
				if order.RealizedPnL < 0 {
					resultStr = "亏损"
				} else {
					resultStr = "盈利"
				}
				entryLabel = "进场"
				exitLabel = "出场"
			} else {
				if order.RealizedPnL < 0 {
					resultStr = "Loss"
				} else {
					resultStr = "Profit"
				}
				entryLabel = "Entry"
				exitLabel = "Exit"
			}
			sb.WriteString(fmt.Sprintf("%d. %s %s | %s %.4f %s %.4f | %s: %+.2f USDT (%+.2f%%) | %s→%s (%s)\n",
				i+1, order.Symbol, order.Side,
				entryLabel, order.EntryPrice, exitLabel, order.ExitPrice,
				resultStr, order.RealizedPnL, order.PnLPct,
				order.EntryTime, order.ExitTime, order.HoldDuration))
		}
		sb.WriteString("\n")
	}

	// Historical trading statistics (helps AI understand past performance)
	if ctx.TradingStats != nil && ctx.TradingStats.TotalTrades > 0 {

		// Win/Loss ratio
		var winLossRatio float64
		if ctx.TradingStats.AvgLoss > 0 {
			winLossRatio = ctx.TradingStats.AvgWin / ctx.TradingStats.AvgLoss
		}

		if lang == LangChinese {
			sb.WriteString("## 历史交易统计\n")
			sb.WriteString(fmt.Sprintf("总交易: %d 笔 | 盈利因子: %.2f | 夏普比率: %.2f | 盈亏比: %.2f\n",
				ctx.TradingStats.TotalTrades,
				ctx.TradingStats.ProfitFactor,
				ctx.TradingStats.SharpeRatio,
				winLossRatio))
			sb.WriteString(fmt.Sprintf("总盈亏: %+.2f USDT | 平均盈利: +%.2f | 平均亏损: -%.2f | 最大回撤: %.1f%%\n",
				ctx.TradingStats.TotalPnL,
				ctx.TradingStats.AvgWin,
				ctx.TradingStats.AvgLoss,
				ctx.TradingStats.MaxDrawdownPct))

			// Performance hints based on profit factor, sharpe, and drawdown
			if ctx.TradingStats.ProfitFactor >= 1.5 && ctx.TradingStats.SharpeRatio >= 1 {
				sb.WriteString("表现: 良好 - 保持当前策略\n")
			} else if ctx.TradingStats.ProfitFactor < 1 {
				sb.WriteString("表现: 需改进 - 提高盈亏比，优化止盈止损\n")
			} else if ctx.TradingStats.MaxDrawdownPct > 30 {
				sb.WriteString("表现: 风险偏高 - 减少仓位，控制回撤\n")
			} else {
				sb.WriteString("表现: 正常 - 有优化空间\n")
			}
		} else {
			sb.WriteString("## Historical Trading Statistics\n")
			sb.WriteString(fmt.Sprintf("Total Trades: %d | Profit Factor: %.2f | Sharpe: %.2f | Win/Loss Ratio: %.2f\n",
				ctx.TradingStats.TotalTrades,
				ctx.TradingStats.ProfitFactor,
				ctx.TradingStats.SharpeRatio,
				winLossRatio))
			sb.WriteString(fmt.Sprintf("Total PnL: %+.2f USDT | Avg Win: +%.2f | Avg Loss: -%.2f | Max Drawdown: %.1f%%\n",
				ctx.TradingStats.TotalPnL,
				ctx.TradingStats.AvgWin,
				ctx.TradingStats.AvgLoss,
				ctx.TradingStats.MaxDrawdownPct))

			// Performance hints based on profit factor, sharpe, and drawdown
			if ctx.TradingStats.ProfitFactor >= 1.5 && ctx.TradingStats.SharpeRatio >= 1 {
				sb.WriteString("Performance: GOOD - maintain current strategy\n")
			} else if ctx.TradingStats.ProfitFactor < 1 {
				sb.WriteString("Performance: NEEDS IMPROVEMENT - improve win/loss ratio, optimize TP/SL\n")
			} else if ctx.TradingStats.MaxDrawdownPct > 30 {
				sb.WriteString("Performance: HIGH RISK - reduce position size, control drawdown\n")
			} else {
				sb.WriteString("Performance: NORMAL - room for optimization\n")
			}
		}
		sb.WriteString("\n")
	}

	// Position information
	if len(ctx.Positions) > 0 {
		if lang == LangChinese {
			sb.WriteString("## 当前持仓\n")
		} else {
			sb.WriteString("## Current Positions\n")
		}
		for i, pos := range ctx.Positions {
			sb.WriteString(e.formatPositionInfo(i+1, pos, ctx, lang))
		}
	} else {
		if lang == LangChinese {
			sb.WriteString("当前持仓: 无\n\n")
		} else {
			sb.WriteString("Current Positions: None\n\n")
		}
	}

	// Candidate coins (exclude coins already in positions to avoid duplicate data)
	positionSymbols := make(map[string]bool)
	for _, pos := range ctx.Positions {
		// Normalize symbol to handle both "ETH" and "ETHUSDT" formats
		normalizedSymbol := market.Normalize(pos.Symbol)
		positionSymbols[normalizedSymbol] = true
	}

	if lang == LangChinese {
		sb.WriteString(fmt.Sprintf("## 候选币种 (%d 个)\n\n", len(ctx.MarketDataMap)))
	} else {
		sb.WriteString(fmt.Sprintf("## Candidate Coins (%d coins)\n\n", len(ctx.MarketDataMap)))
	}
	displayedCount := 0
	for _, coin := range ctx.CandidateCoins {
		// Skip if this coin is already a position (data already shown in positions section)
		normalizedCoinSymbol := market.Normalize(coin.Symbol)
		if positionSymbols[normalizedCoinSymbol] {
			continue
		}

		marketData, hasData := ctx.MarketDataMap[coin.Symbol]
		if !hasData {
			continue
		}
		displayedCount++

		sourceTags := e.formatCoinSourceTag(coin.Sources)
		sb.WriteString(fmt.Sprintf("### %d. %s%s\n\n", displayedCount, coin.Symbol, sourceTags))
		sb.WriteString(e.formatMarketData(marketData))

		if ctx.QuantDataMap != nil {
			if quantData, hasQuant := ctx.QuantDataMap[coin.Symbol]; hasQuant {
				sb.WriteString(e.formatQuantData(quantData))
			}
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Get language for market data formatting
	nofxosLang := nofxos.LangEnglish
	if lang == LangChinese {
		nofxosLang = nofxos.LangChinese
	}

	// OI Ranking data (market-wide open interest changes)
	if ctx.OIRankingData != nil {
		sb.WriteString(nofxos.FormatOIRankingForAI(ctx.OIRankingData, nofxosLang))
	}

	// NetFlow Ranking data (market-wide fund flow)
	if ctx.NetFlowRankingData != nil {
		sb.WriteString(nofxos.FormatNetFlowRankingForAI(ctx.NetFlowRankingData, nofxosLang))
	}

	// Price Ranking data (market-wide gainers/losers)
	if ctx.PriceRankingData != nil {
		sb.WriteString(nofxos.FormatPriceRankingForAI(ctx.PriceRankingData, nofxosLang))
	}

	sb.WriteString("---\n\n")
	if lang == LangChinese {
		sb.WriteString("现在请分析并输出你的决策（思考链 + JSON）\n")
	} else {
		sb.WriteString("Now please analyze and output your decision (Chain of Thought + JSON)\n")
	}

	return sb.String()
}

func (e *StrategyEngine) formatPositionInfo(index int, pos PositionInfo, ctx *Context, lang Language) string {
	var sb strings.Builder

	holdingDuration := ""
	if pos.UpdateTime > 0 {
		durationMs := time.Now().UnixMilli() - pos.UpdateTime
		durationMin := durationMs / (1000 * 60)
		if lang == LangChinese {
			if durationMin < 60 {
				holdingDuration = fmt.Sprintf(" | 持仓时长 %d 分钟", durationMin)
			} else {
				durationHour := durationMin / 60
				durationMinRemainder := durationMin % 60
				holdingDuration = fmt.Sprintf(" | 持仓时长 %d小时 %d分钟", durationHour, durationMinRemainder)
			}
		} else {
			if durationMin < 60 {
				holdingDuration = fmt.Sprintf(" | Holding Duration %d min", durationMin)
			} else {
				durationHour := durationMin / 60
				durationMinRemainder := durationMin % 60
				holdingDuration = fmt.Sprintf(" | Holding Duration %dh %dm", durationHour, durationMinRemainder)
			}
		}
	}

	positionValue := pos.Quantity * pos.MarkPrice
	if positionValue < 0 {
		positionValue = -positionValue
	}

	if lang == LangChinese {
		sb.WriteString(fmt.Sprintf("%d. %s %s | 进场 %.4f 当前 %.4f | 数量 %.4f | 仓位价值 %.2f USDT | 盈亏%+.2f%% | 盈亏金额%+.2f USDT | 峰值盈亏%.2f%% | 杠杆 %dx | 保证金 %.0f | 强平价 %.4f%s\n\n",
			index, pos.Symbol, strings.ToUpper(pos.Side),
			pos.EntryPrice, pos.MarkPrice, pos.Quantity, positionValue, pos.UnrealizedPnLPct, pos.UnrealizedPnL, pos.PeakPnLPct,
			pos.Leverage, pos.MarginUsed, pos.LiquidationPrice, holdingDuration))
	} else {
		sb.WriteString(fmt.Sprintf("%d. %s %s | Entry %.4f Current %.4f | Qty %.4f | Position Value %.2f USDT | PnL%+.2f%% | PnL Amount%+.2f USDT | Peak PnL%.2f%% | Leverage %dx | Margin %.0f | Liq Price %.4f%s\n\n",
			index, pos.Symbol, strings.ToUpper(pos.Side),
			pos.EntryPrice, pos.MarkPrice, pos.Quantity, positionValue, pos.UnrealizedPnLPct, pos.UnrealizedPnL, pos.PeakPnLPct,
			pos.Leverage, pos.MarginUsed, pos.LiquidationPrice, holdingDuration))
	}

	if marketData, ok := ctx.MarketDataMap[pos.Symbol]; ok {
		sb.WriteString(e.formatMarketData(marketData))

		if ctx.QuantDataMap != nil {
			if quantData, hasQuant := ctx.QuantDataMap[pos.Symbol]; hasQuant {
				sb.WriteString(e.formatQuantData(quantData))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (e *StrategyEngine) formatCoinSourceTag(sources []string) string {
	if len(sources) > 1 {
		labels := make([]string, 0, len(sources))
		seen := make(map[string]bool, len(sources))
		for _, source := range sources {
			label := e.formatSingleCoinSource(source)
			if label == "" || seen[label] {
				continue
			}
			seen[label] = true
			labels = append(labels, label)
		}
		if len(labels) == 0 {
			return " (Multiple sources)"
		}
		return " (" + strings.Join(labels, "+") + ")"
	} else if len(sources) == 1 {
		label := e.formatSingleCoinSource(sources[0])
		if label != "" {
			return " (" + label + ")"
		}
	}
	return ""
}

func (e *StrategyEngine) formatSingleCoinSource(source string) string {
	switch source {
	case "ai500":
		return "Score Top"
	case "oi_top":
		return "OI Increase"
	case "oi_low":
		return "OI Decrease"
	case "binance_score":
		return "Binance Score"
	case "binance_oi_top":
		return "Binance OI Increase"
	case "binance_oi_low":
		return "Binance OI Decrease"
	case "static":
		return "Manual selection"
	default:
		return source
	}
}

// ============================================================================
// Market Data Formatting
// ============================================================================

func (e *StrategyEngine) formatMarketData(data *market.Data) string {
	var sb strings.Builder
	indicators := e.config.Indicators

	// 明确标注币种
	sb.WriteString(fmt.Sprintf("=== %s Market Data ===\n\n", data.Symbol))
	sb.WriteString(fmt.Sprintf("current_price = %.4f", data.CurrentPrice))

	if indicators.EnableEMA {
		sb.WriteString(fmt.Sprintf(", current_ema20 = %.3f", data.CurrentEMA20))
	}

	if indicators.EnableMACD {
		sb.WriteString(fmt.Sprintf(", current_macd = %.3f", data.CurrentMACD))
	}

	if indicators.EnableRSI {
		sb.WriteString(fmt.Sprintf(", current_rsi7 = %.3f", data.CurrentRSI7))
	}

	sb.WriteString("\n\n")

	if indicators.EnableOI || indicators.EnableFundingRate {
		sb.WriteString(fmt.Sprintf("Additional data for %s:\n\n", data.Symbol))

		if indicators.EnableOI && data.OpenInterest != nil {
			sb.WriteString(fmt.Sprintf("Open Interest: Latest: %.2f Average: %.2f\n\n",
				data.OpenInterest.Latest, data.OpenInterest.Average))
		}

		if indicators.EnableFundingRate {
			sb.WriteString(fmt.Sprintf("Funding Rate: %.2e\n\n", data.FundingRate))
		}
	}

	if len(data.TimeframeData) > 0 {
		timeframeOrder := []string{"1m", "3m", "5m", "15m", "30m", "1h", "2h", "4h", "6h", "8h", "12h", "1d", "3d", "1w"}
		for _, tf := range timeframeOrder {
			if tfData, ok := data.TimeframeData[tf]; ok {
				sb.WriteString(fmt.Sprintf("=== %s Timeframe (oldest → latest) ===\n\n", strings.ToUpper(tf)))
				e.formatTimeframeSeriesData(&sb, tfData, indicators)
			}
		}
	} else {
		// Compatible with old data format
		if data.IntradaySeries != nil {
			klineConfig := indicators.Klines
			sb.WriteString(fmt.Sprintf("Intraday series (%s intervals, oldest → latest):\n\n", klineConfig.PrimaryTimeframe))

			if len(data.IntradaySeries.MidPrices) > 0 {
				sb.WriteString(fmt.Sprintf("Mid prices: %s\n\n", formatFloatSlice(data.IntradaySeries.MidPrices)))
			}

			if indicators.EnableEMA && len(data.IntradaySeries.EMA20Values) > 0 {
				sb.WriteString(fmt.Sprintf("EMA indicators (20-period): %s\n\n", formatFloatSlice(data.IntradaySeries.EMA20Values)))
			}

			if indicators.EnableMACD && len(data.IntradaySeries.MACDValues) > 0 {
				sb.WriteString(fmt.Sprintf("MACD indicators: %s\n\n", formatFloatSlice(data.IntradaySeries.MACDValues)))
			}

			if indicators.EnableRSI {
				if len(data.IntradaySeries.RSI7Values) > 0 {
					sb.WriteString(fmt.Sprintf("RSI indicators (7-Period): %s\n\n", formatFloatSlice(data.IntradaySeries.RSI7Values)))
				}
				if len(data.IntradaySeries.RSI14Values) > 0 {
					sb.WriteString(fmt.Sprintf("RSI indicators (14-Period): %s\n\n", formatFloatSlice(data.IntradaySeries.RSI14Values)))
				}
			}

			if indicators.EnableVolume && len(data.IntradaySeries.Volume) > 0 {
				sb.WriteString(fmt.Sprintf("Volume: %s\n\n", formatFloatSlice(data.IntradaySeries.Volume)))
			}

			if indicators.EnableATR {
				sb.WriteString(fmt.Sprintf("3m ATR (14-period): %.3f\n\n", data.IntradaySeries.ATR14))
			}
		}

		if data.LongerTermContext != nil && indicators.Klines.EnableMultiTimeframe {
			sb.WriteString(fmt.Sprintf("Longer-term context (%s timeframe):\n\n", indicators.Klines.LongerTimeframe))

			if indicators.EnableEMA {
				sb.WriteString(fmt.Sprintf("20-Period EMA: %.3f vs. 50-Period EMA: %.3f\n\n",
					data.LongerTermContext.EMA20, data.LongerTermContext.EMA50))
			}

			if indicators.EnableATR {
				sb.WriteString(fmt.Sprintf("3-Period ATR: %.3f vs. 14-Period ATR: %.3f\n\n",
					data.LongerTermContext.ATR3, data.LongerTermContext.ATR14))
			}

			if indicators.EnableVolume {
				sb.WriteString(fmt.Sprintf("Current Volume: %.3f vs. Average Volume: %.3f\n\n",
					data.LongerTermContext.CurrentVolume, data.LongerTermContext.AverageVolume))
			}

			if indicators.EnableMACD && len(data.LongerTermContext.MACDValues) > 0 {
				sb.WriteString(fmt.Sprintf("MACD indicators: %s\n\n", formatFloatSlice(data.LongerTermContext.MACDValues)))
			}

			if indicators.EnableRSI && len(data.LongerTermContext.RSI14Values) > 0 {
				sb.WriteString(fmt.Sprintf("RSI indicators (14-Period): %s\n\n", formatFloatSlice(data.LongerTermContext.RSI14Values)))
			}
		}
	}

	return sb.String()
}

func (e *StrategyEngine) formatTimeframeSeriesData(sb *strings.Builder, data *market.TimeframeSeriesData, indicators store.IndicatorConfig) {
	if len(data.Klines) > 0 {
		sb.WriteString("Time(UTC)      Open      High      Low       Close     Volume\n")
		for i, k := range data.Klines {
			t := time.Unix(k.Time/1000, 0).UTC()
			timeStr := t.Format("01-02 15:04")
			marker := ""
			if i == len(data.Klines)-1 {
				marker = "  <- current"
			}
			sb.WriteString(fmt.Sprintf("%-14s %-9.4f %-9.4f %-9.4f %-9.4f %-12.2f%s\n",
				timeStr, k.Open, k.High, k.Low, k.Close, k.Volume, marker))
		}
		sb.WriteString("\n")
	} else if len(data.MidPrices) > 0 {
		sb.WriteString(fmt.Sprintf("Mid prices: %s\n\n", formatFloatSlice(data.MidPrices)))
		if indicators.EnableVolume && len(data.Volume) > 0 {
			sb.WriteString(fmt.Sprintf("Volume: %s\n\n", formatFloatSlice(data.Volume)))
		}
	}

	if indicators.EnableEMA {
		if len(data.EMA20Values) > 0 {
			sb.WriteString(fmt.Sprintf("EMA20: %s\n", formatFloatSlice(data.EMA20Values)))
		}
		if len(data.EMA50Values) > 0 {
			sb.WriteString(fmt.Sprintf("EMA50: %s\n", formatFloatSlice(data.EMA50Values)))
		}
	}

	if indicators.EnableMACD && len(data.MACDValues) > 0 {
		sb.WriteString(fmt.Sprintf("MACD: %s\n", formatFloatSlice(data.MACDValues)))
	}

	if indicators.EnableRSI {
		if len(data.RSI7Values) > 0 {
			sb.WriteString(fmt.Sprintf("RSI7: %s\n", formatFloatSlice(data.RSI7Values)))
		}
		if len(data.RSI14Values) > 0 {
			sb.WriteString(fmt.Sprintf("RSI14: %s\n", formatFloatSlice(data.RSI14Values)))
		}
	}

	if indicators.EnableATR && data.ATR14 > 0 {
		sb.WriteString(fmt.Sprintf("ATR14: %.4f\n", data.ATR14))
	}

	if indicators.EnableBOLL && len(data.BOLLUpper) > 0 {
		sb.WriteString(fmt.Sprintf("BOLL Upper: %s\n", formatFloatSlice(data.BOLLUpper)))
		sb.WriteString(fmt.Sprintf("BOLL Middle: %s\n", formatFloatSlice(data.BOLLMiddle)))
		sb.WriteString(fmt.Sprintf("BOLL Lower: %s\n", formatFloatSlice(data.BOLLLower)))
	}

	sb.WriteString("\n")
}

func (e *StrategyEngine) formatQuantData(data *QuantData) string {
	if data == nil {
		return ""
	}

	indicators := e.config.Indicators
	if !indicators.EnableQuantOI && !indicators.EnableQuantNetflow {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📊 %s Quantitative Data:\n", data.Symbol))

	if len(data.PriceChange) > 0 {
		sb.WriteString("Price Change: ")
		timeframes := []string{"5m", "15m", "1h", "4h", "12h", "24h"}
		parts := []string{}
		for _, tf := range timeframes {
			if v, ok := data.PriceChange[tf]; ok {
				parts = append(parts, fmt.Sprintf("%s: %+.4f%%", tf, v*100))
			}
		}
		sb.WriteString(strings.Join(parts, " | "))
		sb.WriteString("\n")
	}

	if indicators.EnableQuantNetflow && data.Netflow != nil {
		sb.WriteString("Fund Flow (Netflow):\n")
		timeframes := []string{"5m", "15m", "1h", "4h", "12h", "24h"}

		if data.Netflow.Institution != nil {
			if data.Netflow.Institution.Future != nil && len(data.Netflow.Institution.Future) > 0 {
				sb.WriteString("  Institutional Futures:\n")
				for _, tf := range timeframes {
					if v, ok := data.Netflow.Institution.Future[tf]; ok {
						sb.WriteString(fmt.Sprintf("    %s: %s\n", tf, formatFlowValue(v)))
					}
				}
			}
			if data.Netflow.Institution.Spot != nil && len(data.Netflow.Institution.Spot) > 0 {
				sb.WriteString("  Institutional Spot:\n")
				for _, tf := range timeframes {
					if v, ok := data.Netflow.Institution.Spot[tf]; ok {
						sb.WriteString(fmt.Sprintf("    %s: %s\n", tf, formatFlowValue(v)))
					}
				}
			}
		}

		if data.Netflow.Personal != nil {
			if data.Netflow.Personal.Future != nil && len(data.Netflow.Personal.Future) > 0 {
				sb.WriteString("  Retail Futures:\n")
				for _, tf := range timeframes {
					if v, ok := data.Netflow.Personal.Future[tf]; ok {
						sb.WriteString(fmt.Sprintf("    %s: %s\n", tf, formatFlowValue(v)))
					}
				}
			}
			if data.Netflow.Personal.Spot != nil && len(data.Netflow.Personal.Spot) > 0 {
				sb.WriteString("  Retail Spot:\n")
				for _, tf := range timeframes {
					if v, ok := data.Netflow.Personal.Spot[tf]; ok {
						sb.WriteString(fmt.Sprintf("    %s: %s\n", tf, formatFlowValue(v)))
					}
				}
			}
		}
	}

	if indicators.EnableQuantOI && len(data.OI) > 0 {
		for exchange, oiData := range data.OI {
			if len(oiData.Delta) > 0 {
				sb.WriteString(fmt.Sprintf("Open Interest (%s):\n", exchange))
				for _, tf := range []string{"5m", "15m", "1h", "4h", "12h", "24h"} {
					if d, ok := oiData.Delta[tf]; ok {
						sb.WriteString(fmt.Sprintf("    %s: %+.4f%% (%s)\n", tf, d.OIDeltaPercent, formatFlowValue(d.OIDeltaValue)))
					}
				}
			}
		}
	}

	return sb.String()
}

func formatFlowValue(v float64) string {
	sign := ""
	if v >= 0 {
		sign = "+"
	}
	absV := v
	if absV < 0 {
		absV = -absV
	}
	if absV >= 1e9 {
		return fmt.Sprintf("%s%.2fB", sign, v/1e9)
	} else if absV >= 1e6 {
		return fmt.Sprintf("%s%.2fM", sign, v/1e6)
	} else if absV >= 1e3 {
		return fmt.Sprintf("%s%.2fK", sign, v/1e3)
	}
	return fmt.Sprintf("%s%.2f", sign, v)
}

func formatFloatSlice(values []float64) string {
	strValues := make([]string, len(values))
	for i, v := range values {
		strValues[i] = fmt.Sprintf("%.4f", v)
	}
	return "[" + strings.Join(strValues, ", ") + "]"
}

// ============================================================================
// AI Response Parsing
// ============================================================================

func parseFullDecisionResponse(aiResponse string, accountEquity float64, btcEthLeverage, altcoinLeverage int, btcEthPosRatio, altcoinPosRatio, minRiskRewardRatio float64) (*FullDecision, error) {
	cotTrace := extractCoTTrace(aiResponse)

	decisions, err := extractDecisions(aiResponse)
	if err != nil {
		return &FullDecision{
			CoTTrace:  cotTrace,
			Decisions: []Decision{},
		}, fmt.Errorf("failed to extract decisions: %w", err)
	}

	if err := validateDecisions(decisions, accountEquity, btcEthLeverage, altcoinLeverage, btcEthPosRatio, altcoinPosRatio, minRiskRewardRatio); err != nil {
		return &FullDecision{
			CoTTrace:  cotTrace,
			Decisions: decisions,
		}, fmt.Errorf("decision validation failed: %w", err)
	}

	return &FullDecision{
		CoTTrace:  cotTrace,
		Decisions: decisions,
	}, nil
}

func extractCoTTrace(response string) string {
	if match := reReasoningTag.FindStringSubmatch(response); match != nil && len(match) > 1 {
		logger.Infof("✓ Extracted reasoning chain using <reasoning> tag")
		return strings.TrimSpace(match[1])
	}

	if decisionIdx := strings.Index(response, "<decision>"); decisionIdx > 0 {
		logger.Infof("✓ Extracted content before <decision> tag as reasoning chain")
		return strings.TrimSpace(response[:decisionIdx])
	}

	jsonStart := strings.Index(response, "[")
	if jsonStart > 0 {
		logger.Infof("⚠️  Extracted reasoning chain using old format ([ character separator)")
		return strings.TrimSpace(response[:jsonStart])
	}

	return strings.TrimSpace(response)
}

func extractDecisions(response string) ([]Decision, error) {
	s := removeInvisibleRunes(response)
	s = strings.TrimSpace(s)
	s = fixMissingQuotes(s)

	var jsonPart string
	if match := reDecisionTag.FindStringSubmatch(s); match != nil && len(match) > 1 {
		jsonPart = strings.TrimSpace(match[1])
		logger.Infof("✓ Extracted JSON using <decision> tag")
	} else {
		jsonPart = s
		logger.Infof("⚠️  <decision> tag not found, searching JSON in full text")
	}

	jsonPart = fixMissingQuotes(jsonPart)

	if m := reJSONFence.FindStringSubmatch(jsonPart); m != nil && len(m) > 1 {
		jsonContent := strings.TrimSpace(m[1])
		jsonContent = compactArrayOpen(jsonContent)
		jsonContent = fixMissingQuotes(jsonContent)
		if err := validateJSONFormat(jsonContent); err != nil {
			return nil, fmt.Errorf("JSON format validation failed: %w\nJSON content: %s\nFull response:\n%s", err, jsonContent, response)
		}
		var decisions []Decision
		if err := json.Unmarshal([]byte(jsonContent), &decisions); err != nil {
			return nil, fmt.Errorf("JSON parsing failed: %w\nJSON content: %s", err, jsonContent)
		}
		return decisions, nil
	}

	jsonContent := strings.TrimSpace(reJSONArray.FindString(jsonPart))
	if jsonContent == "" {
		logger.Infof("⚠️  [SafeFallback] AI didn't output JSON decision, entering safe wait mode")

		cotSummary := jsonPart
		if len(cotSummary) > 240 {
			cotSummary = cotSummary[:240] + "..."
		}

		fallbackDecision := Decision{
			Symbol:    "ALL",
			Action:    "wait",
			Reasoning: fmt.Sprintf("Model didn't output structured JSON decision, entering safe wait; summary: %s", cotSummary),
		}

		return []Decision{fallbackDecision}, nil
	}

	jsonContent = compactArrayOpen(jsonContent)
	jsonContent = fixMissingQuotes(jsonContent)

	if err := validateJSONFormat(jsonContent); err != nil {
		return nil, fmt.Errorf("JSON format validation failed: %w\nJSON content: %s\nFull response:\n%s", err, jsonContent, response)
	}

	var decisions []Decision
	if err := json.Unmarshal([]byte(jsonContent), &decisions); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %w\nJSON content: %s", err, jsonContent)
	}

	return decisions, nil
}

func fixMissingQuotes(jsonStr string) string {
	jsonStr = strings.ReplaceAll(jsonStr, "\u201c", "\"")
	jsonStr = strings.ReplaceAll(jsonStr, "\u201d", "\"")
	jsonStr = strings.ReplaceAll(jsonStr, "\u2018", "'")
	jsonStr = strings.ReplaceAll(jsonStr, "\u2019", "'")

	jsonStr = strings.ReplaceAll(jsonStr, "［", "[")
	jsonStr = strings.ReplaceAll(jsonStr, "］", "]")
	jsonStr = strings.ReplaceAll(jsonStr, "｛", "{")
	jsonStr = strings.ReplaceAll(jsonStr, "｝", "}")
	jsonStr = strings.ReplaceAll(jsonStr, "：", ":")
	jsonStr = strings.ReplaceAll(jsonStr, "，", ",")

	jsonStr = strings.ReplaceAll(jsonStr, "【", "[")
	jsonStr = strings.ReplaceAll(jsonStr, "】", "]")
	jsonStr = strings.ReplaceAll(jsonStr, "〔", "[")
	jsonStr = strings.ReplaceAll(jsonStr, "〕", "]")
	jsonStr = strings.ReplaceAll(jsonStr, "、", ",")

	jsonStr = strings.ReplaceAll(jsonStr, "　", " ")

	return jsonStr
}

func validateJSONFormat(jsonStr string) error {
	trimmed := strings.TrimSpace(jsonStr)

	if !reArrayHead.MatchString(trimmed) {
		if strings.HasPrefix(trimmed, "[") && !strings.Contains(trimmed[:min(20, len(trimmed))], "{") {
			return fmt.Errorf("not a valid decision array (must contain objects {}), actual content: %s", trimmed[:min(50, len(trimmed))])
		}
		return fmt.Errorf("JSON must start with [{ (whitespace allowed), actual: %s", trimmed[:min(20, len(trimmed))])
	}

	if strings.Contains(jsonStr, "~") {
		return fmt.Errorf("JSON cannot contain range symbol ~, all numbers must be precise single values")
	}

	for i := 0; i < len(jsonStr)-4; i++ {
		if jsonStr[i] >= '0' && jsonStr[i] <= '9' &&
			jsonStr[i+1] == ',' &&
			jsonStr[i+2] >= '0' && jsonStr[i+2] <= '9' &&
			jsonStr[i+3] >= '0' && jsonStr[i+3] <= '9' &&
			jsonStr[i+4] >= '0' && jsonStr[i+4] <= '9' {
			return fmt.Errorf("JSON numbers cannot contain thousand separator comma, found: %s", jsonStr[i:min(i+10, len(jsonStr))])
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func removeInvisibleRunes(s string) string {
	return reInvisibleRunes.ReplaceAllString(s, "")
}

func compactArrayOpen(s string) string {
	return reArrayOpenSpace.ReplaceAllString(strings.TrimSpace(s), "[{")
}

// ============================================================================
// Decision Validation
// ============================================================================

func validateDecisions(decisions []Decision, accountEquity float64, btcEthLeverage, altcoinLeverage int, btcEthPosRatio, altcoinPosRatio, minRiskRewardRatio float64) error {
	for i := range decisions {
		if err := validateDecision(&decisions[i], accountEquity, btcEthLeverage, altcoinLeverage, btcEthPosRatio, altcoinPosRatio, minRiskRewardRatio); err != nil {
			return fmt.Errorf("decision #%d validation failed: %w", i+1, err)
		}
	}
	return nil
}

func validateDecision(d *Decision, accountEquity float64, btcEthLeverage, altcoinLeverage int, btcEthPosRatio, altcoinPosRatio, minRiskRewardRatio float64) error {
	if minRiskRewardRatio <= 0 {
		minRiskRewardRatio = defaultMinRiskRewardRatio
	}

	validActions := map[string]bool{
		"open_long":   true,
		"open_short":  true,
		"close_long":  true,
		"close_short": true,
		"hold":        true,
		"wait":        true,
	}

	if !validActions[d.Action] {
		return fmt.Errorf("invalid action: %s", d.Action)
	}

	if d.Action == "open_long" || d.Action == "open_short" {
		maxLeverage := altcoinLeverage
		posRatio := altcoinPosRatio
		maxPositionValue := accountEquity * posRatio
		if d.Symbol == "BTCUSDT" || d.Symbol == "ETHUSDT" {
			maxLeverage = btcEthLeverage
			posRatio = btcEthPosRatio
			maxPositionValue = accountEquity * posRatio
		}

		if d.Leverage <= 0 {
			return fmt.Errorf("leverage must be greater than 0: %d", d.Leverage)
		}
		if d.Leverage > maxLeverage {
			logger.Infof("⚠️  [Leverage Fallback] %s leverage exceeded (%dx > %dx), auto-adjusting to limit %dx",
				d.Symbol, d.Leverage, maxLeverage, maxLeverage)
			d.Leverage = maxLeverage
		}
		if d.PositionSizeUSD <= 0 {
			return fmt.Errorf("position size must be greater than 0: %.2f", d.PositionSizeUSD)
		}

		const minPositionSizeGeneral = 12.0
		const minPositionSizeBTCETH = 60.0

		if d.Symbol == "BTCUSDT" || d.Symbol == "ETHUSDT" {
			if d.PositionSizeUSD < minPositionSizeBTCETH {
				return fmt.Errorf("%s opening amount too small (%.2f USDT), must be ≥%.2f USDT", d.Symbol, d.PositionSizeUSD, minPositionSizeBTCETH)
			}
		} else {
			if d.PositionSizeUSD < minPositionSizeGeneral {
				return fmt.Errorf("opening amount too small (%.2f USDT), must be ≥%.2f USDT", d.PositionSizeUSD, minPositionSizeGeneral)
			}
		}

		tolerance := maxPositionValue * 0.01
		if d.PositionSizeUSD > maxPositionValue+tolerance {
			if d.Symbol == "BTCUSDT" || d.Symbol == "ETHUSDT" {
				return fmt.Errorf("BTC/ETH single coin position value cannot exceed %.0f USDT (%.1fx account equity), actual: %.0f", maxPositionValue, posRatio, d.PositionSizeUSD)
			} else {
				return fmt.Errorf("altcoin single coin position value cannot exceed %.0f USDT (%.1fx account equity), actual: %.0f", maxPositionValue, posRatio, d.PositionSizeUSD)
			}
		}
		if d.StopLoss <= 0 || d.TakeProfit <= 0 {
			return fmt.Errorf("stop loss and take profit must be greater than 0")
		}

		if d.Action == "open_long" {
			if d.StopLoss >= d.TakeProfit {
				return fmt.Errorf("for long positions, stop loss price must be less than take profit price")
			}
		} else {
			if d.StopLoss <= d.TakeProfit {
				return fmt.Errorf("for short positions, stop loss price must be greater than take profit price")
			}
		}

		var entryPrice float64
		if d.Action == "open_long" {
			entryPrice = d.StopLoss + (d.TakeProfit-d.StopLoss)*0.2
		} else {
			entryPrice = d.StopLoss - (d.StopLoss-d.TakeProfit)*0.2
		}

		var riskPercent, rewardPercent, riskRewardRatio float64
		if d.Action == "open_long" {
			riskPercent = (entryPrice - d.StopLoss) / entryPrice * 100
			rewardPercent = (d.TakeProfit - entryPrice) / entryPrice * 100
			if riskPercent > 0 {
				riskRewardRatio = rewardPercent / riskPercent
			}
		} else {
			riskPercent = (d.StopLoss - entryPrice) / entryPrice * 100
			rewardPercent = (entryPrice - d.TakeProfit) / entryPrice * 100
			if riskPercent > 0 {
				riskRewardRatio = rewardPercent / riskPercent
			}
		}

		if riskRewardRatio < minRiskRewardRatio {
			return fmt.Errorf("risk/reward ratio too low (%.2f:1), must be ≥%.1f:1 [risk: %.2f%% reward: %.2f%%] [stop loss: %.2f take profit: %.2f]",
				riskRewardRatio, minRiskRewardRatio, riskPercent, rewardPercent, d.StopLoss, d.TakeProfit)
		}
	}

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// detectLanguage detects language from text content
// Returns LangChinese if text contains Chinese characters, otherwise LangEnglish
func detectLanguage(text string) Language {
	for _, r := range text {
		if r >= 0x4E00 && r <= 0x9FFF {
			return LangChinese
		}
	}
	return LangEnglish
}
