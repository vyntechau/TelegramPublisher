package cleaner

import (
	"context"
	"log"
	"sync"
	"time"

	"gopkg.in/telebot.v3"
)

// DeleteTask represents a message scheduled for automatic deletion.
type DeleteTask struct {
	ChatID     int64
	MessageIDs []int
	DeleteAt   time.Time
}

// MessageDeleter defines the interface for deleting messages in Telegram.
type MessageDeleter interface {
	Delete(msg telebot.Editable) error
}

// Cleaner manages the timed auto-deletion of copyrighted or sensitive messages.
type Cleaner struct {
	bot      MessageDeleter
	tasks    []*DeleteTask
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	interval time.Duration
}

// New creates a new Cleaner instance with default check interval (2s).
func New(bot MessageDeleter) *Cleaner {
	return NewWithInterval(bot, 2*time.Second)
}

// NewWithInterval creates a Cleaner instance with a custom check interval.
func NewWithInterval(bot MessageDeleter, interval time.Duration) *Cleaner {
	ctx, cancel := context.WithCancel(context.Background())
	return &Cleaner{
		bot:      bot,
		tasks:    make([]*DeleteTask, 0),
		ctx:      ctx,
		cancel:   cancel,
		interval: interval,
	}
}

// Start initiates the background cleanup worker.
func (c *Cleaner) Start() {
	go c.run()
}

// Stop gracefully stops the cleaner.
func (c *Cleaner) Stop() {
	c.cancel()
}

// Schedule queues messages to be deleted after a specified duration.
func (c *Cleaner) Schedule(chatID int64, messageIDs []int, duration time.Duration) {
	if duration <= 0 || len(messageIDs) == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.tasks = append(c.tasks, &DeleteTask{
		ChatID:     chatID,
		MessageIDs: messageIDs,
		DeleteAt:   time.Now().Add(duration),
	})
}

// ScheduleSingle queues a single message for deletion.
func (c *Cleaner) ScheduleSingle(chatID int64, messageID int, duration time.Duration) {
	c.Schedule(chatID, []int{messageID}, duration)
}

func (c *Cleaner) run() {
	interval := c.interval
	if interval <= 0 {
		interval = 2 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.purgeExpired()
		}
	}
}

func (c *Cleaner) purgeExpired() {
	c.mu.Lock()
	now := time.Now()
	var remaining []*DeleteTask

	for _, task := range c.tasks {
		if now.After(task.DeleteAt) || now.Equal(task.DeleteAt) {
			// Process deletion
			if c.bot != nil {
				go func(t *DeleteTask) {
					for _, msgID := range t.MessageIDs {
						msg := &telebot.Message{
							ID:   msgID,
							Chat: &telebot.Chat{ID: t.ChatID},
						}
						if err := c.bot.Delete(msg); err != nil {
							// Ignore "message to delete not found" / already deleted
							log.Printf("[Cleaner] Note: could not delete msg %d in chat %d: %v", msgID, t.ChatID, err)
						}
					}
				}(task)
			}
		} else {
			remaining = append(remaining, task)
		}
	}
	c.tasks = remaining
	c.mu.Unlock()
}
