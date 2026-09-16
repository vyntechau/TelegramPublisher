package bot

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/vyntechau/TelegramPublisher/config"
	"github.com/vyntechau/TelegramPublisher/internal/cleaner"
	"github.com/vyntechau/TelegramPublisher/internal/services/marketing"
	"github.com/vyntechau/TelegramPublisher/internal/services/payment"
	"github.com/vyntechau/TelegramPublisher/internal/services/settings"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"gopkg.in/telebot.v3"
)

// Engine wraps the Telegram bot instance and connected services.
type Engine struct {
	Bot         *telebot.Bot
	Repo        storage.Repository
	Config      *config.Config
	Cleaner     *cleaner.Cleaner
	SettingsSvc *settings.Service
	MarketSvc   *marketing.Service
	PaySvc      *payment.Service
	userLangs   sync.Map
}

func NewEngine(cfg *config.Config, repo storage.Repository) (*Engine, error) {
	if cfg.Bot.Token == "" {
		return nil, fmt.Errorf("bot token is empty")
	}

	pref := telebot.Settings{
		Token:  cfg.Bot.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telebot: %w", err)
	}

	cleanerWorker := cleaner.New(b)
	settingsService := settings.NewService(repo, cfg)
	marketService := marketing.NewService(b, repo)
	payService := payment.NewService(repo, settingsService, b)

	engine := &Engine{
		Bot:         b,
		Repo:        repo,
		Config:      cfg,
		Cleaner:     cleanerWorker,
		SettingsSvc: settingsService,
		MarketSvc:   marketService,
		PaySvc:      payService,
	}

	engine.registerRoutes()
	return engine, nil
}

// Start runs the bot polling and cleaner worker.
func (e *Engine) Start() {
	e.Cleaner.Start()
	log.Printf("[Bot] TelegramPublisher bot started as @%s", e.Bot.Me.Username)
	e.Bot.Start()
}

// Stop gracefully shuts down the bot.
func (e *Engine) Stop() {
	e.Cleaner.Stop()
	e.Bot.Stop()
}

func (e *Engine) registerRoutes() {
	// Middleware: User registration & touch, banned user blocking
	e.Bot.Use(e.UserMiddleware)

	// Command Handlers
	e.Bot.Handle("/start", e.HandleStart)
	e.Bot.Handle("/help", e.HandleHelp)
	e.Bot.Handle("/subscribe", e.HandleSubscribe)
	e.Bot.Handle("/plans", e.HandleSubscribe)
	e.Bot.Handle("/mystatus", e.HandleMyStatus)
	e.Bot.Handle("/lang", e.HandleLanguage)
	e.Bot.Handle("/language", e.HandleLanguage)

	// Admin / Author Commands (Protected by RBAC & dynamic toggle)
	e.Bot.Handle("/stats", e.AdminOnly(e.HandleStats))
	e.Bot.Handle("/analytics", e.AdminOnly(e.HandleStats))
	e.Bot.Handle("/users", e.AdminOnly(e.HandleUsers))
	e.Bot.Handle("/ban", e.AdminOnly(e.HandleBan))
	e.Bot.Handle("/unban", e.AdminOnly(e.HandleUnban))
	e.Bot.Handle("/promote", e.AdminOnly(e.HandlePromote))
	e.Bot.Handle("/demote", e.AdminOnly(e.HandleDemote))
	e.Bot.Handle("/addchannel", e.AdminOnly(e.HandleAddChannel))
	e.Bot.Handle("/delchannel", e.AdminOnly(e.HandleDelChannel))
	e.Bot.Handle("/channels", e.AdminOnly(e.HandleListChannels))
	e.Bot.Handle("/reports", e.AdminOnly(e.HandleListReports))
	e.Bot.Handle("/broadcast", e.AdminOnly(e.HandleBroadcastCommand))
	e.Bot.Handle("/setttl", e.AdminOnly(e.HandleSetTTL))

	// Media Upload & File ID Inspector (Author / Admin)
	e.Bot.Handle(telebot.OnPhoto, e.AuthorOnly(e.HandleMediaUpload))
	e.Bot.Handle(telebot.OnVideo, e.AuthorOnly(e.HandleMediaUpload))
	e.Bot.Handle(telebot.OnDocument, e.AuthorOnly(e.HandleMediaUpload))
	e.Bot.Handle(telebot.OnAnimation, e.AuthorOnly(e.HandleMediaUpload))

	// Callback Query Handlers (Reactions, FSub check, Broken File Reports, Language selection)
	e.Bot.Handle(telebot.OnCallback, e.HandleCallbackQuery)

	// Persistent Keyboard Text Triggers
	e.Bot.Handle("💎 VIP Subscription", e.HandleSubscribe)
	e.Bot.Handle("👤 My Profile", e.HandleMyStatus)
	e.Bot.Handle("ℹ️ Help", e.HandleHelp)
	e.Bot.Handle("🌐 Language", e.HandleLanguage)
	e.Bot.Handle("🚨 Report Broken", func(c telebot.Context) error {
		return c.Send("🚩 *Reporting Broken Content*:\nClick the '🚨 Report Broken' button under any media post, or send the post link/slug to the admin.", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	})
	e.Bot.Handle("↗️ Share Bot", func(c telebot.Context) error {
		botUsername := "bot"
		if e.Bot.Me != nil && e.Bot.Me.Username != "" {
			botUsername = e.Bot.Me.Username
		}
		shareURL := fmt.Sprintf("https://t.me/share/url?url=https://t.me/%s&text=Check+out+this+bot+for+fast+media+delivery!", botUsername)
		inlineMarkup := &telebot.ReplyMarkup{}
		btnShare := inlineMarkup.URL("↗️ Forward / Share Bot", shareURL)
		inlineMarkup.Inline(inlineMarkup.Row(btnShare))
		return c.Send("Share TelegramPublisher with friends or channels:", inlineMarkup)
	})
}
