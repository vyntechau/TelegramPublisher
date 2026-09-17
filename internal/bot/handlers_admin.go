package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/vyntechau/TelegramPublisher/internal/services/marketing"
	"github.com/vyntechau/TelegramPublisher/internal/services/settings"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"gopkg.in/telebot.v3"
)

// HandleStats returns executive analytics overview.
func (e *Engine) HandleStats(c telebot.Context) error {
	ctx := context.Background()
	summary, err := e.Repo.GetAnalyticsSummary(ctx)
	if err != nil {
		return c.Send(fmt.Sprintf("Error fetching analytics: %v", err))
	}

	report := fmt.Sprintf("📊 *Executive Analytics & Metrics Report*\n\n"+
		"👥 *User Growth & Retention:*\n"+
		"• Total Users: `%d`\n"+
		"• Daily Active (DAU): `%d`\n"+
		"• Weekly Active (WAU): `%d`\n"+
		"• Monthly Active (MAU): `%d`\n"+
		"• Blocked Bot: `%d`\n"+
		"• Banned: `%d`\n\n"+
		"🎬 *Content & Engagement:*\n"+
		"• Total Posts: `%d`\n"+
		"• Total Views: `%d`\n"+
		"• Likes: `👍 %d` | Dislikes: `👎 %d`\n\n"+
		"🚩 *Broken File Reports:*\n"+
		"• Pending: `%d` | Resolved: `%d`\n\n"+
		"💰 *Financial & VIP Subscriptions:*\n"+
		"• Active VIP Subscribers: `%d`\n"+
		"• Total Gross Revenue: `$%.2f USDT`\n\n"+
		"_For visual charts and CSV exports, visit the Web Dashboard._",
		summary.TotalUsers, summary.ActiveUsersDaily, summary.ActiveUsersWeekly, summary.ActiveUsersMonthly,
		summary.BlockedUsers, summary.BannedUsers, summary.TotalPosts, summary.TotalViews,
		summary.TotalLikes, summary.TotalDislikes, summary.PendingReports, summary.ResolvedReports,
		summary.ActiveSubscribers, summary.TotalRevenueCrypto)

	return c.Send(report, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// HandleUsers lists registered users.
func (e *Engine) HandleUsers(c telebot.Context) error {
	ctx := context.Background()
	users, total, err := e.Repo.ListUsers(ctx, storage.UserFilter{Limit: 15})
	if err != nil {
		return c.Send(fmt.Sprintf("Error listing users: %v", err))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("👥 *Users Directory* (Showing %d of %d)\n\n", len(users), total))
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
	channels, err := e.Repo.ListChannels(ctx, false)
	if err != nil {
		return c.Send(fmt.Sprintf("Error listing channels: %v", err))
	}

	if len(channels) == 0 {
		return c.Send("No force-subscription channels configured.\nUse `/addchannel` to register one.", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	var sb strings.Builder
	sb.WriteString("📢 *Configured Force-Subscription Channels:*\n\n")
	for _, ch := range channels {
		sb.WriteString(fmt.Sprintf("• *%s* (`%d`)\n  Link: %s | Required: `%t`\n", ch.Title, ch.TelegramID, ch.InviteLink, ch.IsRequired))
	}

	return c.Send(sb.String(), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// HandleListReports lists broken file reports.
func (e *Engine) HandleListReports(c telebot.Context) error {
	ctx := context.Background()
	reports, total, err := e.Repo.ListReports(ctx, storage.ReportStatusPending, 0, 10, 0)
	if err != nil {
		return c.Send(fmt.Sprintf("Error fetching reports: %v", err))
	}

	if total == 0 {
		return c.Send("✅ *No Pending Reports!* All media links are working properly.", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🚩 *Pending Broken File Tickets* (%d total)\n\n", total))

	for _, rep := range reports {
		sb.WriteString(fmt.Sprintf("• *Ticket #%d* | Post: `%s` | Reason: `%s`\n  FileID: `%s`\n",
			rep.ID, rep.Post.Slug, rep.Reason, rep.Post.FileID))
	}
	sb.WriteString("\n_Manage and resolve tickets directly from the Web Dashboard._")

	return c.Send(sb.String(), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// HandleBroadcastCommand sends a broadcast to all active users.
func (e *Engine) HandleBroadcastCommand(c telebot.Context) error {
	content := strings.TrimSpace(c.Data())
	if content == "" {
		return c.Send("Usage: `/broadcast <message_content>`\nSupports standard Markdown.", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	}

	progressMsg, _ := e.Bot.Send(c.Chat(), "🚀 *Broadcasting in progress...*", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})

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

	summary := fmt.Sprintf("✅ *Broadcast Completed!*\n\n"+
		"• *Targeted*: `%d`\n"+
		"• *Delivered*: `%d`\n"+
		"• *Blocked by User (Flagged in DB)*: `%d`\n"+
		"• *Failed*: `%d`\n"+
		"• *Duration*: `%d ms`",
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
	kbMode := e.SettingsSvc.GetString(ctx, settings.KeyKeyboardMode, "both")
	autoDeleteSec := e.SettingsSvc.GetInt(ctx, settings.KeyAutoDeleteSeconds, 120)
	forceSubEnabled := e.SettingsSvc.GetBool(ctx, settings.KeyForceSubEnabled, true)
	miniAppEnabled := e.SettingsSvc.GetBool(ctx, settings.KeyMiniAppEnabled, true)
	miniAppURL := e.SettingsSvc.GetString(ctx, settings.KeyMiniAppURL, "http://localhost:8080")
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

	return c.Send(overview, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

