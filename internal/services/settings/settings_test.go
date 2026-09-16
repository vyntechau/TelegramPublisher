package settings_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/vyntechau/TelegramPublisher/config"
	"github.com/vyntechau/TelegramPublisher/internal/services/settings"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"github.com/vyntechau/TelegramPublisher/internal/storage/sqlite"
)

type errorSettingsRepo struct {
	storage.Repository
}

func (e *errorSettingsRepo) ListSettings(ctx context.Context) (map[string]string, error) {
	return nil, errors.New("simulated list settings error")
}

func (e *errorSettingsRepo) GetSetting(ctx context.Context, key string) (string, error) {
	return "", errors.New("simulated get setting error")
}

func TestSettingsService(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "settings_test_*")
	defer os.RemoveAll(tempDir)

	repo := sqlite.New(filepath.Join(tempDir, "test.db"))
	_ = repo.Init(context.Background())
	defer repo.Close()

	cfg := &config.Config{}
	svc := settings.NewService(repo, cfg)
	ctx := context.Background()

	// 1. GetBool test cases
	// Default fallback when not present
	if !svc.GetBool(ctx, settings.KeyWebEnabled, true) {
		t.Errorf("expected default web_enabled true")
	}
	if svc.GetBool(ctx, settings.KeyWebEnabled, false) {
		t.Errorf("expected fallback false")
	}

	// Truthy strings
	for _, val := range []string{"true", "1", "yes", "on", "TRUE", "Yes"} {
		_ = svc.Set(ctx, "bool_test_key", val, "")
		if !svc.GetBool(ctx, "bool_test_key", false) {
			t.Errorf("expected value '%s' to resolve to true", val)
		}
	}

	// Falsy strings
	for _, val := range []string{"false", "0", "no", "off", "FALSE", "No"} {
		_ = svc.Set(ctx, "bool_test_key", val, "")
		if svc.GetBool(ctx, "bool_test_key", true) {
			t.Errorf("expected value '%s' to resolve to false", val)
		}
	}

	// Invalid bool string fallback
	_ = svc.Set(ctx, "bool_test_key", "invalid_bool", "")
	if !svc.GetBool(ctx, "bool_test_key", true) {
		t.Errorf("expected fallback true for invalid bool string")
	}

	// 2. GetInt test cases
	if svc.GetInt(ctx, settings.KeyAutoDeleteSeconds, 120) != 120 {
		t.Errorf("expected fallback 120 for missing key")
	}
	_ = svc.Set(ctx, settings.KeyAutoDeleteSeconds, "300", "5 minutes")
	if svc.GetInt(ctx, settings.KeyAutoDeleteSeconds, 120) != 300 {
		t.Errorf("expected auto_delete_seconds 300")
	}
	_ = svc.Set(ctx, settings.KeyAutoDeleteSeconds, "not_an_int", "")
	if svc.GetInt(ctx, settings.KeyAutoDeleteSeconds, 120) != 120 {
		t.Errorf("expected fallback 120 for invalid int string")
	}

	// 3. GetString test cases
	if svc.GetString(ctx, settings.KeyAPIURL, "http://fallback") != "http://fallback" {
		t.Errorf("expected fallback string")
	}
	_ = svc.Set(ctx, settings.KeyAPIURL, "https://api.vyntech.cloud", "")
	if svc.GetString(ctx, settings.KeyAPIURL, "http://fallback") != "https://api.vyntech.cloud" {
		t.Errorf("expected configured API URL")
	}

	// 4. GetAll with DB settings
	all, err := svc.GetAll(ctx)
	if err != nil || len(all) == 0 {
		t.Fatalf("expected all settings map, got %v", err)
	}
	if all[settings.KeyAPIURL] != "https://api.vyntech.cloud" {
		t.Errorf("expected map key api_url https://api.vyntech.cloud, got %s", all[settings.KeyAPIURL])
	}
}

func TestSettingsServiceErrorRepo(t *testing.T) {
	mockErrRepo := &errorSettingsRepo{}
	cfg := &config.Config{}
	svc := settings.NewService(mockErrRepo, cfg)
	ctx := context.Background()

	// GetBool / GetInt / GetString with repo error should return fallback
	if !svc.GetBool(ctx, "any_key", true) {
		t.Errorf("expected fallback true on repo error")
	}
	if svc.GetInt(ctx, "any_key", 42) != 42 {
		t.Errorf("expected fallback 42 on repo error")
	}
	if svc.GetString(ctx, "any_key", "default_val") != "default_val" {
		t.Errorf("expected fallback default_val on repo error")
	}

	// GetAll with repo error
	all, err := svc.GetAll(ctx)
	if err != nil {
		t.Fatalf("expected GetAll to succeed even if repo ListSettings fails, got: %v", err)
	}
	if len(all) == 0 {
		t.Errorf("expected default populated settings in GetAll, got empty map")
	}
}
