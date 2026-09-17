package bot

import (
	"context"
	"fmt"
	"net/url"
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
		userLang := e.GetUserLang(c)
		return c.Send(i18n.T(userLang, "post_not_found"), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
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

	userLang := e.GetUserLang(c)

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
			btnReport := inlineMarkup.Data(i18n.T(userLang, "btn_report"), "report_broken", fmt.Sprintf("%d", post.ID))
			actionBtns = append(actionBtns, btnReport)
		}
		if showForward {
			botUsername := "bot"
			if e.Bot.Me != nil && e.Bot.Me.Username != "" {
				botUsername = e.Bot.Me.Username
			}
			shareText := url.QueryEscape(i18n.T(userLang, "share_caption"))
			shareURL := fmt.Sprintf("https://t.me/share/url?url=https://t.me/%s?start=%s&text=%s", botUsername, post.Slug, shareText)
			btnShare := inlineMarkup.URL(i18n.T(userLang, "btn_share"), shareURL)
			actionBtns = append(actionBtns, btnShare)
		}
		if len(actionBtns) > 0 {
			inlineRows = append(inlineRows, inlineMarkup.Row(actionBtns...))
		}

		// Row 3: Open in Mini App
		if showMiniApp && miniAppEnabled && miniAppURL != "" {
			btnMiniApp := inlineMarkup.WebApp(i18n.T(userLang, "btn_mini_app"), &telebot.WebApp{URL: miniAppURL})
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
		warningText := e.SettingsSvc.GetString(ctx, settings.KeyCopyrightWarningText, "")
		countdown := i18n.T(userLang, "timer_countdown", ttl)
		fullWarning := countdown
		if warningText != "" {
			fullWarning = fmt.Sprintf("%s\n%s", warningText, countdown)
		}
		timerMsg, err := e.Bot.Send(c.Chat(), fullWarning, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
		if err == nil && timerMsg != nil {
			e.Cleaner.Schedule(c.Chat().ID, []int{sentMsg.ID, timerMsg.ID}, time.Duration(ttl)*time.Second)
		} else {
			e.Cleaner.ScheduleSingle(c.Chat().ID, sentMsg.ID, time.Duration(ttl)*time.Second)
		}
	}

	return nil
}

// GetUserRole returns the role string for the current user.
func (e *Engine) GetUserRole(c telebot.Context) string {
	if c == nil || c.Sender() == nil {
		return storage.RoleUser
	}
	sender := c.Sender()
	if sender.ID == e.Config.Bot.OwnerID && e.Config.Bot.OwnerID != 0 {
		return storage.RoleOwner
	}
	ctx := context.Background()
	user, err := e.Repo.GetUserByTelegramID(ctx, sender.ID)
	if err == nil && user != nil && user.Role != "" {
		return user.Role
	}
	return storage.RoleUser
}

func (e *Engine) getRolePersistentKeyboard(c telebot.Context) *telebot.ReplyMarkup {
	ctx := context.Background()
	userLang := e.GetUserLang(c)
	role := e.GetUserRole(c)

	miniAppURL := e.SettingsSvc.GetString(ctx, settings.KeyMiniAppURL, "http://localhost:8080")
	miniAppEnabled := e.SettingsSvc.GetBool(ctx, settings.KeyMiniAppEnabled, true)

	replyMarkup := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
	}

	btnMiniAppText := i18n.T(userLang, "btn_mini_app")
	if len(btnMiniAppText) > 22 {
		btnMiniAppText = "🚀 Mini App"
	}

	switch role {
	case storage.RoleOwner, storage.RoleAdmin:
		var r1 []telebot.Btn
		if miniAppEnabled && miniAppURL != "" {
			r1 = append(r1, replyMarkup.WebApp(btnMiniAppText, &telebot.WebApp{URL: miniAppURL}))
		}
		r1 = append(r1, replyMarkup.Text(i18n.T(userLang, "btn_admin_analytics")))

		r2 := []telebot.Btn{
			replyMarkup.Text(i18n.T(userLang, "btn_admin_broadcast")),
			replyMarkup.Text(i18n.T(userLang, "btn_admin_reports")),
		}

		r3 := []telebot.Btn{
			replyMarkup.Text(i18n.T(userLang, "btn_admin_users")),
			replyMarkup.Text(i18n.T(userLang, "btn_admin_channels")),
		}

		r4 := []telebot.Btn{
			replyMarkup.Text(i18n.T(userLang, "btn_admin_upload")),
			replyMarkup.Text(i18n.T(userLang, "btn_admin_settings")),
		}

		r5 := []telebot.Btn{
			replyMarkup.Text(i18n.T(userLang, "btn_my_profile")),
			replyMarkup.Text(i18n.T(userLang, "btn_help")),
			replyMarkup.Text(i18n.T(userLang, "btn_language")),
		}

		replyMarkup.Reply(
			replyMarkup.Row(r1...),
			replyMarkup.Row(r2...),
			replyMarkup.Row(r3...),
			replyMarkup.Row(r4...),
			replyMarkup.Row(r5...),
		)

	case storage.RoleAuthor:
		var r1 []telebot.Btn
		if miniAppEnabled && miniAppURL != "" {
			r1 = append(r1, replyMarkup.WebApp(btnMiniAppText, &telebot.WebApp{URL: miniAppURL}))
		}
		r1 = append(r1, replyMarkup.Text(i18n.T(userLang, "btn_admin_upload")))

		r2 := []telebot.Btn{
			replyMarkup.Text(i18n.T(userLang, "btn_author_reports")),
			replyMarkup.Text(i18n.T(userLang, "btn_admin_analytics")),
		}

		r3 := []telebot.Btn{
			replyMarkup.Text(i18n.T(userLang, "btn_my_profile")),
			replyMarkup.Text(i18n.T(userLang, "btn_help")),
			replyMarkup.Text(i18n.T(userLang, "btn_language")),
		}

		replyMarkup.Reply(
			replyMarkup.Row(r1...),
			replyMarkup.Row(r2...),
			replyMarkup.Row(r3...),
		)

	default: // Regular User
		var r1 []telebot.Btn
		if miniAppEnabled && miniAppURL != "" {
			r1 = append(r1, replyMarkup.WebApp(btnMiniAppText, &telebot.WebApp{URL: miniAppURL}))
		}
		r1 = append(r1, replyMarkup.Text(i18n.T(userLang, "btn_vip_sub")))

		r2 := []telebot.Btn{
			replyMarkup.Text(i18n.T(userLang, "btn_report")),
			replyMarkup.Text(i18n.T(userLang, "btn_share")),
		}

		r3 := []telebot.Btn{
			replyMarkup.Text(i18n.T(userLang, "btn_my_profile")),
			replyMarkup.Text(i18n.T(userLang, "btn_help")),
			replyMarkup.Text(i18n.T(userLang, "btn_language")),
		}

		replyMarkup.Reply(
			replyMarkup.Row(r1...),
			replyMarkup.Row(r2...),
			replyMarkup.Row(r3...),
		)
	}

	return replyMarkup
}

