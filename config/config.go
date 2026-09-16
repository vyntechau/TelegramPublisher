package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// InsecureDefaultJWTSecret is the fallback secret used only in development/testing.
const InsecureDefaultJWTSecret = "vyntech-telegram-publisher-super-secret-key-32b"

var botTokenRegex = regexp.MustCompile(`^\d+:[A-Za-z0-9_-]+$`)

// Config holds the infrastructure bootstrapping configuration.
// Dynamic operational settings (feature toggles, auto-delete TTL, warnings, URLs, AZPays keys)
// are stored and managed directly in the database 'settings' table and editable via Admin UI.
type Config struct {
	App      AppConfig      `yaml:"app"`
	Bot      BotConfig      `yaml:"bot"`
	Database DatabaseConfig `yaml:"database"`
}

// AppConfig holds core HTTP server networking and master security signing secret.
type AppConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Env       string `yaml:"env"`        // development, production, test
	JWTSecret string `yaml:"jwt_secret"` // Cryptographic master secret for JWT authentication
}

// BotConfig holds Telegram API initial connection credentials.
type BotConfig struct {
	Token   string `yaml:"token"`    // Bot token from @BotFather
	OwnerID int64  `yaml:"owner_id"` // Root owner Telegram ID
}

// DatabaseConfig holds database connection configuration via a single unified URL / DSN.
type DatabaseConfig struct {
	URL      string `yaml:"url"`       // Unified connection string (e.g., postgres://..., mysql://..., sqlite://..., or file path)
	Type     string `yaml:"type"`      // Auto-inferred from URL scheme or explicitly specified (sqlite, postgres, mysql, mariadb)
	DSN      string `yaml:"dsn"`       // Normalized DSN passed to the database driver
	FilePath string `yaml:"file_path"` // SQLite database file path
}

// Load loads infrastructure configuration with precedence:
// 1. OS Environment Variables (Highest priority)
// 2. .env file
// 3. config.yaml or config.yml (Lowest priority)
// 4. Built-in secure defaults
func Load(configPath string) (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Host:      "0.0.0.0",
			Port:      8080,
			Env:       "production",
			JWTSecret: InsecureDefaultJWTSecret,
		},
		Bot: BotConfig{
			Token:   "",
			OwnerID: 0,
		},
		Database: DatabaseConfig{
			URL: "sqlite://data/publisher.db",
		},
	}

	// 1. Try loading YAML config
	yamlFiles := []string{configPath, "config.yaml", "config.yml", "config/config.yaml"}
	for _, yf := range yamlFiles {
		if yf == "" {
			continue
		}
		if data, err := os.ReadFile(yf); err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse YAML config %s: %w", yf, err)
			}
			break
		}
	}

	// 2. Try loading .env file (does not override already set OS env)
	_ = godotenv.Load(".env", "config/.env")

	// 3. Override with Environment Variables
	overrideFromEnv(cfg)

	// 4. Auto-detect database type and normalize connection DSN
	inferredType, dsn, filePath := parseDatabaseURL(cfg.Database.URL)
	if cfg.Database.Type == "" {
		cfg.Database.Type = inferredType
	}
	cfg.Database.DSN = dsn
	cfg.Database.FilePath = filePath

	// 5. Auto-generate secure cryptographic key if empty in non-production
	if cfg.App.JWTSecret == "" {
		if cfg.App.Env != "production" {
			cfg.App.JWTSecret = generateSecureRandomKey(32)
		}
	}

	return cfg, nil
}

// Validate performs security sanity and integrity checks on the bootstrapping configuration.
func (c *Config) Validate() error {
	// 1. App Validation
	if c.App.Port < 1 || c.App.Port > 65535 {
		return fmt.Errorf("invalid app port: %d (must be between 1 and 65535)", c.App.Port)
	}

	// In production, enforce strong JWT secret and disallow known default key
	if strings.ToLower(c.App.Env) == "production" {
		if c.App.JWTSecret == InsecureDefaultJWTSecret || len(c.App.JWTSecret) < 32 {
			return errors.New("security error: in production, JWT_SECRET must be explicitly provided and at least 32 characters long")
		}
	}

	// 2. Bot Validation (if token provided)
	if c.Bot.Token != "" {
		if !botTokenRegex.MatchString(c.Bot.Token) {
			return fmt.Errorf("invalid bot token format: must match '<bot_id>:<token_string>'")
		}
	}

	// 3. Database Validation
	dbType := strings.ToLower(c.Database.Type)
	switch dbType {
	case "sqlite", "sqlite3":
		if strings.TrimSpace(c.Database.FilePath) == "" && strings.TrimSpace(c.Database.URL) == "" {
			return errors.New("sqlite database requires a non-empty file path or DATABASE_URL")
		}
	case "postgres", "postgresql", "pgsql", "mysql", "mariadb":
		if strings.TrimSpace(c.Database.URL) == "" && strings.TrimSpace(c.Database.DSN) == "" {
			return fmt.Errorf("%s database requires a valid DATABASE_URL / DB_DSN", dbType)
		}
	default:
		return fmt.Errorf("unsupported database type: '%s' (supported: sqlite, postgres, mysql, mariadb)", c.Database.Type)
	}

	return nil
}

