package storage_test

import (
	"strings"
	"testing"

	"github.com/vyntechau/TelegramPublisher/config"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
)

func TestNewRepository(t *testing.T) {
	tests := []struct {
		dbType      string
		expectErr   string
	}{
		{"", "use sqlite.New(cfg.FilePath) directly"},
		{"sqlite", "use sqlite.New(cfg.FilePath) directly"},
		{"sqlite3", "use sqlite.New(cfg.FilePath) directly"},
		{"postgres", "use postgres.New(cfg) directly"},
		{"postgresql", "use postgres.New(cfg) directly"},
		{"pgsql", "use postgres.New(cfg) directly"},
		{"mysql", "use mysql.New(cfg) directly"},
		{"mariadb", "use mysql.New(cfg) directly"},
		{"unknown_db", "unsupported database type: unknown_db"},
	}

	for _, tt := range tests {
		t.Run("type_"+tt.dbType, func(t *testing.T) {
			cfg := config.DatabaseConfig{Type: tt.dbType}
			repo, err := storage.NewRepository(cfg)
			if repo != nil {
				t.Errorf("expected nil repo, got %+v", repo)
			}
			if err == nil || !strings.Contains(err.Error(), tt.expectErr) {
				t.Errorf("expected error containing %q, got %v", tt.expectErr, err)
			}
		})
	}
}
