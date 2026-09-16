package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vyntechau/TelegramPublisher/internal/storage"
	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db       *sql.DB
	filePath string
}

func New(filePath string) *SQLiteRepository {
	if filePath == "" {
		filePath = "data/publisher.db"
	}
	return &SQLiteRepository{filePath: filePath}
}

var sqlOpen = sql.Open

func (r *SQLiteRepository) Init(ctx context.Context) error {
	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create sqlite dir: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)", r.filePath)
	db, err := sqlOpen("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open sqlite database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	r.db = db

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		telegram_id INTEGER UNIQUE NOT NULL,
		username TEXT DEFAULT '',
		first_name TEXT DEFAULT '',
		role TEXT DEFAULT 'user',
		status TEXT DEFAULT 'active',
		last_active_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		slug TEXT UNIQUE NOT NULL,
		file_id TEXT NOT NULL,
		file_unique_id TEXT DEFAULT '',
		file_type TEXT DEFAULT 'video',
		caption TEXT DEFAULT '',
		author_id INTEGER NOT NULL,
		auto_delete_seconds INTEGER DEFAULT 120,
		views_count INTEGER DEFAULT 0,
		likes_count INTEGER DEFAULT 0,
		dislikes_count INTEGER DEFAULT 0,
		is_protected BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		reaction_type TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(post_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		reported_by INTEGER NOT NULL,
		reason TEXT NOT NULL,
		details TEXT DEFAULT '',
		status TEXT DEFAULT 'pending',
		resolved_by INTEGER,
		resolution_note TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS views (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		ip_hash TEXT DEFAULT '',
		viewed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS analytics_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_type TEXT NOT NULL,
		user_id INTEGER NOT NULL,
		post_id INTEGER,
		metadata_json TEXT DEFAULT '{}',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS channels (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		telegram_id INTEGER UNIQUE NOT NULL,
		title TEXT DEFAULT '',
		invite_link TEXT NOT NULL,
		is_required BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS subscriptions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		plan_id TEXT NOT NULL,
		tier TEXT DEFAULT 'vip',
		status TEXT DEFAULT 'pending',
		azpays_invoice_id TEXT UNIQUE NOT NULL,
		amount_crypto REAL DEFAULT 0.0,
		currency TEXT DEFAULT 'USDT',
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		description TEXT DEFAULT '',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_posts_slug ON posts(slug);
	CREATE INDEX IF NOT EXISTS idx_users_tg ON users(telegram_id);
	CREATE INDEX IF NOT EXISTS idx_likes_post ON likes(post_id);
	CREATE INDEX IF NOT EXISTS idx_reports_post ON reports(post_id);
	CREATE INDEX IF NOT EXISTS idx_events_type ON analytics_events(event_type);
	`

	_, err = r.db.ExecContext(ctx, schema)
	return err
}

func (r *SQLiteRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// Users
func (r *SQLiteRepository) UpsertUser(ctx context.Context, u *storage.User) error {
	query := `
	INSERT INTO users (telegram_id, username, first_name, role, status, last_active_at, created_at)
	VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(telegram_id) DO UPDATE SET
		username = excluded.username,
		first_name = excluded.first_name,
		last_active_at = CURRENT_TIMESTAMP;
	`
	_, err := r.db.ExecContext(ctx, query, u.TelegramID, u.Username, u.FirstName, u.Role, u.Status)
	return err
}

func (r *SQLiteRepository) GetUserByTelegramID(ctx context.Context, tgID int64) (*storage.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, telegram_id, username, first_name, role, status, last_active_at, created_at FROM users WHERE telegram_id = ?`, tgID)
	u := &storage.User{}
	err := row.Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.Role, &u.Status, &u.LastActiveAt, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *SQLiteRepository) GetUserByID(ctx context.Context, id int64) (*storage.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, telegram_id, username, first_name, role, status, last_active_at, created_at FROM users WHERE id = ?`, id)
	u := &storage.User{}
	err := row.Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.Role, &u.Status, &u.LastActiveAt, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *SQLiteRepository) UpdateUserRole(ctx context.Context, tgID int64, role string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET role = ? WHERE telegram_id = ?`, role, tgID)
	return err
}

func (r *SQLiteRepository) UpdateUserStatus(ctx context.Context, tgID int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET status = ? WHERE telegram_id = ?`, status, tgID)
	return err
}

func (r *SQLiteRepository) TouchUserActive(ctx context.Context, tgID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET last_active_at = CURRENT_TIMESTAMP WHERE telegram_id = ?`, tgID)
	return err
}

