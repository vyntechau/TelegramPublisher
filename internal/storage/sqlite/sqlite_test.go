package sqlite_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"github.com/vyntechau/TelegramPublisher/internal/storage/sqlite"
)

func TestSQLiteRepositoryFull(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tg_pub_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "subfolder", "test.db")
	repo := sqlite.New(dbPath)
	ctx := context.Background()

	if err := repo.Init(ctx); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer repo.Close()

	// --- 1. User Management ---
	now := time.Now()
	user := &storage.User{
		TelegramID:   1001,
		Username:     "testuser",
		FirstName:    "Tester",
		Role:         storage.RoleUser,
		Status:     storage.StatusActive,
		LastActiveAt: now,
		CreatedAt:    now,
	}
	if err := repo.UpsertUser(ctx, user); err != nil {
		t.Fatalf("failed to upsert user: %v", err)
	}

	// Update existing user on conflict
	user.Username = "updated_testuser"
	user.FirstName = "Updated Tester"
	if err := repo.UpsertUser(ctx, user); err != nil {
		t.Fatalf("failed to update user on conflict: %v", err)
	}

	fetchedUser, err := repo.GetUserByTelegramID(ctx, 1001)
	if err != nil || fetchedUser == nil {
		t.Fatalf("failed to fetch user: %v", err)
	}
	if fetchedUser.Username != "updated_testuser" {
		t.Errorf("expected username 'updated_testuser', got: %s", fetchedUser.Username)
	}

	// Lookup by ID
	userByID, err := repo.GetUserByID(ctx, fetchedUser.ID)
	if err != nil || userByID == nil {
		t.Fatalf("failed to fetch user by id: %v", err)
	}
	if userByID.TelegramID != 1001 {
		t.Errorf("expected telegram_id 1001, got: %d", userByID.TelegramID)
	}

	// User not found
	if _, err := repo.GetUserByTelegramID(ctx, 999999); err == nil {
		t.Errorf("expected error for non-existent telegram id, got nil")
	}
	if _, err := repo.GetUserByID(ctx, 999999); err == nil {
		t.Errorf("expected error for non-existent user id, got nil")
	}

	// Update Role, Status, Touch
	if err := repo.UpdateUserRole(ctx, 1001, storage.RoleAdmin); err != nil {
		t.Errorf("failed to update role: %v", err)
	}
	if err := repo.UpdateUserStatus(ctx, 1001, storage.StatusActive); err != nil {
		t.Errorf("failed to update status: %v", err)
	}
	if err := repo.TouchUserActive(ctx, 1001); err != nil {
		t.Errorf("failed to touch user active: %v", err)
	}

	// ListUsers with multiple filters
	activeAfter := now.Add(-1 * time.Hour)
	regAfter := now.Add(-1 * time.Hour)
	regBefore := now.Add(1 * time.Hour)
	users, count, err := repo.ListUsers(ctx, storage.UserFilter{
		Role:             storage.RoleAdmin,
		Status:           storage.StatusActive,
		ActiveAfter:      &activeAfter,
		RegisteredAfter:  &regAfter,
		RegisteredBefore: &regBefore,
		Limit:            0, // fallback default limit 50
		Offset:           0,
	})
	if err != nil || count != 1 || len(users) != 1 {
		t.Errorf("expected 1 filtered user, got count=%d len=%d err=%v", count, len(users), err)
	}

	// --- 2. Post Management ---
	post := &storage.Post{
		Slug:              "p_test123",
		FileID:            "BAACAgIAAxkBAAI...",
		FileUniqueID:      "AQAD...",
		FileType:          storage.FileTypeVideo,
		Caption:           "Test Caption",
		AuthorID:          1001,
		AutoDeleteSeconds: 120,
		IsProtected:       true,
	}
	if err := repo.CreatePost(ctx, post); err != nil {
		t.Fatalf("failed to create post: %v", err)
	}
	if post.ID == 0 {
		t.Errorf("expected generated post ID > 0")
	}

	// Get Post By Slug, ID, FileID
	pSlug, err := repo.GetPostBySlug(ctx, "p_test123")
	if err != nil || pSlug == nil {
		t.Fatalf("failed to get post by slug: %v", err)
	}
	pID, err := repo.GetPostByID(ctx, post.ID)
	if err != nil || pID == nil {
		t.Fatalf("failed to get post by id: %v", err)
	}
	pFileID, err := repo.GetPostByFileID(ctx, "BAACAgIAAxkBAAI...")
	if err != nil || pFileID == nil {
		t.Fatalf("failed to get post by file_id: %v", err)
	}

	// Not found lookups
	if _, err := repo.GetPostBySlug(ctx, "nonexistent"); err == nil {
		t.Errorf("expected error for missing slug, got nil")
	}
	if _, err := repo.GetPostByID(ctx, 999999); err == nil {
		t.Errorf("expected error for missing post id, got nil")
	}
	if _, err := repo.GetPostByFileID(ctx, "nonexistent_file"); err == nil {
		t.Errorf("expected error for missing file id, got nil")
	}

	// Update Post
	pSlug.Caption = "Updated Caption"
	pSlug.AutoDeleteSeconds = 240
	if err := repo.UpdatePost(ctx, pSlug); err != nil {
		t.Fatalf("failed to update post: %v", err)
	}

	// List Posts (with author filter and author=0)
	posts, pCount, err := repo.ListPosts(ctx, 1001, 0, 0)
	if err != nil || pCount != 1 || len(posts) != 1 {
		t.Errorf("expected 1 post listed by author, got count=%d len=%d", pCount, len(posts))
	}
	allPosts, _, err := repo.ListPosts(ctx, 0, 10, 0)
	if err != nil || len(allPosts) != 1 {
		t.Errorf("expected 1 post listed for all authors, got %d", len(allPosts))
	}

	// Views count increment
	if err := repo.IncrementPostViews(ctx, post.ID, 1001, "127.0.0.1"); err != nil {
		t.Fatalf("failed to increment views: %v", err)
	}

	// --- 3. Reactions (Full state matrix) ---
	// 3a. New Like
	if err := repo.SetReaction(ctx, post.ID, 1001, storage.ReactionLike); err != nil {
		t.Fatalf("failed to set like reaction: %v", err)
	}
	reaction, err := repo.GetUserReaction(ctx, post.ID, 1001)
	if err != nil || reaction != storage.ReactionLike {
		t.Errorf("expected reaction 'like', got: %s", reaction)
	}

	// 3b. Switch from Like to Dislike
	if err := repo.SetReaction(ctx, post.ID, 1001, storage.ReactionDislike); err != nil {
		t.Fatalf("failed to switch to dislike reaction: %v", err)
	}
	reaction, err = repo.GetUserReaction(ctx, post.ID, 1001)
	if err != nil || reaction != storage.ReactionDislike {
		t.Errorf("expected reaction 'dislike', got: %s", reaction)
	}

	// 3c. Switch back from Dislike to Like
	if err := repo.SetReaction(ctx, post.ID, 1001, storage.ReactionLike); err != nil {
		t.Fatalf("failed to switch back to like reaction: %v", err)
	}

	// 3d. Toggle off Like (same reaction again)
	if err := repo.SetReaction(ctx, post.ID, 1001, storage.ReactionLike); err != nil {
		t.Fatalf("failed to toggle off like reaction: %v", err)
	}
	reaction, err = repo.GetUserReaction(ctx, post.ID, 1001)
	if err != nil || reaction != "" {
		t.Errorf("expected empty reaction after toggle off, got: %s", reaction)
	}

	// 3e. New Dislike -> Toggle off Dislike
	if err := repo.SetReaction(ctx, post.ID, 1001, storage.ReactionDislike); err != nil {
		t.Fatalf("failed to set dislike reaction: %v", err)
	}
	if err := repo.SetReaction(ctx, post.ID, 1001, storage.ReactionDislike); err != nil {
		t.Fatalf("failed to toggle off dislike reaction: %v", err)
	}

	// Get reaction for non-existent like returns empty string without error
	nonExistentReaction, err := repo.GetUserReaction(ctx, post.ID, 999999)
	if err != nil || nonExistentReaction != "" {
		t.Errorf("expected empty reaction for non-existent user, got: %s", nonExistentReaction)
	}

	// --- 4. Reports ---
	report := &storage.Report{
		PostID:     post.ID,
		ReportedBy: 1001,
		Reason:     "broken_stream",
		Details:    "Video playback error",
	}
	if err := repo.CreateReport(ctx, report); err != nil {
		t.Fatalf("failed to create report: %v", err)
	}
	if report.ID == 0 {
		t.Errorf("expected generated report ID > 0")
	}

	// Get Report by ID
	repByID, err := repo.GetReportByID(ctx, report.ID)
	if err != nil || repByID == nil {
		t.Fatalf("failed to get report by id: %v", err)
	}
	if repByID.Reason != "broken_stream" {
		t.Errorf("expected reason 'broken_stream', got: %s", repByID.Reason)
	}

	// Missing report lookup
	if _, err := repo.GetReportByID(ctx, 999999); err == nil {
		t.Errorf("expected error for missing report id, got nil")
	}

	// Update Report Status
	if err := repo.UpdateReportStatus(ctx, report.ID, storage.ReportStatusResolved, 1001, "Fixed"); err != nil {
		t.Fatalf("failed to update report status: %v", err)
	}

	// List Reports (with status and author filters, and default limit <= 0)
	reports, rCount, err := repo.ListReports(ctx, storage.ReportStatusResolved, 1001, 0, 0)
	if err != nil || rCount != 1 || len(reports) != 1 {
		t.Errorf("expected 1 resolved report, got count=%d len=%d", rCount, len(reports))
	}

	// List Reports without filters
	allReps, _, err := repo.ListReports(ctx, "", 0, 10, 0)
	if err != nil || len(allReps) != 1 {
		t.Errorf("expected 1 report in all reports list, got %d", len(allReps))
	}

	// --- 5. Channels ---
	channel := &storage.Channel{
		TelegramID: -1001234567890,
		Title:      "Official Channel",
		InviteLink: "https://t.me/+AbCdEf",
		IsRequired: true,
	}
	if err := repo.AddChannel(ctx, channel); err != nil {
		t.Fatalf("failed to add channel: %v", err)
	}
	if channel.ID == 0 {
		t.Errorf("expected generated channel ID > 0")
	}

	// Conflict update channel
	channel.Title = "Updated Official Channel"
	if err := repo.AddChannel(ctx, channel); err != nil {
		t.Fatalf("failed to update channel on conflict: %v", err)
	}

	// Get Channel by Telegram ID
	chByTG, err := repo.GetChannelByTelegramID(ctx, -1001234567890)
	if err != nil || chByTG == nil {
		t.Fatalf("failed to get channel by telegram id: %v", err)
	}
	if chByTG.Title != "Updated Official Channel" {
		t.Errorf("expected channel title 'Updated Official Channel', got: %s", chByTG.Title)
	}

	// Missing channel lookup
	if _, err := repo.GetChannelByTelegramID(ctx, -999999); err == nil {
		t.Errorf("expected error for non-existent channel, got nil")
	}

	// List channels (required only vs all)
	chListReq, err := repo.ListChannels(ctx, true)
	if err != nil || len(chListReq) != 1 {
		t.Errorf("expected 1 required channel, got: %d", len(chListReq))
	}
	chListAll, err := repo.ListChannels(ctx, false)
	if err != nil || len(chListAll) != 1 {
		t.Errorf("expected 1 channel total, got: %d", len(chListAll))
	}

	// Remove channel
	if err := repo.RemoveChannel(ctx, channel.ID); err != nil {
		t.Fatalf("failed to remove channel: %v", err)
	}

	// --- 6. Subscriptions ---
	sub := &storage.Subscription{
		UserID:          1001,
		PlanID:          "vip_30d",
		Tier:            "vip",
		Status:          storage.SubStatusActive,
		AZPaysInvoiceID: "az_test_invoice_123",
		AmountCrypto:    9.99,
		Currency:        "USDT",
		ExpiresAt:       time.Now().Add(30 * 24 * time.Hour),
	}
	if err := repo.CreateSubscription(ctx, sub); err != nil {
		t.Fatalf("failed to create subscription: %v", err)
	}
	if sub.ID == 0 {
		t.Errorf("expected generated sub ID > 0")
	}

	// Lookup by invoice ID
	subByInv, err := repo.GetSubscriptionByInvoice(ctx, "az_test_invoice_123")
	if err != nil || subByInv == nil {
		t.Fatalf("failed to get sub by invoice: %v", err)
	}
	if subByInv.AmountCrypto != 9.99 {
		t.Errorf("expected amount 9.99, got: %f", subByInv.AmountCrypto)
	}

	// Missing sub lookup
	if _, err := repo.GetSubscriptionByInvoice(ctx, "nonexistent_invoice"); err == nil {
		t.Errorf("expected error for missing invoice, got nil")
	}
	if _, err := repo.GetActiveSubscription(ctx, 999999); err == nil {
		t.Errorf("expected error for missing active subscription, got nil")
	}

	// Update subscription status
	newExpires := time.Now().Add(60 * 24 * time.Hour)
	if err := repo.UpdateSubscriptionStatus(ctx, "az_test_invoice_123", storage.SubStatusActive, newExpires); err != nil {
		t.Fatalf("failed to update subscription status: %v", err)
	}

	// List subscriptions (default limit <= 0)
	subs, sCount, err := repo.ListSubscriptions(ctx, 0, 0)
	if err != nil || sCount != 1 || len(subs) != 1 {
		t.Errorf("expected 1 subscription listed, got count=%d len=%d", sCount, len(subs))
	}

	// --- 7. Settings ---
	if err := repo.SetSetting(ctx, "site_name", "Telegram Publisher", "Main site title"); err != nil {
		t.Fatalf("failed to set setting: %v", err)
	}
	// Conflict update
	if err := repo.SetSetting(ctx, "site_name", "Telegram Publisher Pro", "Updated title"); err != nil {
		t.Fatalf("failed to update setting on conflict: %v", err)
	}

	val, err := repo.GetSetting(ctx, "site_name")
	if err != nil || val != "Telegram Publisher Pro" {
		t.Errorf("expected setting 'Telegram Publisher Pro', got: %s", val)
	}

	// Missing setting returns empty string without error
	missingVal, err := repo.GetSetting(ctx, "nonexistent_setting_key")
	if err != nil || missingVal != "" {
		t.Errorf("expected empty string for missing setting, got: %s", missingVal)
	}

	settingsMap, err := repo.ListSettings(ctx)
	if err != nil || len(settingsMap) == 0 {
		t.Fatalf("failed to list settings: %v", err)
	}
	if settingsMap["site_name"] != "Telegram Publisher Pro" {
		t.Errorf("expected map value match")
	}

	// --- 8. Events & Analytics ---
	event := &storage.AnalyticsEvent{
		EventType:    "post_viewed",
		UserID:       1001,
		PostID:       &post.ID,
		MetadataJSON: `{"client":"telegram_miniapp"}`,
	}
	if err := repo.RecordEvent(ctx, event); err != nil {
		t.Fatalf("failed to record event: %v", err)
	}

	summary, err := repo.GetAnalyticsSummary(ctx)
	if err != nil {
		t.Fatalf("failed to get analytics summary: %v", err)
	}
	if summary.TotalUsers != 1 {
		t.Errorf("expected TotalUsers=1, got: %d", summary.TotalUsers)
	}
	if summary.TotalPosts != 1 {
		t.Errorf("expected TotalPosts=1, got: %d", summary.TotalPosts)
	}
	if summary.ActiveSubscribers != 1 {
		t.Errorf("expected ActiveSubscribers=1, got: %d", summary.ActiveSubscribers)
	}
	if summary.PostsByType[storage.FileTypeVideo] != 1 {
		t.Errorf("expected 1 video post in summary, got: %d", summary.PostsByType[storage.FileTypeVideo])
	}
	if summary.UsersByRole[storage.RoleAdmin] != 1 {
		t.Errorf("expected 1 admin user in summary, got: %d", summary.UsersByRole[storage.RoleAdmin])
	}

	// --- 9. Delete Post ---
	if err := repo.DeletePost(ctx, post.ID); err != nil {
		t.Fatalf("failed to delete post: %v", err)
	}
	if _, err := repo.GetPostByID(ctx, post.ID); err == nil {
		t.Errorf("expected error after deleting post, got nil")
	}
}

