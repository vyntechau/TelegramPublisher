package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/vyntechau/TelegramPublisher/internal/i18n"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"gopkg.in/telebot.v3"
)

// HandleCallbackQuery routes inline keyboard callback queries.
func (e *Engine) HandleCallbackQuery(c telebot.Context) error {
	data := strings.TrimSpace(c.Callback().Data)
	parts := strings.Split(data, "|")
	action := parts[0]

	switch {
	case strings.HasPrefix(action, "setlang_"):
		langCode := strings.TrimPrefix(action, "setlang_")
		e.userLangs.Store(c.Sender().ID, langCode)
		langMeta := i18n.GetLanguageMeta(langCode)
		c.Respond(&telebot.CallbackResponse{Text: fmt.Sprintf("Language: %s", langMeta.NativeName)})
		return c.Send(fmt.Sprintf(i18n.T(langCode, "lang_changed"), langMeta.NativeName), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	case action == "cmd_language":
		return e.HandleLanguage(c)
	case action == "react_like" || action == "react_dislike":
		return e.handleReactionCallback(c, action, parts)
	case action == "report_broken":
		return e.handleReportBrokenCallback(c, parts)
	case action == "submit_report":
		return e.handleSubmitReportReason(c, parts)
	case action == "resolve_report":
		return e.handleResolveReportCallback(c, parts)
	case action == "fsub_check":
		return e.handleFSubCheckCallback(c, parts)
	case action == "cmd_subscribe":
		return e.HandleSubscribe(c)
	case action == "cmd_help":
		return e.HandleHelp(c)
	case action == "buy_plan":
		return e.handleBuyPlanCallback(c, parts)
	default:
		return c.Respond(&telebot.CallbackResponse{Text: "Unknown action"})
	}
}

func (e *Engine) handleReactionCallback(c telebot.Context, action string, parts []string) error {
	if len(parts) < 2 {
		return c.Respond(&telebot.CallbackResponse{Text: "Invalid post reference"})
	}
	postID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Invalid post ID"})
	}

	reaction := storage.ReactionLike
	if action == "react_dislike" {
		reaction = storage.ReactionDislike
	}

	ctx := context.Background()
	sender := c.Sender()

	if err := e.Repo.SetReaction(ctx, postID, sender.ID, reaction); err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Error recording reaction"})
	}

	_ = e.Repo.RecordEvent(ctx, &storage.AnalyticsEvent{
		EventType: "reaction",
		UserID:    sender.ID,
		PostID:    &postID,
	})

	// Fetch updated counts
	post, err := e.Repo.GetPostByID(ctx, postID)
	if err == nil && post != nil {
		inlineMarkup := &telebot.ReplyMarkup{}
		btnLike := inlineMarkup.Data(fmt.Sprintf("👍 %d", post.LikesCount), "react_like", fmt.Sprintf("%d", post.ID))
		btnDislike := inlineMarkup.Data(fmt.Sprintf("👎 %d", post.DislikesCount), "react_dislike", fmt.Sprintf("%d", post.ID))
		btnReport := inlineMarkup.Data("🚩 Report Broken", "report_broken", fmt.Sprintf("%d", post.ID))

		inlineMarkup.Inline(
			inlineMarkup.Row(btnLike, btnDislike),
			inlineMarkup.Row(btnReport),
		)
		_, _ = e.Bot.EditReplyMarkup(c.Callback().Message, inlineMarkup)
	}

	return c.Respond(&telebot.CallbackResponse{Text: "Reaction recorded!"})
}

func (e *Engine) handleReportBrokenCallback(c telebot.Context, parts []string) error {
	if len(parts) < 2 {
		return c.Respond(&telebot.CallbackResponse{Text: "Invalid post reference"})
	}
	postID := parts[1]

	markup := &telebot.ReplyMarkup{}
	btn1 := markup.Data("📵 Video Won't Play", "submit_report", postID, "broken_stream")
	btn2 := markup.Data("📁 Corrupted File", "submit_report", postID, "corrupted_file")
	btn3 := markup.Data("⚠️ Incorrect / Wrong Media", "submit_report", postID, "wrong_content")
	btnCancel := markup.Data("✖ Cancel", "cancel_action")

	markup.Inline(
		markup.Row(btn1),
		markup.Row(btn2),
		markup.Row(btn3),
		markup.Row(btnCancel),
	)

	return c.Send("🚩 *Report Broken Content*\nPlease select the issue you encountered with this file:", markup, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

func (e *Engine) handleSubmitReportReason(c telebot.Context, parts []string) error {
	if len(parts) < 3 {
		return c.Respond(&telebot.CallbackResponse{Text: "Invalid report parameters"})
	}
	postID, _ := strconv.ParseInt(parts[1], 10, 64)
	reason := parts[2]

	ctx := context.Background()
	sender := c.Sender()

	report := &storage.Report{
		PostID:     postID,
		ReportedBy: sender.ID,
		Reason:     reason,
		Details:    "Reported via inline bot button",
		Status:     storage.ReportStatusPending,
	}

	if err := e.Repo.CreateReport(ctx, report); err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Failed to submit report."})
	}

	// Record Analytics
	_ = e.Repo.RecordEvent(ctx, &storage.AnalyticsEvent{
		EventType: "report_submitted",
		UserID:    sender.ID,
		PostID:    &postID,
	})

	_ = e.Bot.Delete(c.Callback().Message)
	return c.Respond(&telebot.CallbackResponse{Text: "✅ Report submitted! Admins & author notified.", ShowAlert: true})
}

func (e *Engine) handleResolveReportCallback(c telebot.Context, parts []string) error {
	if len(parts) < 2 {
		return c.Respond(&telebot.CallbackResponse{Text: "Invalid report ID"})
	}
	reportID, _ := strconv.ParseInt(parts[1], 10, 64)
	ctx := context.Background()

	_ = e.Repo.UpdateReportStatus(ctx, reportID, storage.ReportStatusResolved, c.Sender().ID, "Resolved by admin")
	return c.Respond(&telebot.CallbackResponse{Text: "✅ Report marked as resolved!"})
}
