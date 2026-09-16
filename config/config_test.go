package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyntechau/TelegramPublisher/config"
)

func clearEnvs() {
	envKeys := []string{
		"APP_HOST", "APP_PORT", "APP_ENV", "JWT_SECRET",
		"TELEGRAM_BOT_TOKEN", "BOT_TOKEN", "TELEGRAM_OWNER_ID",
		"DATABASE_URL", "DB_URL", "DB_DSN", "DB_FILE_PATH", "DB_TYPE",
	}
	for _, k := range envKeys {
		_ = os.Unsetenv(k)
	}
}

func TestConfigDefaultsAndEnvOverrides(t *testing.T) {
	clearEnvs()
	defer clearEnvs()

	_ = os.Setenv("APP_HOST", "127.0.0.1")
	_ = os.Setenv("APP_PORT", "9090")
	_ = os.Setenv("APP_ENV", "development")
	_ = os.Setenv("JWT_SECRET", "custom-dev-secret-key-at-least-32-bytes")
	_ = os.Setenv("TELEGRAM_BOT_TOKEN", "987654321:XYZ-abc_12345")
	_ = os.Setenv("TELEGRAM_OWNER_ID", "123456789")
	_ = os.Setenv("DB_TYPE", "SQLITE")
	_ = os.Setenv("DATABASE_URL", "sqlite://data/override.db")

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	if cfg.App.Host != "127.0.0.1" {
		t.Errorf("expected APP_HOST=127.0.0.1, got: %s", cfg.App.Host)
	}
	if cfg.App.Port != 9090 {
		t.Errorf("expected APP_PORT=9090, got: %d", cfg.App.Port)
	}
	if cfg.App.Env != "development" {
		t.Errorf("expected APP_ENV=development, got: %s", cfg.App.Env)
	}
	if cfg.App.JWTSecret != "custom-dev-secret-key-at-least-32-bytes" {
		t.Errorf("expected JWT_SECRET to match env, got: %s", cfg.App.JWTSecret)
	}
	if cfg.Bot.Token != "987654321:XYZ-abc_12345" {
		t.Errorf("expected TELEGRAM_BOT_TOKEN to match env, got: %s", cfg.Bot.Token)
	}
	if cfg.Database.Type != "sqlite" {
		t.Errorf("expected DB_TYPE=sqlite (lowercase), got: %s", cfg.Database.Type)
	}
	if cfg.Bot.OwnerID != 123456789 {
		t.Errorf("expected TELEGRAM_OWNER_ID=123456789, got: %d", cfg.Bot.OwnerID)
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid config, got: %v", err)
	}
}

func TestEnvFallbackKeys(t *testing.T) {
	clearEnvs()
	defer clearEnvs()

	// Test fallback BOT_TOKEN, DB_URL, and invalid int parsing branches
	_ = os.Setenv("APP_PORT", "not-a-number")
	_ = os.Setenv("BOT_TOKEN", "111222:fallback_token")
	_ = os.Setenv("TELEGRAM_OWNER_ID", "invalid-owner-id")
	_ = os.Setenv("DB_URL", "sqlite://data/fallback_db_url.db")

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	if cfg.App.Port != 8080 {
		t.Errorf("expected default port 8080 on invalid int, got: %d", cfg.App.Port)
	}
	if cfg.Bot.Token != "111222:fallback_token" {
		t.Errorf("expected BOT_TOKEN fallback, got: %s", cfg.Bot.Token)
	}
	if cfg.Bot.OwnerID != 0 {
		t.Errorf("expected default owner_id 0 on invalid parse, got: %d", cfg.Bot.OwnerID)
	}
	if cfg.Database.FilePath != "data/fallback_db_url.db" {
		t.Errorf("expected DB_URL fallback, got: %s", cfg.Database.FilePath)
	}

	// Test DB_DSN fallback
	clearEnvs()
	_ = os.Setenv("DB_DSN", "mysql://root:secret@tcp(127.0.0.1:3306)/testdb")
	cfg, err = config.Load("")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Database.Type != "mysql" {
		t.Errorf("expected mysql type from DB_DSN, got: %s", cfg.Database.Type)
	}

	// Test DB_FILE_PATH fallback
	clearEnvs()
	_ = os.Setenv("DB_FILE_PATH", "custom_file.db")
	cfg, err = config.Load("")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Database.FilePath != "custom_file.db" {
		t.Errorf("expected custom_file.db from DB_FILE_PATH, got: %s", cfg.Database.FilePath)
	}
}

