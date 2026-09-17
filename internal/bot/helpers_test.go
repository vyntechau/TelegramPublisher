package bot

import (
	"testing"
)

func TestIsValidPublicURL(t *testing.T) {
	tests := []struct {
		url          string
		requireHTTPS bool
		want         bool
	}{
		{"http://localhost:8080", false, false},
		{"http://localhost:8080", true, false},
		{"https://localhost:8080", true, false},
		{"http://127.0.0.1:8080", false, false},
		{"http://0.0.0.0:8080", false, false},
		{"http://myhost.local", false, false},
		{"http://myhost.internal", false, false},
		{"http://app", false, false},
		{"", false, false},
		{"   ", true, false},
		{"invalid-url", false, false},
		{"http://example.com", false, true},
		{"http://example.com", true, false},
		{"https://example.com", true, true},
		{"https://sub.domain.com/app?foo=bar", true, true},
	}

	for _, tt := range tests {
		got := isValidPublicURL(tt.url, tt.requireHTTPS)
		if got != tt.want {
			t.Errorf("isValidPublicURL(%q, %v) = %v, want %v", tt.url, tt.requireHTTPS, got, tt.want)
		}
	}
}
