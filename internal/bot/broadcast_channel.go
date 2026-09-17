package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/vyntechau/TelegramPublisher/internal/i18n"
	"github.com/vyntechau/TelegramPublisher/internal/services/settings"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"gopkg.in/telebot.v3"
)

// BroadcastPostToChannels automatically broadcasts a newly created post to all configured target channels.
func (e *Engine) BroadcastPostToChannels(ctx context.Context, post *storage.Post) {
	if e.Bot == nil {
		return
	}

	autoPostEnabled := e.SettingsSvc.GetBool(ctx, settings.KeyAutoPostEnabled, false)
	if !autoPostEnabled {
		return
	}

	rawChannels := e.SettingsSvc.GetString(ctx, settings.KeyAutoPostChannels, "")
	if strings.TrimSpace(rawChannels) == "" {
		return
	}

	postFormat := e.SettingsSvc.GetString(ctx, settings.KeyAutoPostFormat, "teaser_with_button")
	miniAppURL := e.SettingsSvc.GetString(ctx, settings.KeyMiniAppURL, "http://localhost:8080")
	miniAppEnabled := e.SettingsSvc.GetBool(ctx, settings.KeyMiniAppEnabled, true)

	botUsername := "bot"
	if e.Bot.Me != nil && e.Bot.Me.Username != "" {
		botUsername = e.Bot.Me.Username
	}
	watchURL := fmt.Sprintf("https://t.me/%s?start=%s", botUsername, post.Slug)

	defaultLang := e.SettingsSvc.GetString(ctx, settings.KeyDefaultLanguage, "en")

	// Build Inline Keyboard for Channel Post
	inlineMarkup := &telebot.ReplyMarkup{}
	btnWatch := inlineMarkup.URL(i18n.T(defaultLang, "btn_watch_media"), watchURL)
	var rows []telebot.Row
	rows = append(rows, inlineMarkup.Row(btnWatch))

	if miniAppEnabled && miniAppURL != "" {
		btnMiniApp := inlineMarkup.WebApp(i18n.T(defaultLang, "btn_mini_app"), &telebot.WebApp{URL: miniAppURL})
		rows = append(rows, inlineMarkup.Row(btnMiniApp))
	}
	inlineMarkup.Inline(rows...)

	// Split channels by comma, newline, or semicolon
	channels := parseChannelList(rawChannels)

	go func() {
		for _, chStr := range channels {
			var recipient telebot.Recipient
			if strings.HasPrefix(chStr, "-100") || strings.HasPrefix(chStr, "-") {
				if id, err := strconv.ParseInt(chStr, 10, 64); err == nil {
					recipient = &telebot.Chat{ID: id}
				}
			} else if id, err := strconv.ParseInt(chStr, 10, 64); err == nil {
				recipient = &telebot.Chat{ID: id}
			} else {
				username := strings.TrimPrefix(chStr, "@")
				recipient = &telebot.Chat{Username: username}
			}

			if recipient == nil {
				continue
			}

			var sendErr error
			if postFormat == "full_media" {
				file := telebot.File{FileID: post.FileID}
				caption := post.Caption
				if caption == "" {
					caption = "🎬 New media post published!"
				}
				switch post.FileType {
				case storage.FileTypePhoto:
					_, sendErr = e.Bot.Send(recipient, &telebot.Photo{File: file, Caption: caption}, inlineMarkup)
				case storage.FileTypeDocument:
					_, sendErr = e.Bot.Send(recipient, &telebot.Document{File: file, Caption: caption}, inlineMarkup)
				case storage.FileTypeAnimation:
					_, sendErr = e.Bot.Send(recipient, &telebot.Animation{File: file, Caption: caption}, inlineMarkup)
				default:
					_, sendErr = e.Bot.Send(recipient, &telebot.Video{File: file, Caption: caption}, inlineMarkup)
				}
			} else {
				teaserText := fmt.Sprintf("🎬 <b>New Media Published!</b>\n\n"+
					"📝 <b>Caption</b>: %s\n"+
					"⏳ <b>Copyright Protection</b>: Media auto-purges in 2 minutes after delivery.\n\n"+
					"👇 <i>Click below to watch the full media in our bot:</i>",
					escapeHTML(post.Caption))
				_, sendErr = e.Bot.Send(recipient, teaserText, inlineMarkup, &telebot.SendOptions{ParseMode: telebot.ModeHTML})
			}

			if sendErr != nil {
				log.Printf("[AutoPost] Failed to send post %s to channel %s: %v", post.Slug, chStr, sendErr)
			} else {
				log.Printf("[AutoPost] Successfully sent post %s to channel %s", post.Slug, chStr)
			}
		}
	}()
}

func parseChannelList(raw string) []string {
	var list []string
	clean := strings.ReplaceAll(raw, "\n", ",")
	clean = strings.ReplaceAll(clean, ";", ",")
	for _, part := range strings.Split(clean, ",") {
		ch := strings.TrimSpace(part)
		if ch != "" {
			list = append(list, ch)
		}
	}
	return list
}

func escapeHTML(text string) string {
	if text == "" {
		return "<i>(No caption)</i>"
	}
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(text)
}