func TestLoadWithYAMLFiles(t *testing.T) {
	clearEnvs()
	defer clearEnvs()

	tmpDir := t.TempDir()

	// 1. Valid YAML load via explicit configPath
	validYAML := `
app:
  host: "10.0.0.1"
  port: 8888
  env: "development"
  jwt_secret: ""
bot:
  token: "555:bot-token"
  owner_id: 999
database:
  url: "sqlite://tmp/test.db"
`
	validPath := filepath.Join(tmpDir, "valid_config.yaml")
	if err := os.WriteFile(validPath, []byte(validYAML), 0600); err != nil {
		t.Fatalf("failed to write temp yaml: %v", err)
	}

	cfg, err := config.Load(validPath)
	if err != nil {
		t.Fatalf("unexpected error loading valid yaml: %v", err)
	}
	if cfg.App.Host != "10.0.0.1" || cfg.App.Port != 8888 {
		t.Errorf("yaml values not applied: %+v", cfg.App)
	}
	// Verify auto-generated random key when jwt_secret is empty in development
	if len(cfg.App.JWTSecret) < 32 {
		t.Errorf("expected auto-generated secure JWT secret, got: %s", cfg.App.JWTSecret)
	}

	// 2. Invalid YAML syntax error handling
	invalidYAML := `
app:
  host: [invalid: unclosed
`
	invalidPath := filepath.Join(tmpDir, "invalid_config.yaml")
	if err := os.WriteFile(invalidPath, []byte(invalidYAML), 0600); err != nil {
		t.Fatalf("failed to write invalid yaml: %v", err)
	}

	_, err = config.Load(invalidPath)
	if err == nil {
		t.Fatalf("expected error loading invalid yaml, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse YAML config") {
		t.Errorf("expected parse error message, got: %v", err)
	}
}

func TestValidationRules(t *testing.T) {
	// Base valid config
	base := func() *config.Config {
		return &config.Config{
			App: config.AppConfig{
				Host:      "0.0.0.0",
				Port:      8080,
				Env:       "production",
				JWTSecret: "a-secure-production-jwt-signing-secret-with-32-chars",
			},
			Bot: config.BotConfig{
				Token:   "123456789:ABCdefGhIJklmnOPqrstUVwxyz",
				OwnerID: 12345,
			},
			Database: config.DatabaseConfig{
				Type:     "sqlite",
				FilePath: "data/test.db",
				URL:      "sqlite://data/test.db",
			},
		}
	}

	// 1. Port boundaries
	c := base()
	c.App.Port = 0
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "invalid app port") {
		t.Errorf("expected port 0 error, got: %v", err)
	}
	c.App.Port = 65536
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "invalid app port") {
		t.Errorf("expected port 65536 error, got: %v", err)
	}

	// 2. Production JWT secret validation
	c = base()
	c.App.JWTSecret = config.InsecureDefaultJWTSecret
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "JWT_SECRET must be explicitly provided") {
		t.Errorf("expected default secret error in production, got: %v", err)
	}

	c = base()
	c.App.JWTSecret = "short-secret"
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "at least 32 characters") {
		t.Errorf("expected short secret error in production, got: %v", err)
	}

	// Non-production allows shorter secrets without error
	c = base()
	c.App.Env = "development"
	c.App.JWTSecret = "short-secret"
	if err := c.Validate(); err != nil {
		t.Errorf("expected non-production to allow dev secrets, got: %v", err)
	}

	// 3. Bot Token Format validation
	c = base()
	c.Bot.Token = "invalid-token-no-colon"
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "invalid bot token format") {
		t.Errorf("expected bot token format error, got: %v", err)
	}
	c.Bot.Token = "" // empty is allowed
	if err := c.Validate(); err != nil {
		t.Errorf("expected empty bot token to be valid, got: %v", err)
	}

	// 4. Database SQLite validation
	c = base()
	c.Database.Type = "sqlite3"
	c.Database.FilePath = ""
	c.Database.URL = ""
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "sqlite database requires") {
		t.Errorf("expected empty sqlite error, got: %v", err)
	}

	// 5. Database Postgres / MySQL validation
	for _, dbType := range []string{"postgres", "postgresql", "pgsql", "mysql", "mariadb"} {
		c = base()
		c.Database.Type = dbType
		c.Database.URL = ""
		c.Database.DSN = ""
		if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "requires a valid DATABASE_URL") {
			t.Errorf("expected %s validation error, got: %v", dbType, err)
		}

		c.Database.URL = "postgres://user:pass@localhost:5432/db"
		if err := c.Validate(); err != nil {
			t.Errorf("expected valid %s config, got: %v", dbType, err)
		}
	}

	// 6. Unsupported Database Type
	c = base()
	c.Database.Type = "oracle"
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "unsupported database type") {
		t.Errorf("expected unsupported db error, got: %v", err)
	}
}

