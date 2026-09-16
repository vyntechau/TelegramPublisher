package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/vyntechau/TelegramPublisher/config"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
)

type MySQLRepository struct {
	db  *sql.DB
	cfg config.DatabaseConfig
}

func New(cfg config.DatabaseConfig) *MySQLRepository {
	return &MySQLRepository{cfg: cfg}
}

func (r *MySQLRepository) Init(ctx context.Context) error {
	dsn := r.cfg.URL
	if dsn == "" {
		dsn = r.cfg.DSN
	}
	if dsn == "" {
		return fmt.Errorf("mysql connection URL or DSN is required")
	}

	dsn = normalizeMySQLDSN(dsn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open mysql database: %w", err)
	}

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	r.db = db

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		telegram_id BIGINT UNIQUE NOT NULL,
		username VARCHAR(255) DEFAULT '',
		first_name VARCHAR(255) DEFAULT '',
		role VARCHAR(50) DEFAULT 'user',
		status VARCHAR(50) DEFAULT 'active',
		last_active_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS posts (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		slug VARCHAR(255) UNIQUE NOT NULL,
		file_id TEXT NOT NULL,
		file_unique_id VARCHAR(255) DEFAULT '',
		file_type VARCHAR(50) DEFAULT 'video',
		caption TEXT,
		author_id BIGINT NOT NULL,
		auto_delete_seconds INT DEFAULT 120,
		views_count INT DEFAULT 0,
		likes_count INT DEFAULT 0,
		dislikes_count INT DEFAULT 0,
		is_protected BOOLEAN DEFAULT TRUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS likes (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		post_id BIGINT NOT NULL,
		user_id BIGINT NOT NULL,
		reaction_type VARCHAR(50) NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE KEY uq_post_user (post_id, user_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS reports (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		post_id BIGINT NOT NULL,
		reported_by BIGINT NOT NULL,
		reason VARCHAR(255) NOT NULL,
		details TEXT,
		status VARCHAR(50) DEFAULT 'pending',
		resolved_by BIGINT,
		resolution_note TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS views (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		post_id BIGINT NOT NULL,
		user_id BIGINT NOT NULL,
		ip_hash VARCHAR(255) DEFAULT '',
		viewed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS analytics_events (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		event_type VARCHAR(100) NOT NULL,
		user_id BIGINT NOT NULL,
		post_id BIGINT,
		metadata_json TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS channels (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		telegram_id BIGINT UNIQUE NOT NULL,
		title VARCHAR(255) DEFAULT '',
		invite_link TEXT NOT NULL,
		is_required BOOLEAN DEFAULT TRUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS subscriptions (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT NOT NULL,
		plan_id VARCHAR(100) NOT NULL,
		tier VARCHAR(50) DEFAULT 'vip',
		status VARCHAR(50) DEFAULT 'pending',
		azpays_invoice_id VARCHAR(255) UNIQUE NOT NULL,
		amount_crypto DOUBLE DEFAULT 0.0,
		currency VARCHAR(50) DEFAULT 'USDT',
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS settings (
		` + "`key`" + ` VARCHAR(255) PRIMARY KEY,
		value TEXT NOT NULL,
		description TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`

	_, err = r.db.ExecContext(ctx, schema)
	return err
}

func (r *MySQLRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func (r *MySQLRepository) UpsertUser(ctx context.Context, u *storage.User) error {
	query := `
	INSERT INTO users (telegram_id, username, first_name, role, status, last_active_at, created_at)
	VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	ON DUPLICATE KEY UPDATE
		username = VALUES(username),
		first_name = VALUES(first_name),
		last_active_at = NOW();
	`
	_, err := r.db.ExecContext(ctx, query, u.TelegramID, u.Username, u.FirstName, u.Role, u.Status)
	return err
}

func (r *MySQLRepository) GetUserByTelegramID(ctx context.Context, tgID int64) (*storage.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, telegram_id, username, first_name, role, status, last_active_at, created_at FROM users WHERE telegram_id = ?`, tgID)
	u := &storage.User{}
	err := row.Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.Role, &u.Status, &u.LastActiveAt, &u.CreatedAt)
	return u, err
}

func (r *MySQLRepository) GetUserByID(ctx context.Context, id int64) (*storage.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, telegram_id, username, first_name, role, status, last_active_at, created_at FROM users WHERE id = ?`, id)
	u := &storage.User{}
	err := row.Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.Role, &u.Status, &u.LastActiveAt, &u.CreatedAt)
	return u, err
}

func (r *MySQLRepository) UpdateUserRole(ctx context.Context, tgID int64, role string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET role = ? WHERE telegram_id = ?`, role, tgID)
	return err
}

func (r *MySQLRepository) UpdateUserStatus(ctx context.Context, tgID int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET status = ? WHERE telegram_id = ?`, status, tgID)
	return err
}

func (r *MySQLRepository) TouchUserActive(ctx context.Context, tgID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET last_active_at = NOW() WHERE telegram_id = ?`, tgID)
	return err
}

func (r *MySQLRepository) ListUsers(ctx context.Context, f storage.UserFilter) ([]*storage.User, int64, error) {
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
		where += " AND last_active_at >= ?"
		args = append(args, *f.ActiveAfter)
	}

	var count int64
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM users %s", where), args...).Scan(&count); err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := f.Offset

	query := fmt.Sprintf("SELECT id, telegram_id, username, first_name, role, status, last_active_at, created_at FROM users %s ORDER BY created_at DESC LIMIT ? OFFSET ?", where)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
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

func (r *MySQLRepository) CreatePost(ctx context.Context, p *storage.Post) error {
	query := `
	INSERT INTO posts (slug, file_id, file_unique_id, file_type, caption, author_id, auto_delete_seconds, is_protected, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW());
	`
	res, err := r.db.ExecContext(ctx, query, p.Slug, p.FileID, p.FileUniqueID, p.FileType, p.Caption, p.AuthorID, p.AutoDeleteSeconds, p.IsProtected)
	if err != nil {
		return err
	}
	p.ID, _ = res.LastInsertId()
	return nil
}

func (r *MySQLRepository) GetPostBySlug(ctx context.Context, slug string) (*storage.Post, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, slug, file_id, file_unique_id, file_type, caption, author_id, auto_delete_seconds, views_count, likes_count, dislikes_count, is_protected, created_at FROM posts WHERE slug = ?`, slug)
	p := &storage.Post{}
	err := row.Scan(&p.ID, &p.Slug, &p.FileID, &p.FileUniqueID, &p.FileType, &p.Caption, &p.AuthorID, &p.AutoDeleteSeconds, &p.ViewsCount, &p.LikesCount, &p.DislikesCount, &p.IsProtected, &p.CreatedAt)
	return p, err
}

func (r *MySQLRepository) GetPostByID(ctx context.Context, id int64) (*storage.Post, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, slug, file_id, file_unique_id, file_type, caption, author_id, auto_delete_seconds, views_count, likes_count, dislikes_count, is_protected, created_at FROM posts WHERE id = ?`, id)
	p := &storage.Post{}
	err := row.Scan(&p.ID, &p.Slug, &p.FileID, &p.FileUniqueID, &p.FileType, &p.Caption, &p.AuthorID, &p.AutoDeleteSeconds, &p.ViewsCount, &p.LikesCount, &p.DislikesCount, &p.IsProtected, &p.CreatedAt)
	return p, err
}

func (r *MySQLRepository) GetPostByFileID(ctx context.Context, fileID string) (*storage.Post, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, slug, file_id, file_unique_id, file_type, caption, author_id, auto_delete_seconds, views_count, likes_count, dislikes_count, is_protected, created_at FROM posts WHERE file_id = ?`, fileID)
	p := &storage.Post{}
	err := row.Scan(&p.ID, &p.Slug, &p.FileID, &p.FileUniqueID, &p.FileType, &p.Caption, &p.AuthorID, &p.AutoDeleteSeconds, &p.ViewsCount, &p.LikesCount, &p.DislikesCount, &p.IsProtected, &p.CreatedAt)
	return p, err
}

func (r *MySQLRepository) UpdatePost(ctx context.Context, p *storage.Post) error {
	_, err := r.db.ExecContext(ctx, `UPDATE posts SET file_id = ?, file_type = ?, caption = ?, auto_delete_seconds = ?, is_protected = ? WHERE id = ?`, p.FileID, p.FileType, p.Caption, p.AutoDeleteSeconds, p.IsProtected, p.ID)
	return err
}

func (r *MySQLRepository) DeletePost(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM posts WHERE id = ?`, id)
	return err
}

func (r *MySQLRepository) ListPosts(ctx context.Context, authorID int64, limit, offset int) ([]*storage.Post, int64, error) {
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
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
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

func (r *MySQLRepository) IncrementPostViews(ctx context.Context, postID, userID int64, ipHash string) error {
	_, _ = r.db.ExecContext(ctx, `INSERT INTO views (post_id, user_id, ip_hash, viewed_at) VALUES (?, ?, ?, NOW())`, postID, userID, ipHash)
	_, err := r.db.ExecContext(ctx, `UPDATE posts SET views_count = views_count + 1 WHERE id = ?`, postID)
	return err
}

func (r *MySQLRepository) SetReaction(ctx context.Context, postID, userID int64, reaction string) error {
	var cur string
	err := r.db.QueryRowContext(ctx, `SELECT reaction_type FROM likes WHERE post_id = ? AND user_id = ?`, postID, userID).Scan(&cur)
	if err == sql.ErrNoRows {
		_, err = r.db.ExecContext(ctx, `INSERT INTO likes (post_id, user_id, reaction_type, created_at) VALUES (?, ?, ?, NOW())`, postID, userID, reaction)
		if err == nil {
			if reaction == storage.ReactionLike {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET likes_count = likes_count + 1 WHERE id = ?`, postID)
			} else {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET dislikes_count = dislikes_count + 1 WHERE id = ?`, postID)
			}
		}
		return err
	} else if err == nil {
		if cur == reaction {
			_, _ = r.db.ExecContext(ctx, `DELETE FROM likes WHERE post_id = ? AND user_id = ?`, postID, userID)
			if reaction == storage.ReactionLike {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET likes_count = GREATEST(0, likes_count - 1) WHERE id = ?`, postID)
			} else {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET dislikes_count = GREATEST(0, dislikes_count - 1) WHERE id = ?`, postID)
			}
		} else {
			_, _ = r.db.ExecContext(ctx, `UPDATE likes SET reaction_type = ? WHERE post_id = ? AND user_id = ?`, reaction, postID, userID)
			if reaction == storage.ReactionLike {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET likes_count = likes_count + 1, dislikes_count = GREATEST(0, dislikes_count - 1) WHERE id = ?`, postID)
			} else {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET dislikes_count = dislikes_count + 1, likes_count = GREATEST(0, likes_count - 1) WHERE id = ?`, postID)
			}
		}
	}
	return nil
}

func (r *MySQLRepository) GetUserReaction(ctx context.Context, postID, userID int64) (string, error) {
	var reaction string
	err := r.db.QueryRowContext(ctx, `SELECT reaction_type FROM likes WHERE post_id = ? AND user_id = ?`, postID, userID).Scan(&reaction)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return reaction, err
}

func (r *MySQLRepository) CreateReport(ctx context.Context, rep *storage.Report) error {
	query := `
	INSERT INTO reports (post_id, reported_by, reason, details, status, created_at, updated_at)
	VALUES (?, ?, ?, ?, 'pending', NOW(), NOW());
	`
	res, err := r.db.ExecContext(ctx, query, rep.PostID, rep.ReportedBy, rep.Reason, rep.Details)
	if err != nil {
		return err
	}
	rep.ID, _ = res.LastInsertId()
	return nil
}

func (r *MySQLRepository) GetReportByID(ctx context.Context, id int64) (*storage.Report, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, post_id, reported_by, reason, details, status, resolved_by, resolution_note, created_at, updated_at FROM reports WHERE id = ?`, id)
	rep := &storage.Report{}
	err := row.Scan(&rep.ID, &rep.PostID, &rep.ReportedBy, &rep.Reason, &rep.Details, &rep.Status, &rep.ResolvedBy, &rep.ResolutionNote, &rep.CreatedAt, &rep.UpdatedAt)
	return rep, err
}

func (r *MySQLRepository) UpdateReportStatus(ctx context.Context, id int64, status string, resolvedBy int64, note string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE reports SET status = ?, resolved_by = ?, resolution_note = ?, updated_at = NOW() WHERE id = ?`, status, resolvedBy, note, id)
	return err
}

