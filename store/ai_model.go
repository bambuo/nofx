package store

import (
	"cmp"
	"context"
	"fmt"
	"nofx/crypto"
	"nofx/ent"
	entaimodel "nofx/ent/aimodel"
	"nofx/logger"
	"strings"
	"time"
)

// AIModelStore AI model storage
type AIModelStore struct {
	ec *ent.Client
}

// AIModel AI model configuration
type AIModel struct {
	ID              string                `json:"id"`
	UserID          string                `json:"user_id"`
	Name            string                `json:"name"`
	Provider        string                `json:"provider"`
	Enabled         bool                  `json:"enabled"`
	APIKey          crypto.EncryptedString `json:"apiKey"`
	CustomAPIURL    string                `json:"customApiUrl"`
	CustomModelName string                `json:"customModelName"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

// fromEntAIModel converts ent.AIModel to store.AIModel
func fromEntAIModel(ea *ent.AIModel) *AIModel {
	if ea == nil {
		return nil
	}
	return &AIModel{
		ID:              ea.ID,
		UserID:          ea.UserID,
		Name:            ea.Name,
		Provider:        ea.Provider,
		Enabled:         ea.Enabled,
		APIKey:          ea.APIKey,
		CustomAPIURL:    ea.CustomAPIURL,
		CustomModelName: ea.CustomModelName,
		CreatedAt:       ea.CreatedAt,
		UpdatedAt:       ea.UpdatedAt,
	}
}

// NewAIModelStore creates a new AIModelStore
func NewAIModelStore() *AIModelStore {
	return &AIModelStore{}
}

func (s *AIModelStore) initTables() error {
	// ent handles schema migration
	return nil
}

func (s *AIModelStore) initDefaultData() error {
	return nil
}

// List retrieves the user's AI model list
func (s *AIModelStore) List(userID string) ([]*AIModel, error) {
	if s.ec == nil {
		return nil, fmt.Errorf("ent client not available")
	}
	ctx := context.Background()
	models, err := s.ec.AIModel.Query().Where(entaimodel.UserIDEQ(userID)).Order(ent.Asc(entaimodel.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*AIModel, len(models))
	for i, m := range models {
		result[i] = fromEntAIModel(m)
	}
	return result, nil
}

// Get retrieves a single AI model
func (s *AIModelStore) Get(userID, modelID string) (*AIModel, error) {
	if s.ec == nil {
		return nil, fmt.Errorf("ent client not available")
	}
	if modelID == "" {
		return nil, fmt.Errorf("model ID cannot be empty")
	}
	ctx := context.Background()
	candidates := []string{}
	if userID != "" {
		candidates = append(candidates, userID)
	}
	if userID != "default" {
		candidates = append(candidates, "default")
	}
	if len(candidates) == 0 {
		candidates = append(candidates, "default")
	}
	for _, uid := range candidates {
		m, err := s.ec.AIModel.Query().
			Where(entaimodel.And(entaimodel.UserIDEQ(uid), entaimodel.IDEQ(modelID))).
			Only(ctx)
		if err == nil {
			return fromEntAIModel(m), nil
		}
		if !ent.IsNotFound(err) {
			return nil, err
		}
	}
	return nil, fmt.Errorf("model not found")
}

// GetByID retrieves an AI model by ID only (for debate engine)
func (s *AIModelStore) GetByID(modelID string) (*AIModel, error) {
	if s.ec == nil {
		return nil, fmt.Errorf("ent client not available")
	}
	if modelID == "" {
		return nil, fmt.Errorf("model ID cannot be empty")
	}
	ctx := context.Background()
	m, err := s.ec.AIModel.Get(ctx, modelID)
	if err != nil {
		return nil, err
	}
	return fromEntAIModel(m), nil
}

// GetDefault retrieves the default enabled AI model
func (s *AIModelStore) GetDefault(userID string) (*AIModel, error) {
	if s.ec == nil {
		return nil, fmt.Errorf("ent client not available")
	}
	ctx := context.Background()
	userID = cmp.Or(userID, "default")
	model, err := s.firstEnabled(ctx, userID)
	if err == nil {
		return model, nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}
	if userID != "default" {
		return s.firstEnabled(ctx, "default")
	}
	return nil, fmt.Errorf("please configure an available AI model in the system first")
}

func (s *AIModelStore) firstEnabled(ctx context.Context, userID string) (*AIModel, error) {
	m, err := s.ec.AIModel.Query().
		Where(entaimodel.And(entaimodel.UserIDEQ(userID), entaimodel.EnabledEQ(true))).
		Order(ent.Desc(entaimodel.FieldUpdatedAt), ent.Asc(entaimodel.FieldID)).
		First(ctx)
	if err != nil {
		return nil, err
	}
	return fromEntAIModel(m), nil
}

// Update updates AI model, creates if not exists
// IMPORTANT: If apiKey is empty string, the existing API key will be preserved (not overwritten)
func (s *AIModelStore) Update(userID, id string, enabled bool, apiKey, customAPIURL, customModelName string) error {
	if s.ec == nil {
		return fmt.Errorf("ent client not available")
	}
	ctx := context.Background()

	// Try exact ID match first
	existing, err := s.ec.AIModel.Query().
		Where(entaimodel.And(entaimodel.UserIDEQ(userID), entaimodel.IDEQ(id))).
		Only(ctx)
	if err == nil {
		// Update existing model
		upd := s.ec.AIModel.UpdateOne(existing).
			SetEnabled(enabled).
			SetCustomAPIURL(customAPIURL).
			SetCustomModelName(customModelName)
		if apiKey != "" {
			upd.SetAPIKey(crypto.EncryptedString(apiKey))
		}
		return upd.Exec(ctx)
	}
	if !ent.IsNotFound(err) {
		return err
	}

	// Try legacy logic compatibility: use id as provider to search
	provider := id
	existing, err = s.ec.AIModel.Query().
		Where(entaimodel.And(entaimodel.UserIDEQ(userID), entaimodel.ProviderEQ(provider))).
		Only(ctx)
	if err == nil {
		logger.Warnf("⚠️ Using legacy provider matching to update model: %s -> %s", provider, existing.ID)
		upd := s.ec.AIModel.UpdateOne(existing).
			SetEnabled(enabled).
			SetCustomAPIURL(customAPIURL).
			SetCustomModelName(customModelName)
		if apiKey != "" {
			upd.SetAPIKey(crypto.EncryptedString(apiKey))
		}
		return upd.Exec(ctx)
	}
	if !ent.IsNotFound(err) {
		return err
	}

	// Create new record
	if provider == id && (provider == "deepseek" || provider == "qwen") {
		provider = id
	} else {
		parts := strings.Split(id, "_")
		if len(parts) >= 2 {
			provider = parts[len(parts)-1]
		} else {
			provider = id
		}
	}

	// Try to get name from existing model with same provider
	var name string
	refModel, err := s.ec.AIModel.Query().
		Where(entaimodel.ProviderEQ(provider)).
		First(ctx)
	if err == nil {
		name = refModel.Name
	} else {
		if provider == "deepseek" {
			name = "DeepSeek AI"
		} else if provider == "qwen" {
			name = "Qwen AI"
		} else {
			name = provider + " AI"
		}
	}

	newModelID := id
	if id == provider {
		newModelID = fmt.Sprintf("%s_%s", userID, provider)
	}

	logger.Infof("✓ Creating new AI model configuration: ID=%s, Provider=%s, Name=%s", newModelID, provider, name)
	_, err = s.ec.AIModel.Create().
		SetID(newModelID).
		SetUserID(userID).
		SetName(name).
		SetProvider(provider).
		SetEnabled(enabled).
		SetAPIKey(crypto.EncryptedString(apiKey)).
		SetCustomAPIURL(customAPIURL).
		SetCustomModelName(customModelName).
		Save(ctx)
	return err
}

// Create creates an AI model
func (s *AIModelStore) Create(userID, id, name, provider string, enabled bool, apiKey, customAPIURL string) error {
	if s.ec == nil {
		return fmt.Errorf("ent client not available")
	}
	ctx := context.Background()
	// Check if already exists (FirstOrOrCreate semantics)
	exists, err := s.ec.AIModel.Query().
		Where(entaimodel.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = s.ec.AIModel.Create().
		SetID(id).
		SetUserID(userID).
		SetName(name).
		SetProvider(provider).
		SetEnabled(enabled).
		SetAPIKey(crypto.EncryptedString(apiKey)).
		SetCustomAPIURL(customAPIURL).
		Save(ctx)
	return err
}
