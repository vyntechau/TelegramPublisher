package analytics_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vyntechau/TelegramPublisher/internal/analytics"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"github.com/vyntechau/TelegramPublisher/internal/storage/sqlite"
)

type errorRepo struct {
	storage.Repository
}

func (e *errorRepo) ListUsers(ctx context.Context, filter storage.UserFilter) ([]*storage.User, int64, error) {
	return nil, 0, errors.New("simulated list users error")
}

func (e *errorRepo) ListPosts(ctx context.Context, authorID int64, limit, offset int) ([]*storage.Post, int64, error) {
	return nil, 0, errors.New("simulated list posts error")
}

func (e *errorRepo) GetAnalyticsSummary(ctx context.Context) (*storage.AnalyticsSummary, error) {
	return nil, errors.New("simulated analytics summary error")
}

func TestAnalyticsService(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "analytics_test_*")
	defer os.RemoveAll(tempDir)

	repo := sqlite.New(filepath.Join(tempDir, "test.db"))
	_ = repo.Init(context.Background())
	defer repo.Close()

	svc := analytics.NewService(repo)
	ctx := context.Background()

	now := time.Now()
	_ = repo.UpsertUser(ctx, &storage.User{
		TelegramID:   1001,
		Username:     "test_analyst",
		FirstName:    "Analyst",
		Role:         storage.RoleAdmin,
		Status:       storage.StatusActive,
		LastActiveAt: now,
		CreatedAt:    now,
	})

	_ = repo.CreatePost(ctx, &storage.Post{
		Slug:              "p_ana1",
		FileID:            "BAACAg...",
		FileType:          storage.FileTypeVideo,
		AuthorID:          1001,
		AutoDeleteSeconds: 120,
		ViewsCount:        10,
		LikesCount:        5,
		DislikesCount:     1,
		IsProtected:       true,
		CreatedAt:         now,
	})

	overview, err := svc.GetOverview(ctx)
	if err != nil {
		t.Fatalf("expected overview, got: %v", err)
	}
	if overview.TotalUsers != 1 || overview.TotalPosts != 1 {
		t.Errorf("unexpected overview metrics: %+v", overview)
	}

	usersCSV, err := svc.ExportUsersCSV(ctx, storage.UserFilter{})
	if err != nil || !strings.Contains(string(usersCSV), "test_analyst") {
		t.Errorf("expected users CSV containing test_analyst, got: %s", string(usersCSV))
	}

	postsCSV, err := svc.ExportPostsCSV(ctx, 0)
	if err != nil || !strings.Contains(string(postsCSV), "p_ana1") {
		t.Errorf("expected posts CSV containing p_ana1, got: %s", string(postsCSV))
	}
}

func TestAnalyticsServiceErrors(t *testing.T) {
	mockErrRepo := &errorRepo{}
	svc := analytics.NewService(mockErrRepo)
	ctx := context.Background()

	if _, err := svc.GetOverview(ctx); err == nil {
		t.Errorf("expected error from GetOverview, got nil")
	}

	if _, err := svc.ExportUsersCSV(ctx, storage.UserFilter{}); err == nil {
		t.Errorf("expected error from ExportUsersCSV, got nil")
	}

	if _, err := svc.ExportPostsCSV(ctx, 1001); err == nil {
		t.Errorf("expected error from ExportPostsCSV, got nil")
	}
}
