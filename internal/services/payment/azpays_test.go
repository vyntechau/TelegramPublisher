package payment_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vyntechau/TelegramPublisher/config"
	"github.com/vyntechau/TelegramPublisher/internal/services/payment"
	"github.com/vyntechau/TelegramPublisher/internal/services/settings"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"github.com/vyntechau/TelegramPublisher/internal/storage/sqlite"
	"gopkg.in/telebot.v3"
)

type mockStorageRepo struct {
	storage.Repository
	createSubErr    error
	getSubErr       error
	updateSubErr    error
}

func (m *mockStorageRepo) CreateSubscription(ctx context.Context, sub *storage.Subscription) error {
	if m.createSubErr != nil {
		return m.createSubErr
	}
	return m.Repository.CreateSubscription(ctx, sub)
}

func (m *mockStorageRepo) GetSubscriptionByInvoice(ctx context.Context, invoiceID string) (*storage.Subscription, error) {
	if m.getSubErr != nil {
		return nil, m.getSubErr
	}
	return m.Repository.GetSubscriptionByInvoice(ctx, invoiceID)
}

func (m *mockStorageRepo) UpdateSubscriptionStatus(ctx context.Context, invoiceID, status string, expiresAt time.Time) error {
	if m.updateSubErr != nil {
		return m.updateSubErr
	}
	return m.Repository.UpdateSubscriptionStatus(ctx, invoiceID, status, expiresAt)
}

func setupTestEnv(t *testing.T) (*sqlite.SQLiteRepository, *settings.Service, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "pay_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	repo := sqlite.New(filepath.Join(tempDir, "test.db"))
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("failed to init sqlite repo: %v", err)
	}

	settingsSvc := settings.NewService(repo, &config.Config{})

	cleanup := func() {
		_ = repo.Close()
		_ = os.RemoveAll(tempDir)
	}

	return repo, settingsSvc, cleanup
}

