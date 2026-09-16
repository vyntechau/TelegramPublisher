package marketing

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"gopkg.in/telebot.v3"
)

// BroadcastPayload represents the broadcast request parameters.
type BroadcastPayload struct {
	Format       string             `json:"format"` // text, html, markdown, raw_json
	Content      string             `json:"content"`
	RawJSON      string             `json:"raw_json,omitempty"`
	Filter       storage.UserFilter `json:"filter"`
	EnablePin    bool               `json:"enable_pin"`
	DisableWebPreview bool          `json:"disable_web_preview"`
}

// BroadcastResult contains execution metrics from a broadcast run.
type BroadcastResult struct {
	TotalTargeted int64     `json:"total_targeted"`
	TotalSent     int64     `json:"total_sent"`
	TotalBlocked  int64     `json:"total_blocked"`
	TotalFailed   int64     `json:"total_failed"`
	StartedAt     time.Time `json:"started_at"`
	CompletedAt   time.Time `json:"completed_at"`
	DurationMs    int64     `json:"duration_ms"`
}

// BotSender defines the interface for delivering Telegram messages.
type BotSender interface {
	Send(to telebot.Recipient, what interface{}, opts ...interface{}) (*telebot.Message, error)
}

// Service handles marketing campaigns and safe message dispatching.
type Service struct {
	bot  BotSender
	repo storage.Repository
}

func NewService(bot BotSender, repo storage.Repository) *Service {
	return &Service{
		bot:  bot,
		repo: repo,
	}
}

// RawJSONMessage represents custom Telegram payload with inline buttons.
type RawJSONMessage struct {
	Text         string           `json:"text"`
	Photo        string           `json:"photo,omitempty"`
	Video        string           `json:"video,omitempty"`
	ParseMode    string           `json:"parse_mode,omitempty"`
	InlineMarkup [][]RawButton    `json:"inline_keyboard,omitempty"`
}

type RawButton struct {
	Text   string `json:"text"`
	URL    string `json:"url,omitempty"`
	Data   string `json:"callback_data,omitempty"`
	WebApp string `json:"web_app_url,omitempty"`
}

// Dispatch executes the broadcast campaign across targeted users with rate limiting and block detection.
func (s *Service) Dispatch(ctx context.Context, payload BroadcastPayload) (*BroadcastResult, error) {
	result := &BroadcastResult{
		StartedAt: time.Now(),
	}

	// Fetch users matching filter
	// Default only active status if not specified
	if payload.Filter.Status == "" {
		payload.Filter.Status = storage.StatusActive
	}
	payload.Filter.Limit = 100000 // broad sweep

	users, total, err := s.repo.ListUsers(ctx, payload.Filter)
	if err != nil {
		return nil, err
	}
	result.TotalTargeted = total

	if len(users) == 0 {
		result.CompletedAt = time.Now()
		result.DurationMs = result.CompletedAt.Sub(result.StartedAt).Milliseconds()
		return result, nil
	}

	// Prepare message options
	sendOpts := &telebot.SendOptions{
		DisableWebPagePreview: payload.DisableWebPreview,
	}

	switch strings.ToLower(payload.Format) {
	case "html":
		sendOpts.ParseMode = telebot.ModeHTML
	case "markdown", "markdownv2":
		sendOpts.ParseMode = telebot.ModeMarkdownV2
	}

	// Rate limiter: 25 messages per second token bucket
	limiter := time.NewTicker(40 * time.Millisecond)
	defer limiter.Stop()

	var mu sync.Mutex

broadcastLoop:
	for _, user := range users {
		select {
		case <-ctx.Done():
			break broadcastLoop
		case <-limiter.C:
			target := &telebot.Chat{ID: user.TelegramID}

			var sendErr error

			if payload.Format == "raw_json" && payload.RawJSON != "" {
				sendErr = s.sendRawJSON(target, payload.RawJSON)
			} else {
				_, sendErr = s.bot.Send(target, payload.Content, sendOpts)
			}

			mu.Lock()
			if sendErr != nil {
				errStr := sendErr.Error()
				if strings.Contains(errStr, "blocked by the user") || strings.Contains(errStr, "user is deactivated") || strings.Contains(errStr, "chat not found") {
					result.TotalBlocked++
					// Automatically flag user as blocked_by_user in database
					go func(tgID int64) {
						_ = s.repo.UpdateUserStatus(context.Background(), tgID, storage.StatusBlockedByUser)
						_ = s.repo.RecordEvent(context.Background(), &storage.AnalyticsEvent{
							EventType:    "bot_block",
							UserID:       tgID,
							MetadataJSON: `{"reason":"telegram_403_blocked"}`,
						})
					}(user.TelegramID)
				} else {
					result.TotalFailed++
					log.Printf("[Marketing] Failed to deliver to %d: %v", user.TelegramID, sendErr)
				}
			} else {
				result.TotalSent++
			}
			mu.Unlock()
		}
	}

	result.CompletedAt = time.Now()
	result.DurationMs = result.CompletedAt.Sub(result.StartedAt).Milliseconds()

	// Log broadcast event
	_ = s.repo.RecordEvent(ctx, &storage.AnalyticsEvent{
		EventType: "broadcast_completed",
		UserID:    0,
		MetadataJSON: mustJSON(map[string]interface{}{
			"format":        payload.Format,
			"targeted":      result.TotalTargeted,
			"sent":          result.TotalSent,
			"blocked":       result.TotalBlocked,
			"failed":        result.TotalFailed,
			"duration_ms":   result.DurationMs,
		}),
	})

	return result, nil
}

func (s *Service) sendRawJSON(target *telebot.Chat, rawJSON string) error {
	var msg RawJSONMessage
	if err := json.Unmarshal([]byte(rawJSON), &msg); err != nil {
		return err
	}

	opts := &telebot.SendOptions{}
	if msg.ParseMode == "HTML" || msg.ParseMode == "html" {
		opts.ParseMode = telebot.ModeHTML
	} else if strings.HasPrefix(strings.ToLower(msg.ParseMode), "markdown") {
		opts.ParseMode = telebot.ModeMarkdownV2
	}

	if len(msg.InlineMarkup) > 0 {
		markup := &telebot.ReplyMarkup{}
		var markupRows []telebot.Row
		for _, row := range msg.InlineMarkup {
			var btns []telebot.Btn
			for _, b := range row {
				btn := telebot.Btn{Text: b.Text}
				if b.URL != "" {
					btn.URL = b.URL
				} else if b.Data != "" {
					btn.Data = b.Data
				} else if b.WebApp != "" {
					btn.WebApp = &telebot.WebApp{URL: b.WebApp}
				}
				btns = append(btns, btn)
			}
			markupRows = append(markupRows, btns)
		}
		markup.Inline(markupRows...)
		opts.ReplyMarkup = markup
	}

	if msg.Photo != "" {
		photo := &telebot.Photo{File: telebot.File{FileID: msg.Photo}, Caption: msg.Text}
		_, err := s.bot.Send(target, photo, opts)
		return err
	} else if msg.Video != "" {
		video := &telebot.Video{File: telebot.File{FileID: msg.Video}, Caption: msg.Text}
		_, err := s.bot.Send(target, video, opts)
		return err
	}

	_, err := s.bot.Send(target, msg.Text, opts)
	return err
}

func mustJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
