package settings

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/vyntechau/TelegramPublisher/config"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
)

// Setting keys constants stored in the database 'settings' table.
const (
	KeyWebEnabled           = "web_enabled"
	KeyAdminWebEnabled      = "admin_web_enabled"
	KeyBotAdminEnabled      = "bot_admin_enabled"
	KeyMiniAppEnabled       = "mini_app_enabled"
	KeyMiniAppURL           = "mini_app_url"
	KeyAutoDeleteSeconds    = "auto_delete_seconds"
	KeyForceSubEnabled      = "force_sub_enabled"
	KeyCopyrightWarningText = "copyright_warning_text"
	KeyBotToken             = "bot_token"
	KeyCustomButtonsJSON    = "custom_buttons_json"

	// Telegram Bot Keyboard & Post Button Layout Modes
	KeyKeyboardMode       = "keyboard_mode"        // "inline", "persistent", "both"
	KeyShowForwardButton  = "show_forward_button"  // bool: true/false
	KeyShowReportButton   = "show_report_button"   // bool: true/false
	KeyShowReactions      = "show_reactions"       // bool: true/false
	KeyShowMiniAppButton  = "show_mini_app_button" // bool: true/false

	// Auto-Post to Channels Settings
	KeyAutoPostEnabled  = "auto_post_enabled"  // bool: true/false
	KeyAutoPostChannels = "auto_post_channels" // comma/newline-separated list of target channel IDs or @usernames
	KeyAutoPostFormat   = "auto_post_format"   // "teaser_with_button", "full_media"

	// Multi-Provider Crypto Subscription Gateway Settings (AzPays, Coinbase Commerce, NOWPayments.io)
	KeySubscriptionEnabled     = "subscription_enabled"
	KeySubscriptionGateway     = "subscription_gateway" // "azpays", "coinbase", "nowpayments"
	KeySubscriptionAPIKey      = "subscription_api_key"
	KeySubscriptionSecretKey   = "subscription_secret_key"
	KeySubscriptionCallbackURL = "subscription_callback_url"
	KeySubscriptionSuccessURL  = "subscription_success_url"

	// Legacy aliases maintained for backwards compatibility
	KeyAzpaysEnabled     = "azpays_enabled"
	KeyAzpaysAPIKey      = "azpays_api_key"
	KeyAzpaysSecretKey   = "azpays_secret_key"
	KeyAzpaysCallbackURL = "azpays_callback_url"
	KeyAzpaysSuccessURL  = "azpays_success_url"
	KeyAzpaysBaseURL     = "azpays_base_url"

	// Setup / Onboarding State Keys
	KeyAPIURL              = "api_url"              // Public / client API base URL
	KeyOnboardingStep      = "onboarding_step"      // Last active onboarding step (1-5)
	KeyOnboardingCompleted = "onboarding_completed" // bool: true/false

	// Localization & Internationalization Settings
	KeyDefaultLanguage    = "default_language"    // e.g. "en", "fa", "ar", "ru", "es", "de", "zh"
	KeySupportedLanguages = "supported_languages" // comma-separated list of enabled languages, e.g. "en,fa,ar,ru,es,de,zh"
)

// Service provides dynamic settings retrieval from the database settings table with secure fallbacks.
type Service struct {
	repo storage.Repository
	cfg  *config.Config
	mu   sync.RWMutex
}

func NewService(repo storage.Repository, cfg *config.Config) *Service {
	return &Service{
		repo: repo,
		cfg:  cfg,
	}
}

