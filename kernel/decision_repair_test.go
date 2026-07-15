package kernel

import (
	"nofx/market"
	"nofx/mcp"
	"nofx/store"
	"strings"
	"testing"
	"time"
)

type repairTestClient struct {
	responses []string
	requests  []*mcp.Request
}

func (c *repairTestClient) SetAPIKey(apiKey string, customURL string, customModel string) {}
func (c *repairTestClient) SetTimeout(timeout time.Duration)                              {}
func (c *repairTestClient) CallWithMessages(systemPrompt, userPrompt string) (string, error) {
	return "", nil
}
func (c *repairTestClient) CallWithRequest(req *mcp.Request) (string, error) {
	c.requests = append(c.requests, req)
	if len(c.responses) == 0 {
		return "", nil
	}
	response := c.responses[0]
	c.responses = c.responses[1:]
	return response, nil
}

func TestGetFullDecisionRepairsUnstructuredAIResponse(t *testing.T) {
	config := store.GetDefaultStrategyConfig("en")
	engine := NewStrategyEngine(&config)
	client := &repairTestClient{
		responses: []string{
			"The market is noisy. I would wait and keep the existing position unchanged.",
			`<reasoning>
- Original response did not contain a trade setup.
</reasoning>
<decision>
[
  {"symbol":"ALL","action":"wait","confidence":0,"reasoning":"No structured actionable decision was present in the original response."}
]
</decision>`,
		},
	}

	ctx := &Context{
		CurrentTime:    "2026-07-15 10:00:00 UTC",
		RuntimeMinutes: 10,
		CallCount:      1,
		Account: AccountInfo{
			TotalEquity:      100,
			AvailableBalance: 80,
		},
		CandidateCoins: []CandidateCoin{{Symbol: "BTCUSDT", Sources: []string{"static"}}},
		MarketDataMap: map[string]*market.Data{
			"BTCUSDT": {Symbol: "BTCUSDT", CurrentPrice: 100000},
		},
		OITopDataMap: map[string]*OITopData{},
	}

	decision, err := GetFullDecisionWithStrategy(ctx, client, engine, "balanced")
	if err != nil {
		t.Fatalf("GetFullDecisionWithStrategy() error = %v", err)
	}
	if len(client.requests) != 2 {
		t.Fatalf("expected initial request plus repair request, got %d", len(client.requests))
	}
	if decisionUsesSafeFallback(decision) {
		t.Fatalf("expected repaired decision to replace safe fallback: %+v", decision.Decisions)
	}
	if len(decision.Decisions) != 1 || decision.Decisions[0].Action != "wait" || decision.Decisions[0].Symbol != "ALL" {
		t.Fatalf("unexpected decisions: %+v", decision.Decisions)
	}
	if !strings.Contains(decision.RawResponse, "FORMAT_REPAIR_RESPONSE") {
		t.Fatalf("raw response should include repair response, got: %s", decision.RawResponse)
	}
}
