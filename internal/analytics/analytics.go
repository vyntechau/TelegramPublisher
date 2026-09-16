package analytics

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/vyntechau/TelegramPublisher/internal/storage"
)

// Service provides metrics calculation and reporting queries.
type Service struct {
	repo storage.Repository
}

func NewService(repo storage.Repository) *Service {
	return &Service{repo: repo}
}

// GetOverview returns real-time KPI overview metrics.
func (s *Service) GetOverview(ctx context.Context) (*storage.AnalyticsSummary, error) {
	return s.repo.GetAnalyticsSummary(ctx)
}

// ExportUsersCSV generates CSV formatted export of registered users.
func (s *Service) ExportUsersCSV(ctx context.Context, filter storage.UserFilter) ([]byte, error) {
	filter.Limit = 50000
	users, _, err := s.repo.ListUsers(ctx, filter)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	// Header
	_ = writer.Write([]string{"ID", "TelegramID", "Username", "FirstName", "Role", "Status", "LastActiveAt", "CreatedAt"})

	for _, u := range users {
		_ = writer.Write([]string{
			fmt.Sprintf("%d", u.ID),
			fmt.Sprintf("%d", u.TelegramID),
			u.Username,
			u.FirstName,
			u.Role,
			u.Status,
			u.LastActiveAt.Format(time.RFC3339),
			u.CreatedAt.Format(time.RFC3339),
		})
	}
	writer.Flush()

	return buf.Bytes(), nil
}

// ExportPostsCSV generates CSV export of all posts with engagement metrics.
func (s *Service) ExportPostsCSV(ctx context.Context, authorID int64) ([]byte, error) {
	posts, _, err := s.repo.ListPosts(ctx, authorID, 50000, 0)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	_ = writer.Write([]string{"ID", "Slug", "FileType", "AuthorID", "AutoDeleteSec", "Views", "Likes", "Dislikes", "Protected", "CreatedAt"})

	for _, p := range posts {
		_ = writer.Write([]string{
			fmt.Sprintf("%d", p.ID),
			p.Slug,
			p.FileType,
			fmt.Sprintf("%d", p.AuthorID),
			fmt.Sprintf("%d", p.AutoDeleteSeconds),
			fmt.Sprintf("%d", p.ViewsCount),
			fmt.Sprintf("%d", p.LikesCount),
			fmt.Sprintf("%d", p.DislikesCount),
			fmt.Sprintf("%t", p.IsProtected),
			p.CreatedAt.Format(time.RFC3339),
		})
	}
	writer.Flush()

	return buf.Bytes(), nil
}
