package bot

import (
	"context"
	"fmt"

	"github.com/vyntechau/TelegramPublisher/internal/i18n"
	"github.com/vyntechau/TelegramPublisher/internal/services/payment"
	"gopkg.in/telebot.v3"
)

// HandleSubscribe presents available VIP subscription tiers.
func (e *Engine) HandleSubscribe(c telebot.Context) error {
	userLang := e.GetUserLang(c)
	markup := &telebot.ReplyMarkup{}
	var rows []telebot.Row

	for _, p := range payment.DefaultPlans {
		btnText := fmt.Sprintf("💎 %s ($%.2f)", p.Name, p.PriceUSD)
		btn := markup.Data(btnText, "buy_plan", p.ID)
		rows = append(rows, markup.Row(btn))
	}

	markup.Inline(rows...)

	text := fmt.Sprintf("%s\n\n%s\n\n%s",
		i18n.T(userLang, "sub_title"),
		i18n.T(userLang, "sub_benefits"),
		i18n.T(userLang, "sub_select_plan"),
	)

	return c.Send(text, markup, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

func (e *Engine) handleBuyPlanCallback(c telebot.Context, parts []string) error {
	userLang := e.GetUserLang(c)
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
	btnPay := markup.URL(i18n.T(userLang, "btn_pay_crypto", invoice.Amount), invoice.PaymentURL)
	markup.Inline(markup.Row(btnPay))

	msg := fmt.Sprintf("%s\n\n"+
		"• *%s*: `%s`\n"+
		"• *%s*: `$%.2f USDT`\n"+
		"• *%s*: `%s`\n\n"+
		"%s",
		i18n.T(userLang, "invoice_title"),
		i18n.T(userLang, "invoice_id_label"), invoice.InvoiceID,
		i18n.T(userLang, "invoice_amount_label"), invoice.Amount,
		i18n.T(userLang, "invoice_expires_label"), invoice.ExpiresAt.Format("2006-01-02 15:04"),
		i18n.T(userLang, "invoice_desc"))

	return c.Send(msg, markup, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}
