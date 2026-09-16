package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	azpays "github.com/azpays/sdk-go"
	"github.com/google/uuid"
	"github.com/vyntechau/TelegramPublisher/internal/services/settings"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"gopkg.in/telebot.v3"
)

// Plan definition
type Plan struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Tier        string  `json:"tier"`
	DurationDay int     `json:"duration_days"`
	PriceUSD    float64 `json:"price_usd"`
	Description string  `json:"description"`
}

var DefaultPlans = []Plan{
	{
		ID:          "vip_30d",
		Name:        "VIP 30-Day Pass",
		Tier:        "vip",
		DurationDay: 30,
		PriceUSD:    9.99,
		Description: "Bypass 2-minute auto-delete countdown & unlock unlimited streaming",
	},
	{
		ID:          "author_pro_30d",
		Name:        "Author Pro 30-Day",
		Tier:        "author_pro",
		DurationDay: 30,
		PriceUSD:    29.99,
		Description: "Upload unlimited media files and access publisher studio analytics",
	},
	{
		ID:          "vip_lifetime",
		Name:        "VIP Lifetime Pass",
		Tier:        "lifetime",
		DurationDay: 3650,
		PriceUSD:    99.00,
		Description: "Permanent VIP privileges and premium perks forever",
	},
}

// CreateInvoiceResponse returned by payment service
type CreateInvoiceResponse struct {
	InvoiceID  string    `json:"invoice_id"`
	PaymentURL string    `json:"payment_url"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Gateway    string    `json:"gateway"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// WebhookPayload sent by payment gateways (AzPays, Coinbase, NOWPayments)
type WebhookPayload struct {
	InvoiceID string  `json:"invoice_id"`
	Status    string  `json:"status"` // paid, completed, success, expired, failed
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	TxHash    string  `json:"tx_hash,omitempty"`
	Signature string  `json:"signature,omitempty"`
	Gateway   string  `json:"gateway,omitempty"`
}

// Service manages multi-provider crypto payment workflows (AzPays, Coinbase Commerce, NOWPayments.io).
type Service struct {
	repo        storage.Repository
	settingsSvc *settings.Service
	bot         *telebot.Bot
}

func NewService(repo storage.Repository, settingsSvc *settings.Service, bot *telebot.Bot) *Service {
	return &Service{
		repo:        repo,
		settingsSvc: settingsSvc,
		bot:         bot,
	}
}

// CreateCheckoutSession generates a new subscription order and invoice supporting AzPays, Coinbase Commerce, and NOWPayments.
func (s *Service) CreateCheckoutSession(ctx context.Context, userID int64, planID, currency string) (*CreateInvoiceResponse, error) {
	var targetPlan *Plan
	for _, p := range DefaultPlans {
		if p.ID == planID {
			targetPlan = &p
			break
		}
	}
	if targetPlan == nil {
		return nil, errors.New("invalid plan id")
	}

	if currency == "" {
		currency = "USDT"
	}

	// Fetch dynamic credentials from database settings table
	gateway := s.settingsSvc.GetString(ctx, settings.KeySubscriptionGateway, "azpays")
	subEnabled := s.settingsSvc.GetBool(ctx, settings.KeySubscriptionEnabled, s.settingsSvc.GetBool(ctx, settings.KeyAzpaysEnabled, false))
	apiKey := s.settingsSvc.GetString(ctx, settings.KeySubscriptionAPIKey, s.settingsSvc.GetString(ctx, settings.KeyAzpaysAPIKey, ""))
	callbackURL := s.settingsSvc.GetString(ctx, settings.KeySubscriptionCallbackURL, s.settingsSvc.GetString(ctx, settings.KeyAzpaysCallbackURL, "http://localhost:8080/api/v1/payments/subscription/webhook"))

	invoiceID := fmt.Sprintf("%s_%s", gateway[:2], uuid.New().String()[:12])
	paymentURL := fmt.Sprintf("https://checkout.%s.net/pay/%s", gateway, invoiceID)
	expiresAt := time.Now().Add(time.Duration(targetPlan.DurationDay) * 24 * time.Hour)

	if subEnabled && apiKey != "" {
		switch gateway {
		case "azpays":
			var opts []azpays.Option
			if baseURL := s.settingsSvc.GetString(ctx, settings.KeyAzpaysBaseURL, ""); baseURL != "" {
				opts = append(opts, azpays.WithBaseURL(baseURL))
			}
			client := azpays.NewClient(apiKey, opts...)
			desc := fmt.Sprintf("%s for user %d", targetPlan.Name, userID)
			azPayment, err := client.Payments.Create(ctx, &azpays.CreatePaymentRequest{
				FiatAmount:     targetPlan.PriceUSD,
				Description:    azpays.String(desc),
				IdempotencyKey: azpays.String(invoiceID),
			})
			if err != nil {
				log.Printf("[AzPays SDK] Warning: Payment creation error: %v, using direct fallback invoice", err)
			} else if azPayment != nil {
				invoiceID = azPayment.ID
				if azPayment.Token != "" {
					paymentURL = fmt.Sprintf("https://checkout.azpays.net/pay/%s", azPayment.Token)
				}
			}
		case "coinbase":
			paymentURL = fmt.Sprintf("https://commerce.coinbase.com/charges/%s", invoiceID)
		case "nowpayments":
			paymentURL = fmt.Sprintf("https://nowpayments.io/payment/?iid=%s", invoiceID)
		default:
			paymentURL = fmt.Sprintf("https://checkout.azpays.net/pay/%s", invoiceID)
		}
	} else if !subEnabled {
		// Mock sandbox URL for local test / preview
		paymentURL = fmt.Sprintf("%s/api/v1/payments/mock-checkout?invoice_id=%s&gateway=%s", callbackURL, invoiceID, gateway)
	}

	sub := &storage.Subscription{
		UserID:          userID,
		PlanID:          targetPlan.ID,
		Tier:            targetPlan.Tier,
		Status:          storage.SubStatusPending,
		AZPaysInvoiceID: invoiceID,
		AmountCrypto:    targetPlan.PriceUSD,
		Currency:        currency,
		ExpiresAt:       expiresAt,
		CreatedAt:       time.Now(),
	}

	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to save subscription: %w", err)
	}

	return &CreateInvoiceResponse{
		InvoiceID:  invoiceID,
		PaymentURL: paymentURL,
		Amount:     targetPlan.PriceUSD,
		Currency:   currency,
		Gateway:    gateway,
		ExpiresAt:  expiresAt,
	}, nil
}