func (r *SQLiteRepository) ListUsers(ctx context.Context, f storage.UserFilter) ([]*storage.User, int64, error) {
	where := "WHERE 1=1"
	var args []interface{}

	if f.Role != "" {
		where += " AND role = ?"
		args = append(args, f.Role)
	}
	if f.Status != "" {
		where += " AND status = ?"
		args = append(args, f.Status)
	}
	if f.ActiveAfter != nil {
		where += " AND datetime(last_active_at) >= datetime(?)"
		args = append(args, f.ActiveAfter.UTC().Format("2006-01-02 15:04:05"))
	}
	if f.RegisteredAfter != nil {
		where += " AND datetime(created_at) >= datetime(?)"
		args = append(args, f.RegisteredAfter.UTC().Format("2006-01-02 15:04:05"))
	}
	if f.RegisteredBefore != nil {
		where += " AND datetime(created_at) <= datetime(?)"
		args = append(args, f.RegisteredBefore.UTC().Format("2006-01-02 15:04:05"))
	}

	var count int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", where)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&count); err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := f.Offset

	query := fmt.Sprintf("SELECT id, telegram_id, username, first_name, role, status, last_active_at, created_at FROM users %s ORDER BY created_at DESC LIMIT ? OFFSET ?", where)
	queryArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*storage.User
	for rows.Next() {
		u := &storage.User{}
		if err := rows.Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.Role, &u.Status, &u.LastActiveAt, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, count, nil
}

// Posts
func (r *SQLiteRepository) CreatePost(ctx context.Context, p *storage.Post) error {
	query := `
	INSERT INTO posts (slug, file_id, file_unique_id, file_type, caption, author_id, auto_delete_seconds, is_protected, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP);
	`
	res, err := r.db.ExecContext(ctx, query, p.Slug, p.FileID, p.FileUniqueID, p.FileType, p.Caption, p.AuthorID, p.AutoDeleteSeconds, p.IsProtected)
	if err != nil {
		return err
	}
	p.ID, _ = res.LastInsertId()
	return nil
}

func (r *SQLiteRepository) GetPostBySlug(ctx context.Context, slug string) (*storage.Post, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, slug, file_id, file_unique_id, file_type, caption, author_id, auto_delete_seconds, views_count, likes_count, dislikes_count, is_protected, created_at FROM posts WHERE slug = ?`, slug)
	p := &storage.Post{}
	err := row.Scan(&p.ID, &p.Slug, &p.FileID, &p.FileUniqueID, &p.FileType, &p.Caption, &p.AuthorID, &p.AutoDeleteSeconds, &p.ViewsCount, &p.LikesCount, &p.DislikesCount, &p.IsProtected, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *SQLiteRepository) GetPostByID(ctx context.Context, id int64) (*storage.Post, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, slug, file_id, file_unique_id, file_type, caption, author_id, auto_delete_seconds, views_count, likes_count, dislikes_count, is_protected, created_at FROM posts WHERE id = ?`, id)
	p := &storage.Post{}
	err := row.Scan(&p.ID, &p.Slug, &p.FileID, &p.FileUniqueID, &p.FileType, &p.Caption, &p.AuthorID, &p.AutoDeleteSeconds, &p.ViewsCount, &p.LikesCount, &p.DislikesCount, &p.IsProtected, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *SQLiteRepository) GetPostByFileID(ctx context.Context, fileID string) (*storage.Post, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, slug, file_id, file_unique_id, file_type, caption, author_id, auto_delete_seconds, views_count, likes_count, dislikes_count, is_protected, created_at FROM posts WHERE file_id = ?`, fileID)
	p := &storage.Post{}
	err := row.Scan(&p.ID, &p.Slug, &p.FileID, &p.FileUniqueID, &p.FileType, &p.Caption, &p.AuthorID, &p.AutoDeleteSeconds, &p.ViewsCount, &p.LikesCount, &p.DislikesCount, &p.IsProtected, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *SQLiteRepository) UpdatePost(ctx context.Context, p *storage.Post) error {
	query := `
	UPDATE posts SET file_id = ?, file_type = ?, caption = ?, auto_delete_seconds = ?, is_protected = ?
	WHERE id = ?;
	`
	_, err := r.db.ExecContext(ctx, query, p.FileID, p.FileType, p.Caption, p.AutoDeleteSeconds, p.IsProtected, p.ID)
	return err
}

func (r *SQLiteRepository) DeletePost(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM posts WHERE id = ?`, id)
	return err
}

