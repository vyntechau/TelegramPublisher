package storage

import (
	"fmt"
	"strings"

	"github.com/vyntechau/TelegramPublisher/config"
)

// NewRepository creates a new Repository based on configuration database type.
// Supported: sqlite, postgres, mysql, mariadb
func NewRepository(cfg config.DatabaseConfig) (Repository, error) {
	switch strings.ToLower(cfg.Type) {
	case "sqlite", "sqlite3", "":
		return nil, fmt.Errorf("use sqlite.New(cfg.FilePath) directly or call via main initialization")
	case "postgres", "postgresql", "pgsql":
		return nil, fmt.Errorf("use postgres.New(cfg) directly")
	case "mysql", "mariadb":
		return nil, fmt.Errorf("use mysql.New(cfg) directly")
	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: sqlite, postgres, mysql, mariadb)", cfg.Type)
	}
}
