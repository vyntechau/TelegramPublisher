package marketing_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/vyntechau/TelegramPublisher/internal/services/marketing"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"github.com/vyntechau/TelegramPublisher/internal/storage/sqlite"
	"gopkg.in/telebot.v3"
)

type mockSender struct {
	mu          sync.Mutex
	sendFunc    func(to telebot.Recipient, what interface{}, opts ...interface{}) (*telebot.Message, error)
	sentCount   int
	sentWhat    []interface{}
	sentOptions []*telebot.SendOptions
}

func (m *mockSender) Send(to telebot.Recipient, what interface{}, opts ...interface{}) (*telebot.Message, error) {
	m.mu.Lock()
	m.sentCount++
	m.sentWhat = append(m.sentWhat, what)
	for _, opt := range opts {
		if so, ok := opt.(*telebot.SendOptions); ok {
			m.sentOptions = append(m.sentOptions, so)
		}
	}
	fn := m.sendFunc
	m.mu.Unlock()

	if fn != nil {
		return fn(to, what, opts...)
	}
	return &telebot.Message{}, nil
}

type errorListUserRepo struct {
	storage.Repository
}

func (e *errorListUserRepo) ListUsers(ctx context.Context, filter storage.UserFilter) ([]*storage.User, int64, error) {
	return nil, 0, errors.New("simulated list users error")
}

func TestMarketingDispatchEmptyAndErrors(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "mkt_test_*")
	defer os.RemoveAll(tempDir)

	repo := sqlite.New(filepath.Join(tempDir, "test.db"))
	_ = repo.Init(context.Background())
	defer repo.Close()

	mock := &mockSender{}
	svc := marketing.NewService(mock, repo)
	ctx := context.Background()

	// 1. Empty list dispatch
	res, err := svc.Dispatch(ctx, marketing.BroadcastPayload{
		Format:  "html",
		Content: "<b>Test</b>",
		Filter:  storage.UserFilter{Role: "admin"},
	})
	if err != nil {
		t.Fatalf("expected dispatch with no matching users to succeed, got: %v", err)
	}
	if res.TotalTargeted != 0 {
		t.Errorf("expected 0 targeted, got: %d", res.TotalTargeted)
	}

	// 2. Repo ListUsers error
	errSvc := marketing.NewService(mock, &errorListUserRepo{})
	if _, err := errSvc.Dispatch(ctx, marketing.BroadcastPayload{}); err == nil {
		t.Errorf("expected error from Dispatch when ListUsers fails, got nil")
	}
}

func TestMarketingDispatchSuccessAndBlockDetection(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "mkt_test_*")
	defer os.RemoveAll(tempDir)

	repo := sqlite.New(filepath.Join(tempDir, "test.db"))
	_ = repo.Init(context.Background())
	defer repo.Close()

	ctx := context.Background()
	_ = repo.UpsertUser(ctx, &storage.User{TelegramID: 101, Status: storage.StatusActive})
	_ = repo.UpsertUser(ctx, &storage.User{TelegramID: 102, Status: storage.StatusActive})
	_ = repo.UpsertUser(ctx, &storage.User{TelegramID: 103, Status: storage.StatusActive})
	_ = repo.UpsertUser(ctx, &storage.User{TelegramID: 104, Status: storage.StatusActive})

	mock := &mockSender{
		sendFunc: func(to telebot.Recipient, what interface{}, opts ...interface{}) (*telebot.Message, error) {
			id := to.Recipient()
			switch id {
			case "101":
				return &telebot.Message{}, nil
			case "102":
				return nil, errors.New("Forbidden: bot was blocked by the user")
			case "103":
				return nil, errors.New("Bad Request: user is deactivated")
			case "104":
				return nil, errors.New("Telegram network failure 500")
			}
			return &telebot.Message{}, nil
		},
	}

	svc := marketing.NewService(mock, repo)

	res, err := svc.Dispatch(ctx, marketing.BroadcastPayload{
		Format:            "markdown",
		Content:           "*Marketing Campaign*",
		DisableWebPreview: true,
	})
	if err != nil {
		t.Fatalf("expected dispatch to complete, got err: %v", err)
	}

	if res.TotalTargeted != 4 {
		t.Errorf("expected 4 targeted, got %d", res.TotalTargeted)
	}
	if res.TotalSent != 1 {
		t.Errorf("expected 1 sent, got %d", res.TotalSent)
	}
	if res.TotalBlocked != 2 {
		t.Errorf("expected 2 blocked, got %d", res.TotalBlocked)
	}
	if res.TotalFailed != 1 {
		t.Errorf("expected 1 failed, got %d", res.TotalFailed)
	}

	// Also test HTML and MarkdownV2 formats
	_, _ = svc.Dispatch(ctx, marketing.BroadcastPayload{
		Format:  "html",
		Content: "<b>HTML</b>",
	})
	_, _ = svc.Dispatch(ctx, marketing.BroadcastPayload{
		Format:  "markdownv2",
		Content: "MarkdownV2",
	})
}

