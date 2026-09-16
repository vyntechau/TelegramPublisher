package i18n

import (
	"fmt"
	"strings"
)

// Language represents metadata about a supported language.
type Language struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"native_name"`
	Flag       string `json:"flag"`
	IsRTL      bool   `json:"is_rtl"`
}

// SupportedLanguages list of all available languages in the system.
var SupportedLanguages = []Language{
	{Code: "en", Name: "English", NativeName: "English", Flag: "🇺🇸", IsRTL: false},
	{Code: "fa", Name: "Persian", NativeName: "فارسی", Flag: "🇮🇷", IsRTL: true},
	{Code: "ar", Name: "Arabic", NativeName: "العربية", Flag: "🇸🇦", IsRTL: true},
	{Code: "ru", Name: "Russian", NativeName: "Русский", Flag: "🇷🇺", IsRTL: false},
	{Code: "es", Name: "Spanish", NativeName: "Español", Flag: "🇪🇸", IsRTL: false},
	{Code: "de", Name: "German", NativeName: "Deutsch", Flag: "🇩🇪", IsRTL: false},
	{Code: "zh", Name: "Chinese", NativeName: "中文", Flag: "🇨🇳", IsRTL: false},
}

var translations = map[string]map[string]string{
	"en": {
		"welcome_title":            "👋 *Welcome to TelegramPublisher!*",
		"welcome_fast_delivery":     "⚡ *Fast & Protected Content Delivery*",
		"welcome_desc":             "🔒 Media links are protected with automated expiration timers.\n💎 Get VIP pass with `/subscribe` for unlimited streaming without auto-delete limits.\n\nSample Open Source project for *VynTech Cloud* infrastructure.",
		"btn_mini_app":             "🚀 Open Mini App & Media Catalog",
		"btn_vip_sub":              "💎 VIP Subscription (No Auto-Delete)",
		"btn_help":                 "ℹ️ Help & FAQ",
		"btn_share":                "↗️ Share Bot",
		"btn_language":             "🌐 Language / زبان",
		"btn_report":               "🚨 Report Broken",
		"btn_my_profile":           "👤 My Profile",
		"lang_select_prompt":       "🌐 *Choose your preferred language:*",
		"lang_changed":             "✅ Language changed to *%s*.",
		"help_title":               "📖 *TelegramPublisher Commands Guide*",
		"help_user_cmds":           "*For Users:*\n• `/start` - Launch bot & open Mini App\n• `/subscribe` - Unlock VIP subscription passes\n• `/mystatus` - Check VIP subscription status\n• `/language` or `/lang` - Change language\n• `/help` - Show this guide",
		"help_admin_cmds":          "\n*For Admins & Authors:*\n• `/stats` - View performance KPI & analytics\n• `/broadcast <msg>` - Send message to all active users\n• `/setttl <sec>` - Change media auto-delete timer\n• `/channels` - Manage auto-posting destination channels\n• `/reports` - Review broken media reports",
		"sub_title":                "💎 *TelegramPublisher VIP Membership*",
		"sub_benefits":             "✨ *Benefits:*\n• ⚡ Instant direct streaming\n• 🛡️ *Zero auto-delete timers* — media stays forever\n• 🚀 High-speed priority cloud delivery\n• 👑 Exclusive VIP badge & direct author support",
		"sub_select_plan":          "Select a crypto subscription plan below (AzPays, Coinbase, NOWPayments):",
		"fsub_required":            "📢 *Channel Membership Required*",
		"fsub_desc":                "To access this media, please join our official channel first, then tap **Check Membership** below.",
		"btn_join_channel":         "📢 Join Channel",
		"btn_check_membership":     "🔄 Check Membership",
		"fsub_verified":            "✅ Thank you for joining! Sending your media now...",
		"err_not_joined":           "❌ You haven't joined all required channels yet. Please join and try again.",
		"media_expired":            "⏳ *This media link has expired.* Please request it again from the bot or Mini App.",
		"err_unauthorized":         "⛔ You are not authorized to run this administrative command.",
	},
	"fa": {
		"welcome_title":            "👋 *به ربات تلگرام پابلیشر خوش آمدید!*",
		"welcome_fast_delivery":     "⚡ *ارسال سریع و محتوای محافظت شده*",
		"welcome_desc":             "🔒 لینک‌های مدیا دارای تایمر حذف خودکار هستند.\n💎 با دستور `/subscribe` اشتراک VIP دریافت کنید تا بدون محدودیت و بدون حذف فایل دانلود کنید.\n\nپروژه متن‌باز مبتنی بر زیرساخت ابری *VynTech Cloud*.",
		"btn_mini_app":             "🚀 باز کردن مینی‌اپ و کاتالوگ",
		"btn_vip_sub":              "💎 اشتراک ویژه VIP (بدون حذف خودکار)",
		"btn_help":                 "ℹ️ راهنما و سوالات متداول",
		"btn_share":                "↗️ اشتراک‌گذاری ربات",
		"btn_language":             "🌐 تغییر زبان / Language",
		"btn_report":               "🚨 گزارش خرابی لینک",
		"btn_my_profile":           "👤 پروفایل من",
		"lang_select_prompt":       "🌐 *لطفاً زبان مورد نظر خود را انتخاب کنید:*",
		"lang_changed":             "✅ زبان به *%s* تغییر یافت.",
		"help_title":               "📖 *راهنمای دستورات ربات*",
		"help_user_cmds":           "*دستورات عمومی:*\n• `/start` - اجرای ربات و باز کردن مینی‌اپ\n• `/subscribe` - خرید و فعال‌سازی اشتراک VIP\n• `/mystatus` - بررسی وضعیت اشتراک VIP\n• `/lang` یا `/language` - تغییر زبان ربات\n• `/help` - مشاهده راهنما",
		"help_admin_cmds":          "\n*دستورات مدیران و نویسندگان:*\n• `/stats` - مشاهده آمار و تحلیل سیستم\n• `/broadcast <متن>` - ارسال پیام همگانی\n• `/setttl <ثانیه>` - تنظیم زمان حذف خودکار مدیا\n• `/channels` - مدیریت کانال‌های مقصد انتشار\n• `/reports` - بررسی گزارش‌های خرابی",
		"sub_title":                "💎 *عضویت ویژه VIP تلگرام پابلیشر*",
		"sub_benefits":             "✨ *مزایای اشتراک:*\n• ⚡ پخش آنی و مستقیم\n• 🛡️ *بدون حذف خودکار* — فایل‌ها همیشه باقی می‌مانند\n• 🚀 دانلود با بالاترین سرعت از سرور ابری\n• 👑 نشان ویژه VIP و دسترسی به محتوای اختصاصی",
		"sub_select_plan":          "یکی از پلن‌های پرداخت ارزی زیر را انتخاب کنید (AzPays, Coinbase, NOWPayments):",
		"fsub_required":            "📢 *عضویت در کانال الزامی است*",
		"fsub_desc":                "برای دسترسی به این مدیا، لطفاً ابتدا در کانال رسمی ما عضو شوید و سپس دکمه **بررسی عضویت** را بزنید.",
		"btn_join_channel":         "📢 عضویت در کانال",
		"btn_check_membership":     "🔄 بررسی عضویت",
		"fsub_verified":            "✅ عضویت شما تایید شد! فایل مدیا ارسال می‌شود...",
		"media_expired":            "⏳ *این لینک مدیا منقضی شده است.* لطفاً مجدداً از مینی‌اپ درخواست دهید.",
		"err_not_joined":           "❌ شما هنوز در کانال‌های تعیین‌شده عضو نشده‌اید.",
		"err_unauthorized":         "⛔ شما دسترسی لازم برای اجرای این دستور مدیریتی را ندارید.",
	},
	"ar": {
		"welcome_title":            "👋 *مرحباً بك في TelegramPublisher!*",
		"welcome_fast_delivery":     "⚡ *توصيل سريع ومحتوى محمي*",
		"welcome_desc":             "🔒 روابط الوسائط محمية بمؤقتات الحذف التلقائي.\n💎 احصل على عضوية VIP عبر `/subscribe` للبث غير المحدود دون حذف.\n\nمشروع مفتوح المصدر مدعوم من *VynTech Cloud*.",
		"btn_mini_app":             "🚀 فتح التطبيق المصغر والكتالوج",
		"btn_vip_sub":              "💎 اشتراك VIP (بدون حذف تلقائي)",
		"btn_help":                 "ℹ️ المساعدة والأسئلة الشائعة",
		"btn_share":                "↗️ مشاركة البوت",
		"btn_language":             "🌐 تغيير اللغة / Language",
		"btn_report":               "🚨 الإبلاغ عن رابط معطل",
		"btn_my_profile":           "👤 ملفي الشخصي",
		"lang_select_prompt":       "🌐 *يرجى اختيار لغتك المفضلة:*",
		"lang_changed":             "✅ تم تغيير اللغة إلى *%s*.",
		"help_title":               "📖 *دليل أوامر البوت*",
		"help_user_cmds":           "*للمستخدمين:*\n• `/start` - تشغيل البوت وفتح التطبيق المصغر\n• `/subscribe` - فتح اشتراكات VIP\n• `/mystatus` - التحقق من حالة الاشتراك\n• `/lang` - تغيير لغة البوت\n• `/help` - عرض هذا الدليل",
		"help_admin_cmds":          "\n*للمسؤولين:*\n• `/stats` - عرض الإحصائيات\n• `/broadcast <رسالة>` - إذاعة رسالة للجميع\n• `/channels` - إدارة القنوات\n• `/reports` - مراجعة البلاغات",
		"sub_title":                "💎 *عضوية TelegramPublisher VIP المميزة*",
		"sub_benefits":             "✨ *المزايا:*\n• ⚡ بث فوري ومباشر\n• 🛡️ *بدون حذف تلقائي* — الوسائط تبقى للأبد\n• 🚀 سرعة فائقة من السحابة\n• 👑 شارة VIP ودعم مباشر",
		"sub_select_plan":          "اختر خطة الاشتراك بالعملات المشفرة أدناه:",
		"fsub_required":            "📢 *الانضمام إلى القناة مطلوب*",
		"fsub_desc":                "للوصول إلى هذه الوسائط، يرجى الانضمام إلى قناتنا أولاً ثم النقر على **التحقق من العضوية**.",
		"btn_join_channel":         "📢 انضم للقناة",
		"btn_check_membership":     "🔄 تحقق من العضوية",
		"fsub_verified":            "✅ شكراً لانضمامك! يتم إرسال الوسائط الآن...",
		"media_expired":            "⏳ *انتهت صلاحية هذا الرابط.*",
		"err_not_joined":           "❌ لم تنضم إلى جميع القنوات المطلوبة بعد.",
		"err_unauthorized":         "⛔ غير مصرح لك بتنفيذ هذا الأمر.",
	},
	"ru": {
		"welcome_title":            "👋 *Добро пожаловать в TelegramPublisher!*",
		"welcome_fast_delivery":     "⚡ *Быстрая и защищенная доставка контента*",
		"welcome_desc":             "🔒 Медиа-ссылки защищены автоматическим таймером удаления.\n💎 Оформите VIP-подписку с помощью `/subscribe` для просмотра без ограничений.\n\nПроект с открытым исходным кодом для инфраструктуры *VynTech Cloud*.",
		"btn_mini_app":             "🚀 Открыть Mini App и каталог",
		"btn_vip_sub":              "💎 VIP Подписка (Без автоудаления)",
		"btn_help":                 "ℹ️ Помощь и FAQ",
		"btn_share":                "↗️ Поделиться ботом",
		"btn_language":             "🌐 Язык / Language",
		"btn_report":               "🚨 Сообщить о проблеме",
		"btn_my_profile":           "👤 Мой профиль",
		"lang_select_prompt":       "🌐 *Выберите предпочитаемый язык:*",
		"lang_changed":             "✅ Язык изменен на *%s*.",
		"help_title":               "📖 *Руководство по командам*",
		"help_user_cmds":           "*Для пользователей:*\n• `/start` - Запустить бота и открыть Mini App\n• `/subscribe` - Оформить VIP-подписку\n• `/mystatus` - Проверить статус подписки\n• `/lang` - Сменить язык\n• `/help` - Показать справку",
		"help_admin_cmds":          "\n*Для администраторов:*\n• `/stats` - Аналитика и метрики\n• `/broadcast <текст>` - Рассылка всем пользователям\n• `/setttl <сек>` - Настройка времени автоудаления",
		"sub_title":                "💎 *VIP-членство TelegramPublisher*",
		"sub_benefits":             "✨ *Преимущества:*\n• ⚡ Мгновенный просмотр\n• 🛡️ *Без автоудаления* — файлы сохраняются навсегда\n• 🚀 Максимальная скорость облака",
		"sub_select_plan":          "Выберите тариф криптоподписки:",
		"fsub_required":            "📢 *Требуется подписка на канал*",
		"fsub_desc":                "Для доступа к медиа подпишитесь на наш канал и нажмите кнопку **Проверить подписку**.",
		"btn_join_channel":         "📢 Подписаться на канал",
		"btn_check_membership":     "🔄 Проверить подписку",
		"fsub_verified":            "✅ Спасибо за подписку! Отправляем медиа...",
		"media_expired":            "⏳ *Срок действия этой ссылки истек.*",
		"err_not_joined":           "❌ Вы еще не подписались на обязательные каналы.",
		"err_unauthorized":         "⛔ У вас нет прав для выполнения этой команды.",
	},
	"es": {
		"welcome_title":            "👋 *¡Bienvenido a TelegramPublisher!*",
		"welcome_fast_delivery":     "⚡ *Entrega de contenido rápida y protegida*",
		"welcome_desc":             "🔒 Los enlaces multimedia están protegidos con temporizadores de autoeliminación.\n💎 Obtén pase VIP con `/subscribe` para streaming ilimitado sin límites de tiempo.\n\nProyecto de código abierto para la infraestructura *VynTech Cloud*.",
		"btn_mini_app":             "🚀 Abrir Mini App y Catálogo",
		"btn_vip_sub":              "💎 Suscripción VIP (Sin Autoeliminación)",
		"btn_help":                 "ℹ️ Ayuda y Preguntas Frecuentes",
		"btn_share":                "↗️ Compartir Bot",
		"btn_language":             "🌐 Idioma / Language",
		"btn_report":               "🚨 Reportar Enlace Caído",
		"btn_my_profile":           "👤 Mi Perfil",
		"lang_select_prompt":       "🌐 *Selecciona tu idioma preferido:*",
		"lang_changed":             "✅ Idioma cambiado a *%s*.",
		"help_title":               "📖 *Guía de Comandos*",
		"help_user_cmds":           "*Para usuarios:*\n• `/start` - Iniciar bot y abrir Mini App\n• `/subscribe` - Desbloquear pases VIP\n• `/mystatus` - Ver estado de suscripción\n• `/lang` - Cambiar idioma\n• `/help` - Ver ayuda",
		"help_admin_cmds":          "\n*Para administradores:*\n• `/stats` - Estadísticas y analíticas\n• `/broadcast <mensaje>` - Enviar mensaje masivo\n• `/channels` - Gestionar canales",
		"sub_title":                "💎 *Membresía VIP de TelegramPublisher*",
		"sub_benefits":             "✨ *Beneficios:*\n• ⚡ Streaming directo instantáneo\n• 🛡️ *Sin autoeliminación* — los archivos quedan para siempre\n• 🚀 Máxima velocidad en la nube",
		"sub_select_plan":          "Selecciona un plan de suscripción cripto:",
		"fsub_required":            "📢 *Se requiere suscripción al canal*",
		"fsub_desc":                "Para acceder a este contenido, únete a nuestro canal oficial y pulsa **Verificar Suscripción**.",
		"btn_join_channel":         "📢 Unirse al Canal",
		"btn_check_membership":     "🔄 Verificar Suscripción",
		"fsub_verified":            "✅ ¡Gracias por unirte! Enviando contenido...",
		"media_expired":            "⏳ *Este enlace multimedia ha caducado.*",
		"err_not_joined":           "❌ Aún no te has unido a los canales requeridos.",
		"err_unauthorized":         "⛔ No estás autorizado para ejecutar este comando.",
	},
	"de": {
		"welcome_title":            "👋 *Willkommen bei TelegramPublisher!*",
		"welcome_fast_delivery":     "⚡ *Schnelle & geschützte Medienbereitstellung*",
		"welcome_desc":             "🔒 Medienlinks sind mit automatischer Löschung geschützt.\n💎 Holen Sie sich den VIP-Pass mit `/subscribe` für unbegrenztes Streaming ohne Löschfristen.\n\nOpen-Source-Projekt für *VynTech Cloud* Infrastruktur.",
		"btn_mini_app":             "🚀 Mini App & Katalog öffnen",
		"btn_vip_sub":              "💎 VIP-Abonnement (Kein automatisches Löschen)",
		"btn_help":                 "ℹ️ Hilfe & FAQ",
		"btn_share":                "↗️ Bot teilen",
		"btn_language":             "🌐 Sprache / Language",
		"btn_report":               "🚨 Fehlerhaften Link melden",
		"btn_my_profile":           "👤 Mein Profil",
		"lang_select_prompt":       "🌐 *Wählen Sie Ihre bevorzugte Sprache:*",
		"lang_changed":             "✅ Sprache geändert zu *%s*.",
		"help_title":               "📖 *Befehlsübersicht*",
		"help_user_cmds":           "*Für Benutzer:*\n• `/start` - Bot starten & Mini App öffnen\n• `/subscribe` - VIP-Pässe freischalten\n• `/mystatus` - VIP-Status prüfen\n• `/lang` - Sprache ändern\n• `/help` - Hilfe anzeigen",
		"help_admin_cmds":          "\n*Für Administratoren:*\n• `/stats` - Statistiken anzeigen\n• `/broadcast <Text>` - Nachricht an alle senden",
		"sub_title":                "💎 *TelegramPublisher VIP-Mitgliedschaft*",
		"sub_benefits":             "✨ *Vorteile:*\n• ⚡ Sofortiges Streaming\n• 🛡️ *Kein automatisches Löschen*\n• 🚀 Höchste Cloud-Geschwindigkeit",
		"sub_select_plan":          "Wählen Sie einen Krypto-Abo-Plan:",
		"fsub_required":            "📢 *Kanalmitgliedschaft erforderlich*",
		"fsub_desc":                "Um auf diese Medien zuzugreifen, treten Sie bitte unserem Kanal bei.",
		"btn_join_channel":         "📢 Kanal beitreten",
		"btn_check_membership":     "🔄 Mitgliedschaft prüfen",
		"fsub_verified":            "✅ Vielen Dank! Medien werden gesendet...",
		"media_expired":            "⏳ *Dieser Medienlink ist abgelaufen.*",
		"err_not_joined":           "❌ Sie sind den erforderlichen Kanälen noch nicht beigetreten.",
		"err_unauthorized":         "⛔ Sie sind nicht berechtigt, diesen Befehl auszuführen.",
	},
	"zh": {
		"welcome_title":            "👋 *欢迎使用 TelegramPublisher！*",
		"welcome_fast_delivery":     "⚡ *快速且受保护的内容分发*",
		"welcome_desc":             "🔒 媒体链接受自动定时删除保护。\n💎 使用 `/subscribe` 获取 VIP 会员，享受无时间限制点播。\n\n*VynTech Cloud* 云基础设施开源项目。",
		"btn_mini_app":             "🚀 打开 Mini App 和媒体目录",
		"btn_vip_sub":              "💎 VIP 订阅（永不自动删除）",
		"btn_help":                 "ℹ️ 帮助与常见问题",
		"btn_share":                "↗️ 分享机器人",
		"btn_language":             "🌐 语言 / Language",
		"btn_report":               "🚨 报告损坏链接",
		"btn_my_profile":           "👤 我的个人资料",
		"lang_select_prompt":       "🌐 *请选择您的首选语言：*",
		"lang_changed":             "✅ 语言已更改为 *%s*。",
		"help_title":               "📖 *机器人指令指南*",
		"help_user_cmds":           "*用户指令：*\n• `/start` - 启动机器人并打开小程序\n• `/subscribe` - 解锁 VIP 订阅\n• `/mystatus` - 查询订阅状态\n• `/lang` - 切换语言\n• `/help` - 显示帮助",
		"help_admin_cmds":          "\n*管理员指令：*\n• `/stats` - 查看数据与统计\n• `/broadcast <内容>` - 发送全员广播",
		"sub_title":                "💎 *TelegramPublisher VIP 尊贵会员*",
		"sub_benefits":             "✨ *会员特权：*\n• ⚡ 极速直链点播\n• 🛡️ *无自动删除限制* — 文件永久保存\n• 🚀 专享高速云端通道",
		"sub_select_plan":          "请选择下方的加密货币订阅方案：",
		"fsub_required":            "📢 *必须先加入频道*",
		"fsub_desc":                "如需访问此媒体内容，请先加入我们的官方频道，然后点击**验证加入**。",
		"btn_join_channel":         "📢 加入频道",
		"btn_check_membership":     "🔄 验证加入",
		"fsub_verified":            "✅ 感谢您的加入！正在发送媒体内容...",
		"media_expired":            "⏳ *该媒体链接已过期。*",
		"err_not_joined":           "❌ 您尚未加入所有指定的频道。",
		"err_unauthorized":         "⛔ 您无权执行此管理指令。",
	},
}

// T translates a key into the target language with optional formatting parameters, falling back to English.
func T(langCode, key string, args ...any) string {
	langCode = strings.ToLower(strings.TrimSpace(langCode))
	if langMap, exists := translations[langCode]; exists {
		if text, ok := langMap[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(text, args...)
			}
			return text
		}
	}

	// Fallback to English
	if enMap, exists := translations["en"]; exists {
		if text, ok := enMap[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(text, args...)
			}
			return text
		}
	}

	return key
}

// GetLanguageMeta returns metadata for a language code or default English.
func GetLanguageMeta(code string) Language {
	code = strings.ToLower(strings.TrimSpace(code))
	for _, l := range SupportedLanguages {
		if l.Code == code {
			return l
		}
	}
	return SupportedLanguages[0] // English
}

// IsSupported returns true if the language code is in the supported list.
func IsSupported(code string) bool {
	code = strings.ToLower(strings.TrimSpace(code))
	for _, l := range SupportedLanguages {
		if l.Code == code {
			return true
		}
	}
	return false
}

// FilterSupported returns only languages present in the comma-separated enabled list.
func FilterSupported(enabledCSV string) []Language {
	if enabledCSV == "" {
		return SupportedLanguages
	}
	parts := strings.Split(enabledCSV, ",")
	allowed := make(map[string]bool)
	for _, p := range parts {
		allowed[strings.ToLower(strings.TrimSpace(p))] = true
	}

	var result []Language
	for _, l := range SupportedLanguages {
		if allowed[l.Code] {
			result = append(result, l)
		}
	}
	if len(result) == 0 {
		return SupportedLanguages
	}
	return result
}