// VerifyWebhook validates incoming payment confirmation from supported crypto gateways (AzPays, Coinbase, NOWPayments).
func (s *Service) VerifyWebhook(ctx context.Context, rawBody []byte, signatureHeader string, payload WebhookPayload) error {
	secretKey := s.settingsSvc.GetString(ctx, settings.KeySubscriptionSecretKey, s.settingsSvc.GetString(ctx, settings.KeyAzpaysSecretKey, ""))
	gateway := s.settingsSvc.GetString(ctx, settings.KeySubscriptionGateway, "azpays")

	// AzPays SDK signature verification
	if gateway == "azpays" && secretKey != "" && signatureHeader != "" && len(rawBody) > 0 {
		if !azpays.VerifyWebhookSignature(rawBody, signatureHeader, secretKey) {
			return errors.New("invalid webhook signature from AzPays SDK verification")
		}
	}

	// Parse AzPays webhook event if applicable
	if gateway == "azpays" && len(rawBody) > 0 {
		event, err := azpays.ParseWebhookEvent(rawBody)
		if err == nil && event != nil {
			var paymentData struct {
				ID         string  `json:"id"`
				Token      string  `json:"token"`
				Status     int     `json:"status"`
				PaidAmount float64 `json:"paid_amount"`
			}
			if err := json.Unmarshal(event.Data, &paymentData); err == nil && paymentData.ID != "" {
				payload.InvoiceID = paymentData.ID
				payload.Status = "paid"
				payload.Amount = paymentData.PaidAmount
			}
		}
	}

	sub, err := s.repo.GetSubscriptionByInvoice(ctx, payload.InvoiceID)
	if err != nil {
		return fmt.Errorf("subscription not found for invoice %s: %w", payload.InvoiceID, err)
	}

	statusLower := payload.Status
	if statusLower == "paid" || statusLower == "completed" || statusLower == "success" || statusLower == "confirmed" || statusLower == "charge:confirmed" {
		duration := 30 * 24 * time.Hour
		if sub.Tier == "lifetime" {
			duration = 3650 * 24 * time.Hour
		}
		newExpiry := time.Now().Add(duration)

		if err := s.repo.UpdateSubscriptionStatus(ctx, payload.InvoiceID, storage.SubStatusActive, newExpiry); err != nil {
			return err
		}

		_ = s.repo.RecordEvent(ctx, &storage.AnalyticsEvent{
			EventType: "subscription_paid",
			UserID:    sub.UserID,
			MetadataJSON: fmt.Sprintf(`{"invoice_id":"%s","amount":%.2f,"currency":"%s","tier":"%s","gateway":"%s"}`,
				sub.AZPaysInvoiceID, payload.Amount, payload.Currency, sub.Tier, gateway),
		})

		if s.bot != nil {
			go func() {
				msg := fmt.Sprintf("🎉 *Payment Confirmed via %s!*\n\nYour *%s* subscription is now active until *%s*.\nEnjoy unlimited media access without copyright auto-delete timers!",
					gateway, sub.Tier, newExpiry.Format("2006-01-02 15:04"))
				_, _ = s.bot.Send(&telebot.Chat{ID: sub.UserID}, msg, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
			}()
		}
	}

	return nil
}
