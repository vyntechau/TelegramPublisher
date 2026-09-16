package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/vyntechau/TelegramPublisher/config"
	"github.com/vyntechau/TelegramPublisher/internal/analytics"
	"github.com/vyntechau/TelegramPublisher/internal/auth"
	"github.com/vyntechau/TelegramPublisher/internal/services/marketing"
	"github.com/vyntechau/TelegramPublisher/internal/services/payment"
	"github.com/vyntechau/TelegramPublisher/internal/services/settings"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
)

var generateJWT = func(s *auth.Service, user *storage.User) (string, error) {
	return s.GenerateJWT(user)
}

var getAllSettings = func(s *settings.Service, ctx context.Context) (map[string]string, error) {
	return s.GetAll(ctx)
}

// Server handles REST API endpoints.
type Server struct {
	repo         storage.Repository
	authSvc      *auth.Service
	settingsSvc  *settings.Service
	analyticsSvc *analytics.Service
	marketSvc    *marketing.Service
	paySvc       *payment.Service
	cfg          *config.Config
}

func NewServer(repo storage.Repository, authSvc *auth.Service, settingsSvc *settings.Service, analyticsSvc *analytics.Service, marketSvc *marketing.Service, paySvc *payment.Service, cfg *config.Config) *Server {
	return &Server{
		repo:         repo,
		authSvc:      authSvc,
		settingsSvc:  settingsSvc,
		analyticsSvc: analyticsSvc,
		marketSvc:    marketSvc,
		paySvc:       paySvc,
		cfg:          cfg,
	}
}

// RegisterRoutes sets up all REST API routes on the mux.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	// Auth
	mux.HandleFunc("/api/v1/auth/telegram", s.handleTelegramAuth)
	mux.HandleFunc("/api/v1/auth/me", s.authMiddleware(s.handleAuthMe))

	// Analytics (Admin/Owner only)
	mux.HandleFunc("/api/v1/analytics/overview", s.requireRoleMiddleware(storage.RoleAdmin, storage.RoleOwner)(s.handleAnalyticsOverview))
	mux.HandleFunc("/api/v1/analytics/export/users.csv", s.requireRoleMiddleware(storage.RoleAdmin, storage.RoleOwner)(s.handleExportUsersCSV))
	mux.HandleFunc("/api/v1/analytics/export/posts.csv", s.requireRoleMiddleware(storage.RoleAdmin, storage.RoleOwner)(s.handleExportPostsCSV))

	// Posts
	mux.HandleFunc("/api/v1/posts", s.handlePosts)
	mux.HandleFunc("/api/v1/posts/", s.handlePostByIDOrSlug)

	// Reactions (Authenticated users)
	mux.HandleFunc("/api/v1/posts/react", s.authMiddleware(s.handlePostReact))

	// Reports
	mux.HandleFunc("/api/v1/reports", s.handleReports)
	mux.HandleFunc("/api/v1/reports/", s.requireRoleMiddleware(storage.RoleAuthor, storage.RoleAdmin, storage.RoleOwner)(s.handleReportUpdate))

	// Marketing / Broadcast (Admin/Owner only)
	mux.HandleFunc("/api/v1/marketing/broadcast", s.requireRoleMiddleware(storage.RoleAdmin, storage.RoleOwner)(s.handleMarketingBroadcast))

	// Channels
	mux.HandleFunc("/api/v1/channels", s.handleChannels)
	mux.HandleFunc("/api/v1/channels/", s.requireRoleMiddleware(storage.RoleAdmin, storage.RoleOwner)(s.handleChannelByID))

	// Users (Admin/Owner only)
	mux.HandleFunc("/api/v1/users", s.requireRoleMiddleware(storage.RoleAdmin, storage.RoleOwner)(s.handleUsers))
	mux.HandleFunc("/api/v1/users/", s.requireRoleMiddleware(storage.RoleAdmin, storage.RoleOwner)(s.handleUserByID))

	// Settings (Admin/Owner only)
	mux.HandleFunc("/api/v1/settings", s.requireRoleMiddleware(storage.RoleAdmin, storage.RoleOwner)(s.handleSettings))

	// Setup Status (Public onboarding check)
	mux.HandleFunc("/api/v1/setup/status", s.handleSetupStatus)

	// Payments & Subscriptions (AzPays, Coinbase, NOWPayments)
	mux.HandleFunc("/api/v1/payments/plans", s.handlePaymentPlans)
	mux.HandleFunc("/api/v1/payments/checkout", s.authMiddleware(s.handlePaymentCheckout))
	mux.HandleFunc("/api/v1/payments/subscription/webhook", s.handleAZPaysWebhook)
	mux.HandleFunc("/api/v1/payments/webhook", s.handleAZPaysWebhook)
	mux.HandleFunc("/api/v1/payments/azpays/webhook", s.handleAZPaysWebhook)
	mux.HandleFunc("/api/v1/payments/mock-checkout", s.handleMockCheckout)
	mux.HandleFunc("/api/v1/payments/azpays/mock-checkout", s.handleMockCheckout)
}