func (r *SQLiteRepository) ListPosts(ctx context.Context, authorID int64, limit, offset int) ([]*storage.Post, int64, error) {
	where := ""
	var args []interface{}
	if authorID > 0 {
		where = "WHERE author_id = ?"
		args = append(args, authorID)
	}

	var count int64
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM posts %s", where), args...).Scan(&count); err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	query := fmt.Sprintf("SELECT id, slug, file_id, file_unique_id, file_type, caption, author_id, auto_delete_seconds, views_count, likes_count, dislikes_count, is_protected, created_at FROM posts %s ORDER BY created_at DESC LIMIT ? OFFSET ?", where)
	queryArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []*storage.Post
	for rows.Next() {
		p := &storage.Post{}
		if err := rows.Scan(&p.ID, &p.Slug, &p.FileID, &p.FileUniqueID, &p.FileType, &p.Caption, &p.AuthorID, &p.AutoDeleteSeconds, &p.ViewsCount, &p.LikesCount, &p.DislikesCount, &p.IsProtected, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		posts = append(posts, p)
	}
	return posts, count, nil
}

func (r *SQLiteRepository) IncrementPostViews(ctx context.Context, postID, userID int64, ipHash string) error {
	_, _ = r.db.ExecContext(ctx, `INSERT INTO views (post_id, user_id, ip_hash, viewed_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)`, postID, userID, ipHash)
	_, err := r.db.ExecContext(ctx, `UPDATE posts SET views_count = views_count + 1 WHERE id = ?`, postID)
	return err
}

// Reactions / Likes
func (r *SQLiteRepository) SetReaction(ctx context.Context, postID, userID int64, reaction string) error {
	var currentReaction string
	err := r.db.QueryRowContext(ctx, `SELECT reaction_type FROM likes WHERE post_id = ? AND user_id = ?`, postID, userID).Scan(&currentReaction)

	if err == sql.ErrNoRows {
		// New reaction
		_, err = r.db.ExecContext(ctx, `INSERT INTO likes (post_id, user_id, reaction_type, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)`, postID, userID, reaction)
		if err != nil {
			return err
		}
		if reaction == storage.ReactionLike {
			_, _ = r.db.ExecContext(ctx, `UPDATE posts SET likes_count = likes_count + 1 WHERE id = ?`, postID)
		} else if reaction == storage.ReactionDislike {
			_, _ = r.db.ExecContext(ctx, `UPDATE posts SET dislikes_count = dislikes_count + 1 WHERE id = ?`, postID)
		}
	} else if err == nil {
		if currentReaction == reaction {
			// Toggle off
			_, _ = r.db.ExecContext(ctx, `DELETE FROM likes WHERE post_id = ? AND user_id = ?`, postID, userID)
			if reaction == storage.ReactionLike {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET likes_count = MAX(0, likes_count - 1) WHERE id = ?`, postID)
			} else if reaction == storage.ReactionDislike {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET dislikes_count = MAX(0, dislikes_count - 1) WHERE id = ?`, postID)
			}
		} else {
			// Change reaction
			_, _ = r.db.ExecContext(ctx, `UPDATE likes SET reaction_type = ? WHERE post_id = ? AND user_id = ?`, reaction, postID, userID)
			if reaction == storage.ReactionLike {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET likes_count = likes_count + 1, dislikes_count = MAX(0, dislikes_count - 1) WHERE id = ?`, postID)
			} else {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET dislikes_count = dislikes_count + 1, likes_count = MAX(0, likes_count - 1) WHERE id = ?`, postID)
			}
		}
	} else {
		return err
	}
	return nil
}