func (e *Engine) sendPersistentKeyboard(c telebot.Context) error {
	replyMarkup := e.getRolePersistentKeyboard(c)
	userLang := e.GetUserLang(c)
	prompt := fmt.Sprintf("⌨️ %s", i18n.T(userLang, "welcome_fast_delivery"))
	return c.Send(prompt, replyMarkup, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

func (e *Engine) GetUserLang(c telebot.Context) string {
	ctx := context.Background()
	defaultLang := e.SettingsSvc.GetString(ctx, settings.KeyDefaultLanguage, "en")
	supportedCSV := e.SettingsSvc.GetString(ctx, settings.KeySupportedLanguages, "en,fa,ar,ru,es,de,zh")
	supported := i18n.FilterSupported(supportedCSV)

	// Ensure defaultLang is active among supported languages
	isDefaultSupported := false
	for _, s := range supported {
		if s.Code == defaultLang {
			isDefaultSupported = true
			break
		}
	}
	if !isDefaultSupported && len(supported) > 0 {
		defaultLang = supported[0].Code
	}

	if c == nil || c.Sender() == nil {
		return defaultLang
	}

	// 1. Check user's manual in-memory session override (via /lang or /language)
	if val, ok := e.userLangs.Load(c.Sender().ID); ok {
		if langStr, valid := val.(string); valid && langStr != "" {
			for _, s := range supported {
				if s.Code == langStr {
					return langStr
				}
			}
		}
	}

	// 2. Default to platform system default language
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
		replyMarkup := e.getRolePersistentKeyboard(c)
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
	userLang := e.GetUserLang(c)
	user, err := e.Repo.GetUserByTelegramID(ctx, c.Sender().ID)
	if err != nil {
		return c.Send(i18n.T(userLang, "err_user_status"))
	}

	sub, _ := e.Repo.GetActiveSubscription(ctx, c.Sender().ID)
	vipStatus := i18n.T(userLang, "profile_free_tier")
	if sub != nil {
		vipStatus = i18n.T(userLang, "profile_active_vip", sub.Tier, sub.ExpiresAt.Format("2006-01-02 15:04"))
	}

	text := fmt.Sprintf("%s\n\n"+
		"• *%s*: `%d`\n"+
		"• *%s*: @%s\n"+
		"• *%s*: `%s`\n"+
		"• *%s*: `%s`\n"+
		"• *%s*: %s\n"+
		"• *%s*: %s",
		i18n.T(userLang, "profile_title"),
		i18n.T(userLang, "profile_tg_id"), user.TelegramID,
		i18n.T(userLang, "profile_username"), user.Username,
		i18n.T(userLang, "profile_role"), user.Role,
		i18n.T(userLang, "profile_status"), user.Status,
		i18n.T(userLang, "profile_vip_sub"), vipStatus,
		i18n.T(userLang, "profile_member_since"), user.CreatedAt.Format("2006-01-02"))

	return c.Send(text, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}
