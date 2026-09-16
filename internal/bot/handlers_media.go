package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vyntechau/TelegramPublisher/internal/services/settings"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"gopkg.in/telebot.v3"
)

// HandleMediaUpload processes media sent by authors/admins and creates a new post with File ID details.
func (e *Engine) HandleMediaUpload(c telebot.Context) error {
	ctx := context.Background()
	sender := c.Sender()

	var fileID string
	var fileUniqueID string
	var fileType string
	var caption string

	msg := c.Message()
	caption = msg.Caption

	if msg.Video != nil {
		fileID = msg.Video.FileID
		fileUniqueID = msg.Video.UniqueID
		fileType = storage.FileTypeVideo
	} else if msg.Photo != nil {
		fileID = msg.Photo.FileID
		fileUniqueID = msg.Photo.UniqueID
		fileType = storage.FileTypePhoto
	} else if msg.Document != nil {
		fileID = msg.Document.FileID
		fileUniqueID = msg.Document.UniqueID
		fileType = storage.FileTypeDocument
	} else if msg.Animation != nil {
		fileID = msg.Animation.FileID
		fileUniqueID = msg.Animation.UniqueID
		fileType = storage.FileTypeAnimation
	} else {
		return c.Send("Unsupported media format.")
	}

	// Generate clean slug
	slug := "p_" + uuid.New().String()[:8]
	defaultTTL := e.SettingsSvc.GetInt(ctx, settings.KeyAutoDeleteSeconds, 120)

	post := &storage.Post{
		Slug:              slug,
		FileID:            fileID,
		FileUniqueID:      fileUniqueID,
		FileType:          fileType,
		Caption:           caption,
		AuthorID:          sender.ID,
		AutoDeleteSeconds: defaultTTL,
		IsProtected:       true,
		CreatedAt:         time.Now(),
	}

	if err := e.Repo.CreatePost(ctx, post); err != nil {
		return c.Send(fmt.Sprintf("❌ Error creating post: %v", err))
	}

	// Record Analytics event
	_ = e.Repo.RecordEvent(ctx, &storage.AnalyticsEvent{
		EventType: "post_created",
		UserID:    sender.ID,
		PostID:    &post.ID,
	})

	// Broadcast to configured target channels if enabled
	e.BroadcastPostToChannels(ctx, post)

	botUsername := e.Bot.Me.Username
	shareLink := fmt.Sprintf("https://t.me/%s?start=%s", botUsername, slug)

	card := fmt.Sprintf("✅ *Media Uploaded & Registered Successfully!*\n\n"+
		"🔗 *Sharable Link*: `%s`\n\n"+
		"📋 *File ID Details (For Admins & Authors)*:\n"+
		"• *File ID*: `%s`\n"+
		"• *File Unique ID*: `%s`\n"+
		"• *Type*: `%s`\n"+
		"• *Slug*: `%s`\n"+
		"• *Auto-Delete TTL*: `%d seconds (2 min)`\n\n"+
		"Users clicking your link will receive this content protected with auto-delete.",
		shareLink, fileID, fileUniqueID, fileType, slug, defaultTTL)

	inlineMarkup := &telebot.ReplyMarkup{}
	btnTest := inlineMarkup.URL("🚀 Test Deep Link", shareLink)
	inlineMarkup.Inline(inlineMarkup.Row(btnTest))

	return c.Send(card, inlineMarkup, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}
