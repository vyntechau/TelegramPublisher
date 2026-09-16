package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vyntechau/TelegramPublisher/config"
	"github.com/vyntechau/TelegramPublisher/internal/analytics"
	"github.com/vyntechau/TelegramPublisher/internal/auth"
	"github.com/vyntechau/TelegramPublisher/internal/services/marketing"
	"github.com/vyntechau/TelegramPublisher/internal/services/payment"
	"github.com/vyntechau/TelegramPublisher/internal/services/settings"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"github.com/vyntechau/TelegramPublisher/internal/storage/sqlite"
	"gopkg.in/telebot.v3"
)

type mockSender struct{}

func (m *mockSender) Send(to telebot.Recipient, what interface{}, opts ...interface{}) (*telebot.Message, error) {
	return &telebot.Message{}, nil
}

type errReader struct{}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

type mockRepo struct {
	storage.Repository
	listPostsErr          error
	createPostErr         error
	listReportsErr        error
	createReportErr       error
	updateReportStatusErr error
	listChannelsErr       error
	addChannelErr         error
	listUsersErr          error
	getUserByIDErr        error
	setReactionErr        error
	analyticsSummaryErr   error
}

func (m *mockRepo) ListPosts(ctx context.Context, authorID int64, limit, offset int) ([]*storage.Post, int64, error) {
	if m.listPostsErr != nil {
		return nil, 0, m.listPostsErr
	}
	return m.Repository.ListPosts(ctx, authorID, limit, offset)
}

func (m *mockRepo) CreatePost(ctx context.Context, post *storage.Post) error {
	if m.createPostErr != nil {
		return m.createPostErr
	}
	return m.Repository.CreatePost(ctx, post)
}

func (m *mockRepo) ListReports(ctx context.Context, status string, postID int64, limit, offset int) ([]*storage.Report, int64, error) {
	if m.listReportsErr != nil {
		return nil, 0, m.listReportsErr
	}
	return m.Repository.ListReports(ctx, status, postID, limit, offset)
}

func (m *mockRepo) CreateReport(ctx context.Context, rep *storage.Report) error {
	if m.createReportErr != nil {
		return m.createReportErr
	}
	return m.Repository.CreateReport(ctx, rep)
}

func (m *mockRepo) UpdateReportStatus(ctx context.Context, id int64, status string, resolvedBy int64, resolutionNote string) error {
	if m.updateReportStatusErr != nil {
		return m.updateReportStatusErr
	}
	return m.Repository.UpdateReportStatus(ctx, id, status, resolvedBy, resolutionNote)
}

func (m *mockRepo) ListChannels(ctx context.Context, activeOnly bool) ([]*storage.Channel, error) {
	if m.listChannelsErr != nil {
		return nil, m.listChannelsErr
	}
	return m.Repository.ListChannels(ctx, activeOnly)
}

func (m *mockRepo) AddChannel(ctx context.Context, channel *storage.Channel) error {
	if m.addChannelErr != nil {
		return m.addChannelErr
	}
	return m.Repository.AddChannel(ctx, channel)
}

func (m *mockRepo) ListUsers(ctx context.Context, filter storage.UserFilter) ([]*storage.User, int64, error) {
	if m.listUsersErr != nil {
		return nil, 0, m.listUsersErr
	}
	return m.Repository.ListUsers(ctx, filter)
}

func (m *mockRepo) GetUserByID(ctx context.Context, id int64) (*storage.User, error) {
	if m.getUserByIDErr != nil {
		return nil, m.getUserByIDErr
	}
	return m.Repository.GetUserByID(ctx, id)
}

func (m *mockRepo) SetReaction(ctx context.Context, postID, userID int64, reaction string) error {
	if m.setReactionErr != nil {
		return m.setReactionErr
	}
	return m.Repository.SetReaction(ctx, postID, userID, reaction)
}

func (m *mockRepo) GetAnalyticsSummary(ctx context.Context) (*storage.AnalyticsSummary, error) {
	if m.analyticsSummaryErr != nil {
		return nil, m.analyticsSummaryErr
	}
	return m.Repository.GetAnalyticsSummary(ctx)
}

