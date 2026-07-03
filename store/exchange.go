package store

import (
	"context"
	"fmt"
	"nofx/crypto"
	"nofx/ent"
	entexchange "nofx/ent/exchange"
	"nofx/logger"
	"time"

	"github.com/google/uuid"
)

// ExchangeStore exchange storage
type ExchangeStore struct {
	ec *ent.Client
}

// Exchange exchange configuration
type Exchange struct {
	ID                      string          `gorm:"primaryKey" json:"id"`
	ExchangeType            string          `gorm:"column:exchange_type;not null;default:''" json:"exchange_type"`
	AccountName             string          `gorm:"column:account_name;not null;default:''" json:"account_name"`
	UserID                  string          `gorm:"column:user_id;not null;default:default;index" json:"user_id"`
	Name                    string          `gorm:"not null" json:"name"`
	Type                    string          `gorm:"not null" json:"type"` // "cex" or "dex"
	Enabled                 bool            `gorm:"default:false" json:"enabled"`
	APIKey                  crypto.EncryptedString `gorm:"column:api_key;default:''" json:"apiKey"`
	SecretKey               crypto.EncryptedString `gorm:"column:secret_key;default:''" json:"secretKey"`
	Passphrase              crypto.EncryptedString `gorm:"column:passphrase;default:''" json:"passphrase"`
	Testnet                 bool            `gorm:"default:false" json:"testnet"`
	HyperliquidWalletAddr   string          `gorm:"column:hyperliquid_wallet_addr;default:''" json:"hyperliquidWalletAddr"`
	AsterUser               string          `gorm:"column:aster_user;default:''" json:"asterUser"`
	AsterSigner             string          `gorm:"column:aster_signer;default:''" json:"asterSigner"`
	AsterPrivateKey         crypto.EncryptedString `gorm:"column:aster_private_key;default:''" json:"asterPrivateKey"`
	LighterWalletAddr       string          `gorm:"column:lighter_wallet_addr;default:''" json:"lighterWalletAddr"`
	LighterPrivateKey       crypto.EncryptedString `gorm:"column:lighter_private_key;default:''" json:"lighterPrivateKey"`
	LighterAPIKeyPrivateKey crypto.EncryptedString `gorm:"column:lighter_api_key_private_key;default:''" json:"lighterAPIKeyPrivateKey"`
	LighterAPIKeyIndex      int             `gorm:"column:lighter_api_key_index;default:0" json:"lighterAPIKeyIndex"`
	CreatedAt               time.Time       `json:"created_at"`
	UpdatedAt               time.Time       `json:"updated_at"`
}

// fromEntExchange converts ent.Exchange to store.Exchange
func fromEntExchange(e *ent.Exchange) *Exchange {
	if e == nil {
		return nil
	}
	return &Exchange{
		ID:                      e.ID,
		ExchangeType:            e.ExchangeType,
		AccountName:             e.AccountName,
		UserID:                  e.UserID,
		Name:                    e.Name,
		Type:                    e.Type,
		Enabled:                 e.Enabled,
		APIKey:                  e.APIKey,
		SecretKey:               e.SecretKey,
		Passphrase:              e.Passphrase,
		Testnet:                 e.Testnet,
		HyperliquidWalletAddr:   e.HyperliquidWalletAddr,
		AsterUser:               e.AsterUser,
		AsterSigner:             e.AsterSigner,
		AsterPrivateKey:         e.AsterPrivateKey,
		LighterWalletAddr:       e.LighterWalletAddr,
		LighterPrivateKey:       e.LighterPrivateKey,
		LighterAPIKeyPrivateKey: e.LighterAPIKeyPrivateKey,
		LighterAPIKeyIndex:      e.LighterAPIKeyIndex,
		CreatedAt:               e.CreatedAt,
		UpdatedAt:               e.UpdatedAt,
	}
}

// NewExchangeStore creates a new ExchangeStore
func NewExchangeStore() *ExchangeStore {
	return &ExchangeStore{}
}

