package storage

import (
	"context"
	"time"
)

// User Role constants
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleAuthor = "author"
	RoleUser   = "user"
)

// User Status constants
const (
	StatusActive        = "active"
	StatusBanned        = "banned"
	StatusBlockedByUser = "blocked_by_user"
)

// File Type constants
const (
	FileTypeVideo     = "video"
	FileTypePhoto     = "photo"
	FileTypeDocument  = "document"
	FileTypeAnimation = "animation"
	FileTypeAudio     = "audio"
)

// Report Status constants
const (
	ReportStatusPending    = "pending"
	ReportStatusInProgress = "in_progress"
	ReportStatusResolved   = "resolved"
	ReportStatusDismissed  = "dismissed"
)

// Reaction Type constants
const (
	ReactionLike     = "like"
	ReactionDislike  = "dislike"
	ReactionFavorite = "favorite"
)

// Subscription Status constants
const (
	SubStatusActive  = "active"
	SubStatusExpired = "expired"
	SubStatusPending = "pending"
)

// User represents a registered bot or mini-app user.
type User struct {
	ID           int64     `json:"id" db:"id" bson:"_id,omitempty"`
	TelegramID   int64     `json:"telegram_id" db:"telegram_id" bson:"telegram_id"`
	Username     string    `json:"username" db:"username" bson:"username"`
	FirstName    string    `json:"first_name" db:"first_name" bson:"first_name"`
	Role         string    `json:"role" db:"role" bson:"role"`
	Status       string    `json:"status" db:"status" bson:"status"`
	LastActiveAt time.Time `json:"last_active_at" db:"last_active_at" bson:"last_active_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at" bson:"created_at"`
}

// Post represents a published media item.
type Post struct {
	ID                int64     `json:"id" db:"id" bson:"_id,omitempty"`
	Slug              string    `json:"slug" db:"slug" bson:"slug"`
	FileID            string    `json:"file_id" db:"file_id" bson:"file_id"`
	FileUniqueID      string    `json:"file_unique_id" db:"file_unique_id" bson:"file_unique_id"`
	FileType          string    `json:"file_type" db:"file_type" bson:"file_type"`
	Caption           string    `json:"caption" db:"caption" bson:"caption"`
	AuthorID          int64     `json:"author_id" db:"author_id" bson:"author_id"`
	AutoDeleteSeconds int       `json:"auto_delete_seconds" db:"auto_delete_seconds" bson:"auto_delete_seconds"`
	ViewsCount        int       `json:"views_count" db:"views_count" bson:"views_count"`
	LikesCount        int       `json:"likes_count" db:"likes_count" bson:"likes_count"`
	DislikesCount     int       `json:"dislikes_count" db:"dislikes_count" bson:"dislikes_count"`
	IsProtected       bool      `json:"is_protected" db:"is_protected" bson:"is_protected"`
	CreatedAt         time.Time `json:"created_at" db:"created_at" bson:"created_at"`
}

// Like represents a user's reaction to a post.
type Like struct {
	ID           int64     `json:"id" db:"id" bson:"_id,omitempty"`
	PostID       int64     `json:"post_id" db:"post_id" bson:"post_id"`
	UserID       int64     `json:"user_id" db:"user_id" bson:"user_id"`
	ReactionType string    `json:"reaction_type" db:"reaction_type" bson:"reaction_type"`
	CreatedAt    time.Time `json:"created_at" db:"created_at" bson:"created_at"`
}

// Report represents a user-submitted broken media report.
type Report struct {
	ID             int64      `json:"id" db:"id" bson:"_id,omitempty"`
	PostID         int64      `json:"post_id" db:"post_id" bson:"post_id"`
	ReportedBy     int64      `json:"reported_by" db:"reported_by" bson:"reported_by"`
	Reason         string     `json:"reason" db:"reason" bson:"reason"`
	Details        string     `json:"details" db:"details" bson:"details"`
	Status         string     `json:"status" db:"status" bson:"status"`
	ResolvedBy     *int64     `json:"resolved_by,omitempty" db:"resolved_by" bson:"resolved_by,omitempty"`
	ResolutionNote string     `json:"resolution_note" db:"resolution_note" bson:"resolution_note"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at" bson:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at" bson:"updated_at"`
	Post           *Post      `json:"post,omitempty" bson:"-"`
	Reporter       *User      `json:"reporter,omitempty" bson:"-"`
}

