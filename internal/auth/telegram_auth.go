package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
)

// TelegramUserPayload represents the parsed 'user' object from Telegram initData.
type TelegramUserPayload struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
}

// Claims represents JWT claims for authenticated users and admins.
type Claims struct {
	TelegramID int64  `json:"telegram_id"`
	UserID     int64  `json:"user_id"`
	Username   string `json:"username"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

// Service provides Telegram WebApp validation and JWT token issuance.
type Service struct {
	botToken  string
	jwtSecret []byte
	repo      storage.Repository
}

func NewService(botToken, jwtSecret string, repo storage.Repository) *Service {
	return &Service{
		botToken:  botToken,
		jwtSecret: []byte(jwtSecret),
		repo:      repo,
	}
}

// ValidateInitData validates raw initData string from Telegram Mini App using HMAC-SHA256.
func (s *Service) ValidateInitData(rawInitData string) (*TelegramUserPayload, error) {
	if rawInitData == "" {
		return nil, errors.New("empty init data")
	}

	values, err := url.ParseQuery(rawInitData)
	if err != nil {
		return nil, fmt.Errorf("invalid init data query: %w", err)
	}

	receivedHash := values.Get("hash")
	if receivedHash == "" {
		return nil, errors.New("missing hash in init data")
	}

	values.Del("hash")

	// Check auth_date (reject if older than 24h)
	authDateStr := values.Get("auth_date")
	if authDateStr == "" {
		return nil, errors.New("missing auth_date")
	}
	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, errors.New("invalid auth_date")
	}
	if time.Now().Unix()-authDate > 86400 {
		return nil, errors.New("init data has expired (older than 24h)")
	}

	// Sort keys alphabetically
	var keys []string
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, values.Get(k)))
	}
	dataCheckString := strings.Join(pairs, "\n")

	// Secret key = HMAC_SHA256("WebAppData", botToken)
	secretKeyHmac := hmac.New(sha256.New, []byte("WebAppData"))
	secretKeyHmac.Write([]byte(s.botToken))
	secretKey := secretKeyHmac.Sum(nil)

	// Calculate hash
	checkHmac := hmac.New(sha256.New, secretKey)
	checkHmac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(checkHmac.Sum(nil))

	if !hmac.Equal([]byte(receivedHash), []byte(expectedHash)) {
		return nil, errors.New("invalid init data signature")
	}

	// Parse User JSON
	userJSON := values.Get("user")
	if userJSON == "" {
		return nil, errors.New("user field missing in init data")
	}

	var tgUser TelegramUserPayload
	if err := json.Unmarshal([]byte(userJSON), &tgUser); err != nil {
		return nil, fmt.Errorf("failed to parse user json: %w", err)
	}

	return &tgUser, nil
}

// GenerateJWT creates a signed JWT token for a user.
func (s *Service) GenerateJWT(user *storage.User) (string, error) {
	claims := Claims{
		TelegramID: user.TelegramID,
		UserID:     user.ID,
		Username:   user.Username,
		Role:       user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "TelegramPublisher",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateJWT verifies a JWT token string and returns parsed claims.
func (s *Service) ValidateJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}
