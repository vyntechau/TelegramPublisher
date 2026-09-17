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
	e.Bot.Handle("/menu", e.sendPersistentKeyboard)
	e.Bot.Handle("/admin", e.AdminOnly(e.HandleBotSettingsOverview))
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

	// Role-Based Persistent Keyboard Text Triggers (Multi-Language Support)
	// 1. Executive Analytics / Stats
	analyticsTriggers := []string{
		"📊 Executive Analytics", "📊 آمار و تحلیل", "📊 الإحصائيات التنفيذية",
		"📊 Аналитика", "📊 Analítica Ejecutiva", "📊 Analysen & KPIs", "📊 核心数据分析",
	}
	for _, t := range analyticsTriggers {
		e.Bot.Handle(t, e.AdminOnly(e.HandleStats))
	}

	// 2. Broadcast Announcement
	broadcastTriggers := []string{
		"📢 Broadcast", "📢 ارسال همگانی", "📢 الإذاعة العامة",
		"📢 Рассылка", "📢 Transmisión Masiva", "📢 Rundschreiben", "📢 全员广播消息",
	}
	for _, t := range broadcastTriggers {
		e.Bot.Handle(t, e.AdminOnly(func(c telebot.Context) error {
			return c.Send("📢 *Broadcast Announcement*:\nUse `/broadcast <your_message>` to send an update to all active users.", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
		}))
	}

	// 3. Broken Content Reports & Tickets
	reportTriggers := []string{
		"🚩 Broken Tickets", "🚩 Content Reports", "🚩 گزارش‌های خرابی", "🚩 گزارش‌های محتوا",
		"🚩 تذاكر البلاغات", "🚩 تقارير المحتوى", "🚩 Жалобы на файлы", "🚩 Отчеты по контенту",
		"🚩 Reportes de Archivos", "🚩 Reportes de Contenido", "🚩 Fehler-Tickets", "🚩 Inhaltsberichte",
		"🚩 故障文件工单", "🚩 媒体内容报告",
	}
	for _, t := range reportTriggers {
		e.Bot.Handle(t, e.AuthorOnly(e.HandleListReports))
	}

	// 4. Users Directory
	userTriggers := []string{
		"👥 Users Directory", "👥 فهرست کاربران", "👥 دليل المستخدمين",
		"👥 Пользователи", "👥 Directorio de Usuarios", "👥 Benutzerverzeichnis", "👥 平台用户目录",
	}
	for _, t := range userTriggers {
		e.Bot.Handle(t, e.AdminOnly(e.HandleUsers))
	}

	// 5. ForceSub Channels
	channelTriggers := []string{
		"📢 ForceSub Channels", "📢 کانال‌های عضویت", "📢 قنوات الاشتراك",
		"📢 Каналы подписки", "📢 Canales Obligatorios", "📢 Pflicht-Kanäle", "📢 强制关注频道",
	}
	for _, t := range channelTriggers {
		e.Bot.Handle(t, e.AdminOnly(e.HandleListChannels))
	}

	// 6. Upload Media Content
	uploadTriggers := []string{
		"➕ Upload Content", "➕ آپلود محتوا", "➕ رفع محتوى",
		"➕ Загрузить контент", "➕ Subir Contenido", "➕ Inhalt hochladen", "➕ 上传媒体内容",
	}
	for _, t := range uploadTriggers {
		e.Bot.Handle(t, e.AuthorOnly(func(c telebot.Context) error {
			return c.Send("📤 *Upload Media*: Send or forward any photo, video, document, or animation to this chat to register a new protected post.", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
		}))
	}

	// 7. Bot Settings Overview
	settingsTriggers := []string{
		"⚙️ Bot Settings", "⚙️ تنظیمات ربات", "⚙️ إعدادات البوت",
		"⚙️ Настройки бота", "⚙️ Ajustes del Bot", "⚙️ Bot-Einstellungen", "⚙️ 机器人系统设置",
	}
	for _, t := range settingsTriggers {
		e.Bot.Handle(t, e.AdminOnly(e.HandleBotSettingsOverview))
	}

	// Standard User Button Triggers
	vipTriggers := []string{
		"💎 VIP Subscription", "💎 VIP Subscription (No Auto-Delete)", "💎 اشتراک ویژه VIP (بدون حذف خودکار)",
		"💎 اشتراك VIP (بدون حذف تلقائي)", "💎 VIP Подписка (Без автоудаления)", "💎 Suscripción VIP (Sin Autoeliminación)",
		"💎 VIP-Abonnement (Kein automatisches Löschen)", "💎 VIP 订阅（永不自动删除）",
	}
	for _, t := range vipTriggers {
		e.Bot.Handle(t, e.HandleSubscribe)
	}

	profileTriggers := []string{
		"👤 My Profile", "👤 پروفایل من", "👤 ملفي الشخصي", "👤 Мой профиль", "👤 Mi Perfil", "👤 Mein Profil", "👤 我的个人资料",
	}
	for _, t := range profileTriggers {
		e.Bot.Handle(t, e.HandleMyStatus)
	}

	helpTriggers := []string{
		"ℹ️ Help", "ℹ️ Help & FAQ", "ℹ️ راهنما و سوالات متداول", "ℹ️ المساعدة والأسئلة الشائعة",
		"ℹ️ Помощь и FAQ", "ℹ️ Ayuda y Preguntas Frecuentes", "ℹ️ Hilfe & FAQ", "ℹ️ 帮助与常见问题",
	}
	for _, t := range helpTriggers {
		e.Bot.Handle(t, e.HandleHelp)
	}

	langTriggers := []string{
		"🌐 Language", "🌐 Language / زبان", "🌐 تغییر زبان / Language", "🌐 تغيير اللغة / Language",
		"🌐 Язык / Language", "🌐 Idioma / Language", "🌐 Sprache / Language", "🌐 语言 / Language",
	}
	for _, t := range langTriggers {
		e.Bot.Handle(t, e.HandleLanguage)
	}

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

