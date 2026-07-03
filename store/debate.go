package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"nofx/ent"
	entdebateparticipant "nofx/ent/debateparticipant"
	entdebatemessage "nofx/ent/debatemessage"
	entdebatesession "nofx/ent/debatesession"
	entdebatevote "nofx/ent/debatevote"
)

// DebateStatus represents the status of a debate session
type DebateStatus string

const (
	DebateStatusPending   DebateStatus = "pending"
	DebateStatusRunning   DebateStatus = "running"
	DebateStatusVoting    DebateStatus = "voting"
	DebateStatusCompleted DebateStatus = "completed"
	DebateStatusCancelled DebateStatus = "cancelled"
)

// DebatePersonality represents AI personality types
type DebatePersonality string

const (
	PersonalityBull        DebatePersonality = "bull"         // Aggressive Bull - looks for long opportunities
	PersonalityBear        DebatePersonality = "bear"         // Cautious Bear - skeptical, focuses on risks
	PersonalityAnalyst     DebatePersonality = "analyst"      // Data Analyst - pure technical analysis
	PersonalityContrarian  DebatePersonality = "contrarian"   // Contrarian - challenges majority opinion
	PersonalityRiskManager DebatePersonality = "risk_manager" // Risk Manager - focuses on position sizing
)

// PersonalityColors maps personalities to colors for UI
var PersonalityColors = map[DebatePersonality]string{
	PersonalityBull:        "#22C55E", // Green
	PersonalityBear:        "#EF4444", // Red
	PersonalityAnalyst:     "#3B82F6", // Blue
	PersonalityContrarian:  "#F59E0B", // Amber
	PersonalityRiskManager: "#8B5CF6", // Purple
}

// PersonalityEmojis maps personalities to emojis
var PersonalityEmojis = map[DebatePersonality]string{
	PersonalityBull:        "🐂",
	PersonalityBear:        "🐻",
	PersonalityAnalyst:     "📊",
	PersonalityContrarian:  "🔄",
	PersonalityRiskManager: "🛡️",
}

// DebateDecision represents a trading decision from the debate
type DebateDecision struct {
	Action          string  `json:"action"`            // open_long/open_short/close_long/close_short/hold/wait
	Symbol          string  `json:"symbol"`            // Trading pair
	Confidence      int     `json:"confidence"`        // 0-100
	Leverage        int     `json:"leverage"`          // Recommended leverage
	PositionPct     float64 `json:"position_pct"`      // Position size as percentage of equity (0.0-1.0)
	PositionSizeUSD float64 `json:"position_size_usd"` // Position size in USD (calculated from pct)
	StopLoss        float64 `json:"stop_loss"`         // Stop loss price
	TakeProfit      float64 `json:"take_profit"`       // Take profit price
	Reasoning       string  `json:"reasoning"`         // Brief reasoning

	// Execution tracking
	Executed   bool      `json:"executed"`              // Whether this decision was executed
	ExecutedAt time.Time `json:"executed_at,omitempty"` // When it was executed
	OrderID    string    `json:"order_id,omitempty"`    // Exchange order ID
	Error      string    `json:"error,omitempty"`       // Execution error if any
}

