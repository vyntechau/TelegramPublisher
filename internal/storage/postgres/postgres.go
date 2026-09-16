package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/vyntechau/TelegramPublisher/config"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
)

type PostgresRepository struct {
	db  *sql.DB
	cfg config.DatabaseConfig
}

func New(cfg config.DatabaseConfig) *PostgresRepository {
	return &PostgresRepository{cfg: cfg}
}

func (r *PostgresRepository) Init(ctx context.Context) error {
	dsn := r.cfg.URL
	if dsn == "" {
		dsn = r.cfg.DSN
	}
	if dsn == "" {
		return fmt.Errorf("postgres connection URL or DSN is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open postgres database: %w", err)
	}

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	r.db = db

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		telegram_id BIGINT UNIQUE NOT NULL,
		username VARCHAR(255) DEFAULT '',
		first_name VARCHAR(255) DEFAULT '',
		role VARCHAR(50) DEFAULT 'user',
		status VARCHAR(50) DEFAULT 'active',
		last_active_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS posts (
		id BIGSERIAL PRIMARY KEY,
		slug VARCHAR(255) UNIQUE NOT NULL,
		file_id TEXT NOT NULL,
		file_unique_id VARCHAR(255) DEFAULT '',
		file_type VARCHAR(50) DEFAULT 'video',
		caption TEXT DEFAULT '',
		author_id BIGINT NOT NULL,
		auto_delete_seconds INT DEFAULT 120,
		views_count INT DEFAULT 0,
		likes_count INT DEFAULT 0,
		dislikes_count INT DEFAULT 0,
		is_protected BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS likes (
		id BIGSERIAL PRIMARY KEY,
		post_id BIGINT NOT NULL,
		user_id BIGINT NOT NULL,
		reaction_type VARCHAR(50) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		UNIQUE(post_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS reports (
		id BIGSERIAL PRIMARY KEY,
		post_id BIGINT NOT NULL,
		reported_by BIGINT NOT NULL,
		reason VARCHAR(255) NOT NULL,
		details TEXT DEFAULT '',
		status VARCHAR(50) DEFAULT 'pending',
		resolved_by BIGINT,
		resolution_note TEXT DEFAULT '',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS views (
		id BIGSERIAL PRIMARY KEY,
		post_id BIGINT NOT NULL,
		user_id BIGINT NOT NULL,
		ip_hash VARCHAR(255) DEFAULT '',
		viewed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS analytics_events (
		id BIGSERIAL PRIMARY KEY,
		event_type VARCHAR(100) NOT NULL,
		user_id BIGINT NOT NULL,
		post_id BIGINT,
		metadata_json TEXT DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS channels (
		id BIGSERIAL PRIMARY KEY,
		telegram_id BIGINT UNIQUE NOT NULL,
		title VARCHAR(255) DEFAULT '',
		invite_link TEXT NOT NULL,
		is_required BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS subscriptions (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		plan_id VARCHAR(100) NOT NULL,
		tier VARCHAR(50) DEFAULT 'vip',
		status VARCHAR(50) DEFAULT 'pending',
		azpays_invoice_id VARCHAR(255) UNIQUE NOT NULL,
		amount_crypto DOUBLE PRECISION DEFAULT 0.0,
		currency VARCHAR(50) DEFAULT 'USDT',
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS settings (
		key VARCHAR(255) PRIMARY KEY,
		value TEXT NOT NULL,
		description TEXT DEFAULT '',
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_pg_posts_slug ON posts(slug);
	CREATE INDEX IF NOT EXISTS idx_pg_users_tg ON users(telegram_id);
	`

	_, err = r.db.ExecContext(ctx, schema)
	return err
}

func (r *PostgresRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func (r *PostgresRepository) UpsertUser(ctx context.Context, u *storage.User) error {
	query := `
	INSERT INTO users (telegram_id, username, first_name, role, status, last_active_at, created_at)
	VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	ON CONFLICT (telegram_id) DO UPDATE SET
		username = EXCLUDED.username,
		first_name = EXCLUDED.first_name,
		last_active_at = NOW();
	`
	_, err := r.db.ExecContext(ctx, query, u.TelegramID, u.Username, u.FirstName, u.Role, u.Status)
	return err
}

func (r *PostgresRepository) GetUserByTelegramID(ctx context.Context, tgID int64) (*storage.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, telegram_id, username, first_name, role, status, last_active_at, created_at FROM users WHERE telegram_id = $1`, tgID)
	u := &storage.User{}
	err := row.Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.Role, &u.Status, &u.LastActiveAt, &u.CreatedAt)
	return u, err
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id int64) (*storage.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, telegram_id, username, first_name, role, status, last_active_at, created_at FROM users WHERE id = $1`, id)
	u := &storage.User{}
	err := row.Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.Role, &u.Status, &u.LastActiveAt, &u.CreatedAt)
	return u, err
}

func (r *PostgresRepository) UpdateUserRole(ctx context.Context, tgID int64, role string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET role = $1 WHERE telegram_id = $2`, role, tgID)
	return err
}

func (r *PostgresRepository) UpdateUserStatus(ctx context.Context, tgID int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET status = $1 WHERE telegram_id = $2`, status, tgID)
	return err
}

func (r *PostgresRepository) TouchUserActive(ctx context.Context, tgID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET last_active_at = NOW() WHERE telegram_id = $1`, tgID)
	return err
}

func (r *PostgresRepository) ListUsers(ctx context.Context, f storage.UserFilter) ([]*storage.User, int64, error) {
	where := "WHERE 1=1"
	var args []interface{}
	idx := 1

	if f.Role != "" {
		where += fmt.Sprintf(" AND role = $%d", idx)
		args = append(args, f.Role)
		idx++
	}
	if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, f.Status)
		idx++
	}
	if f.ActiveAfter != nil {
		where += fmt.Sprintf(" AND last_active_at >= $%d", idx)
		args = append(args, *f.ActiveAfter)
		idx++
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

	query := fmt.Sprintf("SELECT id, telegram_id, username, first_name, role, status, last_active_at, created_at FROM users %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d", where, idx, idx+1)
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

func (r *PostgresRepository) CreatePost(ctx context.Context, p *storage.Post) error {
	query := `
	INSERT INTO posts (slug, file_id, file_unique_id, file_type, caption, author_id, auto_delete_seconds, is_protected, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	RETURNING id;
	`
	return r.db.QueryRowContext(ctx, query, p.Slug, p.FileID, p.FileUniqueID, p.FileType, p.Caption, p.AuthorID, p.AutoDeleteSeconds, p.IsProtected).Scan(&p.ID)
}

func (r *PostgresRepository) GetPostBySlug(ctx context.Context, slug string) (*storage.Post, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, slug, file_id, file_unique_id, file_type, caption, author_id, auto_delete_seconds, views_count, likes_count, dislikes_count, is_protected, created_at FROM posts WHERE slug = $1`, slug)
	p := &storage.Post{}
	err := row.Scan(&p.ID, &p.Slug, &p.FileID, &p.FileUniqueID, &p.FileType, &p.Caption, &p.AuthorID, &p.AutoDeleteSeconds, &p.ViewsCount, &p.LikesCount, &p.DislikesCount, &p.IsProtected, &p.CreatedAt)
	return p, err
}

func (r *PostgresRepository) GetPostByID(ctx context.Context, id int64) (*storage.Post, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, slug, file_id, file_unique_id, file_type, caption, author_id, auto_delete_seconds, views_count, likes_count, dislikes_count, is_protected, created_at FROM posts WHERE id = $1`, id)
	p := &storage.Post{}
	err := row.Scan(&p.ID, &p.Slug, &p.FileID, &p.FileUniqueID, &p.FileType, &p.Caption, &p.AuthorID, &p.AutoDeleteSeconds, &p.ViewsCount, &p.LikesCount, &p.DislikesCount, &p.IsProtected, &p.CreatedAt)
	return p, err
}

func (r *PostgresRepository) GetPostByFileID(ctx context.Context, fileID string) (*storage.Post, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, slug, file_id, file_unique_id, file_type, caption, author_id, auto_delete_seconds, views_count, likes_count, dislikes_count, is_protected, created_at FROM posts WHERE file_id = $1`, fileID)
	p := &storage.Post{}
	err := row.Scan(&p.ID, &p.Slug, &p.FileID, &p.FileUniqueID, &p.FileType, &p.Caption, &p.AuthorID, &p.AutoDeleteSeconds, &p.ViewsCount, &p.LikesCount, &p.DislikesCount, &p.IsProtected, &p.CreatedAt)
	return p, err
}

func (r *PostgresRepository) UpdatePost(ctx context.Context, p *storage.Post) error {
	_, err := r.db.ExecContext(ctx, `UPDATE posts SET file_id = $1, file_type = $2, caption = $3, auto_delete_seconds = $4, is_protected = $5 WHERE id = $6`, p.FileID, p.FileType, p.Caption, p.AutoDeleteSeconds, p.IsProtected, p.ID)
	return err
}

func (r *PostgresRepository) DeletePost(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM posts WHERE id = $1`, id)
	return err
}

func (r *PostgresRepository) ListPosts(ctx context.Context, authorID int64, limit, offset int) ([]*storage.Post, int64, error) {
	where := ""
	var args []interface{}
	idx := 1
	if authorID > 0 {
		where = fmt.Sprintf("WHERE author_id = $%d", idx)
		args = append(args, authorID)
		idx++
	}

	var count int64
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM posts %s", where), args...).Scan(&count); err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	query := fmt.Sprintf("SELECT id, slug, file_id, file_unique_id, file_type, caption, author_id, auto_delete_seconds, views_count, likes_count, dislikes_count, is_protected, created_at FROM posts %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d", where, idx, idx+1)
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

func (r *PostgresRepository) IncrementPostViews(ctx context.Context, postID, userID int64, ipHash string) error {
	_, _ = r.db.ExecContext(ctx, `INSERT INTO views (post_id, user_id, ip_hash, viewed_at) VALUES ($1, $2, $3, NOW())`, postID, userID, ipHash)
	_, err := r.db.ExecContext(ctx, `UPDATE posts SET views_count = views_count + 1 WHERE id = $1`, postID)
	return err
}

func (r *PostgresRepository) SetReaction(ctx context.Context, postID, userID int64, reaction string) error {
	var cur string
	err := r.db.QueryRowContext(ctx, `SELECT reaction_type FROM likes WHERE post_id = $1 AND user_id = $2`, postID, userID).Scan(&cur)
	if err == sql.ErrNoRows {
		_, err = r.db.ExecContext(ctx, `INSERT INTO likes (post_id, user_id, reaction_type, created_at) VALUES ($1, $2, $3, NOW())`, postID, userID, reaction)
		if err == nil {
			if reaction == storage.ReactionLike {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET likes_count = likes_count + 1 WHERE id = $1`, postID)
			} else {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET dislikes_count = dislikes_count + 1 WHERE id = $1`, postID)
			}
		}
		return err
	} else if err == nil {
		if cur == reaction {
			_, _ = r.db.ExecContext(ctx, `DELETE FROM likes WHERE post_id = $1 AND user_id = $2`, postID, userID)
			if reaction == storage.ReactionLike {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET likes_count = GREATEST(0, likes_count - 1) WHERE id = $1`, postID)
			} else {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET dislikes_count = GREATEST(0, dislikes_count - 1) WHERE id = $1`, postID)
			}
		} else {
			_, _ = r.db.ExecContext(ctx, `UPDATE likes SET reaction_type = $1 WHERE post_id = $2 AND user_id = $3`, reaction, postID, userID)
			if reaction == storage.ReactionLike {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET likes_count = likes_count + 1, dislikes_count = GREATEST(0, dislikes_count - 1) WHERE id = $1`, postID)
			} else {
				_, _ = r.db.ExecContext(ctx, `UPDATE posts SET dislikes_count = dislikes_count + 1, likes_count = GREATEST(0, likes_count - 1) WHERE id = $1`, postID)
			}
		}
	}
	return nil
}

func (r *PostgresRepository) GetUserReaction(ctx context.Context, postID, userID int64) (string, error) {
	var reaction string
	err := r.db.QueryRowContext(ctx, `SELECT reaction_type FROM likes WHERE post_id = $1 AND user_id = $2`, postID, userID).Scan(&reaction)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return reaction, err
}

func (r *PostgresRepository) CreateReport(ctx context.Context, rep *storage.Report) error {
	query := `
	INSERT INTO reports (post_id, reported_by, reason, details, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, 'pending', NOW(), NOW())
	RETURNING id;
	`
	return r.db.QueryRowContext(ctx, query, rep.PostID, rep.ReportedBy, rep.Reason, rep.Details).Scan(&rep.ID)
}

func (r *PostgresRepository) GetReportByID(ctx context.Context, id int64) (*storage.Report, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, post_id, reported_by, reason, details, status, resolved_by, resolution_note, created_at, updated_at FROM reports WHERE id = $1`, id)
	rep := &storage.Report{}
	err := row.Scan(&rep.ID, &rep.PostID, &rep.ReportedBy, &rep.Reason, &rep.Details, &rep.Status, &rep.ResolvedBy, &rep.ResolutionNote, &rep.CreatedAt, &rep.UpdatedAt)
	return rep, err
}

func (r *PostgresRepository) UpdateReportStatus(ctx context.Context, id int64, status string, resolvedBy int64, note string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE reports SET status = $1, resolved_by = $2, resolution_note = $3, updated_at = NOW() WHERE id = $4`, status, resolvedBy, note, id)
	return err
}

func (r *PostgresRepository) ListReports(ctx context.Context, status string, authorID int64, limit, offset int) ([]*storage.Report, int64, error) {
	where := "WHERE 1=1"
	var args []interface{}
	idx := 1
	if status != "" {
		where += fmt.Sprintf(" AND r.status = $%d", idx)
		args = append(args, status)
		idx++
	}
	if authorID > 0 {
		where += fmt.Sprintf(" AND p.author_id = $%d", idx)
		args = append(args, authorID)
		idx++
	}

	var count int64
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM reports r JOIN posts p ON r.post_id = p.id %s", where), args...).Scan(&count); err != nil {
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
	%s ORDER BY r.created_at DESC LIMIT $%d OFFSET $%d`, where, idx, idx+1)
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

func (r *PostgresRepository) AddChannel(ctx context.Context, ch *storage.Channel) error {
	query := `
	INSERT INTO channels (telegram_id, title, invite_link, is_required, created_at)
	VALUES ($1, $2, $3, $4, NOW())
	ON CONFLICT (telegram_id) DO UPDATE SET title = EXCLUDED.title, invite_link = EXCLUDED.invite_link, is_required = EXCLUDED.is_required
	RETURNING id;
	`
	return r.db.QueryRowContext(ctx, query, ch.TelegramID, ch.Title, ch.InviteLink, ch.IsRequired).Scan(&ch.ID)
}

func (r *PostgresRepository) RemoveChannel(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM channels WHERE id = $1`, id)
	return err
}

func (r *PostgresRepository) ListChannels(ctx context.Context, onlyRequired bool) ([]*storage.Channel, error) {
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

func (r *PostgresRepository) GetChannelByTelegramID(ctx context.Context, tgID int64) (*storage.Channel, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, telegram_id, title, invite_link, is_required, created_at FROM channels WHERE telegram_id = $1`, tgID)
	ch := &storage.Channel{}
	err := row.Scan(&ch.ID, &ch.TelegramID, &ch.Title, &ch.InviteLink, &ch.IsRequired, &ch.CreatedAt)
	return ch, err
}

func (r *PostgresRepository) CreateSubscription(ctx context.Context, sub *storage.Subscription) error {
	query := `
	INSERT INTO subscriptions (user_id, plan_id, tier, status, azpays_invoice_id, amount_crypto, currency, expires_at, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	RETURNING id;
	`
	return r.db.QueryRowContext(ctx, query, sub.UserID, sub.PlanID, sub.Tier, sub.Status, sub.AZPaysInvoiceID, sub.AmountCrypto, sub.Currency, sub.ExpiresAt).Scan(&sub.ID)
}

func (r *PostgresRepository) GetSubscriptionByInvoice(ctx context.Context, invoiceID string) (*storage.Subscription, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, plan_id, tier, status, azpays_invoice_id, amount_crypto, currency, expires_at, created_at FROM subscriptions WHERE azpays_invoice_id = $1`, invoiceID)
	s := &storage.Subscription{}
	err := row.Scan(&s.ID, &s.UserID, &s.PlanID, &s.Tier, &s.Status, &s.AZPaysInvoiceID, &s.AmountCrypto, &s.Currency, &s.ExpiresAt, &s.CreatedAt)
	return s, err
}

func (r *PostgresRepository) GetActiveSubscription(ctx context.Context, userID int64) (*storage.Subscription, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, plan_id, tier, status, azpays_invoice_id, amount_crypto, currency, expires_at, created_at FROM subscriptions WHERE user_id = $1 AND status = 'active' AND expires_at > NOW() ORDER BY expires_at DESC LIMIT 1`, userID)
	s := &storage.Subscription{}
	err := row.Scan(&s.ID, &s.UserID, &s.PlanID, &s.Tier, &s.Status, &s.AZPaysInvoiceID, &s.AmountCrypto, &s.Currency, &s.ExpiresAt, &s.CreatedAt)
	return s, err
}

func (r *PostgresRepository) UpdateSubscriptionStatus(ctx context.Context, invoiceID, status string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE subscriptions SET status = $1, expires_at = $2 WHERE azpays_invoice_id = $3`, status, expiresAt, invoiceID)
	return err
}

func (r *PostgresRepository) ListSubscriptions(ctx context.Context, limit, offset int) ([]*storage.Subscription, int64, error) {
	var count int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM subscriptions`).Scan(&count); err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, user_id, plan_id, tier, status, azpays_invoice_id, amount_crypto, currency, expires_at, created_at FROM subscriptions ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
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

func (r *PostgresRepository) GetSetting(ctx context.Context, key string) (string, error) {
	var val string
	err := r.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = $1`, key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return val, err
}

func (r *PostgresRepository) SetSetting(ctx context.Context, key, value, desc string) error {
	query := `
	INSERT INTO settings (key, value, description, updated_at)
	VALUES ($1, $2, $3, NOW())
	ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, description = EXCLUDED.description, updated_at = NOW();
	`
	_, err := r.db.ExecContext(ctx, query, key, value, desc)
	return err
}

func (r *PostgresRepository) ListSettings(ctx context.Context) (map[string]string, error) {
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

func (r *PostgresRepository) RecordEvent(ctx context.Context, ev *storage.AnalyticsEvent) error {
	query := `INSERT INTO analytics_events (event_type, user_id, post_id, metadata_json, created_at) VALUES ($1, $2, $3, $4, NOW())`
	_, err := r.db.ExecContext(ctx, query, ev.EventType, ev.UserID, ev.PostID, ev.MetadataJSON)
	return err
}

func (r *PostgresRepository) GetAnalyticsSummary(ctx context.Context) (*storage.AnalyticsSummary, error) {
	summary := &storage.AnalyticsSummary{
		PostsByType: make(map[string]int64),
		UsersByRole: make(map[string]int64),
	}

	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&summary.TotalUsers)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE status = 'blocked_by_user'`).Scan(&summary.BlockedUsers)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE status = 'banned'`).Scan(&summary.BannedUsers)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE last_active_at >= NOW() - INTERVAL '1 day'`).Scan(&summary.ActiveUsersDaily)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE last_active_at >= NOW() - INTERVAL '7 days'`).Scan(&summary.ActiveUsersWeekly)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE last_active_at >= NOW() - INTERVAL '30 days'`).Scan(&summary.ActiveUsersMonthly)

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
