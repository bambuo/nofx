// Package store provides unified database storage layer
// All database operations should go through this package
package store

import (
	"context"
	"database/sql"
	"fmt"
	"nofx/ent"
	"nofx/logger"
	"strings"
	"sync"
)

// Store unified data storage interface
type Store struct {
	db     *sql.DB    // Raw sql.DB for system_config and legacy compatibility
	dbType DBType     // Database type
	ec     *ent.Client // Ent client

	// Sub-stores (lazy initialization)
	user     *UserStore
	aiModel  *AIModelStore
	exchange *ExchangeStore
	trader   *TraderStore
	decision *DecisionStore
	backtest *BacktestStore
	position *PositionStore
	strategy *StrategyStore
	equity   *EquityStore
	order    *OrderStore
	grid     *GridStore

	mu sync.RWMutex
}

// New creates new Store instance (SQLite mode)
func New(dbPath string) (*Store, error) {
	dsn := ensureFKSQLite(dbPath)
	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	ec, err := ent.Open("sqlite3", dsn)
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to create ent client: %w", err)
	}
	if err := ec.Schema.Create(context.Background()); err != nil {
		ec.Close()
		sqlDB.Close()
		return nil, fmt.Errorf("failed to run schema migration: %w", err)
	}
	s := &Store{db: sqlDB, dbType: DBTypeSQLite, ec: ec}
	if err := s.initTables(); err != nil {
		ec.Close()
		return nil, fmt.Errorf("failed to initialize table structure: %w", err)
	}
	if err := s.initDefaultData(); err != nil {
		ec.Close()
		return nil, fmt.Errorf("failed to initialize default data: %w", err)
	}
	logger.Infof("✅ Database initialized (Ent, SQLite)")
	return s, nil
}

// NewWithConfig creates new Store instance with provided database configuration
func NewWithConfig(cfg DBConfig) (*Store, error) {
	ec, err := initEntClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	dbType := DBTypeSQLite
	if cfg.Type == DBTypePostgres {
		dbType = DBTypePostgres
	}
	// Open sql.DB for system_config and legacy compatibility
	var sqlDB *sql.DB
	if cfg.Type == DBTypeSQLite && cfg.Path != "" {
		dsn := ensureFKSQLite(cfg.Path)
		sqlDB, _ = sql.Open("sqlite3", dsn)
	}
	s := &Store{db: sqlDB, dbType: dbType, ec: ec}
	if err := s.initTables(); err != nil {
		ec.Close()
		if sqlDB != nil {
			sqlDB.Close()
		}
		return nil, fmt.Errorf("failed to initialize table structure: %w", err)
	}
	if err := s.initDefaultData(); err != nil {
		ec.Close()
		if sqlDB != nil {
			sqlDB.Close()
		}
		return nil, fmt.Errorf("failed to initialize default data: %w", err)
	}
	logger.Infof("✅ Database initialized (Ent)")
	return s, nil
}

// NewFromGorm is deprecated after GORM removal
// Use New or NewWithConfig instead
func NewFromGorm(gdb *sql.DB) (*Store, error) {
	return nil, fmt.Errorf("NewFromGorm is deprecated, use New or NewWithConfig")
}

// NewFromDB creates Store from existing database connection
func NewFromDB(db *sql.DB) *Store {
	return &Store{db: db, dbType: DBTypeSQLite}
}

// initTables initializes tables not managed by ent schema
func (s *Store) initTables() error {
	// Create system_config table for legacy key-value config
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS system_config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create system_config table: %w", err)
	}
	return nil
}

// initDefaultData initializes default data
func (s *Store) initDefaultData() error {
	// Migrate old decision_account_snapshots data to new trader_equity_snapshots table
	if migrated, err := s.Equity().MigrateFromDecision(); err != nil {
		logger.Warnf("failed to migrate equity data: %v", err)
	} else if migrated > 0 {
		logger.Infof("✅ Migrated %d equity records to new table", migrated)
	}
	return nil
}

// User gets user storage
func (s *Store) User() *UserStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.user == nil {
		s.user = NewUserStore()
		s.user.ec = s.ec
	}
	return s.user
}

// AIModel gets AI model storage
func (s *Store) AIModel() *AIModelStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.aiModel == nil {
		s.aiModel = NewAIModelStore()
		s.aiModel.ec = s.ec
	}
	return s.aiModel
}

