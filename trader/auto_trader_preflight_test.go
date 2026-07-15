package trader

import (
	"math"
	"nofx/kernel"
	"nofx/store"
	"testing"
)

func TestCalculateRiskReward(t *testing.T) {
	tests := []struct {
		name       string
		side       string
		entry      float64
		stopLoss   float64
		takeProfit float64
		wantRisk   float64
		wantReward float64
		wantRR     float64
	}{
		{
			name:       "long",
			side:       "LONG",
			entry:      100,
			stopLoss:   95,
			takeProfit: 115,
			wantRisk:   5,
			wantReward: 15,
			wantRR:     3,
		},
		{
			name:       "short",
			side:       "SHORT",
			entry:      100,
			stopLoss:   105,
			takeProfit: 85,
			wantRisk:   5,
			wantReward: 15,
			wantRR:     3,
		},
		{
			name:       "invalid long target",
			side:       "LONG",
			entry:      100,
			stopLoss:   105,
			takeProfit: 115,
			wantRisk:   -5,
			wantReward: 15,
			wantRR:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			risk, reward, rr := calculateRiskReward(tt.side, tt.entry, tt.stopLoss, tt.takeProfit)
			assertFloatClose(t, "risk", risk, tt.wantRisk)
			assertFloatClose(t, "reward", reward, tt.wantReward)
			assertFloatClose(t, "risk reward", rr, tt.wantRR)
		})
	}
}

func TestBuildOpenOrderPreflight(t *testing.T) {
	at := &AutoTrader{
		config: AutoTraderConfig{
			StrategyConfig: &store.StrategyConfig{
				RiskControl: store.RiskControlConfig{
					MaxPositions:                 3,
					BTCETHMaxPositionValueRatio:  5,
					AltcoinMaxPositionValueRatio: 1,
					MaxMarginUsage:               0.9,
					MinPositionSize:              12,
					MinRiskRewardRatio:           3,
				},
			},
		},
	}

	decision := &kernel.Decision{
		Symbol:          "SOLUSDT",
		Action:          "open_long",
		Leverage:        5,
		PositionSizeUSD: 50,
		StopLoss:        95,
		TakeProfit:      115,
		Confidence:      80,
	}
	positions := []map[string]interface{}{
		{
			"symbol":      "BTCUSDT",
			"markPrice":   100000.0,
			"positionAmt": 0.001,
			"leverage":    5.0,
		},
	}
	balance := map[string]interface{}{
		"availableBalance": 80.0,
		"totalEquity":      100.0,
	}

	got := at.buildOpenOrderPreflight("LONG", decision, positions, balance, 100, 50, 0.5)

	assertFloatClose(t, "initial margin", got.InitialMargin, 10)
	assertFloatClose(t, "current margin used", got.CurrentMarginUsed, 20)
	assertFloatClose(t, "projected margin used", got.ProjectedMarginUsed, 30)
	assertFloatClose(t, "projected margin pct", got.ProjectedMarginPct, 30)
	assertFloatClose(t, "max position value", got.MaxPositionValue, 100)
	assertFloatClose(t, "risk reward", got.RiskRewardRatio, 3)
	if got.CurrentPositions != 1 || got.MaxPositions != 3 {
		t.Fatalf("positions = %d/%d, want 1/3", got.CurrentPositions, got.MaxPositions)
	}
	if got.MinPositionSize != 12 || got.MaxMarginPct != 90 {
		t.Fatalf("risk config summary mismatch: min=%.2f maxMargin=%.2f", got.MinPositionSize, got.MaxMarginPct)
	}
}

func assertFloatClose(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.000001 {
		t.Fatalf("%s = %.8f, want %.8f", name, got, want)
	}
}
