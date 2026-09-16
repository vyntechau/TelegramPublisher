package scalar_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyntechau/TelegramPublisher/internal/api/scalar"
)

func TestScalarRegister(t *testing.T) {
	tempDir := t.TempDir()
	openAPIFile := filepath.Join(tempDir, "openapi.json")
	_ = os.WriteFile(openAPIFile, []byte(`{"openapi":"3.0.0"}`), 0644)

	mux := http.NewServeMux()
	scalar.Register(mux, openAPIFile)

	// Test 1: GET /docs
	reqDocs := httptest.NewRequest(http.MethodGet, "/docs", nil)
	recDocs := httptest.NewRecorder()
	mux.ServeHTTP(recDocs, reqDocs)

	if recDocs.Code != http.StatusOK {
		t.Fatalf("expected 200 for /docs, got %d", recDocs.Code)
	}
	if !strings.Contains(recDocs.Body.String(), "Scalar") {
		t.Errorf("expected body to contain 'Scalar', got: %s", recDocs.Body.String())
	}

	// Test 2: GET /docs/openapi.json (existing file)
	reqJSON := httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil)
	recJSON := httptest.NewRecorder()
	mux.ServeHTTP(recJSON, reqJSON)

	if recJSON.Code != http.StatusOK {
		t.Fatalf("expected 200 for /docs/openapi.json, got %d", recJSON.Code)
	}
	if !strings.Contains(recJSON.Body.String(), "3.0.0") {
		t.Errorf("expected body to contain '3.0.0', got: %s", recJSON.Body.String())
	}

	// Test 3: GET /docs/openapi.json fallback (file doesn't exist)
	muxFallback := http.NewServeMux()
	scalar.Register(muxFallback, filepath.Join(tempDir, "nonexistent.json"))
	recFallback := httptest.NewRecorder()
	muxFallback.ServeHTTP(recFallback, reqJSON)
	if recFallback.Code != http.StatusOK {
		t.Fatalf("expected 200 for fallback, got %d", recFallback.Code)
	}
}