func (r *MySQLRepository) ListReports(ctx context.Context, status string, authorID int64, limit, offset int) ([]*storage.Report, int64, error) {
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
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
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

func (r *MySQLRepository) AddChannel(ctx context.Context, ch *storage.Channel) error {
	query := `
	INSERT INTO channels (telegram_id, title, invite_link, is_required, created_at)
	VALUES (?, ?, ?, ?, NOW())
	ON DUPLICATE KEY UPDATE title = VALUES(title), invite_link = VALUES(invite_link), is_required = VALUES(is_required);
	`
	res, err := r.db.ExecContext(ctx, query, ch.TelegramID, ch.Title, ch.InviteLink, ch.IsRequired)
	if err != nil {
		return err
	}
	ch.ID, _ = res.LastInsertId()
	return nil
}

func (r *MySQLRepository) RemoveChannel(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM channels WHERE id = ?`, id)
	return err
}

func (r *MySQLRepository) ListChannels(ctx context.Context, onlyRequired bool) ([]*storage.Channel, error) {
	where := ""
	if onlyRequired {
		where = "WHERE is_required = TRUE"
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

func (r *MySQLRepository) GetChannelByTelegramID(ctx context.Context, tgID int64) (*storage.Channel, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, telegram_id, title, invite_link, is_required, created_at FROM channels WHERE telegram_id = ?`, tgID)
	ch := &storage.Channel{}
	err := row.Scan(&ch.ID, &ch.TelegramID, &ch.Title, &ch.InviteLink, &ch.IsRequired, &ch.CreatedAt)
	return ch, err
}