// Helpers
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

// Auth Middleware
func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.WriteHeader(http.StatusOK)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			errorResponse(w, http.StatusUnauthorized, "Missing or invalid Authorization header")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := s.authSvc.ValidateJWT(tokenStr)
		if err != nil {
			errorResponse(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Attach to request context
		ctx := r.Context()
		ctx = setContextClaims(ctx, claims)
		next(w, r.WithContext(ctx))
	}
}

// Role Authorization Middleware
func (s *Server) requireRoleMiddleware(allowedRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return s.authMiddleware(s.requireRole(next, allowedRoles...))
	}
}

func (s *Server) requireRole(next http.HandlerFunc, allowedRoles ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := getContextClaims(r.Context())
		if claims == nil {
			errorResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		authorized := false
		for _, role := range allowedRoles {
			if claims.Role == role || claims.Role == storage.RoleOwner {
				authorized = true
				break
			}
		}

		if !authorized {
			errorResponse(w, http.StatusForbidden, "Forbidden: insufficient permissions for role "+claims.Role)
			return
		}

		next(w, r)
	}
}

// Auth Endpoints
func (s *Server) handleTelegramAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		InitData string `json:"init_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	tgUser, err := s.authSvc.ValidateInitData(req.InitData)
	if err != nil {
		// Fallback for dev / mock testing
		if s.cfg.App.Env == "development" || s.cfg.Bot.Token == "" {
			tgUser = &auth.TelegramUserPayload{
				ID:        s.cfg.Bot.OwnerID,
				FirstName: "DevAdmin",
				Username:  "admin",
			}
		} else {
			errorResponse(w, http.StatusUnauthorized, fmt.Sprintf("Authentication failed: %v", err))
			return
		}
	}

	ctx := r.Context()
	user, err := s.repo.GetUserByTelegramID(ctx, tgUser.ID)
	if err != nil {
		role := storage.RoleUser
		if tgUser.ID == s.cfg.Bot.OwnerID && s.cfg.Bot.OwnerID != 0 {
			role = storage.RoleOwner
		}
		user = &storage.User{
			TelegramID: tgUser.ID,
			Username:   tgUser.Username,
			FirstName:  tgUser.FirstName,
			Role:       role,
			Status:     storage.StatusActive,
		}
		_ = s.repo.UpsertUser(ctx, user)
		user, _ = s.repo.GetUserByTelegramID(ctx, tgUser.ID)
	}

	token, err := generateJWT(s.authSvc, user)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func (s *Server) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	claims := getContextClaims(r.Context())
	if claims == nil {
		errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	user, err := s.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		errorResponse(w, http.StatusNotFound, "User not found")
		return
	}
	sub, _ := s.repo.GetActiveSubscription(r.Context(), user.TelegramID)
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"user":         user,
		"subscription": sub,
	})
}

// Analytics Endpoints
func (s *Server) handleAnalyticsOverview(w http.ResponseWriter, r *http.Request) {
	summary, err := s.analyticsSvc.GetOverview(r.Context())
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, summary)
}

func (s *Server) handleExportUsersCSV(w http.ResponseWriter, r *http.Request) {
	csvData, err := s.analyticsSvc.ExportUsersCSV(r.Context(), storage.UserFilter{})
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=users_export.csv")
	_, _ = w.Write(csvData)
}

func (s *Server) handleExportPostsCSV(w http.ResponseWriter, r *http.Request) {
	csvData, err := s.analyticsSvc.ExportPostsCSV(r.Context(), 0)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=posts_export.csv")
	_, _ = w.Write(csvData)
}

// Posts Endpoints
func (s *Server) handlePosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		posts, total, err := s.repo.ListPosts(ctx, 0, limit, offset)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"posts": posts,
			"total": total,
		})
	case http.MethodPost:
		// Protected
		claims := getContextClaims(ctx)
		if claims == nil || (claims.Role != storage.RoleAuthor && claims.Role != storage.RoleAdmin && claims.Role != storage.RoleOwner) {
			errorResponse(w, http.StatusForbidden, "Forbidden")
			return
		}
		var post storage.Post
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		post.AuthorID = claims.TelegramID
		if err := s.repo.CreatePost(ctx, &post); err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusCreated, post)
	default:
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handlePostByIDOrSlug(w http.ResponseWriter, r *http.Request) {
	slugOrID := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		post, err := s.repo.GetPostBySlug(ctx, slugOrID)
		if err != nil {
			if id, err := strconv.ParseInt(slugOrID, 10, 64); err == nil {
				post, _ = s.repo.GetPostByID(ctx, id)
			}
		}
		if post == nil {
			errorResponse(w, http.StatusNotFound, "Post not found")
			return
		}
		_ = s.repo.IncrementPostViews(ctx, post.ID, 0, r.RemoteAddr)
		jsonResponse(w, http.StatusOK, post)
	case http.MethodDelete:
		claims := getContextClaims(ctx)
		if claims == nil || (claims.Role != storage.RoleAdmin && claims.Role != storage.RoleOwner) {
			errorResponse(w, http.StatusForbidden, "Forbidden")
			return
		}
		id, _ := strconv.ParseInt(slugOrID, 10, 64)
		_ = s.repo.DeletePost(ctx, id)
		jsonResponse(w, http.StatusOK, map[string]string{"message": "Post deleted"})
	default:
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Reactions
func (s *Server) handlePostReact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	claims := getContextClaims(r.Context())
	if claims == nil {
		errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		PostID   int64  `json:"post_id"`
		Reaction string `json:"reaction"` // like, dislike
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err := s.repo.SetReaction(r.Context(), req.PostID, claims.TelegramID, req.Reaction); err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	post, _ := s.repo.GetPostByID(r.Context(), req.PostID)
	jsonResponse(w, http.StatusOK, post)
}

// Reports
func (s *Server) handleReports(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		status := r.URL.Query().Get("status")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		reports, total, err := s.repo.ListReports(ctx, status, 0, limit, offset)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"reports": reports,
			"total":   total,
		})
	case http.MethodPost:
		var rep storage.Report
		if err := json.NewDecoder(r.Body).Decode(&rep); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		if err := s.repo.CreateReport(ctx, &rep); err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusCreated, rep)
	default:
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleReportUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/reports/")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var req struct {
		Status string `json:"status"`
		Note   string `json:"resolution_note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	claims := getContextClaims(r.Context())
	resolverID := int64(0)
	if claims != nil {
		resolverID = claims.TelegramID
	}

	if err := s.repo.UpdateReportStatus(r.Context(), id, req.Status, resolverID, req.Note); err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Report updated"})
}