// View records unique post views.
type View struct {
	ID       int64     `json:"id" db:"id" bson:"_id,omitempty"`
	PostID   int64     `json:"post_id" db:"post_id" bson:"post_id"`
	UserID   int64     `json:"user_id" db:"user_id" bson:"user_id"`
	IPHash   string    `json:"ip_hash" db:"ip_hash" bson:"ip_hash"`
	ViewedAt time.Time `json:"viewed_at" db:"viewed_at" bson:"viewed_at"`
}

// AnalyticsEvent represents an event stream record.
type AnalyticsEvent struct {
	ID           int64     `json:"id" db:"id" bson:"_id,omitempty"`
	EventType    string    `json:"event_type" db:"event_type" bson:"event_type"`
	UserID       int64     `json:"user_id" db:"user_id" bson:"user_id"`
	PostID       *int64    `json:"post_id,omitempty" db:"post_id" bson:"post_id,omitempty"`
	MetadataJSON string    `json:"metadata_json" db:"metadata_json" bson:"metadata_json"`
	CreatedAt    time.Time `json:"created_at" db:"created_at" bson:"created_at"`
}

// Channel represents a force-sub required Telegram channel.
type Channel struct {
	ID         int64     `json:"id" db:"id" bson:"_id,omitempty"`
	TelegramID int64     `json:"telegram_id" db:"telegram_id" bson:"telegram_id"`
	Title      string    `json:"title" db:"title" bson:"title"`
	InviteLink string    `json:"invite_link" db:"invite_link" bson:"invite_link"`
	IsRequired bool      `json:"is_required" db:"is_required" bson:"is_required"`
	CreatedAt  time.Time `json:"created_at" db:"created_at" bson:"created_at"`
}

// Subscription represents an AZPays crypto subscription pass.
type Subscription struct {
	ID              int64     `json:"id" db:"id" bson:"_id,omitempty"`
	UserID          int64     `json:"user_id" db:"user_id" bson:"user_id"`
	PlanID          string    `json:"plan_id" db:"plan_id" bson:"plan_id"`
	Tier            string    `json:"tier" db:"tier" bson:"tier"`
	Status          string    `json:"status" db:"status" bson:"status"`
	AZPaysInvoiceID string    `json:"azpays_invoice_id" db:"azpays_invoice_id" bson:"azpays_invoice_id"`
	AmountCrypto    float64   `json:"amount_crypto" db:"amount_crypto" bson:"amount_crypto"`
	Currency        string    `json:"currency" db:"currency" bson:"currency"`
	ExpiresAt       time.Time `json:"expires_at" db:"expires_at" bson:"expires_at"`
	CreatedAt       time.Time `json:"created_at" db:"created_at" bson:"created_at"`
}

// Setting represents dynamic key-value storage.
type Setting struct {
	Key         string    `json:"key" db:"key" bson:"_id"`
	Value       string    `json:"value" db:"value" bson:"value"`
	Description string    `json:"description" db:"description" bson:"description"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at" bson:"updated_at"`
}

// UserFilter defines criteria for user queries and marketing broadcasts.
type UserFilter struct {
	Role               string     `json:"role"`
	Status             string     `json:"status"`
	ActiveAfter        *time.Time `json:"active_after"`
	RegisteredAfter    *time.Time `json:"registered_after"`
	RegisteredBefore   *time.Time `json:"registered_before"`
	HasActiveSub       *bool      `json:"has_active_sub"`
	Limit              int        `json:"limit"`
	Offset             int        `json:"offset"`
}