func computeSignatureHeader(payload []byte, secret string) string {
	ts := time.Now().Unix()
	msg := fmt.Sprintf("%d.%s", ts, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("t=%d,v1=%s", ts, sig)
}

func TestCreateCheckoutSession_ValidationAndDefaults(t *testing.T) {
	repo, settingsSvc, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := payment.NewService(repo, settingsSvc, nil)
	ctx := context.Background()

	// 1. Invalid plan ID
	_, err := svc.CreateCheckoutSession(ctx, 123, "invalid_plan_id", "USDT")
	if err == nil || !strings.Contains(err.Error(), "invalid plan id") {
		t.Fatalf("expected 'invalid plan id' error, got: %v", err)
	}

	// 2. Default currency fallback when currency is empty
	resp, err := svc.CreateCheckoutSession(ctx, 123, "vip_30d", "")
	if err != nil {
		t.Fatalf("expected success with default currency, got: %v", err)
	}
	if resp.Currency != "USDT" {
		t.Errorf("expected currency to default to USDT, got: %s", resp.Currency)
	}

	// 3. Sub enabled but apiKey is empty (fallback URL without external gateway call)
	_ = settingsSvc.Set(ctx, settings.KeySubscriptionEnabled, "true", "")
	_ = settingsSvc.Set(ctx, settings.KeySubscriptionAPIKey, "", "")
	respEmptyKey, err := svc.CreateCheckoutSession(ctx, 123, "vip_30d", "USDT")
	if err != nil {
		t.Fatalf("expected checkout creation with empty key, got: %v", err)
	}
	if !strings.HasPrefix(respEmptyKey.PaymentURL, "https://checkout.") {
		t.Errorf("expected direct fallback payment URL, got: %s", respEmptyKey.PaymentURL)
	}

	// 4. Repo failure on CreateSubscription
	mockRepo := &mockStorageRepo{
		Repository:   repo,
		createSubErr: errors.New("simulated db create sub error"),
	}
	svcErr := payment.NewService(mockRepo, settingsSvc, nil)
	_, err = svcErr.CreateCheckoutSession(ctx, 123, "vip_30d", "USDT")
	if err == nil || !strings.Contains(err.Error(), "failed to save subscription") {
		t.Fatalf("expected failed to save subscription error, got: %v", err)
	}
}

func TestCreateCheckoutSession_Gateways(t *testing.T) {
	repo, settingsSvc, cleanup := setupTestEnv(t)
	defer cleanup()

	ctx := context.Background()

	// 1. AzPays gateway with mock server returning valid payment with Token
	var returnEmptyToken bool
	var return500Error bool
	var invCounter int

	mockAzPaysServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if return500Error {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"ok":false,"msg":"internal error"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		invCounter++
		token := "tok_sample_test_999"
		if returnEmptyToken {
			token = ""
		}
		respJSON := fmt.Sprintf(`{"ok":true,"msg":"ok","data":{"id":"az_inv_%d","token":"%s"}}`, 100+invCounter, token)
		_, _ = w.Write([]byte(respJSON))
	}))
	defer mockAzPaysServer.Close()

	_ = settingsSvc.Set(ctx, settings.KeySubscriptionEnabled, "true", "")
	_ = settingsSvc.Set(ctx, settings.KeySubscriptionAPIKey, "test_api_key", "")
	_ = settingsSvc.Set(ctx, settings.KeySubscriptionGateway, "azpays", "")
	_ = settingsSvc.Set(ctx, settings.KeyAzpaysBaseURL, mockAzPaysServer.URL, "")

	svc := payment.NewService(repo, settingsSvc, nil)

	// Test AzPays successful token
	resp, err := svc.CreateCheckoutSession(ctx, 1001, "vip_30d", "USDT")
	if err != nil {
		t.Fatalf("expected successful azpays session, got: %v", err)
	}
	if resp.InvoiceID != "az_inv_101" || !strings.Contains(resp.PaymentURL, "tok_sample_test_999") {
		t.Errorf("unexpected azpays invoice response: %+v", resp)
	}

	// Test AzPays with empty token in response
	returnEmptyToken = true
	respNoTok, err := svc.CreateCheckoutSession(ctx, 1002, "vip_30d", "USDT")
	if err != nil {
		t.Fatalf("expected successful azpays session without token, got: %v", err)
	}
	if respNoTok.InvoiceID != "az_inv_102" {
		t.Errorf("expected invoice id az_inv_102, got: %s", respNoTok.InvoiceID)
	}

	// Test AzPays with server error (falls back to direct invoice)
	return500Error = true
	respErrFallback, err := svc.CreateCheckoutSession(ctx, 1003, "vip_30d", "USDT")
	if err != nil {
		t.Fatalf("expected fallback invoice on azpays error, got: %v", err)
	}
	if !strings.HasPrefix(respErrFallback.PaymentURL, "https://checkout.azpays.net/pay/") {
		t.Errorf("expected fallback URL, got: %s", respErrFallback.PaymentURL)
	}

	// 2. Gateway: coinbase
	_ = settingsSvc.Set(ctx, settings.KeySubscriptionGateway, "coinbase", "")
	respCoinbase, err := svc.CreateCheckoutSession(ctx, 1004, "vip_30d", "USDT")
	if err != nil {
		t.Fatalf("expected coinbase checkout session, got: %v", err)
	}
	if !strings.HasPrefix(respCoinbase.PaymentURL, "https://commerce.coinbase.com/charges/") {
		t.Errorf("expected coinbase payment URL, got: %s", respCoinbase.PaymentURL)
	}

	// 3. Gateway: nowpayments
	_ = settingsSvc.Set(ctx, settings.KeySubscriptionGateway, "nowpayments", "")
	respNow, err := svc.CreateCheckoutSession(ctx, 1005, "vip_30d", "USDT")
	if err != nil {
		t.Fatalf("expected nowpayments checkout session, got: %v", err)
	}
	if !strings.HasPrefix(respNow.PaymentURL, "https://nowpayments.io/payment/?iid=") {
		t.Errorf("expected nowpayments payment URL, got: %s", respNow.PaymentURL)
	}

	// 4. Gateway: unknown default
	_ = settingsSvc.Set(ctx, settings.KeySubscriptionGateway, "custom_gw", "")
	respDefault, err := svc.CreateCheckoutSession(ctx, 1006, "vip_30d", "USDT")
	if err != nil {
		t.Fatalf("expected custom gateway session, got: %v", err)
	}
	if !strings.HasPrefix(respDefault.PaymentURL, "https://checkout.azpays.net/pay/") {
		t.Errorf("expected default checkout URL, got: %s", respDefault.PaymentURL)
	}

	// 5. Subscription Disabled (!subEnabled) -> mock sandbox checkout URL
	_ = settingsSvc.Set(ctx, settings.KeySubscriptionEnabled, "false", "")
	_ = settingsSvc.Set(ctx, settings.KeySubscriptionGateway, "azpays", "")
	_ = settingsSvc.Set(ctx, settings.KeySubscriptionCallbackURL, "https://callback.domain.com", "")
	respDisabled, err := svc.CreateCheckoutSession(ctx, 1007, "vip_30d", "USDT")
	if err != nil {
		t.Fatalf("expected mock sandbox session when sub disabled, got: %v", err)
	}
	if !strings.Contains(respDisabled.PaymentURL, "mock-checkout") {
		t.Errorf("expected mock-checkout url, got: %s", respDisabled.PaymentURL)
	}
}