// List gets user's exchange list
func (s *ExchangeStore) List(userID string) ([]*Exchange, error) {
	exchanges, err := s.ec.Exchange.Query().
		Where(entexchange.UserID(userID)).
		Order(ent.Asc(entexchange.FieldExchangeType), ent.Asc(entexchange.FieldAccountName)).
		All(context.Background())
	if err != nil {
		return nil, err
	}
	result := make([]*Exchange, len(exchanges))
	for i, e := range exchanges {
		result[i] = fromEntExchange(e)
	}
	return result, nil
}

// GetByID gets a specific exchange by UUID
func (s *ExchangeStore) GetByID(userID, id string) (*Exchange, error) {
	e, err := s.ec.Exchange.Query().
		Where(entexchange.And(entexchange.ID(id), entexchange.UserID(userID))).
		Only(context.Background())
	if err != nil {
		return nil, err
	}
	return fromEntExchange(e), nil
}

// getExchangeNameAndType returns the display name and type for an exchange type
func getExchangeNameAndType(exchangeType string) (name string, typ string) {
	switch exchangeType {
	case "binance":
		return "Binance Futures", "cex"
	case "bybit":
		return "Bybit Futures", "cex"
	case "okx":
		return "OKX Futures", "cex"
	case "bitget":
		return "Bitget Futures", "cex"
	case "hyperliquid":
		return "Hyperliquid", "dex"
	case "aster":
		return "Aster DEX", "dex"
	case "lighter":
		return "LIGHTER DEX", "dex"
	default:
		return exchangeType + " Exchange", "cex"
	}
}

// Create creates a new exchange account with UUID
func (s *ExchangeStore) Create(userID, exchangeType, accountName string, enabled bool,
	apiKey, secretKey, passphrase string, testnet bool,
	hyperliquidWalletAddr, asterUser, asterSigner, asterPrivateKey,
	lighterWalletAddr, lighterPrivateKey, lighterApiKeyPrivateKey string, lighterApiKeyIndex int) (string, error) {

	id := uuid.New().String()
	name, typ := getExchangeNameAndType(exchangeType)

	if accountName == "" {
		accountName = "Default"
	}

	logger.Debugf("🔧 ExchangeStore.Create: userID=%s, exchangeType=%s, accountName=%s, id=%s",
		userID, exchangeType, accountName, id)

	q := s.ec.Exchange.Create().
		SetID(id).
		SetExchangeType(exchangeType).
		SetAccountName(accountName).
		SetUserID(userID).
		SetName(name).
		SetType(typ).
		SetEnabled(enabled).
		SetTestnet(testnet).
		SetHyperliquidWalletAddr(hyperliquidWalletAddr).
		SetAsterUser(asterUser).
		SetAsterSigner(asterSigner).
		SetLighterWalletAddr(lighterWalletAddr).
		SetLighterAPIKeyIndex(lighterApiKeyIndex)
	if apiKey != "" {
		q.SetAPIKey(crypto.EncryptedString(apiKey))
	}
	if secretKey != "" {
		q.SetSecretKey(crypto.EncryptedString(secretKey))
	}
	if passphrase != "" {
		q.SetPassphrase(crypto.EncryptedString(passphrase))
	}
	if asterPrivateKey != "" {
		q.SetAsterPrivateKey(crypto.EncryptedString(asterPrivateKey))
	}
	if lighterPrivateKey != "" {
		q.SetLighterPrivateKey(crypto.EncryptedString(lighterPrivateKey))
	}
	if lighterApiKeyPrivateKey != "" {
		q.SetLighterAPIKeyPrivateKey(crypto.EncryptedString(lighterApiKeyPrivateKey))
	}
	created, err := q.Save(context.Background())
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

// Update updates exchange configuration by UUID
func (s *ExchangeStore) Update(userID, id string, enabled bool, apiKey, secretKey, passphrase string, testnet bool,
	hyperliquidWalletAddr, asterUser, asterSigner, asterPrivateKey, lighterWalletAddr, lighterPrivateKey, lighterApiKeyPrivateKey string, lighterApiKeyIndex int) error {

	logger.Debugf("🔧 ExchangeStore.Update: userID=%s, id=%s, enabled=%v", userID, id, enabled)

	upd := s.ec.Exchange.Update().
		Where(entexchange.And(entexchange.ID(id), entexchange.UserID(userID))).
		SetEnabled(enabled).
		SetTestnet(testnet).
		SetHyperliquidWalletAddr(hyperliquidWalletAddr).
		SetAsterUser(asterUser).
		SetAsterSigner(asterSigner).
		SetLighterWalletAddr(lighterWalletAddr).
		SetLighterAPIKeyIndex(lighterApiKeyIndex).
		SetUpdatedAt(time.Now().UTC())

	if apiKey != "" {
		upd.SetAPIKey(crypto.EncryptedString(apiKey))
	}
	if secretKey != "" {
		upd.SetSecretKey(crypto.EncryptedString(secretKey))
	}
	if passphrase != "" {
		upd.SetPassphrase(crypto.EncryptedString(passphrase))
	}
	if asterPrivateKey != "" {
		upd.SetAsterPrivateKey(crypto.EncryptedString(asterPrivateKey))
	}
	if lighterPrivateKey != "" {
		upd.SetLighterPrivateKey(crypto.EncryptedString(lighterPrivateKey))
	}
	if lighterApiKeyPrivateKey != "" {
		upd.SetLighterAPIKeyPrivateKey(crypto.EncryptedString(lighterApiKeyPrivateKey))
	}

	n, err := upd.Save(context.Background())
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("exchange not found: id=%s, userID=%s", id, userID)
	}
	return nil
}

