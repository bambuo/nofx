package kernel

import (
	"fmt"
	"nofx/store"
	"testing"
	"time"
)

// TestLeverageFallback tests automatic correction when leverage exceeds limit
func TestLeverageFallback(t *testing.T) {
	tests := []struct {
		name            string
		decision        Decision
		accountEquity   float64
		btcEthLeverage  int
		altcoinLeverage int
		wantLeverage    int // Expected leverage after correction
		wantError       bool
	}{
		{
			name: "Altcoin leverage exceeded - auto-correct to limit",
			decision: Decision{
				Symbol:          "SOLUSDT",
				Action:          "open_long",
				Leverage:        20, // Exceeds limit
				PositionSizeUSD: 100,
				StopLoss:        50,
				TakeProfit:      200,
			},
			accountEquity:   100,
			btcEthLeverage:  10,
			altcoinLeverage: 5, // Limit 5x
			wantLeverage:    5, // Should be corrected to 5
			wantError:       false,
		},
		{
			name: "BTC leverage exceeded - auto-correct to limit",
			decision: Decision{
				Symbol:          "BTCUSDT",
				Action:          "open_long",
				Leverage:        20, // Exceeds limit
				PositionSizeUSD: 1000,
				StopLoss:        90000,
				TakeProfit:      110000,
			},
			accountEquity:   100,
			btcEthLeverage:  10, // Limit 10x
			altcoinLeverage: 5,
			wantLeverage:    10, // Should be corrected to 10
			wantError:       false,
		},
		{
			name: "Leverage within limit - no correction",
			decision: Decision{
				Symbol:          "ETHUSDT",
				Action:          "open_short",
				Leverage:        5, // Not exceeded
				PositionSizeUSD: 500,
				StopLoss:        4000,
				TakeProfit:      3000,
			},
			accountEquity:   100,
			btcEthLeverage:  10,
			altcoinLeverage: 5,
			wantLeverage:    5, // Stays unchanged
			wantError:       false,
		},
		{
			name: "Leverage is 0 - should error",
			decision: Decision{
				Symbol:          "SOLUSDT",
				Action:          "open_long",
				Leverage:        0, // Invalid
				PositionSizeUSD: 100,
				StopLoss:        50,
				TakeProfit:      200,
			},
			accountEquity:   100,
			btcEthLeverage:  10,
			altcoinLeverage: 5,
			wantLeverage:    0,
			wantError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use default position value ratios for testing (10x for BTC/ETH, 1.5x for altcoins)
			err := validateDecision(&tt.decision, tt.accountEquity, tt.btcEthLeverage, tt.altcoinLeverage, 10.0, 1.5, 3.0)

			// Check error status
			if (err != nil) != tt.wantError {
				t.Errorf("validateDecision() error = %v, wantError %v", err, tt.wantError)
				return
			}

			// If shouldn't error, check if leverage was correctly corrected
			if !tt.wantError && tt.decision.Leverage != tt.wantLeverage {
				t.Errorf("Leverage not corrected: got %d, want %d", tt.decision.Leverage, tt.wantLeverage)
			}
		})
	}
}

func TestValidateDecisionUsesConfiguredRiskRewardRatio(t *testing.T) {
	decision := Decision{
		Symbol:          "SOLUSDT",
		Action:          "open_long",
		Leverage:        3,
		PositionSizeUSD: 100,
		StopLoss:        90,
		TakeProfit:      110,
	}

	if err := validateDecision(&decision, 100, 10, 5, 10.0, 1.5, 5.0); err == nil {
		t.Fatal("expected configured 5.0 risk/reward ratio to reject this decision")
	}

	if err := validateDecision(&decision, 100, 10, 5, 10.0, 1.5, 3.0); err != nil {
		t.Fatalf("expected configured 3.0 risk/reward ratio to accept this decision: %v", err)
	}
}

func TestCoinSourceProviders(t *testing.T) {
	tests := []struct {
		name      string
		providers []string
		apiKey    string
		want      []string
	}{
		{
			name: "default chain is binance only",
			want: []string{coinProviderBinance},
		},
		{
			name:      "explicit binance only",
			providers: []string{coinProviderBinance},
			want:      []string{coinProviderBinance},
		},
		{
			name:      "skips nofxos without custom api key",
			providers: []string{"unknown", coinProviderBinance, coinProviderBinance, coinProviderNofxOS},
			want:      []string{coinProviderBinance},
		},
		{
			name:      "allows nofxos with custom api key",
			providers: []string{coinProviderNofxOS, coinProviderBinance},
			apiKey:    "custom-key",
			want:      []string{coinProviderNofxOS, coinProviderBinance},
		},
		{
			name:      "falls back when all configured providers are unknown",
			providers: []string{"unknown"},
			want:      []string{coinProviderBinance},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &StrategyEngine{
				config: &store.StrategyConfig{
					CoinSource: store.CoinSourceConfig{
						Providers: tt.providers,
					},
					Indicators: store.IndicatorConfig{
						NofxOSAPIKey: tt.apiKey,
					},
				},
			}
			got := engine.coinSourceProviders()
			if len(got) != len(tt.want) {
				t.Fatalf("coinSourceProviders() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("coinSourceProviders() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMarketDataFailureQuarantine(t *testing.T) {
	engine := &StrategyEngine{}
	engine.quarantineSymbol("KORUUSDT")

	if !engine.isSymbolQuarantined("KORUUSDT") {
		t.Fatal("expected symbol to be quarantined")
	}

	engine.quarantinedSymbols["KORUUSDT"] = time.Now().Add(-time.Second)
	if engine.isSymbolQuarantined("KORUUSDT") {
		t.Fatal("expected expired quarantine to be ignored")
	}
}

func TestShouldQuarantineMarketDataError(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{err: fmt.Errorf("KORUUSDT data is stale, possible cache failure"), want: true},
		{err: fmt.Errorf("Primary timeframe 5m K-line data is empty"), want: true},
		{err: fmt.Errorf("temporary http timeout"), want: false},
		{err: nil, want: false},
	}

	for _, tt := range tests {
		if got := shouldQuarantineMarketDataError(tt.err); got != tt.want {
			t.Fatalf("shouldQuarantineMarketDataError(%v) = %v, want %v", tt.err, got, tt.want)
		}
	}
}

// contains checks if string contains substring (helper function)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
