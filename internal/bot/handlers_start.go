package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vyntechau/TelegramPublisher/internal/i18n"
	"github.com/vyntechau/TelegramPublisher/internal/services/settings"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"gopkg.in/telebot.v3"
)

// HandleStart processes /start and /start <slug> deep-link media requests.
func (e *Engine) HandleStart(c telebot.Context) error {
	ctx := context.Background()
	sender := c.Sender()
	payload := strings.TrimSpace(c.Data())

	// If no payload, send welcome menu & Mini App link
	if payload == "" {
		return e.sendWelcomeMenu(c)
	}

	// 1. Force-Sub Check
	forceSubEnabled := e.SettingsSvc.GetBool(ctx, settings.KeyForceSubEnabled, true)
	if forceSubEnabled {
		joinedAll, unjoinedChannels, err := e.checkForceSub(sender.ID)
		if err == nil && !joinedAll {
			return e.sendForceSubGate(c, payload, unjoinedChannels)
		}
	}

	return e.DeliverPost(c, sender, payload)
}

// DeliverPost delivers media to the chat with reaction buttons and 2-minute auto-delete countdown.
func (e *Engine) DeliverPost(c telebot.Context, sender *telebot.User, slug string) error {
	ctx := context.Background()

	// 2. Fetch Post by Slug
	post, err := e.Repo.GetPostBySlug(ctx, slug)
	if err != nil {
		return c.Send("❌ *Post Not Found*: The requested media link is invalid or has expired.", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	// 3. Track View
	_ = e.Repo.IncrementPostViews(ctx, post.ID, sender.ID, "")
	_ = e.Repo.RecordEvent(ctx, &storage.AnalyticsEvent{
		EventType: "view",
		UserID:    sender.ID,
		PostID:    &post.ID,
	})

	// 4. Check if user has active VIP pass (VIP users bypass auto-delete countdown)
	sub, _ := e.Repo.GetActiveSubscription(ctx, sender.ID)
	isVIP := sub != nil && (sub.Tier == "vip" || sub.Tier == "lifetime")

	// 5. Build Dynamic Post Buttons (Inline Keyboard & Persistent Layout)
	kbMode := strings.ToLower(e.SettingsSvc.GetString(ctx, settings.KeyKeyboardMode, "both"))
	showReactions := e.SettingsSvc.GetBool(ctx, settings.KeyShowReactions, true)
	showReport := e.SettingsSvc.GetBool(ctx, settings.KeyShowReportButton, true)
	showForward := e.SettingsSvc.GetBool(ctx, settings.KeyShowForwardButton, true)
	showMiniApp := e.SettingsSvc.GetBool(ctx, settings.KeyShowMiniAppButton, true)
	miniAppURL := e.SettingsSvc.GetString(ctx, settings.KeyMiniAppURL, "http://localhost:8080")
	miniAppEnabled := e.SettingsSvc.GetBool(ctx, settings.KeyMiniAppEnabled, true)

	var inlineMarkup *telebot.ReplyMarkup
	if kbMode == "inline" || kbMode == "both" || kbMode == "" {
		inlineMarkup = &telebot.ReplyMarkup{}
		var inlineRows []telebot.Row

		// Row 1: Reactions (Like / Dislike)
		if showReactions {
			btnLike := inlineMarkup.Data(fmt.Sprintf("👍 %d", post.LikesCount), "react_like", fmt.Sprintf("%d", post.ID))
			btnDislike := inlineMarkup.Data(fmt.Sprintf("👎 %d", post.DislikesCount), "react_dislike", fmt.Sprintf("%d", post.ID))
			inlineRows = append(inlineRows, inlineMarkup.Row(btnLike, btnDislike))
		}

		// Row 2: Action Buttons (Report Broken & Forward / Share)
		var actionBtns []telebot.Btn
		if showReport {
			btnReport := inlineMarkup.Data("🚨 Report Broken", "report_broken", fmt.Sprintf("%d", post.ID))
			actionBtns = append(actionBtns, btnReport)
		}
		if showForward {
			botUsername := "bot"
			if e.Bot.Me != nil && e.Bot.Me.Username != "" {
				botUsername = e.Bot.Me.Username
			}
			shareURL := fmt.Sprintf("https://t.me/share/url?url=https://t.me/%s?start=%s&text=Check+out+this+media+on+TelegramPublisher!", botUsername, post.Slug)
			btnShare := inlineMarkup.URL("↗️ Forward / Share", shareURL)
			actionBtns = append(actionBtns, btnShare)
		}
		if len(actionBtns) > 0 {
			inlineRows = append(inlineRows, inlineMarkup.Row(actionBtns...))
		}

		// Row 3: Open in Mini App
		if showMiniApp && miniAppEnabled && miniAppURL != "" {
			btnMiniApp := inlineMarkup.WebApp("🚀 Open in Mini App", &telebot.WebApp{URL: miniAppURL})
			inlineRows = append(inlineRows, inlineMarkup.Row(btnMiniApp))
		}

		if len(inlineRows) > 0 {
			inlineMarkup.Inline(inlineRows...)
		} else {
			inlineMarkup = nil
		}
	}

	// 6. Deliver Media
	file := telebot.File{FileID: post.FileID}
	caption := post.Caption

	var sentMsg *telebot.Message
	var sendErr error

	var sendOpts []interface{}
	if inlineMarkup != nil {
		sendOpts = append(sendOpts, inlineMarkup)
	}

	switch post.FileType {
	case storage.FileTypeVideo:
		video := &telebot.Video{File: file, Caption: caption}
		sentMsg, sendErr = e.Bot.Send(c.Chat(), video, sendOpts...)
	case storage.FileTypePhoto:
		photo := &telebot.Photo{File: file, Caption: caption}
		sentMsg, sendErr = e.Bot.Send(c.Chat(), photo, sendOpts...)
	case storage.FileTypeDocument:
		doc := &telebot.Document{File: file, Caption: caption}
		sentMsg, sendErr = e.Bot.Send(c.Chat(), doc, sendOpts...)
	case storage.FileTypeAnimation:
		anim := &telebot.Animation{File: file, Caption: caption}
		sentMsg, sendErr = e.Bot.Send(c.Chat(), anim, sendOpts...)
	default:
		video := &telebot.Video{File: file, Caption: caption}
		sentMsg, sendErr = e.Bot.Send(c.Chat(), video, sendOpts...)
	}

	if sendErr != nil {
		return c.Send(fmt.Sprintf("⚠️ *Error delivering media*: %v", sendErr), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	// If persistent keyboard mode is active, send or refresh persistent bottom keyboard
	if kbMode == "persistent" || kbMode == "both" {
		_ = e.sendPersistentKeyboard(c)
	}

	// 7. Auto-Delete Timer (2-Minute TTL for copyright protection)
	ttl := post.AutoDeleteSeconds
	if ttl <= 0 {
		ttl = e.SettingsSvc.GetInt(ctx, settings.KeyAutoDeleteSeconds, 120)
	}

	if !isVIP && ttl > 0 {
		warningText := e.SettingsSvc.GetString(ctx, settings.KeyCopyrightWarningText, "⏳ *Copyright Protection*: This content will be automatically deleted in 2 minutes.")
		timerMsg, err := e.Bot.Send(c.Chat(), fmt.Sprintf("%s\n⏱️ Auto-delete in *%d seconds*.", warningText, ttl), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
		if err == nil && timerMsg != nil {
			e.Cleaner.Schedule(c.Chat().ID, []int{sentMsg.ID, timerMsg.ID}, time.Duration(ttl)*time.Second)
		} else {
			e.Cleaner.ScheduleSingle(c.Chat().ID, sentMsg.ID, time.Duration(ttl)*time.Second)
		}
	}

	return nil
}

func (e *Engine) sendPersistentKeyboard(c telebot.Context) error {
	ctx := context.Background()
	miniAppURL := e.SettingsSvc.GetString(ctx, settings.KeyMiniAppURL, "http://localhost:8080")
	miniAppEnabled := e.SettingsSvc.GetBool(ctx, settings.KeyMiniAppEnabled, true)

	replyMarkup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
	}

	var row1 []telebot.Btn
	if miniAppEnabled && miniAppURL != "" {
		row1 = append(row1, replyMarkup.WebApp("🚀 Mini App", &telebot.WebApp{URL: miniAppURL}))
	}
	row1 = append(row1, replyMarkup.Text("💎 VIP Subscription"))

	row2 := []telebot.Btn{
		replyMarkup.Text("🚨 Report Broken"),
		replyMarkup.Text("↗️ Share Bot"),
	}

	row3 := []telebot.Btn{
		replyMarkup.Text("👤 My Profile"),
		replyMarkup.Text("ℹ️ Help"),
	}

	replyMarkup.Reply(
		replyMarkup.Row(row1...),
		replyMarkup.Row(row2...),
		replyMarkup.Row(row3...),
	)

	return nil
}

func (e *Engine) GetUserLang(c telebot.Context) string {
	if c == nil || c.Sender() == nil {
		return e.SettingsSvc.GetString(context.Background(), settings.KeyDefaultLanguage, "en")
	}

	// 1. Check in-memory session override
	if val, ok := e.userLangs.Load(c.Sender().ID); ok {
		if langStr, valid := val.(string); valid && langStr != "" {
			return langStr
		}
	}

	ctx := context.Background()
	defaultLang := e.SettingsSvc.GetString(ctx, settings.KeyDefaultLanguage, "en")
	supportedCSV := e.SettingsSvc.GetString(ctx, settings.KeySupportedLanguages, "en,fa,ar,ru,es,de,zh")
	supported := i18n.FilterSupported(supportedCSV)

	// 2. Check sender Telegram language code if supported
	if c.Sender().LanguageCode != "" {
		code := strings.ToLower(c.Sender().LanguageCode)
		if len(code) > 2 {
			code = code[:2]
		}
		for _, s := range supported {
			if s.Code == code {
				return code
			}
		}
	}

	return defaultLang
}

// HandleLanguage presents an interactive language switcher keyboard.
func (e *Engine) HandleLanguage(c telebot.Context) error {
	ctx := context.Background()
	userLang := e.GetUserLang(c)
	supportedCSV := e.SettingsSvc.GetString(ctx, settings.KeySupportedLanguages, "en,fa,ar,ru,es,de,zh")
	languages := i18n.FilterSupported(supportedCSV)

	markup := &telebot.ReplyMarkup{}
	var rows []telebot.Row

	var row []telebot.Btn
	for i, l := range languages {
		label := fmt.Sprintf("%s %s", l.Flag, l.NativeName)
		if l.Code == userLang {
			label = fmt.Sprintf("✅ %s", label)
		}
		btn := markup.Data(label, fmt.Sprintf("setlang_%s", l.Code))
		row = append(row, btn)
		if len(row) == 2 || i == len(languages)-1 {
			rows = append(rows, markup.Row(row...))
			row = []telebot.Btn{}
		}
	}

	markup.Inline(rows...)
	prompt := i18n.T(userLang, "lang_select_prompt")
	return c.Send(prompt, markup, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

func (e *Engine) sendWelcomeMenu(c telebot.Context) error {
	ctx := context.Background()
	userLang := e.GetUserLang(c)
	kbMode := strings.ToLower(e.SettingsSvc.GetString(ctx, settings.KeyKeyboardMode, "both"))
	miniAppURL := e.SettingsSvc.GetString(ctx, settings.KeyMiniAppURL, "http://localhost:8080")
	miniAppEnabled := e.SettingsSvc.GetBool(ctx, settings.KeyMiniAppEnabled, true)

	markup := &telebot.ReplyMarkup{}
	var rows []telebot.Row

	if miniAppEnabled && miniAppURL != "" {
		btnMiniApp := markup.WebApp(i18n.T(userLang, "btn_mini_app"), &telebot.WebApp{URL: miniAppURL})
		rows = append(rows, markup.Row(btnMiniApp))
	}

	btnSub := markup.Data(i18n.T(userLang, "btn_vip_sub"), "cmd_subscribe")
	btnHelp := markup.Data(i18n.T(userLang, "btn_help"), "cmd_help")
	rows = append(rows, markup.Row(btnSub, btnHelp))

	btnLang := markup.Data(i18n.T(userLang, "btn_language"), "cmd_language")
	botUsername := "bot"
	if e.Bot.Me != nil && e.Bot.Me.Username != "" {
		botUsername = e.Bot.Me.Username
	}
	shareURL := fmt.Sprintf("https://t.me/share/url?url=https://t.me/%s&text=Check+out+this+bot+for+fast+media+delivery!", botUsername)
	btnShare := markup.URL(i18n.T(userLang, "btn_share"), shareURL)
	rows = append(rows, markup.Row(btnLang, btnShare))

	markup.Inline(rows...)

	welcomeText := fmt.Sprintf("%s\n\n%s\n%s",
		i18n.T(userLang, "welcome_title"),
		i18n.T(userLang, "welcome_fast_delivery"),
		i18n.T(userLang, "welcome_desc"),
	)

	if kbMode == "persistent" || kbMode == "both" {
		replyMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
		var r1 []telebot.Btn
		if miniAppEnabled && miniAppURL != "" {
			r1 = append(r1, replyMarkup.WebApp("🚀 Mini App", &telebot.WebApp{URL: miniAppURL}))
		}
		r1 = append(r1, replyMarkup.Text("💎 VIP Subscription"))
		replyMarkup.Reply(
			replyMarkup.Row(r1...),
			replyMarkup.Row(replyMarkup.Text("🚨 Report Broken"), replyMarkup.Text("🌐 Language")),
			replyMarkup.Row(replyMarkup.Text("👤 My Profile"), replyMarkup.Text("ℹ️ Help")),
		)
		return c.Send(welcomeText, markup, replyMarkup, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	return c.Send(welcomeText, markup, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

func (e *Engine) HandleHelp(c telebot.Context) error {
	userLang := e.GetUserLang(c)
	helpText := fmt.Sprintf("%s\n\n%s%s",
		i18n.T(userLang, "help_title"),
		i18n.T(userLang, "help_user_cmds"),
		i18n.T(userLang, "help_admin_cmds"),
	)

	return c.Send(helpText, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

func (e *Engine) HandleMyStatus(c telebot.Context) error {
	ctx := context.Background()
	user, err := e.Repo.GetUserByTelegramID(ctx, c.Sender().ID)
	if err != nil {
		return c.Send("Could not retrieve user status.")
	}

	sub, _ := e.Repo.GetActiveSubscription(ctx, c.Sender().ID)
	vipStatus := "None (Free Tier)"
	if sub != nil {
		vipStatus = fmt.Sprintf("Active *%s* (Expires %s)", sub.Tier, sub.ExpiresAt.Format("2006-01-02 15:04"))
	}

	text := fmt.Sprintf("👤 *User Profile*\n\n"+
		"• *Telegram ID*: `%d`\n"+
		"• *Username*: @%s\n"+
		"• *Role*: `%s`\n"+
		"• *Status*: `%s`\n"+
		"• *VIP Subscription*: %s\n"+
		"• *Member Since*: %s",
		user.TelegramID, user.Username, user.Role, user.Status, vipStatus, user.CreatedAt.Format("2006-01-02"))

	return c.Send(text, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}
