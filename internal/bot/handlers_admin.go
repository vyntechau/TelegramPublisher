package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/vyntechau/TelegramPublisher/internal/i18n"
	"github.com/vyntechau/TelegramPublisher/internal/services/marketing"
	"github.com/vyntechau/TelegramPublisher/internal/services/settings"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"gopkg.in/telebot.v3"
)

// HandleStats returns executive analytics overview.
func (e *Engine) HandleStats(c telebot.Context) error {
	ctx := context.Background()
	userLang := e.GetUserLang(c)
	summary, err := e.Repo.GetAnalyticsSummary(ctx)
	if err != nil {
		return c.Send(fmt.Sprintf("Error fetching analytics: %v", err))
	}

	report := fmt.Sprintf("%s\n\n"+
		"%s\n"+
		"%s `%d`\n"+
		"%s `%d`\n"+
		"%s `%d`\n"+
		"%s `%d`\n"+
		"%s `%d`\n"+
		"%s `%d`\n\n"+
		"%s\n"+
		"%s `%d`\n"+
		"%s `%d`\n"+
		"%s\n\n"+
		"%s\n"+
		"%s\n\n"+
		"%s\n"+
		"%s `%d`\n"+
		"%s `$%.2f USDT`\n\n"+
		"%s",
		i18n.T(userLang, "stats_title"),
		i18n.T(userLang, "stats_sec_users"),
		i18n.T(userLang, "stats_lbl_total_users"), summary.TotalUsers,
		i18n.T(userLang, "stats_lbl_dau"), summary.ActiveUsersDaily,
		i18n.T(userLang, "stats_lbl_wau"), summary.ActiveUsersWeekly,
		i18n.T(userLang, "stats_lbl_mau"), summary.ActiveUsersMonthly,
		i18n.T(userLang, "stats_lbl_blocked"), summary.BlockedUsers,
		i18n.T(userLang, "stats_lbl_banned"), summary.BannedUsers,
		i18n.T(userLang, "stats_sec_content"),
		i18n.T(userLang, "stats_lbl_total_posts"), summary.TotalPosts,
		i18n.T(userLang, "stats_lbl_total_views"), summary.TotalViews,
		i18n.T(userLang, "stats_lbl_likes_dislikes", summary.TotalLikes, summary.TotalDislikes),
		i18n.T(userLang, "stats_sec_reports"),
		i18n.T(userLang, "stats_lbl_reports_pending_resolved", summary.PendingReports, summary.ResolvedReports),
		i18n.T(userLang, "stats_sec_finance"),
		i18n.T(userLang, "stats_lbl_active_subscribers"), summary.ActiveSubscribers,
		i18n.T(userLang, "stats_lbl_total_revenue"), summary.TotalRevenueCrypto,
		i18n.T(userLang, "stats_footer"))

	return c.Send(report, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// HandleUsers lists registered users.
func (e *Engine) HandleUsers(c telebot.Context) error {
	ctx := context.Background()
	userLang := e.GetUserLang(c)
	users, total, err := e.Repo.ListUsers(ctx, storage.UserFilter{Limit: 15})
	if err != nil {
		return c.Send(fmt.Sprintf("Error listing users: %v", err))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s\n\n", i18n.T(userLang, "admin_users_title", len(users), total)))
	for _, u := range users {
		sb.WriteString(fmt.Sprintf("• `%d` | @%s | Role: `%s` | Status: `%s`\n", u.TelegramID, u.Username, u.Role, u.Status))
	}
	sb.WriteString("\nCommands: `/ban <id>`, `/unban <id>`, `/promote <id>`, `/demote <id>`")

	return c.Send(sb.String(), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// HandleBan bans a user from accessing the bot.
func (e *Engine) HandleBan(c telebot.Context) error {
	payload := strings.TrimSpace(c.Data())
	if payload == "" {
		return c.Send("Usage: `/ban <telegram_id>`", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	tgID, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return c.Send("Invalid telegram ID.")
	}

	ctx := context.Background()
	if err := e.Repo.UpdateUserStatus(ctx, tgID, storage.StatusBanned); err != nil {
		return c.Send(fmt.Sprintf("Error banning user: %v", err))
	}

	return c.Send(fmt.Sprintf("✅ User `%d` has been *banned*.", tgID), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// HandleUnban unbans a user.
func (e *Engine) HandleUnban(c telebot.Context) error {
	payload := strings.TrimSpace(c.Data())
	if payload == "" {
		return c.Send("Usage: `/unban <telegram_id>`", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	tgID, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return c.Send("Invalid telegram ID.")
	}

	ctx := context.Background()
	if err := e.Repo.UpdateUserStatus(ctx, tgID, storage.StatusActive); err != nil {
		return c.Send(fmt.Sprintf("Error unbanning user: %v", err))
	}

	return c.Send(fmt.Sprintf("✅ User `%d` has been *unbanned*.", tgID), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// HandlePromote promotes a user to Author or Admin.
func (e *Engine) HandlePromote(c telebot.Context) error {
	args := c.Args()
	if len(args) < 1 {
		return c.Send("Usage: `/promote <telegram_id> [author|admin]`", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	tgID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return c.Send("Invalid telegram ID.")
	}

	role := storage.RoleAuthor
	if len(args) > 1 && strings.ToLower(args[1]) == "admin" {
		role = storage.RoleAdmin
	}

	ctx := context.Background()
	if err := e.Repo.UpdateUserRole(ctx, tgID, role); err != nil {
		return c.Send(fmt.Sprintf("Error updating role: %v", err))
	}

	return c.Send(fmt.Sprintf("✅ User `%d` is now promoted to *%s*.", tgID, role), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// HandleDemote demotes a user back to regular user.
func (e *Engine) HandleDemote(c telebot.Context) error {
	payload := strings.TrimSpace(c.Data())
	if payload == "" {
		return c.Send("Usage: `/demote <telegram_id>`", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	tgID, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return c.Send("Invalid telegram ID.")
	}

	ctx := context.Background()
	if err := e.Repo.UpdateUserRole(ctx, tgID, storage.RoleUser); err != nil {
		return c.Send(fmt.Sprintf("Error demoting user: %v", err))
	}

	return c.Send(fmt.Sprintf("✅ User `%d` has been demoted to *user*.", tgID), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// HandleAddChannel adds a required Force-Sub channel.
func (e *Engine) HandleAddChannel(c telebot.Context) error {
	args := c.Args()
	if len(args) < 2 {
		return c.Send("Usage: `/addchannel <channel_telegram_id> <invite_link> [title]`", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	chID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return c.Send("Invalid channel ID.")
	}
	link := args[1]
	title := "Official Channel"
	if len(args) > 2 {
		title = strings.Join(args[2:], " ")
	}

	ctx := context.Background()
	channel := &storage.Channel{
		TelegramID: chID,
		Title:      title,
		InviteLink: link,
		IsRequired: true,
	}

	if err := e.Repo.AddChannel(ctx, channel); err != nil {
		return c.Send(fmt.Sprintf("Error adding channel: %v", err))
	}

	return c.Send(fmt.Sprintf("✅ Channel *%s* (`%d`) registered for Force-Subscription.", title, chID), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// HandleDelChannel removes a channel.
func (e *Engine) HandleDelChannel(c telebot.Context) error {
	payload := strings.TrimSpace(c.Data())
	if payload == "" {
		return c.Send("Usage: `/delchannel <channel_id_or_tg_id>`", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	id, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return c.Send("Invalid ID.")
	}

	ctx := context.Background()
	_ = e.Repo.RemoveChannel(ctx, id)
	return c.Send("✅ Channel removed.", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// HandleListChannels lists registered force-sub channels.
func (e *Engine) HandleListChannels(c telebot.Context) error {
	ctx := context.Background()
	userLang := e.GetUserLang(c)
	channels, err := e.Repo.ListChannels(ctx, false)
	if err != nil {
		return c.Send(fmt.Sprintf("Error listing channels: %v", err))
	}

	if len(channels) == 0 {
		return c.Send(i18n.T(userLang, "admin_channels_none"), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	var sb strings.Builder
	sb.WriteString(i18n.T(userLang, "admin_channels_title"))
	for _, ch := range channels {
		sb.WriteString(fmt.Sprintf("• *%s* (`%d`)\n  Link: %s | Required: `%t`\n", ch.Title, ch.TelegramID, ch.InviteLink, ch.IsRequired))
	}

	return c.Send(sb.String(), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// HandleListReports lists broken file reports.
func (e *Engine) HandleListReports(c telebot.Context) error {
	ctx := context.Background()
	userLang := e.GetUserLang(c)
	reports, total, err := e.Repo.ListReports(ctx, storage.ReportStatusPending, 0, 10, 0)
	if err != nil {
		return c.Send(fmt.Sprintf("Error fetching reports: %v", err))
	}

	if total == 0 {
		return c.Send(i18n.T(userLang, "admin_reports_none"), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(i18n.T(userLang, "admin_reports_title"), total))

	for _, rep := range reports {
		sb.WriteString(fmt.Sprintf("• *Ticket #%d* | Post: `%s` | Reason: `%s`\n  FileID: `%s`\n",
			rep.ID, rep.Post.Slug, rep.Reason, rep.Post.FileID))
	}
	sb.WriteString(i18n.T(userLang, "admin_reports_footer"))

	return c.Send(sb.String(), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// HandleBroadcastCommand sends a broadcast to all active users.
func (e *Engine) HandleBroadcastCommand(c telebot.Context) error {
	userLang := e.GetUserLang(c)
	content := strings.TrimSpace(c.Data())
	if content == "" {
		return c.Send("Usage: `/broadcast <message_content>`\nSupports standard Markdown.", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	progressMsg, _ := e.Bot.Send(c.Chat(), i18n.T(userLang, "admin_broadcast_progress"), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})

	ctx := context.Background()
	res, err := e.MarketSvc.Dispatch(ctx, marketing.BroadcastPayload{
		Format:  "markdown",
		Content: content,
		Filter: storage.UserFilter{
			Status: storage.StatusActive,
		},
	})
	if err != nil {
		return c.Send(fmt.Sprintf("Broadcast error: %v", err))
	}

	summary := fmt.Sprintf(i18n.T(userLang, "admin_broadcast_completed"),
		res.TotalTargeted, res.TotalSent, res.TotalBlocked, res.TotalFailed, res.DurationMs)

	if progressMsg != nil {
		_ = e.Bot.Delete(progressMsg)
	}
	return c.Send(summary, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// HandleSetTTL updates default auto-delete time.
func (e *Engine) HandleSetTTL(c telebot.Context) error {
	payload := strings.TrimSpace(c.Data())
	if payload == "" {
		return c.Send("Usage: `/setttl <seconds>` (e.g. `/setttl 120` for 2 minutes)", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	seconds, err := strconv.Atoi(payload)
	if err != nil || seconds < 0 {
		return c.Send("Invalid duration in seconds.")
	}

	ctx := context.Background()
	_ = e.SettingsSvc.Set(ctx, settings.KeyAutoDeleteSeconds, strconv.Itoa(seconds), "Default copyright auto-delete duration in seconds")

	return c.Send(fmt.Sprintf("⏱️ Default copyright auto-delete TTL set to *%d seconds*.", seconds), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// HandleBotSettingsOverview returns dynamic system runtime settings overview.
func (e *Engine) HandleBotSettingsOverview(c telebot.Context) error {
	ctx := context.Background()
	userLang := e.GetUserLang(c)
	kbMode := e.SettingsSvc.GetString(ctx, settings.KeyKeyboardMode, "both")
	autoDeleteSec := e.SettingsSvc.GetInt(ctx, settings.KeyAutoDeleteSeconds, 120)
	forceSubEnabled := e.SettingsSvc.GetBool(ctx, settings.KeyForceSubEnabled, true)
	miniAppEnabled := e.SettingsSvc.GetBool(ctx, settings.KeyMiniAppEnabled, true)
	miniAppURL := e.SettingsSvc.GetString(ctx, settings.KeyMiniAppURL, "http://localhost:3000")
	autoPostEnabled := e.SettingsSvc.GetBool(ctx, settings.KeyAutoPostEnabled, false)
	subEnabled := e.SettingsSvc.GetBool(ctx, settings.KeySubscriptionEnabled, true)
	gateway := e.SettingsSvc.GetString(ctx, settings.KeySubscriptionGateway, "azpays")

	overview := fmt.Sprintf("⚙️ *Bot Runtime Settings & Config*\n\n"+
		"• *Keyboard Layout Mode*: `%s`\n"+
		"• *Auto-Delete TTL*: `%d seconds`\n"+
		"• *Force Channel Sub Gate*: `%t`\n"+
		"• *Mini App Web Player*: `%t` (`%s`)\n"+
		"• *Auto-Posting Channels*: `%t`\n"+
		"• *Crypto VIP Gateways*: `%t` (`%s`)\n\n"+
		"_To modify keys, update settings from the Web Dashboard._",
		kbMode, autoDeleteSec, forceSubEnabled, miniAppEnabled, miniAppURL, autoPostEnabled, subEnabled, gateway)

	menu := &telebot.ReplyMarkup{}
	if user, err := e.Repo.GetUserByTelegramID(ctx, c.Sender().ID); err == nil && e.AuthSvc != nil {
		if token, err := e.AuthSvc.GenerateJWT(user); err == nil {
			cleanURL := strings.TrimRight(miniAppURL, "/")
			dashboardURL := fmt.Sprintf("%s/admin?token=%s", cleanURL, token)
			btnOpen := menu.URL(i18n.T(userLang, "btn_open_dashboard"), dashboardURL)
			menu.Inline(menu.Row(btnOpen))
		}
	}

	return c.Send(overview, menu, telebot.ModeMarkdown)
}

// HandleWebLogin generates an authenticated one-click link and token for the Web Admin Dashboard.
func (e *Engine) HandleWebLogin(c telebot.Context) error {
	ctx := context.Background()
	userLang := e.GetUserLang(c)
	user, err := e.Repo.GetUserByTelegramID(ctx, c.Sender().ID)
	if err != nil {
		return c.Send("❌ User profile not found. Please send /start first.")
	}

	if e.AuthSvc == nil {
		return c.Send("❌ Authentication service unavailable.")
	}

	token, err := e.AuthSvc.GenerateJWT(user)
	if err != nil {
		return c.Send(fmt.Sprintf("❌ Failed to generate session token: %v", err))
	}

	webURL := e.SettingsSvc.GetString(ctx, settings.KeyMiniAppURL, "http://localhost:3000")
	cleanURL := strings.TrimRight(webURL, "/")
	dashboardURL := fmt.Sprintf("%s/admin?token=%s", cleanURL, token)

	menu := &telebot.ReplyMarkup{}
	btnOpen := menu.URL(i18n.T(userLang, "admin_web_login_btn"), dashboardURL)
	menu.Inline(menu.Row(btnOpen))

	msg := fmt.Sprintf(i18n.T(userLang, "admin_web_login_title"), token)

	return c.Send(msg, menu, telebot.ModeMarkdown)
}


