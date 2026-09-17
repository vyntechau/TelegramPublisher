package bot

import (
	"context"
	"strings"

	"github.com/vyntechau/TelegramPublisher/internal/i18n"
	"github.com/vyntechau/TelegramPublisher/internal/services/settings"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"gopkg.in/telebot.v3"
)

// UserMiddleware automatically registers incoming users, updates active timestamp, and blocks banned users.
func (e *Engine) UserMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		sender := c.Sender()
		if sender == nil {
			return next(c)
		}

		ctx := context.Background()

		// Check if owner from config
		role := storage.RoleUser
		if sender.ID == e.Config.Bot.OwnerID && e.Config.Bot.OwnerID != 0 {
			role = storage.RoleOwner
		}

		// Look up existing user
		existing, err := e.Repo.GetUserByTelegramID(ctx, sender.ID)
		if err != nil {
			// New user registration
			user := &storage.User{
				TelegramID: sender.ID,
				Username:   sender.Username,
				FirstName:  sender.FirstName,
				Role:       role,
				Status:     storage.StatusActive,
			}
			_ = e.Repo.UpsertUser(ctx, user)
		} else {
			// Check if banned
			if existing.Status == storage.StatusBanned {
				userLang := e.GetUserLang(c)
				return c.Send(i18n.T(userLang, "err_banned"), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
			}

			// If owner role needs sync
			if sender.ID == e.Config.Bot.OwnerID && existing.Role != storage.RoleOwner {
				_ = e.Repo.UpdateUserRole(ctx, sender.ID, storage.RoleOwner)
			}

			// Touch active & update info
			_ = e.Repo.UpsertUser(ctx, &storage.User{
				TelegramID: sender.ID,
				Username:   sender.Username,
				FirstName:  sender.FirstName,
				Role:       existing.Role,
				Status:     existing.Status,
			})
		}

		return next(c)
	}
}

// AdminOnly middleware restricts handler to Owner and Admin roles, and checks bot_admin_enabled flag.
func (e *Engine) AdminOnly(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		ctx := context.Background()

		// Check if bot administration is toggled on in database settings table
		botAdminEnabled := e.SettingsSvc.GetBool(ctx, settings.KeyBotAdminEnabled, true)
		if !botAdminEnabled {
			userLang := e.GetUserLang(c)
			return c.Send(i18n.T(userLang, "err_admin_disabled"), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
		}

		sender := c.Sender()
		if sender == nil {
			return nil
		}

		if sender.ID == e.Config.Bot.OwnerID && e.Config.Bot.OwnerID != 0 {
			return next(c)
		}

		user, err := e.Repo.GetUserByTelegramID(ctx, sender.ID)
		if err != nil || (user.Role != storage.RoleAdmin && user.Role != storage.RoleOwner) {
			userLang := e.GetUserLang(c)
			return c.Send(i18n.T(userLang, "err_unauthorized"), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
		}

		return next(c)
	}
}

// AuthorOnly middleware allows Owner, Admin, and Author roles to upload media and inspect file IDs.
func (e *Engine) AuthorOnly(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		sender := c.Sender()
		if sender == nil {
			return nil
		}

		if sender.ID == e.Config.Bot.OwnerID && e.Config.Bot.OwnerID != 0 {
			return next(c)
		}

		ctx := context.Background()
		user, err := e.Repo.GetUserByTelegramID(ctx, sender.ID)
		if err != nil || (user.Role != storage.RoleAuthor && user.Role != storage.RoleAdmin && user.Role != storage.RoleOwner) {
			// Check if user has active author_pro subscription
			sub, _ := e.Repo.GetActiveSubscription(ctx, sender.ID)
			if sub != nil && (sub.Tier == "author_pro" || sub.Tier == "lifetime") {
				return next(c)
			}
			userLang := e.GetUserLang(c)
			return c.Send(i18n.T(userLang, "err_author_required"), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
		}

		return next(c)
	}
}

func sanitizeInput(s string) string {
	return strings.TrimSpace(s)
}
