package bot

import (
	"context"
	"fmt"

	"github.com/vyntechau/TelegramPublisher/internal/services/payment"
	"gopkg.in/telebot.v3"
)

// HandleSubscribe presents available VIP subscription tiers.
func (e *Engine) HandleSubscribe(c telebot.Context) error {
	markup := &telebot.ReplyMarkup{}
	var rows []telebot.Row

	for _, p := range payment.DefaultPlans {
		btnText := fmt.Sprintf("💎 %s ($%.2f)", p.Name, p.PriceUSD)
		btn := markup.Data(btnText, "buy_plan", p.ID)
		rows = append(rows, markup.Row(btn))
	}

	markup.Inline(rows...)

	text := "💎 *VIP Subscription Passes (AZPays Crypto Checkout)*\n\n" +
		"Upgrade your experience to enjoy premium benefits:\n" +
		"• ⚡ *No 2-Minute Auto-Delete*: Keep all requested files in your chat permanently\n" +
		"• 🚀 *High-Speed Streaming & Direct Downloads*\n" +
		"• 👑 *Author Pro Privileges*: Upload & publish your own content\n\n" +
		"Select a subscription tier below to proceed with instant crypto checkout:"

	return c.Send(text, markup, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

func (e *Engine) handleBuyPlanCallback(c telebot.Context, parts []string) error {
	if len(parts) < 2 {
		return c.Respond(&telebot.CallbackResponse{Text: "Invalid plan"})
	}
	planID := parts[1]

	ctx := context.Background()
	invoice, err := e.PaySvc.CreateCheckoutSession(ctx, c.Sender().ID, planID, "USDT")
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: fmt.Sprintf("Error creating checkout: %v", err), ShowAlert: true})
	}

	markup := &telebot.ReplyMarkup{}
	btnPay := markup.URL(fmt.Sprintf("💳 Pay $%.2f (USDT / Crypto)", invoice.Amount), invoice.PaymentURL)
	markup.Inline(markup.Row(btnPay))

	msg := fmt.Sprintf("🧾 *Invoice Generated*\n\n"+
		"• *Invoice ID*: `%s`\n"+
		"• *Amount*: `$%.2f USDT`\n"+
		"• *Expires*: `%s`\n\n"+
		"Click the button below to complete payment on AZPays. Your VIP pass will be activated automatically once the transaction is confirmed.",
		invoice.InvoiceID, invoice.Amount, invoice.ExpiresAt.Format("2006-01-02 15:04"))

	return c.Send(msg, markup, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}