// UpdateAccountName updates the account name for an exchange
func (s *ExchangeStore) UpdateAccountName(userID, id, accountName string) error {
	n, err := s.ec.Exchange.Update().
		Where(entexchange.And(entexchange.ID(id), entexchange.UserID(userID))).
		SetAccountName(accountName).
		SetUpdatedAt(time.Now().UTC()).
		Save(context.Background())
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("exchange not found: id=%s, userID=%s", id, userID)
	}
	return nil
}

// Delete deletes an exchange account
func (s *ExchangeStore) Delete(userID, id string) error {
	n, err := s.ec.Exchange.Delete().Where(entexchange.And(entexchange.ID(id), entexchange.UserID(userID))).Exec(context.Background())
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("exchange not found: id=%s, userID=%s", id, userID)
	}
	logger.Infof("🗑️ Deleted exchange: id=%s, userID=%s", id, userID)
	return nil
}

// CreateLegacy creates exchange configuration (legacy API for backward compatibility)
// This method is deprecated, use Create instead
func (s *ExchangeStore) CreateLegacy(userID, id, name, typ string, enabled bool, apiKey, secretKey string, testnet bool,
	hyperliquidWalletAddr, asterUser, asterSigner, asterPrivateKey string) error {

	// Check if this is an old-style ID (exchange type as ID)
	if id == "binance" || id == "bybit" || id == "okx" || id == "bitget" || id == "hyperliquid" || id == "aster" || id == "lighter" {
		_, err := s.Create(userID, id, "Default", enabled, apiKey, secretKey, "", testnet,
			hyperliquidWalletAddr, asterUser, asterSigner, asterPrivateKey, "", "", "", 0)
		return err
	}

	// Otherwise assume it's already a UUID — check if exists first
	exists, err := s.ec.Exchange.Query().Where(entexchange.ID(id)).Exist(context.Background())
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = s.ec.Exchange.Create().
		SetID(id).
		SetUserID(userID).
		SetName(name).
		SetType(typ).
		SetEnabled(enabled).
		SetAPIKey(crypto.EncryptedString(apiKey)).
		SetSecretKey(crypto.EncryptedString(secretKey)).
		SetTestnet(testnet).
		SetHyperliquidWalletAddr(hyperliquidWalletAddr).
		SetAsterUser(asterUser).
		SetAsterSigner(asterSigner).
		SetAsterPrivateKey(crypto.EncryptedString(asterPrivateKey)).
		Save(context.Background())
	return err
}

func (s *ExchangeStore) initTables() error {
	return nil
}

// migrateToMultiAccount migrates old schema (id=exchange_type) to new schema (id=UUID)
func (s *ExchangeStore) migrateToMultiAccount() error {
	return nil
}