// Masked returns a safe-to-log copy of the Config with all secrets redacted.
func (c *Config) Masked() Config {
	masked := *c

	// Mask JWT Secret
	if len(masked.App.JWTSecret) > 0 {
		masked.App.JWTSecret = "***REDACTED***"
	}

	// Mask Bot Token (keep prefix ID for debugging)
	if len(masked.Bot.Token) > 0 {
		parts := strings.SplitN(masked.Bot.Token, ":", 2)
		if len(parts) == 2 {
			masked.Bot.Token = parts[0] + ":***REDACTED***"
		} else {
			masked.Bot.Token = "***REDACTED***"
		}
	}

	// Mask Database URL and DSN
	if len(masked.Database.URL) > 0 {
		masked.Database.URL = maskDSN(masked.Database.URL)
	}
	if len(masked.Database.DSN) > 0 {
		masked.Database.DSN = maskDSN(masked.Database.DSN)
	}

	return masked
}

func maskDSN(dsn string) string {
	// Replaces passwords in format user:password@host or scheme://user:pass@host
	re := regexp.MustCompile(`://([^:]+):([^@]+)@`)
	masked := re.ReplaceAllString(dsn, "://$1:***REDACTED***@")

	reRaw := regexp.MustCompile(`^([^:]+):([^@]+)@tcp\(`)
	return reRaw.ReplaceAllString(masked, "$1:***REDACTED***@tcp(")
}

func parseDatabaseURL(rawURL string) (dbType string, dsn string, filePath string) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "sqlite", "data/publisher.db", "data/publisher.db"
	}

	lower := strings.ToLower(rawURL)
	if strings.HasPrefix(lower, "postgres://") || strings.HasPrefix(lower, "postgresql://") || strings.HasPrefix(lower, "pgsql://") {
		return "postgres", rawURL, ""
	}

	if strings.HasPrefix(lower, "mysql://") || strings.HasPrefix(lower, "mariadb://") {
		return "mysql", rawURL, ""
	}

	if strings.HasPrefix(lower, "sqlite://") {
		path := strings.TrimPrefix(rawURL, "sqlite://")
		return "sqlite", path, path
	}

	if strings.HasPrefix(lower, "sqlite:") {
		path := strings.TrimPrefix(rawURL, "sqlite:")
		return "sqlite", path, path
	}

	if strings.HasPrefix(lower, "file:") {
		path := strings.TrimPrefix(rawURL, "file:")
		return "sqlite", path, path
	}

	if strings.Contains(rawURL, "@tcp(") {
		return "mysql", rawURL, ""
	}

	// File path
	return "sqlite", rawURL, rawURL
}

var randRead = rand.Read

func generateSecureRandomKey(bytesLen int) string {
	b := make([]byte, bytesLen)
	if _, err := randRead(b); err != nil {
		return hex.EncodeToString([]byte("vyntech-fallback-secret-seed-key"))
	}
	return hex.EncodeToString(b)
}

func overrideFromEnv(cfg *Config) {
	// App Bootstrap & Security
	if v := strings.TrimSpace(os.Getenv("APP_HOST")); v != "" {
		cfg.App.Host = v
	}
	if v := strings.TrimSpace(os.Getenv("APP_PORT")); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.App.Port = p
		}
	}
	if v := strings.TrimSpace(os.Getenv("APP_ENV")); v != "" {
		cfg.App.Env = v
	}
	if v := strings.TrimSpace(os.Getenv("JWT_SECRET")); v != "" {
		cfg.App.JWTSecret = v
	}

	// Bot Credentials
	if v := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")); v != "" {
		cfg.Bot.Token = v
	} else if v := strings.TrimSpace(os.Getenv("BOT_TOKEN")); v != "" {
		cfg.Bot.Token = v
	}
	if v := strings.TrimSpace(os.Getenv("TELEGRAM_OWNER_ID")); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			cfg.Bot.OwnerID = id
		}
	}

	// Unified Database Connection URL / DSN (Single env)
	if v := strings.TrimSpace(os.Getenv("DATABASE_URL")); v != "" {
		cfg.Database.URL = v
	} else if v := strings.TrimSpace(os.Getenv("DB_URL")); v != "" {
		cfg.Database.URL = v
	} else if v := strings.TrimSpace(os.Getenv("DB_DSN")); v != "" {
		cfg.Database.URL = v
	} else if v := strings.TrimSpace(os.Getenv("DB_FILE_PATH")); v != "" {
		cfg.Database.URL = v
	}

	if v := strings.TrimSpace(os.Getenv("DB_TYPE")); v != "" {
		cfg.Database.Type = strings.ToLower(v)
	}
}
