package binance

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/url"
	"nofx/security"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultFuturesBaseURL = "https://fapi.binance.com"
	defaultTimeout        = 15 * time.Second
	defaultOIPeriod       = "1h"
	defaultOILookback     = 2
	defaultOIScanLimit    = 60
	minCandidateOIValue   = 15_000_000.0
	minFallbackQuoteValue = 25_000_000.0
	tickerCacheTTL        = 30 * time.Second
	oiChangeCacheTTL      = 2 * time.Minute
)

// FuturesClient reads public Binance USD-M futures market data.
type FuturesClient struct {
	BaseURL string
	Timeout time.Duration

	currentOIValueFunc func(symbol string, lastPrice float64) (float64, error)

	mu             sync.Mutex
	tickersCache   cachedTickers
	oiChangesCache cachedOIChanges
}

type cachedTickers struct {
	values    []ticker24h
	fetchedAt time.Time
}

type cachedOIChanges struct {
	values    []oiChange
	fetchedAt time.Time
}

// NewFuturesClient creates a client for public Binance USD-M futures endpoints.
func NewFuturesClient(baseURL string) *FuturesClient {
	if baseURL == "" {
		baseURL = DefaultFuturesBaseURL
	}
	return &FuturesClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Timeout: defaultTimeout,
	}
}

type ticker24h struct {
	Symbol             string `json:"symbol"`
	PriceChangePercent string `json:"priceChangePercent"`
	LastPrice          string `json:"lastPrice"`
	QuoteVolume        string `json:"quoteVolume"`
	Count              int64  `json:"count"`
}

type oiHistoryPoint struct {
	Symbol               string `json:"symbol"`
	SumOpenInterest      string `json:"sumOpenInterest"`
	SumOpenInterestValue string `json:"sumOpenInterestValue"`
	Timestamp            int64  `json:"timestamp"`
}

type scoredTicker struct {
	Symbol             string
	QuoteVolume        float64
	LastPrice          float64
	CurrentOIValue     float64
	PriceChangePercent float64
	Score              float64
}

type openInterestSnapshot struct {
	OpenInterest string `json:"openInterest"`
	Symbol       string `json:"symbol"`
	Time         int64  `json:"time"`
}

type oiChange struct {
	Symbol             string
	CurrentOIValue     float64
	OIDeltaValue       float64
	OIDeltaPercent     float64
	PriceChangePercent float64
	QuoteVolume        float64
}

// GetTopScoredSymbols returns a liquid, momentum-aware futures coin list.
func (c *FuturesClient) GetTopScoredSymbols(limit int) ([]string, error) {
	if limit <= 0 {
		limit = 30
	}

	changes, err := c.getOIChanges()
	if err == nil {
		symbols := topScoredSymbolsFromOIChanges(changes, limit)
		if len(symbols) > 0 {
			return symbols, nil
		}
	}

	symbols, fallbackErr := c.getTopVolumeSymbols(limit)
	if fallbackErr != nil {
		if err != nil {
			return nil, fmt.Errorf("%w; Binance volume fallback failed: %v", err, fallbackErr)
		}
		return nil, fallbackErr
	}
	return symbols, nil
}

func topScoredSymbolsFromOIChanges(changes []oiChange, limit int) []string {
	scored := make([]scoredTicker, 0, len(changes))
	for _, change := range changes {
		if change.CurrentOIValue < minCandidateOIValue {
			continue
		}
		// Favor symbols with both deep OI and high turnover, then add a modest momentum component.
		score := math.Log10(change.QuoteVolume+1) + math.Log10(change.CurrentOIValue+1)*0.35 + math.Abs(change.PriceChangePercent)*0.08
		scored = append(scored, scoredTicker{
			Symbol:             change.Symbol,
			QuoteVolume:        change.QuoteVolume,
			CurrentOIValue:     change.CurrentOIValue,
			PriceChangePercent: change.PriceChangePercent,
			Score:              score,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].Score == scored[j].Score {
			return scored[i].CurrentOIValue > scored[j].CurrentOIValue
		}
		return scored[i].Score > scored[j].Score
	})

	if len(scored) < limit {
		limit = len(scored)
	}
	symbols := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		symbols = append(symbols, scored[i].Symbol)
	}
	return symbols
}