func (r *SQLiteRepository) GetUserReaction(ctx context.Context, postID, userID int64) (string, error) {
	var reaction string
	err := r.db.QueryRowContext(ctx, `SELECT reaction_type FROM likes WHERE post_id = ? AND user_id = ?`, postID, userID).Scan(&reaction)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return reaction, err
}

// Reports
func (r *SQLiteRepository) CreateReport(ctx context.Context, rep *storage.Report) error {
	query := `
	INSERT INTO reports (post_id, reported_by, reason, details, status, created_at, updated_at)
	VALUES (?, ?, ?, ?, 'pending', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
	`
	res, err := r.db.ExecContext(ctx, query, rep.PostID, rep.ReportedBy, rep.Reason, rep.Details)
	if err != nil {
		return err
	}
	rep.ID, _ = res.LastInsertId()
	return nil
}

func (r *SQLiteRepository) GetReportByID(ctx context.Context, id int64) (*storage.Report, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, post_id, reported_by, reason, details, status, resolved_by, resolution_note, created_at, updated_at FROM reports WHERE id = ?`, id)
	rep := &storage.Report{}
	err := row.Scan(&rep.ID, &rep.PostID, &rep.ReportedBy, &rep.Reason, &rep.Details, &rep.Status, &rep.ResolvedBy, &rep.ResolutionNote, &rep.CreatedAt, &rep.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return rep, nil
}

func (r *SQLiteRepository) UpdateReportStatus(ctx context.Context, id int64, status string, resolvedBy int64, note string) error {
	query := `
	UPDATE reports SET status = ?, resolved_by = ?, resolution_note = ?, updated_at = CURRENT_TIMESTAMP
	WHERE id = ?;
	`
	_, err := r.db.ExecContext(ctx, query, status, resolvedBy, note, id)
	return err
}

func (r *SQLiteRepository) ListReports(ctx context.Context, status string, authorID int64, limit, offset int) ([]*storage.Report, int64, error) {
	where := "WHERE 1=1"
	var args []interface{}
	if status != "" {
		where += " AND r.status = ?"
		args = append(args, status)
	}
	if authorID > 0 {
		where += " AND p.author_id = ?"
		args = append(args, authorID)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM reports r JOIN posts p ON r.post_id = p.id %s", where)
	var count int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&count); err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	query := fmt.Sprintf(`
	SELECT r.id, r.post_id, r.reported_by, r.reason, r.details, r.status, r.resolved_by, r.resolution_note, r.created_at, r.updated_at,
	       p.id, p.slug, p.file_id, p.file_type, p.caption, p.author_id
	FROM reports r
	JOIN posts p ON r.post_id = p.id
	%s ORDER BY r.created_at DESC LIMIT ? OFFSET ?`, where)
	queryArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reports []*storage.Report
	for rows.Next() {
		rep := &storage.Report{Post: &storage.Post{}}
		if err := rows.Scan(&rep.ID, &rep.PostID, &rep.ReportedBy, &rep.Reason, &rep.Details, &rep.Status, &rep.ResolvedBy, &rep.ResolutionNote, &rep.CreatedAt, &rep.UpdatedAt,
			&rep.Post.ID, &rep.Post.Slug, &rep.Post.FileID, &rep.Post.FileType, &rep.Post.Caption, &rep.Post.AuthorID); err != nil {
			return nil, 0, err
		}
		reports = append(reports, rep)
	}
	return reports, count, nil
}

// Channels
func (r *SQLiteRepository) AddChannel(ctx context.Context, ch *storage.Channel) error {
	query := `
	INSERT INTO channels (telegram_id, title, invite_link, is_required, created_at)
	VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(telegram_id) DO UPDATE SET title = excluded.title, invite_link = excluded.invite_link, is_required = excluded.is_required;
	`
	res, err := r.db.ExecContext(ctx, query, ch.TelegramID, ch.Title, ch.InviteLink, ch.IsRequired)
	if err != nil {
		return err
	}
	ch.ID, _ = res.LastInsertId()
	return nil
}

func (r *SQLiteRepository) RemoveChannel(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM channels WHERE id = ?`, id)
	return err
}

