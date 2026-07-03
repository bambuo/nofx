package store

import "fmt"

// DBType database type
type DBType string

const (
	DBTypeSQLite   DBType = "sqlite"
	DBTypePostgres DBType = "postgres"
)

// DBConfig database configuration
type DBConfig struct {
	Type     DBType
	Path     string
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// convertQuery converts query placeholders for different database types
func convertQuery(query string, dbType DBType) string {
	if dbType == DBTypePostgres {
		// Convert ? placeholders to $1, $2, etc.
		result := make([]byte, 0, len(query))
		paramIdx := 1
		for i := 0; i < len(query); i++ {
			if query[i] == '?' {
				result = append(result, []byte(fmt.Sprintf("$%d", paramIdx))...)
				paramIdx++
			} else {
				result = append(result, query[i])
			}
		}
		return string(result)
	}
	return query
}