// GetBool returns a boolean setting from the database settings table.
func (s *Service) GetBool(ctx context.Context, key string, fallback bool) bool {
	val, err := s.repo.GetSetting(ctx, key)
	if err == nil && val != "" {
		switch strings.ToLower(val) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return fallback
}

// GetInt returns an integer setting from the database settings table.
func (s *Service) GetInt(ctx context.Context, key string, fallback int) int {
	val, err := s.repo.GetSetting(ctx, key)
	if err == nil && val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return fallback
}

// GetString returns a string setting from the database settings table.
func (s *Service) GetString(ctx context.Context, key string, fallback string) string {
	val, err := s.repo.GetSetting(ctx, key)
	if err == nil && val != "" {
		return val
	}
	return fallback
}

// Set updates a setting in the database settings table.
func (s *Service) Set(ctx context.Context, key, value, description string) error {
	return s.repo.SetSetting(ctx, key, value, description)
}

// GetAll returns a merged map of all active settings from the database table.
func (s *Service) GetAll(ctx context.Context) (map[string]string, error) {
	dbSettings, err := s.repo.ListSettings(ctx)
	if err != nil {
		dbSettings = make(map[string]string)
	}

	// Subscription gateway resolution with legacy fallback
	subEnabled := s.GetBool(ctx, KeySubscriptionEnabled, s.GetBool(ctx, KeyAzpaysEnabled, false))
	subGateway := s.GetString(ctx, KeySubscriptionGateway, "azpays")
	subAPIKey := s.GetString(ctx, KeySubscriptionAPIKey, s.GetString(ctx, KeyAzpaysAPIKey, ""))
	subSecretKey := s.GetString(ctx, KeySubscriptionSecretKey, s.GetString(ctx, KeyAzpaysSecretKey, ""))
	subCallback := s.GetString(ctx, KeySubscriptionCallbackURL, s.GetString(ctx, KeyAzpaysCallbackURL, "http://localhost:8080/api/v1/payments/subscription/webhook"))
	subSuccess := s.GetString(ctx, KeySubscriptionSuccessURL, s.GetString(ctx, KeyAzpaysSuccessURL, "http://localhost:8080/payment/success"))

	res := map[string]string{
		KeyWebEnabled:              strconv.FormatBool(s.GetBool(ctx, KeyWebEnabled, true)),
		KeyAdminWebEnabled:         strconv.FormatBool(s.GetBool(ctx, KeyAdminWebEnabled, true)),
		KeyBotAdminEnabled:         strconv.FormatBool(s.GetBool(ctx, KeyBotAdminEnabled, true)),
		KeyMiniAppEnabled:          strconv.FormatBool(s.GetBool(ctx, KeyMiniAppEnabled, true)),
		KeyMiniAppURL:              s.GetString(ctx, KeyMiniAppURL, "http://localhost:8080"),
		KeyAutoDeleteSeconds:       strconv.Itoa(s.GetInt(ctx, KeyAutoDeleteSeconds, 120)),
		KeyForceSubEnabled:         strconv.FormatBool(s.GetBool(ctx, KeyForceSubEnabled, true)),
		KeyCopyrightWarningText:    s.GetString(ctx, KeyCopyrightWarningText, "⏳ *Copyright Protection*: This content will be automatically deleted in 2 minutes."),
		KeyCustomButtonsJSON:       s.GetString(ctx, KeyCustomButtonsJSON, `[[{"text":"🌐 Official Website","url":"https://vyntech.cloud"},{"text":"⭐ VIP Subscription","url":"https://t.me/yourbot?start=vip"}]]`),
		KeyKeyboardMode:            s.GetString(ctx, KeyKeyboardMode, "both"),
		KeyShowForwardButton:       strconv.FormatBool(s.GetBool(ctx, KeyShowForwardButton, true)),
		KeyShowReportButton:        strconv.FormatBool(s.GetBool(ctx, KeyShowReportButton, true)),
		KeyShowReactions:           strconv.FormatBool(s.GetBool(ctx, KeyShowReactions, true)),
		KeyShowMiniAppButton:       strconv.FormatBool(s.GetBool(ctx, KeyShowMiniAppButton, true)),
		KeyAutoPostEnabled:         strconv.FormatBool(s.GetBool(ctx, KeyAutoPostEnabled, false)),
		KeyAutoPostChannels:        s.GetString(ctx, KeyAutoPostChannels, ""),
		KeyAutoPostFormat:          s.GetString(ctx, KeyAutoPostFormat, "teaser_with_button"),
		KeySubscriptionEnabled:     strconv.FormatBool(subEnabled),
		KeySubscriptionGateway:     subGateway,
		KeySubscriptionAPIKey:      subAPIKey,
		KeySubscriptionSecretKey:   subSecretKey,
		KeySubscriptionCallbackURL: subCallback,
		KeySubscriptionSuccessURL:  subSuccess,
		KeyAzpaysEnabled:           strconv.FormatBool(subEnabled),
		KeyAzpaysAPIKey:            subAPIKey,
		KeyAzpaysSecretKey:         subSecretKey,
		KeyAzpaysCallbackURL:       subCallback,
		KeyAzpaysSuccessURL:        subSuccess,
		KeyAPIURL:                  s.GetString(ctx, KeyAPIURL, "http://localhost:8080"),
		KeyOnboardingStep:         s.GetString(ctx, KeyOnboardingStep, "1"),
		KeyOnboardingCompleted:    strconv.FormatBool(s.GetBool(ctx, KeyOnboardingCompleted, false)),
	}

	for k, v := range dbSettings {
		res[k] = v
	}

	return res, nil
}