// Marketing Broadcast
func (s *Server) handleMarketingBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var payload marketing.BroadcastPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	res, err := s.marketSvc.Dispatch(r.Context(), payload)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, res)
}

// Channels
func (s *Server) handleChannels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		channels, err := s.repo.ListChannels(ctx, false)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, channels)
	case http.MethodPost:
		var ch storage.Channel
		if err := json.NewDecoder(r.Body).Decode(&ch); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		if err := s.repo.AddChannel(ctx, &ch); err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusCreated, ch)
	default:
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleChannelByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/channels/")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if r.Method == http.MethodDelete {
		_ = s.repo.RemoveChannel(r.Context(), id)
		jsonResponse(w, http.StatusOK, map[string]string{"message": "Channel removed"})
		return
	}
	errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
}

// Users
func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	role := r.URL.Query().Get("role")
	status := r.URL.Query().Get("status")

	users, total, err := s.repo.ListUsers(ctx, storage.UserFilter{
		Limit:  limit,
		Offset: offset,
		Role:   role,
		Status: status,
	})
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"users": users,
		"total": total,
	})
}

func (s *Server) handleUserByID(w http.ResponseWriter, r *http.Request) {
	tgIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	tgID, _ := strconv.ParseInt(tgIDStr, 10, 64)

	var req struct {
		Role   string `json:"role"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	ctx := r.Context()
	if req.Role != "" {
		_ = s.repo.UpdateUserRole(ctx, tgID, req.Role)
	}
	if req.Status != "" {
		_ = s.repo.UpdateUserStatus(ctx, tgID, req.Status)
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "User updated"})
}

// Settings
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		settingsMap, err := getAllSettings(s.settingsSvc, ctx)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, settingsMap)
	case http.MethodPost:
		var req map[string]string
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		for k, v := range req {
			_ = s.settingsSvc.Set(ctx, k, v, "Updated from admin web")
		}
		jsonResponse(w, http.StatusOK, map[string]string{"message": "Settings saved"})
	default:
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Payments & AZPays
func (s *Server) handlePaymentPlans(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, payment.DefaultPlans)
}

func (s *Server) handlePaymentCheckout(w http.ResponseWriter, r *http.Request) {
	claims := getContextClaims(r.Context())
	if claims == nil {
		errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		PlanID   string `json:"plan_id"`
		Currency string `json:"currency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	invoice, err := s.paySvc.CreateCheckoutSession(r.Context(), claims.TelegramID, req.PlanID, req.Currency)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, invoice)
}

func (s *Server) handleAZPaysWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	sigHeader := r.Header.Get("X-Azpays-Signature")
	if sigHeader == "" {
		sigHeader = r.Header.Get("Azpays-Signature")
	}

	var payload payment.WebhookPayload
	_ = json.Unmarshal(bodyBytes, &payload)

	if err := s.paySvc.VerifyWebhook(r.Context(), bodyBytes, sigHeader, payload); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleMockCheckout(w http.ResponseWriter, r *http.Request) {
	invoiceID := r.URL.Query().Get("invoice_id")
	w.Header().Set("Content-Type", "text/html")
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>AZPays Sandbox Checkout</title><style>body{background:#0f172a;color:#fff;font-family:sans-serif;padding:40px;text-align:center;} .btn{background:#3b82f6;color:#fff;padding:12px 24px;border:none;border-radius:8px;cursor:pointer;font-size:16px;margin:10px;}</style></head>
<body>
  <h1>AZPays Crypto Payment (Sandbox)</h1>
  <p>Invoice: <code>%s</code></p>
  <button class="btn" onclick="pay('paid')">✅ Simulate Successful Crypto Payment</button>
  <script>
    function pay(status) {
      fetch('/api/v1/payments/azpays/webhook', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({invoice_id: '%s', status: status, amount: 9.99, currency: 'USDT'})
      }).then(() => alert('Payment completed! Return to Telegram bot.')).catch(e => alert(e));
    }
  </script>
</body>
</html>`, invoiceID, invoiceID)
	_, _ = w.Write([]byte(html))
}

// Setup Status
func (s *Server) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	ctx := r.Context()
	isCompleted := s.settingsSvc.GetBool(ctx, settings.KeyOnboardingCompleted, false)
	stepVal := s.settingsSvc.GetString(ctx, settings.KeyOnboardingStep, "1")
	apiUrlVal := s.settingsSvc.GetString(ctx, settings.KeyAPIURL, "http://localhost:8080")

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"completed": isCompleted,
		"step":      stepVal,
		"api_url":   apiUrlVal,
	})
}

// Request context helpers
type contextKey string

const claimsContextKey contextKey = "jwt_claims"

func setContextClaims(ctx context.Context, claims *auth.Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

func getContextClaims(ctx context.Context) *auth.Claims {
	if val := ctx.Value(claimsContextKey); val != nil {
		if c, ok := val.(*auth.Claims); ok {
			return c
		}
	}
	return nil
}
