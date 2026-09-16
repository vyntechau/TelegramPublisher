package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/vyntechau/TelegramPublisher/internal/storage"
)

type nthCancelContext struct {
	context.Context
	mu        sync.Mutex
	count     int
	threshold int
	doneCh    chan struct{}
}

func newNthCancelContext(threshold int) *nthCancelContext {
	return &nthCancelContext{
		Context:   context.Background(),
		threshold: threshold,
		doneCh:    make(chan struct{}),
	}
}

func (c *nthCancelContext) Done() <-chan struct{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
	if c.count >= c.threshold {
		select {
		case <-c.doneCh:
		default:
			close(c.doneCh)
		}
		return c.doneCh
	}
	return nil
}

func (c *nthCancelContext) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	select {
	case <-c.doneCh:
		return context.Canceled
	default:
		return nil
	}
}

func TestInitSQLOpenError(t *testing.T) {
	origSQLOpen := sqlOpen
	defer func() { sqlOpen = origSQLOpen }()

	sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
		return nil, errors.New("simulated open error")
	}

	tempDir := t.TempDir()
	r := New(filepath.Join(tempDir, "test.db"))
	if err := r.Init(context.Background()); err == nil {
		t.Errorf("expected error from Init when sqlOpen fails, got nil")
	}
}

func TestGetActiveSubscriptionFound(t *testing.T) {
	tempDir := t.TempDir()
	r := New(filepath.Join(tempDir, "test.db"))
	ctx := context.Background()
	if err := r.Init(ctx); err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer r.Close()

	sub := &storage.Subscription{
		UserID:          2002,
		PlanID:          "vip_monthly",
		Tier:            "vip",
		Status:          storage.SubStatusActive,
		AZPaysInvoiceID: "inv_2002",
		AmountCrypto:    15.0,
		Currency:        "USDT",
		ExpiresAt:       time.Now().Add(10 * 24 * time.Hour),
	}
	if err := r.CreateSubscription(ctx, sub); err != nil {
		t.Fatalf("failed to create sub: %v", err)
	}

	active, err := r.GetActiveSubscription(ctx, 2002)
	if err != nil || active == nil {
		t.Fatalf("expected active subscription, got err=%v", err)
	}
	if active.Tier != "vip" {
		t.Errorf("expected tier vip, got %s", active.Tier)
	}
}

func TestScanErrors(t *testing.T) {
	tempDir := t.TempDir()
	r := New(filepath.Join(tempDir, "scan_err.db"))
	ctx := context.Background()
	if err := r.Init(ctx); err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer r.Close()

	// Corrupt column types by inserting incompatible string into integer columns
	_, _ = r.db.ExecContext(ctx, `INSERT INTO users (telegram_id, username) VALUES ('invalid_int_tg', 'bad')`)
	if _, _, err := r.ListUsers(ctx, storage.UserFilter{}); err == nil {
		t.Errorf("expected scan error in ListUsers for corrupt data, got nil")
	}

	_, _ = r.db.ExecContext(ctx, `INSERT INTO posts (slug, file_id, author_id) VALUES ('bad_p', 'f', 'invalid_int_author')`)
	if _, _, err := r.ListPosts(ctx, 0, 10, 0); err == nil {
		t.Errorf("expected scan error in ListPosts for corrupt data, got nil")
	}

	_, _ = r.db.ExecContext(ctx, `INSERT INTO posts (id, slug, file_id, author_id) VALUES (999, 'rep_post', 'f', 1)`)
	_, _ = r.db.ExecContext(ctx, `INSERT INTO reports (post_id, reported_by, reason) VALUES (999, 'bad_int_reporter', 'reason')`)
	if _, _, err := r.ListReports(ctx, "", 0, 10, 0); err == nil {
		t.Errorf("expected scan error in ListReports for corrupt data, got nil")
	}

	_, _ = r.db.ExecContext(ctx, `INSERT INTO channels (telegram_id, invite_link) VALUES ('bad_int_channel_tg', 'link')`)
	if _, err := r.ListChannels(ctx, false); err == nil {
		t.Errorf("expected scan error in ListChannels for corrupt data, got nil")
	}

	_, _ = r.db.ExecContext(ctx, `INSERT INTO subscriptions (user_id, plan_id, azpays_invoice_id, amount_crypto, expires_at) VALUES ('bad_user_int', 'p', 'inv_bad', 'not_a_float', '2026-01-01')`)
	if _, _, err := r.ListSubscriptions(ctx, 10, 0); err == nil {
		t.Errorf("expected scan error in ListSubscriptions for corrupt data, got nil")
	}

	// Corrupt settings table with NULL value to trigger ListSettings scan error
	_, _ = r.db.ExecContext(ctx, `DROP TABLE settings`)
	_, _ = r.db.ExecContext(ctx, `CREATE TABLE settings (key TEXT, value TEXT)`)
	_, _ = r.db.ExecContext(ctx, `INSERT INTO settings (key, value) VALUES ('null_key', NULL)`)
	if _, err := r.ListSettings(ctx); err == nil {
		t.Errorf("expected scan error in ListSettings for NULL value, got nil")
	}
}

func TestSetReactionInsertError(t *testing.T) {
	tempDir := t.TempDir()
	r := New(filepath.Join(tempDir, "react_err.db"))
	ctx := context.Background()
	if err := r.Init(ctx); err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer r.Close()

	// Replace likes table with CHECK constraint that rejects all reactions
	_, _ = r.db.ExecContext(ctx, `DROP TABLE likes`)
	_, _ = r.db.ExecContext(ctx, `CREATE TABLE likes (id INTEGER PRIMARY KEY, post_id INTEGER, user_id INTEGER, reaction_type TEXT CHECK(reaction_type = 'forbidden'), created_at DATETIME)`)

	if err := r.SetReaction(ctx, 10, 10, storage.ReactionLike); err == nil {
		t.Errorf("expected insert error in SetReaction when table constraint fails, got nil")
	}
}

func TestListQueryContextFailurePaths(t *testing.T) {
	tempDir := t.TempDir()
	r := New(filepath.Join(tempDir, "query_err.db"))
	ctx := context.Background()
	if err := r.Init(ctx); err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer r.Close()

	for threshold := 1; threshold <= 10; threshold++ {
		c1 := newNthCancelContext(threshold)
		_, _, _ = r.ListUsers(c1, storage.UserFilter{})

		c2 := newNthCancelContext(threshold)
		_, _, _ = r.ListPosts(c2, 0, 10, 0)

		c3 := newNthCancelContext(threshold)
		_, _, _ = r.ListReports(c3, "", 0, 10, 0)

		c4 := newNthCancelContext(threshold)
		_, _, _ = r.ListSubscriptions(c4, 10, 0)
	}
}