func (r *SQLiteRepository) ListChannels(ctx context.Context, onlyRequired bool) ([]*storage.Channel, error) {
	where := ""
	if onlyRequired {
		where = "WHERE is_required = 1"
	}
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf("SELECT id, telegram_id, title, invite_link, is_required, created_at FROM channels %s ORDER BY id ASC", where))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*storage.Channel
	for rows.Next() {
		ch := &storage.Channel{}
		if err := rows.Scan(&ch.ID, &ch.TelegramID, &ch.Title, &ch.InviteLink, &ch.IsRequired, &ch.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, nil
}

func (r *SQLiteRepository) GetChannelByTelegramID(ctx context.Context, tgID int64) (*storage.Channel, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, telegram_id, title, invite_link, is_required, created_at FROM channels WHERE telegram_id = ?`, tgID)
	ch := &storage.Channel{}
	err := row.Scan(&ch.ID, &ch.TelegramID, &ch.Title, &ch.InviteLink, &ch.IsRequired, &ch.CreatedAt)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

// Subscriptions
func (r *SQLiteRepository) CreateSubscription(ctx context.Context, sub *storage.Subscription) error {
	query := `
	INSERT INTO subscriptions (user_id, plan_id, tier, status, azpays_invoice_id, amount_crypto, currency, expires_at, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP);
	`
	res, err := r.db.ExecContext(ctx, query, sub.UserID, sub.PlanID, sub.Tier, sub.Status, sub.AZPaysInvoiceID, sub.AmountCrypto, sub.Currency, sub.ExpiresAt.UTC().Format("2006-01-02 15:04:05"))
	if err != nil {
		return err
	}
	sub.ID, _ = res.LastInsertId()
	return nil
}

func (r *SQLiteRepository) GetSubscriptionByInvoice(ctx context.Context, invoiceID string) (*storage.Subscription, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, plan_id, tier, status, azpays_invoice_id, amount_crypto, currency, expires_at, created_at FROM subscriptions WHERE azpays_invoice_id = ?`, invoiceID)
	s := &storage.Subscription{}
	err := row.Scan(&s.ID, &s.UserID, &s.PlanID, &s.Tier, &s.Status, &s.AZPaysInvoiceID, &s.AmountCrypto, &s.Currency, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SQLiteRepository) GetActiveSubscription(ctx context.Context, userID int64) (*storage.Subscription, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, plan_id, tier, status, azpays_invoice_id, amount_crypto, currency, expires_at, created_at FROM subscriptions WHERE user_id = ? AND status = 'active' AND datetime(expires_at) > datetime('now') ORDER BY expires_at DESC LIMIT 1`, userID)
	s := &storage.Subscription{}
	err := row.Scan(&s.ID, &s.UserID, &s.PlanID, &s.Tier, &s.Status, &s.AZPaysInvoiceID, &s.AmountCrypto, &s.Currency, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SQLiteRepository) UpdateSubscriptionStatus(ctx context.Context, invoiceID, status string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE subscriptions SET status = ?, expires_at = ? WHERE azpays_invoice_id = ?`, status, expiresAt.UTC().Format("2006-01-02 15:04:05"), invoiceID)
	return err
}

