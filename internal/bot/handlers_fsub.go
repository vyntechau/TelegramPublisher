package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"gopkg.in/telebot.v3"
)

// checkForceSub verifies if user is a member of all required channels.
func (e *Engine) checkForceSub(userID int64) (bool, []*storage.Channel, error) {
	ctx := context.Background()
	channels, err := e.Repo.ListChannels(ctx, true)
	if err != nil || len(channels) == 0 {
		return true, nil, nil
	}

	var unjoined []*storage.Channel
	for _, ch := range channels {
		member, err := e.Bot.ChatMemberOf(&telebot.Chat{ID: ch.TelegramID}, &telebot.User{ID: userID})
		if err != nil || member == nil || (member.Role != telebot.Member && member.Role != telebot.Administrator && member.Role != telebot.Creator) {
			unjoined = append(unjoined, ch)
		}
	}

	return len(unjoined) == 0, unjoined, nil
}

func (e *Engine) sendForceSubGate(c telebot.Context, slug string, unjoined []*storage.Channel) error {
	markup := &telebot.ReplyMarkup{}
	var rows []telebot.Row

	for i, ch := range unjoined {
		btn := markup.URL(fmt.Sprintf("📢 Join Channel %d: %s", i+1, ch.Title), ch.InviteLink)
		rows = append(rows, markup.Row(btn))
	}

	btnTryAgain := markup.Data("🔄 I have joined! Unlock content", "fsub_check", slug)
	rows = append(rows, markup.Row(btnTryAgain))

	markup.Inline(rows...)

	msg := "🔒 *Channel Subscription Required*\n\n" +
		"To unlock and view this media, you must join our official channel(s) first.\n" +
		"Click the buttons below to join, then click *'I have joined'*."

	return c.Send(msg, markup, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

func (e *Engine) handleFSubCheckCallback(c telebot.Context, parts []string) error {
	slug := ""
	if len(parts) > 1 {
		slug = strings.TrimSpace(parts[1])
	}

	joinedAll, unjoined, err := e.checkForceSub(c.Sender().ID)
	if err == nil && joinedAll {
		_ = e.Bot.Delete(c.Callback().Message)
		_ = c.Respond(&telebot.CallbackResponse{Text: "✅ Subscription verified! Unlocking..."})

		// Deliver the post
		if slug != "" {
			return e.DeliverPost(c, c.Sender(), slug)
		}
		return e.sendWelcomeMenu(c)
	}

	markup := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	for i, ch := range unjoined {
		btn := markup.URL(fmt.Sprintf("📢 Join Channel %d: %s", i+1, ch.Title), ch.InviteLink)
		rows = append(rows, markup.Row(btn))
	}
	btnTryAgain := markup.Data("🔄 Try Again", "fsub_check", slug)
	rows = append(rows, markup.Row(btnTryAgain))
	markup.Inline(rows...)

	_, _ = e.Bot.EditReplyMarkup(c.Callback().Message, markup)
	return c.Respond(&telebot.CallbackResponse{Text: "❌ You have not joined all required channels yet.", ShowAlert: true})
}
