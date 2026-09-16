package auth_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vyntechau/TelegramPublisher/internal/auth"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
)

func makeInitData(botToken string, userJSON, authDate string, overrideHash string) string {
	dataCheckString := fmt.Sprintf("auth_date=%s\nquery_id=AAHdF6IQAAAAAN0XohDhrOrc", authDate)
	if userJSON != "" {
		dataCheckString += fmt.Sprintf("\nuser=%s", userJSON)
	}

	hash := overrideHash
	if hash == "" {
		secretKeyHmac := hmac.New(sha256.New, []byte("WebAppData"))
		secretKeyHmac.Write([]byte(botToken))
		secretKey := secretKeyHmac.Sum(nil)

		checkHmac := hmac.New(sha256.New, secretKey)
		checkHmac.Write([]byte(dataCheckString))
		hash = hex.EncodeToString(checkHmac.Sum(nil))
	}

	result := fmt.Sprintf("query_id=AAHdF6IQAAAAAN0XohDhrOrc&auth_date=%s&hash=%s", authDate, hash)
	if userJSON != "" {
		result += fmt.Sprintf("&user=%s", url.QueryEscape(userJSON))
	}
	return result
}

func TestTelegramInitDataValidationAndJWT(t *testing.T) {
	botToken := "123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ"
	jwtSecret := "test-super-secret-key-32b-length"

	svc := auth.NewService(botToken, jwtSecret, nil)

	// Valid Telegram initData
	authDate := fmt.Sprintf("%d", time.Now().Unix())
	userJSON := `{"id":123456,"first_name":"Alice","last_name":"Smith","username":"alice","photo_url":"https://t.me/p.jpg"}`
	rawInitData := makeInitData(botToken, userJSON, authDate, "")

	tgUser, err := svc.ValidateInitData(rawInitData)
	if err != nil {
		t.Fatalf("expected valid initData, got error: %v", err)
	}

	if tgUser.ID != 123456 || tgUser.Username != "alice" || tgUser.LastName != "Smith" {
		t.Errorf("expected user ID 123456 and username 'alice', got %+v", tgUser)
	}

	// JWT Generation & Validation
	user := &storage.User{
		ID:         1,
		TelegramID: 123456,
		Username:   "alice",
		Role:       storage.RoleAuthor,
	}

	token, err := svc.GenerateJWT(user)
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}

	claims, err := svc.ValidateJWT(token)
	if err != nil {
		t.Fatalf("failed to validate JWT: %v", err)
	}

	if claims.TelegramID != 123456 || claims.Role != storage.RoleAuthor || claims.UserID != 1 || claims.Username != "alice" {
		t.Errorf("claims mismatch: %+v", claims)
	}
}

func TestValidateInitDataErrors(t *testing.T) {
	botToken := "123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ"
	jwtSecret := "test-super-secret-key-32b-length"
	svc := auth.NewService(botToken, jwtSecret, nil)

	// 1. Empty init data
	if _, err := svc.ValidateInitData(""); err == nil || err.Error() != "empty init data" {
		t.Errorf("expected empty init data error, got: %v", err)
	}

	// 2. Invalid query encoding
	if _, err := svc.ValidateInitData("%zz&invalid"); err == nil {
		t.Errorf("expected url query parse error, got nil")
	}

	// 3. Missing hash
	if _, err := svc.ValidateInitData("user=%7B%7D&auth_date=123"); err == nil || err.Error() != "missing hash in init data" {
		t.Errorf("expected missing hash error, got: %v", err)
	}

	// 4. Missing auth_date
	if _, err := svc.ValidateInitData("hash=abcdef&user=%7B%7D"); err == nil || err.Error() != "missing auth_date" {
		t.Errorf("expected missing auth_date error, got: %v", err)
	}

	// 5. Invalid auth_date format
	if _, err := svc.ValidateInitData("hash=abcdef&auth_date=not_a_number"); err == nil || err.Error() != "invalid auth_date" {
		t.Errorf("expected invalid auth_date error, got: %v", err)
	}

	// 6. Expired auth_date (> 24 hours ago)
	expiredDate := fmt.Sprintf("%d", time.Now().Unix()-90000)
	if _, err := svc.ValidateInitData(fmt.Sprintf("hash=abcdef&auth_date=%s", expiredDate)); err == nil || err.Error() != "init data has expired (older than 24h)" {
		t.Errorf("expected expired init data error, got: %v", err)
	}

	// 7. Invalid HMAC signature mismatch
	nowDate := fmt.Sprintf("%d", time.Now().Unix())
	invalidSigData := makeInitData(botToken, `{"id":1}`, nowDate, "0000000000000000000000000000000000000000000000000000000000000000")
	if _, err := svc.ValidateInitData(invalidSigData); err == nil || err.Error() != "invalid init data signature" {
		t.Errorf("expected signature mismatch error, got: %v", err)
	}

	// 8. Missing user field
	missingUserData := makeInitData(botToken, "", nowDate, "")
	if _, err := svc.ValidateInitData(missingUserData); err == nil || err.Error() != "user field missing in init data" {
		t.Errorf("expected user missing error, got: %v", err)
	}

	// 9. Malformed user JSON
	malformedUserData := makeInitData(botToken, "{invalid-json-content", nowDate, "")
	if _, err := svc.ValidateInitData(malformedUserData); err == nil {
		t.Errorf("expected json parse error, got nil")
	}
}

func TestValidateJWTErrors(t *testing.T) {
	botToken := "123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ"
	jwtSecret := "test-super-secret-key-32b-length"
	svc := auth.NewService(botToken, jwtSecret, nil)

	// 1. Malformed token string
	if _, err := svc.ValidateJWT("invalid.jwt.token"); err == nil {
		t.Errorf("expected error on malformed token string, got nil")
	}

	// 2. Token signed with wrong algorithm (e.g. None or SigningMethodNone)
	noneToken := jwt.NewWithClaims(jwt.SigningMethodNone, auth.Claims{
		TelegramID: 123456,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	tokenStr, _ := noneToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if _, err := svc.ValidateJWT(tokenStr); err == nil {
		t.Errorf("expected error for non-HMAC signing method, got nil")
	}

	// 3. Expired token
	expiredUser := &storage.User{
		ID:         1,
		TelegramID: 123456,
		Username:   "bob",
		Role:       storage.RoleUser,
	}
	expiredClaims := auth.Claims{
		TelegramID: expiredUser.TelegramID,
		UserID:     expiredUser.ID,
		Username:   expiredUser.Username,
		Role:       expiredUser.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "TelegramPublisher",
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenStr, _ := expiredToken.SignedString([]byte(jwtSecret))

	if _, err := svc.ValidateJWT(expiredTokenStr); err == nil {
		t.Errorf("expected error for expired JWT token, got nil")
	}
}