func TestMarketingDispatchRawJSON(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "mkt_test_*")
	defer os.RemoveAll(tempDir)

	repo := sqlite.New(filepath.Join(tempDir, "test.db"))
	_ = repo.Init(context.Background())
	defer repo.Close()

	ctx := context.Background()
	_ = repo.UpsertUser(ctx, &storage.User{TelegramID: 201, Status: storage.StatusActive})

	mock := &mockSender{}
	svc := marketing.NewService(mock, repo)

	// 1. Photo with inline buttons (URL, Callback Data, WebApp)
	rawJSONPhoto := `{
		"text": "Check this photo",
		"photo": "AgACAgIAAxkBAAI...",
		"parse_mode": "HTML",
		"inline_keyboard": [
			[
				{"text": "Visit Website", "url": "https://vyntech.cloud"},
				{"text": "Click Me", "callback_data": "btn_clicked"},
				{"text": "Open App", "web_app_url": "https://app.vyntech.cloud"}
			]
		]
	}`
	res, err := svc.Dispatch(ctx, marketing.BroadcastPayload{
		Format:  "raw_json",
		RawJSON: rawJSONPhoto,
	})
	if err != nil || res.TotalSent != 1 {
		t.Errorf("expected raw_json photo dispatch success, got err=%v res=%+v", err, res)
	}

	// 2. Video with Markdown parse mode
	rawJSONVideo := `{
		"text": "Check this video",
		"video": "BAACAgIAAxkBAAI...",
		"parse_mode": "markdown"
	}`
	res, err = svc.Dispatch(ctx, marketing.BroadcastPayload{
		Format:  "raw_json",
		RawJSON: rawJSONVideo,
	})
	if err != nil || res.TotalSent != 1 {
		t.Errorf("expected raw_json video dispatch success, got err=%v", err)
	}

	// 3. Plain Text raw_json
	rawJSONText := `{
		"text": "Just plain text",
		"parse_mode": "html"
	}`
	res, err = svc.Dispatch(ctx, marketing.BroadcastPayload{
		Format:  "raw_json",
		RawJSON: rawJSONText,
	})
	if err != nil || res.TotalSent != 1 {
		t.Errorf("expected raw_json text dispatch success, got err=%v", err)
	}

	// 4. Invalid raw_json syntax
	res, err = svc.Dispatch(ctx, marketing.BroadcastPayload{
		Format:  "raw_json",
		RawJSON: "{invalid-json-content",
	})
	if err != nil || res.TotalFailed != 1 {
		t.Errorf("expected raw_json syntax error handled as failed message, got err=%v res=%+v", err, res)
	}
}

func TestMarketingDispatchContextCancellation(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "mkt_test_*")
	defer os.RemoveAll(tempDir)

	repo := sqlite.New(filepath.Join(tempDir, "test.db"))
	_ = repo.Init(context.Background())
	defer repo.Close()

	ctx := context.Background()
	for i := 1; i <= 5; i++ {
		_ = repo.UpsertUser(ctx, &storage.User{TelegramID: int64(300 + i), Status: storage.StatusActive})
	}

	cancelCtx, cancel := context.WithCancel(ctx)

	mock := &mockSender{
		sendFunc: func(to telebot.Recipient, what interface{}, opts ...interface{}) (*telebot.Message, error) {
			cancel() // cancel context on first message delivered
			return &telebot.Message{}, nil
		},
	}
	svc := marketing.NewService(mock, repo)

	res, err := svc.Dispatch(cancelCtx, marketing.BroadcastPayload{
		Format:  "text",
		Content: "Canceled broadcast",
	})
	if err != nil {
		t.Errorf("expected dispatch with canceled context to complete cleanly, got: %v", err)
	}
	if res.TotalSent > 3 {
		t.Errorf("expected early termination after context cancellation, sent: %d", res.TotalSent)
	}
}