func (c *FuturesClient) getTopVolumeSymbols(limit int) ([]string, error) {
	tickers, err := c.getUSDTPerpTickers()
	if err != nil {
		return nil, err
	}

	scored := make([]scoredTicker, 0, len(tickers))
	for _, ticker := range tickers {
		quoteVolume, ok := parsePositiveFloat(ticker.QuoteVolume)
		if !ok || quoteVolume < minFallbackQuoteValue {
			continue
		}
		lastPrice, ok := parsePositiveFloat(ticker.LastPrice)
		if !ok {
			continue
		}
		priceChangePercent, _ := strconv.ParseFloat(ticker.PriceChangePercent, 64)
		score := math.Log10(quoteVolume+1) + math.Abs(priceChangePercent)*0.08
		scored = append(scored, scoredTicker{
			Symbol:             normalizeSymbol(ticker.Symbol),
			QuoteVolume:        quoteVolume,
			LastPrice:          lastPrice,
			PriceChangePercent: priceChangePercent,
			Score:              score,
		})
	}

	if len(scored) == 0 {
		for _, ticker := range tickers {
			quoteVolume, ok := parsePositiveFloat(ticker.QuoteVolume)
			if !ok {
				continue
			}
			lastPrice, ok := parsePositiveFloat(ticker.LastPrice)
			if !ok {
				continue
			}
			priceChangePercent, _ := strconv.ParseFloat(ticker.PriceChangePercent, 64)
			scored = append(scored, scoredTicker{
				Symbol:             normalizeSymbol(ticker.Symbol),
				QuoteVolume:        quoteVolume,
				LastPrice:          lastPrice,
				PriceChangePercent: priceChangePercent,
				Score:              math.Log10(quoteVolume + 1),
			})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].Score == scored[j].Score {
			return scored[i].QuoteVolume > scored[j].QuoteVolume
		}
		return scored[i].Score > scored[j].Score
	})

	verified := c.filterScoredByCurrentOI(scored, limit)
	if len(verified) > 0 {
		return verified, nil
	}

	if len(scored) < limit {
		limit = len(scored)
	}
	symbols := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		symbols = append(symbols, scored[i].Symbol)
	}
	return symbols, nil
}

func (c *FuturesClient) filterScoredByCurrentOI(scored []scoredTicker, limit int) []string {
	if limit <= 0 {
		limit = 30
	}

	symbols := make([]string, 0, limit)
	scanned := 0
	for _, ticker := range scored {
		if scanned >= defaultOIScanLimit {
			break
		}
		scanned++

		oiValue, err := c.fetchCurrentOIValue(ticker.Symbol, ticker.LastPrice)
		if err != nil || oiValue < minCandidateOIValue {
			continue
		}
		symbols = append(symbols, ticker.Symbol)
		if len(symbols) >= limit {
			break
		}
	}
	return symbols
}

func (c *FuturesClient) fetchCurrentOIValue(symbol string, lastPrice float64) (float64, error) {
	if c.currentOIValueFunc != nil {
		return c.currentOIValueFunc(symbol, lastPrice)
	}
	return c.getCurrentOIValue(symbol, lastPrice)
}

func (c *FuturesClient) getCurrentOIValue(symbol string, lastPrice float64) (float64, error) {
	if lastPrice <= 0 {
		return 0, fmt.Errorf("last price is not positive")
	}

	query := url.Values{}
	query.Set("symbol", symbol)

	body, err := c.doGET("/fapi/v1/openInterest", query)
	if err != nil {
		return 0, err
	}

	var snapshot openInterestSnapshot
	if err := json.Unmarshal(body, &snapshot); err != nil {
		return 0, fmt.Errorf("parse Binance current OI for %s: %w", symbol, err)
	}

	openInterest, ok := parsePositiveFloat(snapshot.OpenInterest)
	if !ok {
		return 0, fmt.Errorf("invalid current OI for %s", symbol)
	}
	return openInterest * lastPrice, nil
}

// GetOIIncreasingSymbols returns symbols with the strongest recent OI increase.
func (c *FuturesClient) GetOIIncreasingSymbols(limit int) ([]string, error) {
	changes, err := c.getOIChanges()
	if err != nil {
		return nil, err
	}
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].OIDeltaValue == changes[j].OIDeltaValue {
			return changes[i].QuoteVolume > changes[j].QuoteVolume
		}
		return changes[i].OIDeltaValue > changes[j].OIDeltaValue
	})
	return takeOISymbols(changes, limit, true), nil
}

// GetOIDecreasingSymbols returns symbols with the strongest recent OI decrease.
func (c *FuturesClient) GetOIDecreasingSymbols(limit int) ([]string, error) {
	changes, err := c.getOIChanges()
	if err != nil {
		return nil, err
	}
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].OIDeltaValue == changes[j].OIDeltaValue {
			return changes[i].QuoteVolume > changes[j].QuoteVolume
		}
		return changes[i].OIDeltaValue < changes[j].OIDeltaValue
	})
	return takeOISymbols(changes, limit, false), nil
}

func (c *FuturesClient) getUSDTPerpTickers() ([]ticker24h, error) {
	c.mu.Lock()
	if time.Since(c.tickersCache.fetchedAt) < tickerCacheTTL && len(c.tickersCache.values) > 0 {
		tickers := append([]ticker24h(nil), c.tickersCache.values...)
		c.mu.Unlock()
		return tickers, nil
	}
	c.mu.Unlock()

	body, err := c.doGET("/fapi/v1/ticker/24hr", nil)
	if err != nil {
		return nil, err
	}

	var tickers []ticker24h
	if err := json.Unmarshal(body, &tickers); err != nil {
		return nil, fmt.Errorf("parse Binance 24h tickers: %w", err)
	}

	filtered := make([]ticker24h, 0, len(tickers))
	for _, ticker := range tickers {
		if isUSDTPerpSymbol(ticker.Symbol) {
			filtered = append(filtered, ticker)
		}
	}
	c.mu.Lock()
	c.tickersCache = cachedTickers{
		values:    append([]ticker24h(nil), filtered...),
		fetchedAt: time.Now(),
	}
	c.mu.Unlock()
	return filtered, nil
}