func TestSQLiteNewDefaultPathAndClose(t *testing.T) {
	// New with empty path uses default
	r := sqlite.New("")
	if err := r.Close(); err != nil {
		t.Errorf("unexpected error closing uninitialized repo: %v", err)
	}
}

func TestSQLiteRepositoryClosedErrors(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tg_pub_closed_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := sqlite.New(filepath.Join(tempDir, "closed.db"))
	ctx := context.Background()

	if err := repo.Init(ctx); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	_ = repo.Close()

	// Test error returns on closed DB
	if err := repo.CreatePost(ctx, &storage.Post{Slug: "closed"}); err == nil {
		t.Errorf("expected error on closed repo for CreatePost")
	}
	if _, _, err := repo.ListUsers(ctx, storage.UserFilter{}); err == nil {
		t.Errorf("expected error on closed repo for ListUsers")
	}
	if _, _, err := repo.ListPosts(ctx, 0, 10, 0); err == nil {
		t.Errorf("expected error on closed repo for ListPosts")
	}
	if err := repo.SetReaction(ctx, 1, 1, storage.ReactionLike); err == nil {
		t.Errorf("expected error on closed repo for SetReaction")
	}
	if err := repo.CreateReport(ctx, &storage.Report{}); err == nil {
		t.Errorf("expected error on closed repo for CreateReport")
	}
	if _, _, err := repo.ListReports(ctx, "", 0, 10, 0); err == nil {
		t.Errorf("expected error on closed repo for ListReports")
	}
	if err := repo.AddChannel(ctx, &storage.Channel{}); err == nil {
		t.Errorf("expected error on closed repo for AddChannel")
	}
	if _, err := repo.ListChannels(ctx, false); err == nil {
		t.Errorf("expected error on closed repo for ListChannels")
	}
	if err := repo.CreateSubscription(ctx, &storage.Subscription{}); err == nil {
		t.Errorf("expected error on closed repo for CreateSubscription")
	}
	if _, _, err := repo.ListSubscriptions(ctx, 10, 0); err == nil {
		t.Errorf("expected error on closed repo for ListSubscriptions")
	}
	if _, err := repo.ListSettings(ctx); err == nil {
		t.Errorf("expected error on closed repo for ListSettings")
	}
}

func TestSQLiteInitDirectoryError(t *testing.T) {
	tempFile, err := os.CreateTemp("", "tg_pub_init_file_*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Using a file as parent dir will make MkdirAll fail
	badPath := filepath.Join(tempFile.Name(), "sub", "test.db")
	repo := sqlite.New(badPath)
	if err := repo.Init(context.Background()); err == nil {
		t.Errorf("expected error initializing repo in invalid directory path, got nil")
	}
}