func TestMasking(t *testing.T) {
	// Standard credentials masking
	c := config.Config{
		App: config.AppConfig{
			JWTSecret: "super-secret-key-12345678901234567890",
		},
		Bot: config.BotConfig{
			Token: "123456:sensitive-token-part",
		},
		Database: config.DatabaseConfig{
			URL: "postgres://dbuser:mypassword@localhost:5432/appdb",
			DSN: "dbuser:mypassword@tcp(localhost:3306)/appdb",
		},
	}

	masked := c.Masked()
	if masked.App.JWTSecret != "***REDACTED***" {
		t.Errorf("expected JWTSecret masked, got: %s", masked.App.JWTSecret)
	}
	if masked.Bot.Token != "123456:***REDACTED***" {
		t.Errorf("expected Bot.Token prefix kept and secret masked, got: %s", masked.Bot.Token)
	}
	if strings.Contains(masked.Database.URL, "mypassword") || !strings.Contains(masked.Database.URL, "://dbuser:***REDACTED***@") {
		t.Errorf("expected URL password masked, got: %s", masked.Database.URL)
	}
	if strings.Contains(masked.Database.DSN, "mypassword") || !strings.Contains(masked.Database.DSN, "dbuser:***REDACTED***@tcp(") {
		t.Errorf("expected DSN password masked, got: %s", masked.Database.DSN)
	}

	// Bot token without colon format
	c2 := config.Config{
		Bot: config.BotConfig{
			Token: "singletokenstringwithoutcolon",
		},
	}
	masked2 := c2.Masked()
	if masked2.Bot.Token != "***REDACTED***" {
		t.Errorf("expected fully redacted bot token, got: %s", masked2.Bot.Token)
	}
}

func TestParseDatabaseURLVariations(t *testing.T) {
	clearEnvs()
	defer clearEnvs()

	tests := []struct {
		url          string
		expectedType string
		expectedPath string
	}{
		{"", "sqlite", "data/publisher.db"},
		{"postgres://user:pass@localhost:5432/db", "postgres", ""},
		{"postgresql://user:pass@localhost:5432/db", "postgres", ""},
		{"pgsql://user:pass@localhost:5432/db", "postgres", ""},
		{"mysql://user:pass@localhost:3306/db", "mysql", ""},
		{"mariadb://user:pass@localhost:3306/db", "mysql", ""},
		{"sqlite://data/custom.db", "sqlite", "data/custom.db"},
		{"sqlite:data/short.db", "sqlite", "data/short.db"},
		{"file:data/file.db", "sqlite", "data/file.db"},
		{"user:pass@tcp(localhost:3306)/mysqldb", "mysql", ""},
		{"data/plain_file.db", "sqlite", "data/plain_file.db"},
	}

	for _, tt := range tests {
		_ = os.Setenv("DATABASE_URL", tt.url)
		cfg, err := config.Load("")
		if err != nil {
			t.Fatalf("unexpected error for URL '%s': %v", tt.url, err)
		}
		if cfg.Database.Type != tt.expectedType {
			t.Errorf("URL '%s': expected type '%s', got '%s'", tt.url, tt.expectedType, cfg.Database.Type)
		}
		if tt.expectedPath != "" && cfg.Database.FilePath != tt.expectedPath {
			t.Errorf("URL '%s': expected FilePath '%s', got '%s'", tt.url, tt.expectedPath, cfg.Database.FilePath)
		}
	}
}