func (r *MySQLRepository) CreateSubscription(ctx context.Context, sub *storage.Subscription) error {
	query := `
	INSERT INTO subscriptions (user_id, plan_id, tier, status, azpays_invoice_id, amount_crypto, currency, expires_at, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW());
	`
	res, err := r.db.ExecContext(ctx, query, sub.UserID, sub.PlanID, sub.Tier, sub.Status, sub.AZPaysInvoiceID, sub.AmountCrypto, sub.Currency, sub.ExpiresAt)
	if err != nil {
		return err
	}
	sub.ID, _ = res.LastInsertId()
	return nil
}

func (r *MySQLRepository) GetSubscriptionByInvoice(ctx context.Context, invoiceID string) (*storage.Subscription, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, plan_id, tier, status, azpays_invoice_id, amount_crypto, currency, expires_at, created_at FROM subscriptions WHERE azpays_invoice_id = ?`, invoiceID)
	s := &storage.Subscription{}
	err := row.Scan(&s.ID, &s.UserID, &s.PlanID, &s.Tier, &s.Status, &s.AZPaysInvoiceID, &s.AmountCrypto, &s.Currency, &s.ExpiresAt, &s.CreatedAt)
	return s, err
}

func (r *MySQLRepository) GetActiveSubscription(ctx context.Context, userID int64) (*storage.Subscription, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, plan_id, tier, status, azpays_invoice_id, amount_crypto, currency, expires_at, created_at FROM subscriptions WHERE user_id = ? AND status = 'active' AND expires_at > NOW() ORDER BY expires_at DESC LIMIT 1`, userID)
	s := &storage.Subscription{}
	err := row.Scan(&s.ID, &s.UserID, &s.PlanID, &s.Tier, &s.Status, &s.AZPaysInvoiceID, &s.AmountCrypto, &s.Currency, &s.ExpiresAt, &s.CreatedAt)
	return s, err
}