// AnalyticsSummary contains calculated metrics for dashboard reports.
type AnalyticsSummary struct {
	TotalUsers         int64            `json:"total_users"`
	ActiveUsersDaily   int64            `json:"active_users_daily"`
	ActiveUsersWeekly  int64            `json:"active_users_weekly"`
	ActiveUsersMonthly int64            `json:"active_users_monthly"`
	BlockedUsers       int64            `json:"blocked_users"`
	BannedUsers        int64            `json:"banned_users"`
	TotalPosts         int64            `json:"total_posts"`
	TotalViews         int64            `json:"total_views"`
	TotalLikes         int64            `json:"total_likes"`
	TotalDislikes      int64            `json:"total_dislikes"`
	PendingReports     int64            `json:"pending_reports"`
	ResolvedReports    int64            `json:"resolved_reports"`
	TotalRevenueCrypto float64          `json:"total_revenue_crypto"`
	ActiveSubscribers  int64            `json:"active_subscribers"`
	PostsByType        map[string]int64 `json:"posts_by_type"`
	UsersByRole        map[string]int64 `json:"users_by_role"`
}

// Repository defines the universal storage interface implemented across SQLite, PostgreSQL, MySQL, MariaDB, and MongoDB.
type Repository interface {
	// Lifecycle
	Init(ctx context.Context) error
	Close() error

	// Users
	UpsertUser(ctx context.Context, user *User) error
	GetUserByTelegramID(ctx context.Context, tgID int64) (*User, error)
	GetUserByID(ctx context.Context, id int64) (*User, error)
	UpdateUserRole(ctx context.Context, tgID int64, role string) error
	UpdateUserStatus(ctx context.Context, tgID int64, status string) error
	TouchUserActive(ctx context.Context, tgID int64) error
	ListUsers(ctx context.Context, filter UserFilter) ([]*User, int64, error)

	// Posts
	CreatePost(ctx context.Context, post *Post) error
	GetPostBySlug(ctx context.Context, slug string) (*Post, error)
	GetPostByID(ctx context.Context, id int64) (*Post, error)
	GetPostByFileID(ctx context.Context, fileID string) (*Post, error)
	UpdatePost(ctx context.Context, post *Post) error
	DeletePost(ctx context.Context, id int64) error
	ListPosts(ctx context.Context, authorID int64, limit, offset int) ([]*Post, int64, error)
	IncrementPostViews(ctx context.Context, postID, userID int64, ipHash string) error

	// Reactions / Likes
	SetReaction(ctx context.Context, postID, userID int64, reaction string) error
	GetUserReaction(ctx context.Context, postID, userID int64) (string, error)

	// Reports
	CreateReport(ctx context.Context, report *Report) error
	GetReportByID(ctx context.Context, id int64) (*Report, error)
	UpdateReportStatus(ctx context.Context, id int64, status string, resolvedBy int64, note string) error
	ListReports(ctx context.Context, status string, authorID int64, limit, offset int) ([]*Report, int64, error)

	// Channels
	AddChannel(ctx context.Context, ch *Channel) error
	RemoveChannel(ctx context.Context, id int64) error
	ListChannels(ctx context.Context, onlyRequired bool) ([]*Channel, error)
	GetChannelByTelegramID(ctx context.Context, tgID int64) (*Channel, error)

	// Subscriptions
	CreateSubscription(ctx context.Context, sub *Subscription) error
	GetSubscriptionByInvoice(ctx context.Context, invoiceID string) (*Subscription, error)
	GetActiveSubscription(ctx context.Context, userID int64) (*Subscription, error)
	UpdateSubscriptionStatus(ctx context.Context, invoiceID, status string, expiresAt time.Time) error
	ListSubscriptions(ctx context.Context, limit, offset int) ([]*Subscription, int64, error)

	// Settings
	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key, value, desc string) error
	ListSettings(ctx context.Context) (map[string]string, error)

	// Analytics & Events
	RecordEvent(ctx context.Context, event *AnalyticsEvent) error
	GetAnalyticsSummary(ctx context.Context) (*AnalyticsSummary, error)
}
