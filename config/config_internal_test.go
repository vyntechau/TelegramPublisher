package config

import (
	"errors"
	"testing"
)

func TestParseDatabaseURLEmpty(t *testing.T) {
	dbType, dsn, filePath := parseDatabaseURL("")
	if dbType != "sqlite" || dsn != "data/publisher.db" || filePath != "data/publisher.db" {
		t.Errorf("expected default sqlite on empty URL, got dbType=%s, dsn=%s, filePath=%s", dbType, dsn, filePath)
	}

	dbType, dsn, filePath = parseDatabaseURL("   ")
	if dbType != "sqlite" || dsn != "data/publisher.db" || filePath != "data/publisher.db" {
		t.Errorf("expected default sqlite on whitespace URL, got dbType=%s, dsn=%s, filePath=%s", dbType, dsn, filePath)
	}
}

func TestGenerateSecureRandomKeyFallback(t *testing.T) {
	origRandRead := randRead
	defer func() { randRead = origRandRead }()

	// Force randRead to fail
	randRead = func(b []byte) (n int, err error) {
		return 0, errors.New("simulated entropy failure")
	}

	key := generateSecureRandomKey(32)
	expected := "76796e746563682d66616c6c6261636b2d7365637265742d736565642d6b6579"
	if key != expected {
		t.Errorf("expected fallback key %s, got %s", expected, key)
	}
}