func (r *MySQLRepository) UpdateSubscriptionStatus(ctx context.Context, invoiceID, status string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE subscriptions SET status = ?, expires_at = ? WHERE azpays_invoice_id = ?`, status, expiresAt, invoiceID)
	return err
}

func (r *MySQLRepository) ListSubscriptions(ctx context.Context, limit, offset int) ([]*storage.Subscription, int64, error) {
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

func (r *MySQLRepository) GetSetting(ctx context.Context, key string) (string, error) {
	var val string
	err := r.db.QueryRowContext(ctx, "SELECT value FROM settings WHERE `key` = ?", key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return val, err
}

func (r *MySQLRepository) SetSetting(ctx context.Context, key, value, desc string) error {
	query := "INSERT INTO settings (`key`, value, description, updated_at) VALUES (?, ?, ?, NOW()) ON DUPLICATE KEY UPDATE value = VALUES(value), description = VALUES(description), updated_at = NOW();"
	_, err := r.db.ExecContext(ctx, query, key, value, desc)
	return err
}

func (r *MySQLRepository) ListSettings(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT `key`, value FROM settings")
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

func (r *MySQLRepository) RecordEvent(ctx context.Context, ev *storage.AnalyticsEvent) error {
	query := `INSERT INTO analytics_events (event_type, user_id, post_id, metadata_json, created_at) VALUES (?, ?, ?, ?, NOW())`
	_, err := r.db.ExecContext(ctx, query, ev.EventType, ev.UserID, ev.PostID, ev.MetadataJSON)
	return err
}

func (r *MySQLRepository) GetAnalyticsSummary(ctx context.Context) (*storage.AnalyticsSummary, error) {
	summary := &storage.AnalyticsSummary{
		PostsByType: make(map[string]int64),
		UsersByRole: make(map[string]int64),
	}

	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&summary.TotalUsers)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE status = 'blocked_by_user'`).Scan(&summary.BlockedUsers)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE status = 'banned'`).Scan(&summary.BannedUsers)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE last_active_at >= NOW() - INTERVAL 1 DAY`).Scan(&summary.ActiveUsersDaily)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE last_active_at >= NOW() - INTERVAL 7 DAY`).Scan(&summary.ActiveUsersWeekly)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE last_active_at >= NOW() - INTERVAL 30 DAY`).Scan(&summary.ActiveUsersMonthly)

	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(views_count),0), COALESCE(SUM(likes_count),0), COALESCE(SUM(dislikes_count),0) FROM posts`).Scan(&summary.TotalPosts, &summary.TotalViews, &summary.TotalLikes, &summary.TotalDislikes)

	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM reports WHERE status = 'pending'`).Scan(&summary.PendingReports)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM reports WHERE status = 'resolved'`).Scan(&summary.ResolvedReports)

	_ = r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount_crypto), 0.0) FROM subscriptions WHERE status = 'active'`).Scan(&summary.TotalRevenueCrypto)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM subscriptions WHERE status = 'active' AND expires_at > NOW()`).Scan(&summary.ActiveSubscribers)

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

func normalizeMySQLDSN(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "mysql://") || strings.HasPrefix(raw, "mariadb://") {
		u, err := url.Parse(raw)
		if err == nil {
			user := u.User.Username()
			pass, _ := u.User.Password()
			host := u.Host
			dbname := strings.TrimPrefix(u.Path, "/")
			query := u.RawQuery
			if query == "" {
				query = "parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"
			} else if !strings.Contains(query, "parseTime=") {
				query += "&parseTime=true"
			}

			if pass != "" {
				return fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", user, pass, host, dbname, query)
			} else if user != "" {
				return fmt.Sprintf("%s@tcp(%s)/%s?%s", user, host, dbname, query)
			}
			return fmt.Sprintf("tcp(%s)/%s?%s", host, dbname, query)
		}
	}
	return raw
}