func setupTestServer(t *testing.T) (*Server, *mockRepo, *http.ServeMux, *config.Config, *auth.Service) {
	tempDir := t.TempDir()
	sqliteRepo := sqlite.New(filepath.Join(tempDir, "test.db"))
	_ = sqliteRepo.Init(context.Background())
	t.Cleanup(func() { sqliteRepo.Close() })

	repo := &mockRepo{Repository: sqliteRepo}
	cfg := &config.Config{
		App: config.AppConfig{JWTSecret: "test-jwt-secret-32b-length-required", Env: "development"},
		Bot: config.BotConfig{Token: "bot_tok", OwnerID: 1111},
	}

	authSvc := auth.NewService("bot_tok", cfg.App.JWTSecret, repo)
	settingsSvc := settings.NewService(repo, cfg)
	analyticsSvc := analytics.NewService(repo)
	marketSvc := marketing.NewService(&mockSender{}, repo)
	paySvc := payment.NewService(repo, settingsSvc, nil)

	server := NewServer(repo, authSvc, settingsSvc, analyticsSvc, marketSvc, paySvc, cfg)
	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	return server, repo, mux, cfg, authSvc
}

func makeToken(t *testing.T, authSvc *auth.Service, user *storage.User) string {
	tok, err := authSvc.GenerateJWT(user)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	return tok
}

