package ent

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	sqlite3 "modernc.org/sqlite"
)

func init() {
	// ent expects the driver to be registered as "sqlite3",
	// but modernc.org/sqlite registers as "sqlite"
	sql.Register("sqlite3", &sqlite3.Driver{})
}

// DBConfig holds configuration for ent database connection
// Mirrors store.DBConfig from the existing store package
type DBConfig struct {
	Type     string // "sqlite" or "postgres"
	Path     string // for sqlite
	Host     string // for postgres
	Port     int    // for postgres
	User     string // for postgres
	Password string // for postgres
	DBName   string // for postgres
	SSLMode  string // for postgres
}

// OpenEnt opens an ent client from the given config
func OpenEnt(cfg DBConfig) (*Client, error) {
	switch cfg.Type {
	case "sqlite":
		dsn := cfg.Path
		if !strings.Contains(dsn, "foreign_keys") {
			if strings.Contains(dsn, "?") {
				dsn += "&_pragma=foreign_keys(1)"
			} else {
				dsn += "?_pragma=foreign_keys(1)"
			}
		}
		return Open("sqlite3", dsn)
	case "postgres":
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
		)
		return Open("postgres", dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
}

// DBConfigFromStoreConfig converts a store.DBConfig to ent.DBConfig
// Used for dual-write mode during GORM -> Ent migration
func DBConfigFromStoreConfig(driverType, path, host, user, password, dbname, sslmode string, port int) DBConfig {
	return DBConfig{
		Type:     driverType,
		Path:     path,
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
		SSLMode:  sslmode,
	}
}