func TestVerifyWebhook_AllBranches(t *testing.T) {
	repo, settingsSvc, cleanup := setupTestEnv(t)
	defer cleanup()

	ctx := context.Background()

	// Initialize offline bot to test notification dispatch branch
	bot, err := telebot.NewBot(telebot.Settings{Offline: true})
	if err != nil {
		t.Fatalf("failed to create offline bot: %v", err)
	}

	svc := payment.NewService(repo, settingsSvc, bot)

	_ = settingsSvc.Set(ctx, settings.KeySubscriptionGateway, "azpays", "")
	_ = settingsSvc.Set(ctx, settings.KeySubscriptionSecretKey, "test_secret_key_123", "")

	// Pre-create subscriptions in database:
	// Normal 30-day VIP sub
	subNormal := &storage.Subscription{
		UserID:          501,
		PlanID:          "vip_30d",
		Tier:            "vip",
		Status:          storage.SubStatusPending,
		AZPaysInvoiceID: "inv_vip_normal_1",
		AmountCrypto:    9.99,
		Currency:        "USDT",
		ExpiresAt:       time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:       time.Now(),
	}
	if err := repo.CreateSubscription(ctx, subNormal); err != nil {
		t.Fatalf("failed to create normal sub: %v", err)
	}

	// Lifetime VIP sub
	subLifetime := &storage.Subscription{
		UserID:          502,
		PlanID:          "vip_lifetime",
		Tier:            "lifetime",
		Status:          storage.SubStatusPending,
		AZPaysInvoiceID: "inv_vip_lifetime_2",
		AmountCrypto:    99.00,
		Currency:        "USDT",
		ExpiresAt:       time.Now().Add(3650 * 24 * time.Hour),
		CreatedAt:       time.Now(),
	}
	if err := repo.CreateSubscription(ctx, subLifetime); err != nil {
		t.Fatalf("failed to create lifetime sub: %v", err)
	}

	// 1. Invalid signature test
	badBody := []byte(`{"event":"payment.success"}`)
	err = svc.VerifyWebhook(ctx, badBody, "t=123,v1=invalidsig", payment.WebhookPayload{})
	if err == nil || !strings.Contains(err.Error(), "invalid webhook signature") {
		t.Fatalf("expected invalid webhook signature error, got: %v", err)
	}

	// 2. Valid signature + Valid ParseWebhookEvent with paymentData
	eventPayload := map[string]interface{}{
		"event":     "payment.success",
		"timestamp": time.Now().Unix(),
		"data": map[string]interface{}{
			"id":          "inv_vip_normal_1",
			"token":       "tok_123",
			"status":      1,
			"paid_amount": 9.99,
		},
	}
	eventBytes, _ := json.Marshal(eventPayload)
	sigHeader := computeSignatureHeader(eventBytes, "test_secret_key_123")

	err = svc.VerifyWebhook(ctx, eventBytes, sigHeader, payment.WebhookPayload{})
	if err != nil {
		t.Fatalf("expected successful webhook processing with valid signature and event data, got: %v", err)
	}

	// Verify updated status in DB
	activeSub, err := repo.GetActiveSubscription(ctx, 501)
	if err != nil || activeSub == nil {
		t.Fatalf("expected active sub for user 501, got: %v", err)
	}

	// 3. Lifetime subscription renewal + bot notification
	lifetimePayload := payment.WebhookPayload{
		InvoiceID: "inv_vip_lifetime_2",
		Status:    "confirmed",
		Amount:    99.00,
		Currency:  "USDT",
	}
	err = svc.VerifyWebhook(ctx, nil, "", lifetimePayload)
	if err != nil {
		t.Fatalf("expected successful lifetime webhook, got: %v", err)
	}
	time.Sleep(50 * time.Millisecond) // allow bot goroutine to complete

	activeLifeSub, err := repo.GetActiveSubscription(ctx, 502)
	if err != nil || activeLifeSub == nil {
		t.Fatalf("expected active lifetime sub, got: %v", err)
	}

	// 4. Non-confirmed / unhandled status (e.g. "failed", "expired") returns nil without activating
	unhandledSub := &storage.Subscription{
		UserID:          503,
		PlanID:          "vip_30d",
		Tier:            "vip",
		Status:          storage.SubStatusPending,
		AZPaysInvoiceID: "inv_unhandled_status",
		AmountCrypto:    9.99,
		Currency:        "USDT",
		ExpiresAt:       time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:       time.Now(),
	}
	_ = repo.CreateSubscription(ctx, unhandledSub)

	err = svc.VerifyWebhook(ctx, nil, "", payment.WebhookPayload{
		InvoiceID: "inv_unhandled_status",
		Status:    "failed",
	})
	if err != nil {
		t.Fatalf("expected nil error on unhandled status, got: %v", err)
	}

	// 5. Subscription not found in repo
	err = svc.VerifyWebhook(ctx, nil, "", payment.WebhookPayload{
		InvoiceID: "non_existent_inv",
		Status:    "paid",
	})
	if err == nil || !strings.Contains(err.Error(), "subscription not found") {
		t.Fatalf("expected subscription not found error, got: %v", err)
	}

	// 6. UpdateSubscriptionStatus repo failure
	mockRepoErr := &mockStorageRepo{
		Repository:   repo,
		updateSubErr: errors.New("simulated update sub status error"),
	}
	svcWithRepoErr := payment.NewService(mockRepoErr, settingsSvc, nil)
	err = svcWithRepoErr.VerifyWebhook(ctx, nil, "", payment.WebhookPayload{
		InvoiceID: "inv_vip_normal_1",
		Status:    "paid",
	})
	if err == nil || !strings.Contains(err.Error(), "simulated update sub status error") {
		t.Fatalf("expected simulated update sub status error, got: %v", err)
	}

	// 7. ParseWebhookEvent invalid JSON branch (handled gracefully)
	invalidJSONBody := []byte(`{not valid json}`)
	_ = svc.VerifyWebhook(ctx, invalidJSONBody, "", payment.WebhookPayload{
		InvoiceID: "inv_vip_normal_1",
		Status:    "paid",
	})

	// 8. ParseWebhookEvent with empty ID in data (does not override payload)
	emptyIDEvent := []byte(`{"event":"test","timestamp":123,"data":{"id":"","token":""}}`)
	_ = svc.VerifyWebhook(ctx, emptyIDEvent, "", payment.WebhookPayload{
		InvoiceID: "inv_vip_normal_1",
		Status:    "paid",
	})
}