// DebateSession represents a debate session (API struct)
type DebateSession struct {
	ID              string            `json:"id"`
	UserID          string            `json:"user_id"`
	Name            string            `json:"name"`
	StrategyID      string            `json:"strategy_id"`
	Status          DebateStatus      `json:"status"`
	Symbol          string            `json:"symbol"`           // Primary symbol (for backward compat, may be empty for multi-coin)
	MaxRounds       int               `json:"max_rounds"`
	CurrentRound    int               `json:"current_round"`
	IntervalMinutes int               `json:"interval_minutes"` // Debate interval (5, 15, 30, 60 minutes)
	PromptVariant   string            `json:"prompt_variant"`   // balanced/aggressive/conservative/scalping
	FinalDecision   *DebateDecision   `json:"final_decision,omitempty"`  // Single decision (backward compat)
	FinalDecisions  []*DebateDecision `json:"final_decisions,omitempty"` // Multi-coin decisions
	AutoExecute     bool              `json:"auto_execute"`
	TraderID        string            `json:"trader_id,omitempty"` // Trader to use for auto-execute
	// OI Ranking data options
	EnableOIRanking bool      `json:"enable_oi_ranking"` // Whether to include OI ranking data
	OIRankingLimit  int       `json:"oi_ranking_limit"`  // Number of OI ranking entries (default 10)
	OIDuration      string    `json:"oi_duration"`       // Duration for OI data (1h, 4h, 24h, etc.)
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// DebateSessionDB is the DB model for debate_sessions
type DebateSessionDB struct {
	ID              string       `json:"id"`
	UserID          string       `json:"user_id"`
	Name            string       `json:"name"`
	StrategyID      string       `json:"strategy_id"`
	Status          DebateStatus `json:"status"`
	Symbol          string       `json:"symbol"`
	MaxRounds       int          `json:"max_rounds"`
	CurrentRound    int          `json:"current_round"`
	IntervalMinutes int          `json:"interval_minutes"`
	PromptVariant   string       `json:"prompt_variant"`
	FinalDecision   string       `json:"final_decision,omitempty"` // JSON string
	AutoExecute     bool         `json:"auto_execute"`
	TraderID        string       `json:"trader_id,omitempty"`
	EnableOIRanking bool         `json:"enable_oi_ranking"`
	OIRankingLimit  int          `json:"oi_ranking_limit"`
	OIDuration      string       `json:"oi_duration"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

func (db *DebateSessionDB) toSession() *DebateSession {
	s := &DebateSession{
		ID:              db.ID,
		UserID:          db.UserID,
		Name:            db.Name,
		StrategyID:      db.StrategyID,
		Status:          db.Status,
		Symbol:          db.Symbol,
		MaxRounds:       db.MaxRounds,
		CurrentRound:    db.CurrentRound,
		IntervalMinutes: db.IntervalMinutes,
		PromptVariant:   db.PromptVariant,
		AutoExecute:     db.AutoExecute,
		TraderID:        db.TraderID,
		EnableOIRanking: db.EnableOIRanking,
		OIRankingLimit:  db.OIRankingLimit,
		OIDuration:      db.OIDuration,
		CreatedAt:       db.CreatedAt,
		UpdatedAt:       db.UpdatedAt,
	}

	// Set defaults
	if s.IntervalMinutes == 0 {
		s.IntervalMinutes = 5
	}
	if s.PromptVariant == "" {
		s.PromptVariant = "balanced"
	}
	if s.OIRankingLimit == 0 {
		s.OIRankingLimit = 10
	}
	if s.OIDuration == "" {
		s.OIDuration = "1h"
	}

	// Parse final decision
	if db.FinalDecision != "" {
		var decision DebateDecision
		if json.Unmarshal([]byte(db.FinalDecision), &decision) == nil {
			s.FinalDecision = &decision
		}
	}

	return s
}

// DebateParticipant represents an AI participant in a debate
type DebateParticipant struct {
	ID          string            `json:"id"`
	SessionID   string            `json:"session_id"`
	AIModelID   string            `json:"ai_model_id"`
	AIModelName string            `json:"ai_model_name"`
	Provider    string            `json:"provider"`
	Personality DebatePersonality `json:"personality"`
	Color       string            `json:"color"`
	SpeakOrder  int               `json:"speak_order"`
	CreatedAt   time.Time         `json:"created_at"`
}

// DebateMessage represents a message in the debate
type DebateMessage struct {
	ID          string            `json:"id"`
	SessionID   string            `json:"session_id"`
	Round       int               `json:"round"`
	AIModelID   string            `json:"ai_model_id"`
	AIModelName string            `json:"ai_model_name"`
	Provider    string            `json:"provider"`
	Personality DebatePersonality `json:"personality"`
	MessageType string            `json:"message_type"` // analysis/rebuttal/final/vote
	Content     string            `json:"content"`
	DecisionRaw string            `json:"-"`                       // JSON string in DB
	Decision    *DebateDecision   `json:"decision,omitempty"`      // Parsed for API
	Decisions   []*DebateDecision `json:"decisions,omitempty"`     // Multi-coin decisions
	Confidence  int               `json:"confidence"`
	CreatedAt   time.Time         `json:"created_at"`
}

// DebateVote represents a final vote from an AI (can contain multiple coin decisions)
type DebateVote struct {
	ID            string            `json:"id"`
	SessionID     string            `json:"session_id"`
	AIModelID     string            `json:"ai_model_id"`
	AIModelName   string            `json:"ai_model_name"`
	Action        string            `json:"action"`   // Primary action (backward compat)
	Symbol        string            `json:"symbol"`   // Primary symbol (backward compat)
	Confidence    int               `json:"confidence"`
	Leverage      int               `json:"leverage"`
	PositionPct   float64           `json:"position_pct"`
	StopLossPct   float64           `json:"stop_loss_pct"`
	TakeProfitPct float64           `json:"take_profit_pct"`
	Reasoning     string            `json:"reasoning"`
	Decisions     []*DebateDecision `json:"decisions,omitempty"` // Multi-coin decisions
	CreatedAt     time.Time         `json:"created_at"`
}

// ==================== Ent Conversion Functions ====================

// fromEntDebateSession converts ent.DebateSession to store.DebateSessionDB
func fromEntDebateSession(s *ent.DebateSession) DebateSessionDB {
	return DebateSessionDB{
		ID:              s.ID,
		UserID:          s.UserID,
		Name:            s.Name,
		StrategyID:      s.StrategyID,
		Status:          DebateStatus(s.Status),
		Symbol:          s.Symbol,
		MaxRounds:       s.MaxRounds,
		CurrentRound:    s.CurrentRound,
		IntervalMinutes: s.IntervalMinutes,
		PromptVariant:   s.PromptVariant,
		FinalDecision:   s.FinalDecision,
		AutoExecute:     s.AutoExecute,
		TraderID:        s.TraderID,
		EnableOIRanking: s.EnableOiRanking,
		OIRankingLimit:  s.OiRankingLimit,
		OIDuration:      s.OiDuration,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

// fromEntDebateParticipant converts ent.DebateParticipant to store.DebateParticipant
func fromEntDebateParticipant(p *ent.DebateParticipant) DebateParticipant {
	return DebateParticipant{
		ID:          p.ID,
		SessionID:   p.SessionID,
		AIModelID:   p.AiModelID,
		AIModelName: p.AiModelName,
		Provider:    p.Provider,
		Personality: DebatePersonality(p.Personality),
		Color:       p.Color,
		SpeakOrder:  p.SpeakOrder,
		CreatedAt:   p.CreatedAt,
	}
}

// fromEntDebateMessage converts ent.DebateMessage to store.DebateMessage
func fromEntDebateMessage(m *ent.DebateMessage) DebateMessage {
	return DebateMessage{
		ID:          m.ID,
		SessionID:   m.SessionID,
		Round:       m.Round,
		AIModelID:   m.AiModelID,
		AIModelName: m.AiModelName,
		Provider:    m.Provider,
		Personality: DebatePersonality(m.Personality),
		MessageType: m.MessageType,
		Content:     m.Content,
		DecisionRaw: m.DecisionRaw,
		Confidence:  m.Confidence,
		CreatedAt:   m.CreatedAt,
	}
}

// fromEntDebateVote converts ent.DebateVote to store.DebateVote
func fromEntDebateVote(v *ent.DebateVote) DebateVote {
	return DebateVote{
		ID:            v.ID,
		SessionID:     v.SessionID,
		AIModelID:     v.AiModelID,
		AIModelName:   v.AiModelName,
		Action:        v.Action,
		Symbol:        v.Symbol,
		Confidence:    v.Confidence,
		Leverage:      v.Leverage,
		PositionPct:   v.PositionPct,
		StopLossPct:   v.StopLossPct,
		TakeProfitPct: v.TakeProfitPct,
		Reasoning:     v.Reasoning,
		CreatedAt:     v.CreatedAt,
	}
}

// ==================== Debate Store ====================
type DebateStore struct {
	ec *ent.Client
}

// NewDebateStore creates a new DebateStore
func NewDebateStore() *DebateStore {
	return &DebateStore{}
}

// SetEntClient sets the ent client for this debate store
func (s *DebateStore) SetEntClient(ec *ent.Client) {
	s.ec = ec
}

// InitSchema creates the debate tables
func (s *DebateStore) InitSchema() error {
	return nil
}

// CreateSession creates a new debate session
func (s *DebateStore) CreateSession(session *DebateSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	session.Status = DebateStatusPending
	session.CurrentRound = 0
	if session.IntervalMinutes == 0 {
		session.IntervalMinutes = 5
	}
	if session.PromptVariant == "" {
		session.PromptVariant = "balanced"
	}
	if session.OIRankingLimit == 0 {
		session.OIRankingLimit = 10
	}
	if session.OIDuration == "" {
		session.OIDuration = "1h"
	}

	var finalDecision string
	if session.FinalDecision != nil {
		data, err := json.Marshal(session.FinalDecision)
		if err != nil {
			return err
		}
		finalDecision = string(data)
	}

	ctx := context.Background()
	_, err := s.ec.DebateSession.Create().
		SetID(session.ID).
		SetUserID(session.UserID).
		SetName(session.Name).
		SetStrategyID(session.StrategyID).
		SetStatus(string(session.Status)).
		SetSymbol(session.Symbol).
		SetMaxRounds(session.MaxRounds).
		SetCurrentRound(session.CurrentRound).
		SetIntervalMinutes(session.IntervalMinutes).
		SetPromptVariant(session.PromptVariant).
		SetFinalDecision(finalDecision).
		SetAutoExecute(session.AutoExecute).
		SetTraderID(session.TraderID).
		SetEnableOiRanking(session.EnableOIRanking).
		SetOiRankingLimit(session.OIRankingLimit).
		SetOiDuration(session.OIDuration).
		Save(ctx)
	return err
}

// GetSession gets a debate session by ID
func (s *DebateStore) GetSession(id string) (*DebateSession, error) {
	ctx := context.Background()
	db, err := s.ec.DebateSession.Query().
		Where(entdebatesession.ID(id)).
		First(ctx)
	if err != nil {
		return nil, err
	}
	dbRecord := fromEntDebateSession(db)
	return dbRecord.toSession(), nil
}

// GetSessionsByUser gets all debate sessions for a user
func (s *DebateStore) GetSessionsByUser(userID string) ([]*DebateSession, error) {
	ctx := context.Background()
	dbs, err := s.ec.DebateSession.Query().
		Where(entdebatesession.UserID(userID)).
		Order(ent.Desc(entdebatesession.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	sessions := make([]*DebateSession, len(dbs))
	for i, db := range dbs {
		dbRecord := fromEntDebateSession(db)
		sessions[i] = dbRecord.toSession()
	}
	return sessions, nil
}

// ListAllSessions returns all debate sessions (for cleanup on startup)
func (s *DebateStore) ListAllSessions() ([]*DebateSession, error) {
	ctx := context.Background()
	dbs, err := s.ec.DebateSession.Query().
		Select(entdebatesession.FieldID, entdebatesession.FieldStatus).
		All(ctx)
	if err != nil {
		return nil, err
	}

	sessions := make([]*DebateSession, len(dbs))
	for i, db := range dbs {
		sessions[i] = &DebateSession{ID: db.ID, Status: DebateStatus(db.Status)}
	}
	return sessions, nil
}

// UpdateSessionStatus updates the status of a debate session
func (s *DebateStore) UpdateSessionStatus(id string, status DebateStatus) error {
	ctx := context.Background()
	return s.ec.DebateSession.UpdateOneID(id).
		SetStatus(string(status)).
		Exec(ctx)
}

// UpdateSessionRound updates the current round of a debate session
func (s *DebateStore) UpdateSessionRound(id string, round int) error {
	ctx := context.Background()
	return s.ec.DebateSession.UpdateOneID(id).
		SetCurrentRound(round).
		Exec(ctx)
}

// UpdateSessionFinalDecision updates the final decision of a debate session (single decision)
func (s *DebateStore) UpdateSessionFinalDecision(id string, decision *DebateDecision) error {
	decisionJSON, err := json.Marshal(decision)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return s.ec.DebateSession.UpdateOneID(id).
		SetFinalDecision(string(decisionJSON)).
		SetStatus(string(DebateStatusCompleted)).
		Exec(ctx)
}

// UpdateSessionFinalDecisions updates both single and multi-coin final decisions
func (s *DebateStore) UpdateSessionFinalDecisions(id string, primaryDecision *DebateDecision, allDecisions []*DebateDecision) error {
	primaryJSON, err := json.Marshal(primaryDecision)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return s.ec.DebateSession.UpdateOneID(id).
		SetFinalDecision(string(primaryJSON)).
		SetStatus(string(DebateStatusCompleted)).
		Exec(ctx)
}

// DeleteSession deletes a debate session and all related data
func (s *DebateStore) DeleteSession(id string) error {
	ctx := context.Background()
	// Delete related data first
	s.ec.DebateParticipant.Delete().Where(entdebateparticipant.SessionID(id)).Exec(ctx)
	s.ec.DebateMessage.Delete().Where(entdebatemessage.SessionID(id)).Exec(ctx)
	s.ec.DebateVote.Delete().Where(entdebatevote.SessionID(id)).Exec(ctx)
	return s.ec.DebateSession.DeleteOneID(id).Exec(ctx)
}

// AddParticipant adds a participant to a debate session
func (s *DebateStore) AddParticipant(participant *DebateParticipant) error {
	if participant.ID == "" {
		participant.ID = uuid.New().String()
	}
	if participant.Color == "" {
		if color, ok := PersonalityColors[participant.Personality]; ok {
			participant.Color = color
		} else {
			participant.Color = "#6B7280" // Default gray
		}
	}
	ctx := context.Background()
	_, err := s.ec.DebateParticipant.Create().
		SetID(participant.ID).
		SetSessionID(participant.SessionID).
		SetAiModelID(participant.AIModelID).
		SetAiModelName(participant.AIModelName).
		SetProvider(participant.Provider).
		SetPersonality(string(participant.Personality)).
		SetColor(participant.Color).
		SetSpeakOrder(participant.SpeakOrder).
		Save(ctx)
	return err
}

// GetParticipants gets all participants for a debate session
func (s *DebateStore) GetParticipants(sessionID string) ([]*DebateParticipant, error) {
	ctx := context.Background()
	participants, err := s.ec.DebateParticipant.Query().
		Where(entdebateparticipant.SessionID(sessionID)).
		Order(ent.Asc(entdebateparticipant.FieldSpeakOrder)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*DebateParticipant, len(participants))
	for i, p := range participants {
		converted := fromEntDebateParticipant(p)
		result[i] = &converted
	}
	return result, nil
}

// AddMessage adds a message to a debate session
func (s *DebateStore) AddMessage(msg *DebateMessage) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	if msg.Decision != nil {
		data, err := json.Marshal(msg.Decision)
		if err != nil {
			return err
		}
		msg.DecisionRaw = string(data)
	}
	ctx := context.Background()
	_, err := s.ec.DebateMessage.Create().
		SetID(msg.ID).
		SetSessionID(msg.SessionID).
		SetRound(msg.Round).
		SetAiModelID(msg.AIModelID).
		SetAiModelName(msg.AIModelName).
		SetProvider(msg.Provider).
		SetPersonality(string(msg.Personality)).
		SetMessageType(msg.MessageType).
		SetContent(msg.Content).
		SetDecisionRaw(msg.DecisionRaw).
		SetConfidence(msg.Confidence).
		Save(ctx)
	return err
}

// GetMessages gets all messages for a debate session
func (s *DebateStore) GetMessages(sessionID string) ([]*DebateMessage, error) {
	ctx := context.Background()
	messages, err := s.ec.DebateMessage.Query().
		Where(entdebatemessage.SessionID(sessionID)).
		Order(ent.Asc(entdebatemessage.FieldRound), ent.Asc(entdebatemessage.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*DebateMessage, len(messages))
	for i, msg := range messages {
		converted := fromEntDebateMessage(msg)
		// Parse decision JSON
		if converted.DecisionRaw != "" {
			var decision DebateDecision
			if json.Unmarshal([]byte(converted.DecisionRaw), &decision) == nil {
				converted.Decision = &decision
			}
		}
		result[i] = &converted
	}
	return result, nil
}

// GetMessagesByRound gets messages for a specific round
func (s *DebateStore) GetMessagesByRound(sessionID string, round int) ([]*DebateMessage, error) {
	ctx := context.Background()
	messages, err := s.ec.DebateMessage.Query().
		Where(
			entdebatemessage.SessionID(sessionID),
			entdebatemessage.Round(round),
		).
		Order(ent.Asc(entdebatemessage.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*DebateMessage, len(messages))
	for i, msg := range messages {
		converted := fromEntDebateMessage(msg)
		// Parse decision JSON
		if converted.DecisionRaw != "" {
			var decision DebateDecision
			if json.Unmarshal([]byte(converted.DecisionRaw), &decision) == nil {
				converted.Decision = &decision
			}
		}
		result[i] = &converted
	}
	return result, nil
}

// AddVote adds a vote to a debate session
func (s *DebateStore) AddVote(vote *DebateVote) error {
	if vote.ID == "" {
		vote.ID = uuid.New().String()
	}
	ctx := context.Background()
	_, err := s.ec.DebateVote.Create().
		SetID(vote.ID).
		SetSessionID(vote.SessionID).
		SetAiModelID(vote.AIModelID).
		SetAiModelName(vote.AIModelName).
		SetAction(vote.Action).
		SetSymbol(vote.Symbol).
		SetConfidence(vote.Confidence).
		SetLeverage(vote.Leverage).
		SetPositionPct(vote.PositionPct).
		SetStopLossPct(vote.StopLossPct).
		SetTakeProfitPct(vote.TakeProfitPct).
		SetReasoning(vote.Reasoning).
		Save(ctx)
	return err
}

// GetVotes gets all votes for a debate session
func (s *DebateStore) GetVotes(sessionID string) ([]*DebateVote, error) {
	ctx := context.Background()
	votes, err := s.ec.DebateVote.Query().
		Where(entdebatevote.SessionID(sessionID)).
		Order(ent.Asc(entdebatevote.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*DebateVote, len(votes))
	for i, v := range votes {
		converted := fromEntDebateVote(v)
		result[i] = &converted
	}
	return result, nil
}

// DebateSessionWithDetails combines session with participants and messages
type DebateSessionWithDetails struct {
	*DebateSession
	Participants []*DebateParticipant `json:"participants"`
	Messages     []*DebateMessage     `json:"messages"`
	Votes        []*DebateVote        `json:"votes"`
}

// GetSessionWithDetails gets a session with all related data
func (s *DebateStore) GetSessionWithDetails(id string) (*DebateSessionWithDetails, error) {
	session, err := s.GetSession(id)
	if err != nil {
		return nil, err
	}

	participants, err := s.GetParticipants(id)
	if err != nil {
		return nil, err
	}

	messages, err := s.GetMessages(id)
	if err != nil {
		return nil, err
	}

	votes, err := s.GetVotes(id)
	if err != nil {
		return nil, err
	}

	return &DebateSessionWithDetails{
		DebateSession: session,
		Participants:  participants,
		Messages:      messages,
		Votes:         votes,
	}, nil
}