func (c *FuturesClient) getOIChanges() ([]oiChange, error) {
	c.mu.Lock()
	if time.Since(c.oiChangesCache.fetchedAt) < oiChangeCacheTTL && len(c.oiChangesCache.values) > 0 {
		changes := append([]oiChange(nil), c.oiChangesCache.values...)
		c.mu.Unlock()
		return changes, nil
	}
	c.mu.Unlock()

	tickers, err := c.getUSDTPerpTickers()
	if err != nil {
		return nil, err
	}

	sort.Slice(tickers, func(i, j int) bool {
		left, _ := strconv.ParseFloat(tickers[i].QuoteVolume, 64)
		right, _ := strconv.ParseFloat(tickers[j].QuoteVolume, 64)
		return left > right
	})
	if len(tickers) > defaultOIScanLimit {
		tickers = tickers[:defaultOIScanLimit]
	}

	changes := make([]oiChange, 0, len(tickers))
	for _, ticker := range tickers {
		points, err := c.getOpenInterestHistory(ticker.Symbol, defaultOIPeriod, defaultOILookback)
		if err != nil || len(points) < 2 {
			continue
		}

		sort.Slice(points, func(i, j int) bool {
			return points[i].Timestamp < points[j].Timestamp
		})
		prev := points[len(points)-2]
		curr := points[len(points)-1]
		prevValue, prevOK := parsePositiveFloat(prev.SumOpenInterestValue)
		currValue, currOK := parsePositiveFloat(curr.SumOpenInterestValue)
		if !prevOK || !currOK {
			continue
		}

		priceChangePercent, _ := strconv.ParseFloat(ticker.PriceChangePercent, 64)
		quoteVolume, _ := strconv.ParseFloat(ticker.QuoteVolume, 64)
		changes = append(changes, oiChange{
			Symbol:             normalizeSymbol(ticker.Symbol),
			CurrentOIValue:     currValue,
			OIDeltaValue:       currValue - prevValue,
			OIDeltaPercent:     (currValue - prevValue) / prevValue * 100,
			PriceChangePercent: priceChangePercent,
			QuoteVolume:        quoteVolume,
		})
	}
	c.mu.Lock()
	c.oiChangesCache = cachedOIChanges{
		values:    append([]oiChange(nil), changes...),
		fetchedAt: time.Now(),
	}
	c.mu.Unlock()
	return changes, nil
}

func (c *FuturesClient) getOpenInterestHistory(symbol, period string, limit int) ([]oiHistoryPoint, error) {
	query := url.Values{}
	query.Set("pair", symbol)
	query.Set("contractType", "PERPETUAL")
	query.Set("period", period)
	query.Set("limit", strconv.Itoa(limit))

	body, err := c.doGET("/futures/data/openInterestHist", query)
	if err != nil {
		return nil, err
	}

	var points []oiHistoryPoint
	if err := json.Unmarshal(body, &points); err != nil {
		return nil, fmt.Errorf("parse Binance OI history for %s: %w", symbol, err)
	}
	return points, nil
}

func (c *FuturesClient) doGET(path string, query url.Values) ([]byte, error) {
	rawURL := c.BaseURL + path
	if len(query) > 0 {
		rawURL += "?" + query.Encode()
	}

	resp, err := security.SafeGet(rawURL, c.Timeout)
	if err != nil {
		return nil, fmt.Errorf("request Binance %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read Binance response from %s: %w", path, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Binance %s returned HTTP %d: %s", path, resp.StatusCode, string(body))
	}
	return body, nil
}

func takeOISymbols(changes []oiChange, limit int, increasing bool) []string {
	if limit <= 0 {
		limit = 10
	}
	symbols := make([]string, 0, limit)
	for _, change := range changes {
		if change.CurrentOIValue < minCandidateOIValue {
			continue
		}
		if increasing && change.OIDeltaValue <= 0 {
			continue
		}
		if !increasing && change.OIDeltaValue >= 0 {
			continue
		}
		symbols = append(symbols, change.Symbol)
		if len(symbols) >= limit {
			break
		}
	}
	return symbols
}

func parsePositiveFloat(value string) (float64, bool) {
	parsed, err := strconv.ParseFloat(value, 64)
	return parsed, err == nil && parsed > 0
}

func normalizeSymbol(symbol string) string {
	return strings.ToUpper(strings.TrimSpace(symbol))
}

func isUSDTPerpSymbol(symbol string) bool {
	symbol = normalizeSymbol(symbol)
	if !strings.HasSuffix(symbol, "USDT") {
		return false
	}
	// Exclude common stablecoin pairs that are poor strategy candidates.
	switch symbol {
	case "USDCUSDT", "BUSDUSDT", "FDUSDUSDT", "TUSDUSDT", "USDPUSDT":
		return false
	default:
		return true
	}
}
