package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/vyntechau/TelegramPublisher/config"
	"github.com/vyntechau/TelegramPublisher/internal/analytics"
	"github.com/vyntechau/TelegramPublisher/internal/api/graphql"
	"github.com/vyntechau/TelegramPublisher/internal/api/rest"
	"github.com/vyntechau/TelegramPublisher/internal/api/scalar"
	"github.com/vyntechau/TelegramPublisher/internal/auth"
	"github.com/vyntechau/TelegramPublisher/internal/bot"
	"github.com/vyntechau/TelegramPublisher/internal/services/marketing"
	"github.com/vyntechau/TelegramPublisher/internal/services/payment"
	"github.com/vyntechau/TelegramPublisher/internal/services/settings"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
	"github.com/vyntechau/TelegramPublisher/internal/storage/mysql"
	"github.com/vyntechau/TelegramPublisher/internal/storage/postgres"
	"github.com/vyntechau/TelegramPublisher/internal/storage/sqlite"
)

func main() {
	log.Println("==================================================")
	log.Println("⚡ TelegramPublisher - Open Source Bot & Mini App")
	log.Println("☁️ Official Sample Application for VynTech Cloud")
	log.Println("==================================================")

	// 1. Load Multi-Source Configuration (OS Env > .env > config.yaml)
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("[Config] Failed to load configuration: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("[Config] Configuration security validation failed: %v", err)
	}

	masked := cfg.Masked()
	log.Printf("[Config] Bootstrapping initialized in '%s' environment (Server: %s:%d, DB: %s)",
		masked.App.Env, masked.App.Host, masked.App.Port, masked.Database.Type)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Initialize Database Adapter (sqlite, postgres, mysql, mariadb)
	var repo storage.Repository
	dbType := strings.ToLower(cfg.Database.Type)

	switch dbType {
	case "sqlite", "sqlite3", "":
		log.Printf("[DB] Using SQLite database at: %s", cfg.Database.FilePath)
		repo = sqlite.New(cfg.Database.FilePath)
	case "postgres", "postgresql", "pgsql":
		log.Printf("[DB] Using PostgreSQL database: %s", masked.Database.URL)
		repo = postgres.New(cfg.Database)
	case "mysql", "mariadb":
		log.Printf("[DB] Using MySQL/MariaDB database: %s", masked.Database.URL)
		repo = mysql.New(cfg.Database)
	default:
		log.Fatalf("[DB] Unsupported database type: %s. Supported: sqlite, postgres, mysql, mariadb", cfg.Database.Type)
	}

	if err := repo.Init(ctx); err != nil {
		log.Fatalf("[DB] Database initialization failed: %v", err)
	}
	defer repo.Close()
	log.Println("[DB] Database connected and schema migrated successfully.")

	// 3. Initialize Core Domain Services (Settings & Payment use DB settings table)
	authSvc := auth.NewService(cfg.Bot.Token, cfg.App.JWTSecret, repo)
	settingsSvc := settings.NewService(repo, cfg)
	analyticsSvc := analytics.NewService(repo)

	// 4. Initialize Telegram Bot Engine (if bot token provided)
	var botEngine *bot.Engine
	var marketSvc *marketing.Service
	var paySvc *payment.Service

	if cfg.Bot.Token != "" {
		engine, err := bot.NewEngine(cfg, repo)
		if err != nil {
			log.Printf("[Bot] Warning: Could not initialize bot: %v", err)
			marketSvc = marketing.NewService(nil, repo)
			paySvc = payment.NewService(repo, settingsSvc, nil)
		} else {
			botEngine = engine
			marketSvc = engine.MarketSvc
			paySvc = engine.PaySvc
			go botEngine.Start()
		}
	} else {
		log.Println("[Bot] Notice: BOT_TOKEN is not set. Bot polling disabled, running in API/Web server mode.")
		marketSvc = marketing.NewService(nil, repo)
		paySvc = payment.NewService(repo, settingsSvc, nil)
	}

	// 5. Setup HTTP Router & APIs
	mux := http.NewServeMux()

	// REST API
	restServer := rest.NewServer(repo, authSvc, settingsSvc, analyticsSvc, marketSvc, paySvc, cfg)
	restServer.RegisterRoutes(mux)

	// Scalar API Interactive Documentation (/docs and /docs/openapi.json)
	scalar.Register(mux, "docs/openapi.json")

	// GraphQL API & GraphiQL Playground (/graphql)
	if err := graphql.Register(mux, repo, analyticsSvc, marketSvc); err != nil {
		log.Printf("[GraphQL] Schema registration warning: %v", err)
	}

	// Static Web App / Single Unified React App
	fs := http.FileServer(http.Dir("web/dist"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := "web/dist" + r.URL.Path
		if _, err := os.Stat(path); os.IsNotExist(err) && !strings.Contains(r.URL.Path, ".") {
			http.ServeFile(w, r, "web/dist/index.html")
			return
		}
		fs.ServeHTTP(w, r)
	})

	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, x-backend-url")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		mux.ServeHTTP(w, r)
	})

	serverAddr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	httpServer := &http.Server{
		Addr:         serverAddr,
		Handler:      corsHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("[HTTP] Web Server running at http://%s", serverAddr)
		log.Printf("[HTTP] 📚 Scalar API Docs: http://%s/docs", serverAddr)
		log.Printf("[HTTP] ⚡ GraphQL Playground: http://%s/graphql", serverAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[HTTP] Server failed: %v", err)
		}
	}()

	// 6. Graceful Shutdown Listener
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan

	log.Println("[Server] Shutting down gracefully...")
	if botEngine != nil {
		botEngine.Stop()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = httpServer.Shutdown(shutdownCtx)

	log.Println("[Server] TelegramPublisher stopped.")
}
