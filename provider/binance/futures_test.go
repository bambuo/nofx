package binance

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestIsUSDTPerpSymbol(t *testing.T) {
	tests := []struct {
		symbol string
		want   bool
	}{
		{symbol: "BTCUSDT", want: true},
		{symbol: "ethusdt", want: true},
		{symbol: "BTCUSD", want: false},
		{symbol: "USDCUSDT", want: false},
	}

	for _, tt := range tests {
		if got := isUSDTPerpSymbol(tt.symbol); got != tt.want {
			t.Fatalf("isUSDTPerpSymbol(%q) = %v, want %v", tt.symbol, got, tt.want)
		}
	}
}

func TestTakeOISymbols(t *testing.T) {
	changes := []oiChange{
		{Symbol: "BTCUSDT", CurrentOIValue: minCandidateOIValue + 1, OIDeltaValue: 100},
		{Symbol: "ETHUSDT", CurrentOIValue: minCandidateOIValue + 1, OIDeltaValue: -50},
		{Symbol: "SOLUSDT", CurrentOIValue: minCandidateOIValue + 1, OIDeltaValue: 25},
		{Symbol: "XRPUSDT", CurrentOIValue: minCandidateOIValue + 1, OIDeltaValue: -10},
	}

	increasing := takeOISymbols(changes, 2, true)
	if !reflect.DeepEqual(increasing, []string{"BTCUSDT", "SOLUSDT"}) {
		t.Fatalf("increasing symbols = %v", increasing)
	}

	decreasing := takeOISymbols(changes, 2, false)
	if !reflect.DeepEqual(decreasing, []string{"ETHUSDT", "XRPUSDT"}) {
		t.Fatalf("decreasing symbols = %v", decreasing)
	}
}

func TestTakeOISymbolsFiltersLowOI(t *testing.T) {
	changes := []oiChange{
		{Symbol: "LOWUSDT", CurrentOIValue: minCandidateOIValue - 1, OIDeltaValue: 1000},
		{Symbol: "BTCUSDT", CurrentOIValue: minCandidateOIValue + 1, OIDeltaValue: 100},
	}

	got := takeOISymbols(changes, 2, true)
	if !reflect.DeepEqual(got, []string{"BTCUSDT"}) {
		t.Fatalf("symbols = %v", got)
	}
}

func TestTopScoredSymbolsFromOIChangesFiltersLowOI(t *testing.T) {
	changes := []oiChange{
		{
			Symbol:             "LOWUSDT",
			CurrentOIValue:     minCandidateOIValue - 1,
			OIDeltaValue:       1000,
			PriceChangePercent: 20,
			QuoteVolume:        100_000_000,
		},
	}

	got := topScoredSymbolsFromOIChanges(changes, 10)
	if len(got) != 0 {
		t.Fatalf("symbols = %v, want empty list for low OI candidates", got)
	}
}

func TestGetTopVolumeSymbolsFallback(t *testing.T) {
	client := &FuturesClient{
		tickersCache: cachedTickers{
			fetchedAt: time.Now(),
			values: []ticker24h{
				{Symbol: "LOWUSDT", LastPrice: "1", QuoteVolume: "30000000", PriceChangePercent: "1"},
				{Symbol: "BTCUSDT", LastPrice: "100000", QuoteVolume: "100000000", PriceChangePercent: "2"},
			},
		},
	}

	got, err := client.getTopVolumeSymbols(2)
	if err != nil {
		t.Fatalf("getTopVolumeSymbols() error = %v", err)
	}
	if !reflect.DeepEqual(got, []string{"BTCUSDT", "LOWUSDT"}) {
		t.Fatalf("symbols = %v", got)
	}
}

func TestGetTopVolumeSymbolsFiltersByCurrentOI(t *testing.T) {
	client := &FuturesClient{
		currentOIValueFunc: func(symbol string, lastPrice float64) (float64, error) {
			switch symbol {
			case "BTCUSDT":
				return 20_000_000, nil
			case "LOWUSDT":
				return 100, nil
			default:
				return 0, fmt.Errorf("unexpected symbol: %s", symbol)
			}
		},
		tickersCache: cachedTickers{
			fetchedAt: time.Now(),
			values: []ticker24h{
				{Symbol: "LOWUSDT", LastPrice: "1", QuoteVolume: "300000000", PriceChangePercent: "20"},
				{Symbol: "BTCUSDT", LastPrice: "100000", QuoteVolume: "100000000", PriceChangePercent: "2"},
			},
		},
	}

	got, err := client.getTopVolumeSymbols(2)
	if err != nil {
		t.Fatalf("getTopVolumeSymbols() error = %v", err)
	}
	if !reflect.DeepEqual(got, []string{"BTCUSDT"}) {
		t.Fatalf("symbols = %v", got)
	}
}

func TestParsePositiveFloat(t *testing.T) {
	value, ok := parsePositiveFloat("123.45")
	if !ok || value != 123.45 {
		t.Fatalf("parsePositiveFloat positive = %v, %v", value, ok)
	}

	if _, ok := parsePositiveFloat("0"); ok {
		t.Fatal("parsePositiveFloat(0) should be false")
	}

	if _, ok := parsePositiveFloat("not-a-number"); ok {
		t.Fatal("parsePositiveFloat(invalid) should be false")
	}
}
