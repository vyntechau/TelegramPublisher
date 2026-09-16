package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	gql "github.com/graphql-go/graphql"
	"github.com/vyntechau/TelegramPublisher/internal/analytics"
	"github.com/vyntechau/TelegramPublisher/internal/services/marketing"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"github.com/vyntechau/TelegramPublisher/internal/storage/sqlite"
)

type failReactionRepo struct {
	storage.Repository
}

func (f *failReactionRepo) SetReaction(ctx context.Context, postID, userID int64, reaction string) error {
	return errors.New("reaction failed")
}

func TestGraphQL(t *testing.T) {
	tempDir := t.TempDir()
	repo := sqlite.New(filepath.Join(tempDir, "test.db"))
	_ = repo.Init(context.Background())
	defer repo.Close()

	// Seed data
	_ = repo.UpsertUser(context.Background(), &storage.User{
		TelegramID: 100,
		Username:   "testuser",
		FirstName:  "Tester",
		Role:       storage.RoleUser,
		Status:     storage.StatusActive,
	})

	_ = repo.CreatePost(context.Background(), &storage.Post{
		Slug:              "gql_post",
		FileID:            "BAAC...",
		FileType:          "video",
		AuthorID:          100,
		AutoDeleteSeconds: 60,
	})

	analyticsSvc := analytics.NewService(repo)
	marketSvc := marketing.NewService(nil, repo)

	mux := http.NewServeMux()
	err := Register(mux, repo, analyticsSvc, marketSvc)
	if err != nil {
		t.Fatalf("unexpected error registering graphql: %v", err)
	}

	execQuery := func(t *testing.T, query string, variables map[string]interface{}) *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]interface{}{
			"query":     query,
			"variables": variables,
		})
		req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	// 1. Query analytics
	t.Run("QueryAnalytics", func(t *testing.T) {
		rec := execQuery(t, `{ analytics { total_users total_posts } }`, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 2. Query posts
	t.Run("QueryPosts", func(t *testing.T) {
		rec := execQuery(t, `{ posts(limit: 10, offset: 0) { id slug } }`, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 3. Query post by slug
	t.Run("QueryPostBySlug", func(t *testing.T) {
		rec := execQuery(t, `{ post(slug: "gql_post") { id slug } }`, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 4. Query reports
	t.Run("QueryReports", func(t *testing.T) {
		rec := execQuery(t, `{ reports { id reason status } }`, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 5. Query users
	t.Run("QueryUsers", func(t *testing.T) {
		rec := execQuery(t, `{ users { id telegram_id username } }`, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 6. Mutation reactPost (success)
	t.Run("MutationReactPostSuccess", func(t *testing.T) {
		rec := execQuery(t, `mutation { reactPost(post_id: 1, user_id: 100, reaction: "like") { id likes_count } }`, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 7. Mutation createReport (with details)
	t.Run("MutationCreateReportWithDetails", func(t *testing.T) {
		rec := execQuery(t, `mutation { createReport(post_id: 1, reported_by: 100, reason: "spam", details: "abusive content") { id reason details } }`, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 8. Mutation createReport (without details)
	t.Run("MutationCreateReportWithoutDetails", func(t *testing.T) {
		rec := execQuery(t, `mutation { createReport(post_id: 1, reported_by: 100, reason: "inappropriate") { id reason details } }`, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 9. Mutation reactPost (error in SetReaction)
	t.Run("MutationReactPostError", func(t *testing.T) {
		fRepo := &failReactionRepo{Repository: repo}
		fMux := http.NewServeMux()
		_ = Register(fMux, fRepo, analyticsSvc, marketSvc)

		body, _ := json.Marshal(map[string]interface{}{
			"query": `mutation { reactPost(post_id: 1, user_id: 100, reaction: "like") { id } }`,
		})
		req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		fMux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if _, ok := resp["errors"]; !ok {
			t.Errorf("expected errors in response, got: %s", rec.Body.String())
		}
	})

	// 10. Schema creation error branch
	t.Run("SchemaCreationError", func(t *testing.T) {
		oldNewSchema := newSchema
		defer func() { newSchema = oldNewSchema }()
		newSchema = func(config gql.SchemaConfig) (gql.Schema, error) {
			return gql.Schema{}, errors.New("simulated schema failure")
		}

		err := Register(http.NewServeMux(), repo, analyticsSvc, marketSvc)
		if err == nil {
			t.Fatal("expected schema creation error, got nil")
		}
	})
}
