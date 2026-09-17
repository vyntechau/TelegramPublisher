package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/vyntechau/TelegramPublisher/internal/i18n"
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
	userLang := e.GetUserLang(c)
	markup := &telebot.ReplyMarkup{}
	var rows []telebot.Row

	for i, ch := range unjoined {
		btn := markup.URL(fmt.Sprintf("%s %d: %s", i18n.T(userLang, "btn_join_channel"), i+1, ch.Title), ch.InviteLink)
		rows = append(rows, markup.Row(btn))
	}

	btnTryAgain := markup.Data(i18n.T(userLang, "btn_check_membership"), "fsub_check", slug)
	rows = append(rows, markup.Row(btnTryAgain))

	markup.Inline(rows...)

	msg := fmt.Sprintf("%s\n\n%s", i18n.T(userLang, "fsub_required"), i18n.T(userLang, "fsub_desc"))

	return c.Send(msg, markup, telebot.ModeMarkdown)
}

func (e *Engine) handleFSubCheckCallback(c telebot.Context, parts []string) error {
	userLang := e.GetUserLang(c)
	slug := ""
	if len(parts) > 1 {
		slug = strings.TrimSpace(parts[1])
	}

	joinedAll, unjoined, err := e.checkForceSub(c.Sender().ID)
	if err == nil && joinedAll {
		_ = e.Bot.Delete(c.Callback().Message)
		_ = c.Respond(&telebot.CallbackResponse{Text: i18n.T(userLang, "fsub_verified")})

		// Deliver the post
		if slug != "" {
			return e.DeliverPost(c, c.Sender(), slug)
		}
		return e.sendWelcomeMenu(c)
	}

	markup := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	for i, ch := range unjoined {
		btn := markup.URL(fmt.Sprintf("%s %d: %s", i18n.T(userLang, "btn_join_channel"), i+1, ch.Title), ch.InviteLink)
		rows = append(rows, markup.Row(btn))
	}
	btnTryAgain := markup.Data(i18n.T(userLang, "btn_check_membership"), "fsub_check", slug)
	rows = append(rows, markup.Row(btnTryAgain))
	markup.Inline(rows...)

	_, _ = e.Bot.EditReplyMarkup(c.Callback().Message, markup)
	return c.Respond(&telebot.CallbackResponse{Text: i18n.T(userLang, "err_not_joined"), ShowAlert: true})
}