func (r *SQLiteRepository) ListSubscriptions(ctx context.Context, limit, offset int) ([]*storage.Subscription, int64, error) {
	var count int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM subscriptions`).Scan(&count); err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, user_id, plan_id, tier, status, azpays_invoice_id, amount_crypto, currency, expires_at, created_at FROM subscriptions ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var subs []*storage.Subscription
	for rows.Next() {
		s := &storage.Subscription{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.PlanID, &s.Tier, &s.Status, &s.AZPaysInvoiceID, &s.AmountCrypto, &s.Currency, &s.ExpiresAt, &s.CreatedAt); err != nil {
			return nil, 0, err
		}
		subs = append(subs, s)
	}
	return subs, count, nil
}

// Settings
func (r *SQLiteRepository) GetSetting(ctx context.Context, key string) (string, error) {
	var val string
	err := r.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return val, err
}

func (r *SQLiteRepository) SetSetting(ctx context.Context, key, value, desc string) error {
	query := `
	INSERT INTO settings (key, value, description, updated_at)
	VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(key) DO UPDATE SET value = excluded.value, description = excluded.description, updated_at = CURRENT_TIMESTAMP;
	`
	_, err := r.db.ExecContext(ctx, query, key, value, desc)
	return err
}

func (r *SQLiteRepository) ListSettings(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		res[k] = v
	}
	return res, nil
}

// Analytics & Events
func (r *SQLiteRepository) RecordEvent(ctx context.Context, ev *storage.AnalyticsEvent) error {
	query := `INSERT INTO analytics_events (event_type, user_id, post_id, metadata_json, created_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`
	_, err := r.db.ExecContext(ctx, query, ev.EventType, ev.UserID, ev.PostID, ev.MetadataJSON)
	return err
}

func (r *SQLiteRepository) GetAnalyticsSummary(ctx context.Context) (*storage.AnalyticsSummary, error) {
	summary := &storage.AnalyticsSummary{
		PostsByType: make(map[string]int64),
		UsersByRole: make(map[string]int64),
	}

	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&summary.TotalUsers)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE status = 'blocked_by_user'`).Scan(&summary.BlockedUsers)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE status = 'banned'`).Scan(&summary.BannedUsers)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE last_active_at >= datetime('now', '-1 day')`).Scan(&summary.ActiveUsersDaily)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE last_active_at >= datetime('now', '-7 days')`).Scan(&summary.ActiveUsersWeekly)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE last_active_at >= datetime('now', '-30 days')`).Scan(&summary.ActiveUsersMonthly)

	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(views_count),0), COALESCE(SUM(likes_count),0), COALESCE(SUM(dislikes_count),0) FROM posts`).Scan(&summary.TotalPosts, &summary.TotalViews, &summary.TotalLikes, &summary.TotalDislikes)

	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM reports WHERE status = 'pending'`).Scan(&summary.PendingReports)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM reports WHERE status = 'resolved'`).Scan(&summary.ResolvedReports)

	_ = r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount_crypto), 0.0) FROM subscriptions WHERE status = 'active'`).Scan(&summary.TotalRevenueCrypto)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM subscriptions WHERE status = 'active' AND expires_at > CURRENT_TIMESTAMP`).Scan(&summary.ActiveSubscribers)

	// Posts by type
	rows, err := r.db.QueryContext(ctx, `SELECT file_type, COUNT(*) FROM posts GROUP BY file_type`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ft string
			var cnt int64
			if err := rows.Scan(&ft, &cnt); err == nil {
				summary.PostsByType[ft] = cnt
			}
		}
	}

	// Users by role
	rowsRole, err := r.db.QueryContext(ctx, `SELECT role, COUNT(*) FROM users GROUP BY role`)
	if err == nil {
		defer rowsRole.Close()
		for rowsRole.Next() {
			var role string
			var cnt int64
			if err := rowsRole.Scan(&role, &cnt); err == nil {
				summary.UsersByRole[role] = cnt
			}
		}
	}

	return summary, nil
}