// Exchange gets exchange storage
func (s *Store) Exchange() *ExchangeStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.exchange == nil {
		s.exchange = NewExchangeStore()
		s.exchange.ec = s.ec
	}
	return s.exchange
}

// Trader gets trader storage
func (s *Store) Trader() *TraderStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.trader == nil {
		s.trader = NewTraderStore()
		s.trader.ec = s.ec
	}
	return s.trader
}

// Decision gets decision log storage
func (s *Store) Decision() *DecisionStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.decision == nil {
		s.decision = NewDecisionStore()
		s.decision.ec = s.ec
	}
	return s.decision
}

// Backtest gets backtest data storage
func (s *Store) Backtest() *BacktestStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.backtest == nil {
		s.backtest = NewBacktestStore()
		s.backtest.ec = s.ec
	}
	return s.backtest
}

// Position gets position storage
func (s *Store) Position() *PositionStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.position == nil {
		s.position = NewPositionStore()
		s.position.ec = s.ec
	}
	return s.position
}

// Strategy gets strategy storage
func (s *Store) Strategy() *StrategyStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.strategy == nil {
		s.strategy = NewStrategyStore()
		s.strategy.ec = s.ec
	}
	return s.strategy
}

// Equity gets equity storage
func (s *Store) Equity() *EquityStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.equity == nil {
		s.equity = NewEquityStore()
		s.equity.ec = s.ec
	}
	return s.equity
}

// Order gets order storage
func (s *Store) Order() *OrderStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.order == nil {
		s.order = NewOrderStore()
		s.order.ec = s.ec
	}
	return s.order
}

// Grid gets grid trading storage
func (s *Store) Grid() *GridStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.grid == nil {
		s.grid = NewGridStore()
		s.grid.ec = s.ec
	}
	return s.grid
}

// Close closes database connection
func (s *Store) Close() error {
	if s.ec != nil {
		s.ec.Close()
	}
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GormDB is deprecated after GORM removal
func (s *Store) GormDB() *sql.DB {
	return s.db
}

// EntClient returns the ent client
func (s *Store) EntClient() *ent.Client {
	return s.ec
}

// ensureFKSQLite ensures the SQLite DSN includes foreign_keys pragma
func ensureFKSQLite(path string) string {
	if strings.Contains(path, "foreign_keys") {
		return path
	}
	if strings.Contains(path, "?") {
		return path + "&_pragma=foreign_keys(1)"
	}
	return path + "?_pragma=foreign_keys(1)"
}

// getSQLDB gets the underlying *sql.DB from ent driver
func (s *Store) getSQLDB() (*sql.DB, error) {
	if s.db != nil {
		return s.db, nil
	}
	return nil, fmt.Errorf("database not initialized")
}

// initEntClient initializes ent client from DBConfig
func initEntClient(cfg DBConfig) (*ent.Client, error) {
	ec, err := ent.OpenEnt(ent.DBConfig{
		Type:     string(cfg.Type),
		Path:     cfg.Path,
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		DBName:   cfg.DBName,
		SSLMode:  cfg.SSLMode,
	})
	if err != nil {
		return nil, fmt.Errorf("ent open: %w", err)
	}
	if err := ec.Schema.Create(context.Background()); err != nil {
		ec.Close()
		return nil, fmt.Errorf("ent schema create: %w", err)
	}
	return ec, nil
}

// Driver returns database driver for abstraction (legacy)
func (s *Store) Driver() *sql.DB {
	return s.db
}

// DBType returns current database type
func (s *Store) DBType() DBType {
	return s.dbType
}

// DB gets underlying database connection (for legacy code compatibility)
func (s *Store) DB() *sql.DB {
	return s.db
}

// GetSystemConfig gets a system configuration value by key
func (s *Store) GetSystemConfig(key string) (string, error) {
	if s.db == nil {
		return "", fmt.Errorf("database not initialized")
	}
	var value string
	row := s.db.QueryRow("SELECT value FROM system_config WHERE key = ?", key)
	if err := row.Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// SetSystemConfig sets a system configuration value
func (s *Store) SetSystemConfig(key, value string) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := s.db.Exec(`
		INSERT INTO system_config (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	return err
}

// Transaction is deprecated after GORM removal
func (s *Store) Transaction(fn func(tx *sql.Tx) error) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// TransactionSQL executes transaction with sql.Tx (legacy)
func (s *Store) TransactionSQL(fn func(tx *sql.Tx) error) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