func TestAuthMiddleware(t *testing.T) {
	server, repo, mux, _, authSvc := setupTestServer(t)

	// Pre-create user
	user := &storage.User{TelegramID: 999, Username: "admin", Role: storage.RoleAdmin, Status: storage.StatusActive}
	_ = repo.UpsertUser(context.Background(), user)
	dbUser, _ := repo.GetUserByTelegramID(context.Background(), 999)
	validToken := makeToken(t, authSvc, dbUser)

	t.Run("OptionsMethod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/me", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("MissingAuthHeader", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("InvalidAuthPrefix", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Basic abcdef")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("InvalidToken", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer invalid.jwt.token")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("RequireRoleWithoutClaims", func(t *testing.T) {
		handler := server.requireRole(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}, storage.RoleAdmin)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("RequireRoleForbidden", func(t *testing.T) {
		userRegular := &storage.User{ID: 2, TelegramID: 888, Username: "regular", Role: storage.RoleUser, Status: storage.StatusActive}
		regToken := makeToken(t, authSvc, userRegular)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/overview", nil)
		req.Header.Set("Authorization", "Bearer "+regToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("RequireRoleOwnerAllowed", func(t *testing.T) {
		owner := &storage.User{ID: 3, TelegramID: 777, Username: "owner", Role: storage.RoleOwner, Status: storage.StatusActive}
		ownerToken := makeToken(t, authSvc, owner)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/overview", nil)
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("RequireRoleAdminAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/overview", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestContextClaimsHelper(t *testing.T) {
	ctx := context.Background()
	if getContextClaims(ctx) != nil {
		t.Error("expected nil claims for empty context")
	}

	ctxNonClaims := context.WithValue(ctx, claimsContextKey, "invalid string value")
	if getContextClaims(ctxNonClaims) != nil {
		t.Error("expected nil claims for non-auth.Claims context value")
	}

	claims := &auth.Claims{UserID: 42, Role: storage.RoleAdmin}
	ctxWithClaims := setContextClaims(ctx, claims)
	retrieved := getContextClaims(ctxWithClaims)
	if retrieved == nil || retrieved.UserID != 42 {
		t.Errorf("expected retrieved claims with UserID 42, got %+v", retrieved)
	}
}

func TestTelegramAuth(t *testing.T) {
	server, repo, mux, cfg, _ := setupTestServer(t)

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/telegram", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rec.Code)
		}
	})

	t.Run("InvalidJSONBody", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram", strings.NewReader(`{invalid json`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("ProductionUnauthorized", func(t *testing.T) {
		cfgProd := &config.Config{
			App: config.AppConfig{JWTSecret: "test-jwt-secret-32b-length-required", Env: "production"},
			Bot: config.BotConfig{Token: "real_token", OwnerID: 1111},
		}
		prodAuthSvc := auth.NewService("real_token", cfgProd.App.JWTSecret, repo)
		serverProd := NewServer(repo, prodAuthSvc, nil, nil, nil, nil, cfgProd)
		muxProd := http.NewServeMux()
		serverProd.RegisterRoutes(muxProd)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram", strings.NewReader(`{"init_data":"invalid_data"}`))
		rec := httptest.NewRecorder()
		muxProd.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("DevFallbackNewOwnerUser", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram", strings.NewReader(`{"init_data":""}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var res map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if res["token"] == nil || res["user"] == nil {
			t.Fatalf("expected token and user in response")
		}
	})

	t.Run("DevFallbackExistingUser", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram", strings.NewReader(`{"init_data":""}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("DevFallbackNonOwnerUser", func(t *testing.T) {
		cfgNonOwner := &config.Config{
			App: config.AppConfig{JWTSecret: "test-jwt-secret-32b-length-required", Env: "development"},
			Bot: config.BotConfig{Token: "", OwnerID: 999999}, // different owner ID
		}
		authSvc := auth.NewService("", cfgNonOwner.App.JWTSecret, repo)
		serverNonOwner := NewServer(repo, authSvc, nil, nil, nil, nil, cfgNonOwner)
		muxNonOwner := http.NewServeMux()
		serverNonOwner.RegisterRoutes(muxNonOwner)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram", strings.NewReader(`{"init_data":""}`))
		rec := httptest.NewRecorder()
		muxNonOwner.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("GenerateJWTFailure", func(t *testing.T) {
		origGen := generateJWT
		defer func() { generateJWT = origGen }()
		generateJWT = func(s *auth.Service, user *storage.User) (string, error) {
			return "", errors.New("jwt failure")
		}

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram", strings.NewReader(`{"init_data":""}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	_ = cfg
	_ = server
}

func TestAuthMe(t *testing.T) {
	_, repo, mux, _, authSvc := setupTestServer(t)

	user := &storage.User{TelegramID: 1010, Username: "me_user", Role: storage.RoleUser, Status: storage.StatusActive}
	_ = repo.UpsertUser(context.Background(), user)
	dbUser, _ := repo.GetUserByTelegramID(context.Background(), 1010)
	validToken := makeToken(t, authSvc, dbUser)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("UserNotFound", func(t *testing.T) {
		nonExistentUser := &storage.User{ID: 99999, TelegramID: 99999, Role: storage.RoleUser}
		nonExistentToken := makeToken(t, authSvc, nonExistentUser)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+nonExistentToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("ClaimsNilDirect", func(t *testing.T) {
		server, _, _, _, _ := setupTestServer(t)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		rec := httptest.NewRecorder()
		server.handleAuthMe(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestAnalyticsEndpoints(t *testing.T) {
	_, repo, mux, _, authSvc := setupTestServer(t)

	admin := &storage.User{ID: 1, TelegramID: 1111, Username: "admin", Role: storage.RoleAdmin, Status: storage.StatusActive}
	_ = repo.UpsertUser(context.Background(), admin)
	adminToken := makeToken(t, authSvc, admin)

	t.Run("OverviewSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/overview", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("OverviewError", func(t *testing.T) {
		repo.analyticsSummaryErr = errors.New("analytics summary error")
		defer func() { repo.analyticsSummaryErr = nil }()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/overview", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("ExportUsersCSVSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/export/users.csv", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Header().Get("Content-Type") != "text/csv" {
			t.Errorf("expected text/csv, got %s", rec.Header().Get("Content-Type"))
		}
	})

	t.Run("ExportUsersCSVError", func(t *testing.T) {
		repo.listUsersErr = errors.New("list users error")
		defer func() { repo.listUsersErr = nil }()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/export/users.csv", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("ExportPostsCSVSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/export/posts.csv", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Header().Get("Content-Type") != "text/csv" {
			t.Errorf("expected text/csv, got %s", rec.Header().Get("Content-Type"))
		}
	})

	t.Run("ExportPostsCSVError", func(t *testing.T) {
		repo.listPostsErr = errors.New("list posts error")
		defer func() { repo.listPostsErr = nil }()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/export/posts.csv", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})
}

func TestPostsEndpoints(t *testing.T) {
	_, repo, mux, _, authSvc := setupTestServer(t)

	author := &storage.User{ID: 2, TelegramID: 2222, Username: "author", Role: storage.RoleAuthor, Status: storage.StatusActive}
	user := &storage.User{ID: 3, TelegramID: 3333, Username: "user", Role: storage.RoleUser, Status: storage.StatusActive}
	admin := &storage.User{ID: 1, TelegramID: 1111, Username: "admin", Role: storage.RoleAdmin, Status: storage.StatusActive}
	_ = repo.UpsertUser(context.Background(), author)
	_ = repo.UpsertUser(context.Background(), user)
	_ = repo.UpsertUser(context.Background(), admin)

	authorToken := makeToken(t, authSvc, author)
	userToken := makeToken(t, authSvc, user)
	adminToken := makeToken(t, authSvc, admin)

	// ListPosts GET
	t.Run("ListPostsSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts?limit=10&offset=0", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("ListPostsError", func(t *testing.T) {
		repo.listPostsErr = errors.New("list posts db error")
		defer func() { repo.listPostsErr = nil }()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	// CreatePost POST
	t.Run("CreatePostForbiddenNoAuth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(`{}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("CreatePostForbiddenRoleUser", func(t *testing.T) {
		claims := &auth.Claims{UserID: 3, TelegramID: 3333, Role: storage.RoleUser}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(`{}`))
		req = req.WithContext(setContextClaims(req.Context(), claims))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("CreatePostInvalidJSON", func(t *testing.T) {
		claims := &auth.Claims{UserID: 2, TelegramID: 2222, Role: storage.RoleAuthor}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(`{bad json`))
		req = req.WithContext(setContextClaims(req.Context(), claims))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("CreatePostRepoError", func(t *testing.T) {
		repo.createPostErr = errors.New("create post error")
		defer func() { repo.createPostErr = nil }()

		claims := &auth.Claims{UserID: 2, TelegramID: 2222, Role: storage.RoleAuthor}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(`{"slug":"err_post","file_id":"fid"}`))
		req = req.WithContext(setContextClaims(req.Context(), claims))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("CreatePostSuccess", func(t *testing.T) {
		claims := &auth.Claims{UserID: 2, TelegramID: 2222, Role: storage.RoleAuthor}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(`{"slug":"my_post","file_id":"fid123"}`))
		req = req.WithContext(setContextClaims(req.Context(), claims))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("PostsMethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/posts", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rec.Code)
		}
	})

	// PostByIDOrSlug GET
	t.Run("GetPostBySlugSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/my_post", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("GetPostByIDSuccess", func(t *testing.T) {
		// Fetch existing post by numeric ID
		post, _ := repo.GetPostBySlug(context.Background(), "my_post")
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/posts/%d", post.ID), nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("GetPostNotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/non_existent_slug", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	// PostByIDOrSlug DELETE
	t.Run("DeletePostForbiddenNoAuth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/posts/1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("DeletePostForbiddenRoleUser", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/posts/1", nil)
		claims := &auth.Claims{Role: storage.RoleUser}
		req = req.WithContext(setContextClaims(req.Context(), claims))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("DeletePostSuccessAdmin", func(t *testing.T) {
		post, _ := repo.GetPostBySlug(context.Background(), "my_post")
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/posts/%d", post.ID), nil)
		claims := &auth.Claims{Role: storage.RoleAdmin}
		req = req.WithContext(setContextClaims(req.Context(), claims))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("PostByIDMethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/posts/1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rec.Code)
		}
	})

	_ = authorToken
	_ = userToken
	_ = adminToken
}

func TestPostReact(t *testing.T) {
	_, repo, mux, _, authSvc := setupTestServer(t)

	user := &storage.User{ID: 1, TelegramID: 100, Username: "u", Role: storage.RoleUser, Status: storage.StatusActive}
	_ = repo.UpsertUser(context.Background(), user)
	uToken := makeToken(t, authSvc, user)

	_ = repo.CreatePost(context.Background(), &storage.Post{
		ID:       1,
		Slug:     "react_slug",
		FileID:   "BAAC...",
		FileType: "video",
		AuthorID: 100,
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/react", nil)
		req.Header.Set("Authorization", "Bearer "+uToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rec.Code)
		}
	})

	t.Run("ClaimsNilDirect", func(t *testing.T) {
		server, _, _, _, _ := setupTestServer(t)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/react", strings.NewReader(`{}`))
		rec := httptest.NewRecorder()
		server.handlePostReact(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/react", strings.NewReader(`{bad json`))
		req.Header.Set("Authorization", "Bearer "+uToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("RepoError", func(t *testing.T) {
		repo.setReactionErr = errors.New("failed set reaction")
		defer func() { repo.setReactionErr = nil }()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/react", strings.NewReader(`{"post_id":1,"reaction":"like"}`))
		req.Header.Set("Authorization", "Bearer "+uToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/react", strings.NewReader(`{"post_id":1,"reaction":"like"}`))
		req.Header.Set("Authorization", "Bearer "+uToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestReportsEndpoints(t *testing.T) {
	_, repo, mux, _, authSvc := setupTestServer(t)

	author := &storage.User{ID: 1, TelegramID: 100, Username: "author", Role: storage.RoleAuthor, Status: storage.StatusActive}
	_ = repo.UpsertUser(context.Background(), author)
	authorToken := makeToken(t, authSvc, author)

	t.Run("ListReportsSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports?status=pending&limit=10&offset=0", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("ListReportsError", func(t *testing.T) {
		repo.listReportsErr = errors.New("list reports error")
		defer func() { repo.listReportsErr = nil }()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("CreateReportInvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/reports", strings.NewReader(`{bad json`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("CreateReportRepoError", func(t *testing.T) {
		repo.createReportErr = errors.New("create report error")
		defer func() { repo.createReportErr = nil }()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/reports", strings.NewReader(`{"post_id":1,"reason":"spam"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("CreateReportSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/reports", strings.NewReader(`{"post_id":1,"reason":"spam","details":"inappropriate"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("ReportsMethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/reports", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rec.Code)
		}
	})

	// ReportUpdate
	t.Run("ReportUpdateInvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/reports/1", strings.NewReader(`{bad json`))
		req.Header.Set("Authorization", "Bearer "+authorToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("ReportUpdateRepoError", func(t *testing.T) {
		repo.updateReportStatusErr = errors.New("update report error")
		defer func() { repo.updateReportStatusErr = nil }()

		req := httptest.NewRequest(http.MethodPut, "/api/v1/reports/1", strings.NewReader(`{"status":"resolved","resolution_note":"handled"}`))
		req.Header.Set("Authorization", "Bearer "+authorToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("ReportUpdateSuccessWithClaims", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/reports/1", strings.NewReader(`{"status":"resolved","resolution_note":"handled"}`))
		req.Header.Set("Authorization", "Bearer "+authorToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("ReportUpdateWithoutClaims", func(t *testing.T) {
		server, _, _, _, _ := setupTestServer(t)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/reports/1", strings.NewReader(`{"status":"resolved"}`))
		rec := httptest.NewRecorder()
		server.handleReportUpdate(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestMarketingBroadcast(t *testing.T) {
	_, repo, mux, _, authSvc := setupTestServer(t)

	admin := &storage.User{ID: 1, TelegramID: 1111, Username: "admin", Role: storage.RoleAdmin, Status: storage.StatusActive}
	_ = repo.UpsertUser(context.Background(), admin)
	adminToken := makeToken(t, authSvc, admin)

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/broadcast", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rec.Code)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/broadcast", strings.NewReader(`{bad json`))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("DispatchError", func(t *testing.T) {
		repo.listUsersErr = errors.New("list users failed for dispatch")
		defer func() { repo.listUsersErr = nil }()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/broadcast", strings.NewReader(`{"content":"hello"}`))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("DispatchSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/broadcast", strings.NewReader(`{"content":"hello marketing"}`))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestChannelsEndpoints(t *testing.T) {
	_, repo, mux, _, authSvc := setupTestServer(t)

	admin := &storage.User{ID: 1, TelegramID: 1111, Username: "admin", Role: storage.RoleAdmin, Status: storage.StatusActive}
	_ = repo.UpsertUser(context.Background(), admin)
	adminToken := makeToken(t, authSvc, admin)

	t.Run("ListChannelsSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/channels", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("ListChannelsError", func(t *testing.T) {
		repo.listChannelsErr = errors.New("list channels db error")
		defer func() { repo.listChannelsErr = nil }()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/channels", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("AddChannelInvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/channels", strings.NewReader(`{bad json`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("AddChannelRepoError", func(t *testing.T) {
		repo.addChannelErr = errors.New("add channel db error")
		defer func() { repo.addChannelErr = nil }()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/channels", strings.NewReader(`{"title":"My Channel"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("AddChannelSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/channels", strings.NewReader(`{"title":"My Channel","channel_id":-10012345}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("ChannelsMethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/channels", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rec.Code)
		}
	})

	// ChannelByID
	t.Run("ChannelByIDDeleteSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/channels/123", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("ChannelByIDMethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/123", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rec.Code)
		}
	})
}

func TestUsersEndpoints(t *testing.T) {
	_, repo, mux, _, authSvc := setupTestServer(t)

	admin := &storage.User{ID: 1, TelegramID: 1111, Username: "admin", Role: storage.RoleAdmin, Status: storage.StatusActive}
	_ = repo.UpsertUser(context.Background(), admin)
	adminToken := makeToken(t, authSvc, admin)

	t.Run("ListUsersSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users?limit=10&offset=0&role=user&status=active", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("ListUsersError", func(t *testing.T) {
		repo.listUsersErr = errors.New("list users error")
		defer func() { repo.listUsersErr = nil }()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	// UserByID Update
	t.Run("UserByIDInvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/1111", strings.NewReader(`{bad json`))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("UserByIDSuccessWithFields", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/1111", strings.NewReader(`{"role":"admin","status":"active"}`))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("UserByIDSuccessWithoutFields", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/1111", strings.NewReader(`{}`))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestSettingsEndpoints(t *testing.T) {
	_, repo, mux, _, authSvc := setupTestServer(t)

	admin := &storage.User{ID: 1, TelegramID: 1111, Username: "admin", Role: storage.RoleAdmin, Status: storage.StatusActive}
	_ = repo.UpsertUser(context.Background(), admin)
	adminToken := makeToken(t, authSvc, admin)

	t.Run("GetSettingsSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("GetSettingsError", func(t *testing.T) {
		origGetAll := getAllSettings
		defer func() { getAllSettings = origGetAll }()
		getAllSettings = func(s *settings.Service, ctx context.Context) (map[string]string, error) {
			return nil, errors.New("get all settings error")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("PostSettingsInvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", strings.NewReader(`{bad json`))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("PostSettingsSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", strings.NewReader(`{"key1":"val1","key2":"val2"}`))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("SettingsMethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/settings", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rec.Code)
		}
	})
}

func TestPaymentsEndpoints(t *testing.T) {
	_, repo, mux, _, authSvc := setupTestServer(t)

	user := &storage.User{ID: 1, TelegramID: 12345, Username: "buyer", Role: storage.RoleUser, Status: storage.StatusActive}
	_ = repo.UpsertUser(context.Background(), user)
	userToken := makeToken(t, authSvc, user)

	t.Run("PlansSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/plans", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("CheckoutClaimsNilDirect", func(t *testing.T) {
		server, _, _, _, _ := setupTestServer(t)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/checkout", strings.NewReader(`{}`))
		rec := httptest.NewRecorder()
		server.handlePaymentCheckout(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("CheckoutInvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/checkout", strings.NewReader(`{bad json`))
		req.Header.Set("Authorization", "Bearer "+userToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("CheckoutUnsupportedPlanError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/checkout", strings.NewReader(`{"plan_id":"invalid_plan","currency":"USDT"}`))
		req.Header.Set("Authorization", "Bearer "+userToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("CheckoutSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/checkout", strings.NewReader(`{"plan_id":"vip_30d","currency":"USDT"}`))
		req.Header.Set("Authorization", "Bearer "+userToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("WebhookMethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/webhook", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rec.Code)
		}
	})

	t.Run("WebhookBodyReadError", func(t *testing.T) {
		server, _, _, _, _ := setupTestServer(t)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", errReader{})
		rec := httptest.NewRecorder()
		server.handleAZPaysWebhook(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("WebhookVerifyErrorHeader1", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", strings.NewReader(`{"invoice_id":"inv_1"}`))
		req.Header.Set("X-Azpays-Signature", "invalid_sig")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("WebhookVerifyErrorHeader2", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", strings.NewReader(`{"invoice_id":"inv_1"}`))
		req.Header.Set("Azpays-Signature", "invalid_sig")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("WebhookVerifySuccessInDevMode", func(t *testing.T) {
		// In dev mode when azpays secret is empty, signature is verified or mock is accepted
		// Let's create a subscription first to test full webhook success
		sub := &storage.Subscription{
			UserID:          12345,
			PlanID:          "vip_30d",
			Tier:            "vip",
			Status:          storage.SubStatusPending,
			AZPaysInvoiceID: "inv_test_ok",
			AmountCrypto:    9.99,
			Currency:        "USDT",
			CreatedAt:       time.Now(),
		}
		_ = repo.CreateSubscription(context.Background(), sub)

		body, _ := json.Marshal(payment.WebhookPayload{
			InvoiceID: "inv_test_ok",
			Status:    "paid",
			Amount:    9.99,
			Currency:  "USDT",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/azpays/webhook", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("MockCheckout", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/mock-checkout?invoice_id=inv_123", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "inv_123") {
			t.Errorf("expected html to contain inv_123, got: %s", rec.Body.String())
		}
	})
}

func TestSetupStatus(t *testing.T) {
	_, _, mux, _, _ := setupTestServer(t)

	t.Run("OptionsMethod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/v1/setup/status", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/status", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rec.Code)
		}
	})

	t.Run("GetSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		var res map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if _, ok := res["completed"]; !ok {
			t.Errorf("expected completed key in json response")
		}
	})
}
