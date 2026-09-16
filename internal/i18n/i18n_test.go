package i18n_test

import (
	"strings"
	"testing"

	"github.com/vyntechau/TelegramPublisher/internal/i18n"
)

func TestTranslate(t *testing.T) {
	// 1. Direct language hit without args
	faWelcome := i18n.T("fa", "welcome_title")
	if faWelcome == "" || faWelcome == "welcome_title" {
		t.Errorf("expected Persian translation, got: %s", faWelcome)
	}

	// 2. Direct language hit with formatting args
	enChanged := i18n.T("en", "lang_changed", "Persian")
	if !strings.Contains(enChanged, "Persian") {
		t.Errorf("expected formatted string containing Persian, got: %s", enChanged)
	}

	// 3. Unknown language falls back to English (without args)
	fallbackNoArgs := i18n.T("xx_unknown", "welcome_title")
	if !strings.Contains(fallbackNoArgs, "TelegramPublisher") {
		t.Errorf("expected English fallback, got: %s", fallbackNoArgs)
	}

	// 4. Unknown language falls back to English (with args)
	fallbackWithArgs := i18n.T("xx_unknown", "lang_changed", "Spanish")
	if !strings.Contains(fallbackWithArgs, "Spanish") {
		t.Errorf("expected formatted English fallback, got: %s", fallbackWithArgs)
	}

	// 5. Existing language with missing key falls back to English
	// First let's create a scenario where key is not in "fa" but exists in "en":
	// If all keys exist in both, we can test unknown key which falls through to returning the key.
	unknownKey := i18n.T("fa", "completely_unknown_key_xyz")
	if unknownKey != "completely_unknown_key_xyz" {
		t.Errorf("expected key itself for non-existent key, got: %s", unknownKey)
	}

	// Unknown key with non-existent language
	unknownLangAndKey := i18n.T("unknown_lang", "non_existent_key")
	if unknownLangAndKey != "non_existent_key" {
		t.Errorf("expected key itself, got: %s", unknownLangAndKey)
	}
}

func TestGetLanguageMeta(t *testing.T) {
	// Valid code
	faMeta := i18n.GetLanguageMeta("fa")
	if faMeta.Code != "fa" || !faMeta.IsRTL {
		t.Errorf("unexpected metadata for 'fa': %+v", faMeta)
	}

	// Upper case / whitespace
	enMeta := i18n.GetLanguageMeta(" EN ")
	if enMeta.Code != "en" || enMeta.IsRTL {
		t.Errorf("unexpected metadata for ' EN ': %+v", enMeta)
	}

	// Unknown code defaults to English
	defaultMeta := i18n.GetLanguageMeta("unknown_lang_code")
	if defaultMeta.Code != "en" {
		t.Errorf("expected default English, got: %+v", defaultMeta)
	}
}

func TestIsSupported(t *testing.T) {
	if !i18n.IsSupported("en") {
		t.Errorf("expected 'en' to be supported")
	}
	if !i18n.IsSupported(" FA ") {
		t.Errorf("expected ' FA ' to be supported")
	}
	if i18n.IsSupported("invalid_code") {
		t.Errorf("expected 'invalid_code' to not be supported")
	}
}

func TestFilterSupported(t *testing.T) {
	// Empty CSV returns all
	all := i18n.FilterSupported("")
	if len(all) != len(i18n.SupportedLanguages) {
		t.Errorf("expected all languages, got %d", len(all))
	}

	// Partial filter
	filtered := i18n.FilterSupported("en, fa, zh")
	if len(filtered) != 3 {
		t.Errorf("expected 3 languages, got %d", len(filtered))
	}

	// None matched returns all
	fallback := i18n.FilterSupported("xx, yy, zz")
	if len(fallback) != len(i18n.SupportedLanguages) {
		t.Errorf("expected fallback to all languages when none match, got %d", len(fallback))
	}
}
