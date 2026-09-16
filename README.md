# ⚡ TelegramPublisher

<div align="center">

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react)](https://react.dev)
[![Test Coverage](https://img.shields.io/badge/Coverage-100%25%20Statements-brightgreen?style=flat&logo=go)](https://github.com/vyntechau/TelegramPublisher)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-3.1-6BA539?style=flat&logo=openapiinitiative)](https://openapis.org)
[![Scalar Docs](https://img.shields.io/badge/API%20Docs-Scalar-6366f1)](http://localhost:8080/docs)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

**Enterprise-grade Open Source Telegram media delivery, copyright timer protection, and content publisher platform.**  
*Modular microservice architecture built with Go 1.25+, React 18, and Vite.*

[Key Features](#-key-features) • [Architecture](#-architecture) • [Test Coverage](#-test-coverage--quality-assurance) • [Quick Start](#-quick-start) • [Cloud Deploy (~$1/mo)](#%EF%B8%8F-option-1-deploy-on-vyntech-cloud-1-usd--month) • [API & Docs](#-api--documentation) • [Contributing](#-contributing) • [Security](#-security--vulnerability-reporting)

</div>

---

## 📖 Overview

**TelegramPublisher** is an open-source, production-ready content publishing platform and Telegram bot infrastructure designed for media distribution, automated copyright protection, and digital subscription monetization. 

It solves media expiration challenges by combining **server-side background TTL purge workers**, **cryptographic payment gateways**, a **Telegram Mini App (TMA)**, and a **dual REST + GraphQL API layer** with real-time analytics.

---

## 🌟 Key Features

### ⏳ Automated Copyright Protection (Self-Destruct TTL)
- Schedules automatic deletion of sensitive or copyrighted media messages from chats after a configurable countdown (default: 120 seconds).
- Zero-leakage background cleaner worker continuously purges expired content.
- VIP subscribers automatically receive exemptions for permanent media access.

### 🌐 7-Language Internationalization (i18n) & Native RTL
- Full multi-language support: **English (en)**, **Persian (fa)**, **Arabic (ar)**, **Russian (ru)**, **Spanish (es)**, **German (de)**, and **Chinese (zh)**.
- First-class Right-to-Left (**RTL**) layout rendering for Persian and Arabic.
- Dynamic language switcher inside Telegram bot keyboards and Mini App UI.

### 💳 Multi-Provider Crypto Subscriptions
- Built-in integrations for **AzPays**, **Coinbase Commerce**, and **NOWPayments**.
- Instant subscription checkout for **USDT**, **TON**, **BTC**, **ETH**, and other cryptocurrencies.
- Webhook signature verification (HMAC-SHA256) and simulated developer sandbox checkout.
- Automated tier upgrades (VIP 30-Day, Author Pro 30-Day, Lifetime VIP Pass).

### 📱 Unified Role-Based Web App (React + Vite + Tailwind)
- **User (Telegram Mini App)**: Media catalog explorer, audio/video streaming, like/dislike reactions, broken file reporting, VIP pass checkout.
- **Author (Publisher Studio)**: Upload media, inspect Telegram `file_id`s & hashes, monitor personal post views, and track ticket resolutions.
- **Admin & Owner (Intelligence Suite)**: Executive KPI dashboards, moderation (ban/unban/role promotion), force-sub channels, ticket SLA queues, and runtime toggle settings.

### 📚 Dual API Layer & Interactive Docs
- **REST API (OpenAPI 3.1)**: Fully documented endpoints rendered interactively with modern **Scalar UI** at `/docs`.
- **GraphQL API**: Rich GraphQL schema and **GraphiQL Playground** at `/graphql`.

### 📢 Targeted Marketing & Safe Broadcaster
- Dispatch messages formatted as Plain Text, MarkdownV2, HTML, or **Raw JSON Telegram Payloads** (custom inline keyboards, WebApp buttons, media groups).
- Granular audience filters: All users, active within last 7 days, specific roles, active VIP status.
- **Automated Block Detection**: Identifies Telegram 403 Forbidden errors and marks user profiles as `blocked_by_user` to preserve bot deliverability.

### 🛡️ Force-Subscription Channel Gatekeeper
- Validates user membership in mandatory Telegram channels before unlocking protected media links.
- Seamless inline verification workflow with *"🔄 Check Membership"* retry buttons.

### 🚩 Content Reporting & SLA Resolution
- Built-in reporting system allowing users to flag broken media links or copyright concerns directly to publisher queues.

---

## 📐 Architecture

```
TelegramPublisher/
├── cmd/bot/main.go                 # Unified Server Entrypoint (Bot + HTTP + React SPA)
├── config/                         # Multi-source configuration (.env, YAML, ENV vars)
├── docs/openapi.json               # OpenAPI 3.1 Schema Specification
├── internal/
│   ├── analytics/                  # Metrics aggregation & CSV export engine
│   ├── api/
│   │   ├── graphql/                # GraphQL schema & GraphiQL playground (/graphql)
│   │   ├── rest/                   # REST API controllers (/api/v1/...)
│   │   └── scalar/                 # Modern Scalar Docs renderer (/docs)
│   ├── auth/                       # Telegram WebApp initData HMAC-SHA256 & JWT auth
│   ├── bot/                        # Telegram Bot engine, RBAC & command handlers
│   ├── cleaner/                    # 2-Minute copyright auto-delete worker
│   ├── i18n/                       # 7-Language dictionary & RTL localization engine
│   ├── services/
│   │   ├── marketing/              # Targeted broadcast engine & block detector
│   │   ├── payment/                # Crypto gateway service (AzPays, Coinbase, NOWPayments)
│   │   └── settings/               # Dynamic feature toggles & runtime settings
│   └── storage/                    # Repository interfaces & database adapters
│       ├── factory.go              # Database engine factory
│       ├── sqlite/                 # SQLite (Pure Go CGO-free via modernc.org/sqlite)
│       ├── postgres/               # PostgreSQL 14+ via pgx/v5
│       └── mysql/                  # MySQL 8+ & MariaDB
├── web/                            # Modern React 18 + Vite + Tailwind CSS Frontend
│   └── src/
│       ├── views/UserMiniApp.tsx   # Telegram Mini App
│       ├── views/AuthorStudio.tsx  # Publisher Studio & File Inspector
│       └── views/AdminDashboard.tsx# Admin Intelligence Suite
├── deploy/
│   ├── docker-compose.yml          # Multi-database local testing
│   └── vyntech-cloud.yaml          # Production cloud deployment manifest
├── Dockerfile                      # Production multi-stage container build
├── CONTRIBUTING.md                 # Open source contributor guide
├── CODE_OF_CONDUCT.md              # Community standards
├── SECURITY.md                     # Vulnerability reporting policy
└── LICENSE                         # MIT License
```

---

## 🧪 Test Coverage & Quality Assurance

TelegramPublisher maintains rigorous testing standards with **100% statement coverage** across all active core packages and API layers:

| Package | Path | Statement Coverage | Status |
| :--- | :--- | :---: | :---: |
| **API REST Layer** | `internal/api/rest` | **100.0%** | PASS |
| **API GraphQL Layer** | `internal/api/graphql` | **100.0%** | PASS |
| **API Scalar Documentation** | `internal/api/scalar` | **100.0%** | PASS |
| **Crypto Payment Service** | `internal/services/payment` | **100.0%** | PASS |
| **Marketing & Broadcaster** | `internal/services/marketing` | **100.0%** | PASS |
| **Settings & Toggles** | `internal/services/settings` | **100.0%** | PASS |
| **Cleaner Worker (TTL)** | `internal/cleaner` | **100.0%** | PASS |
| **Storage (SQLite Engine)** | `internal/storage/sqlite` | **100.0%** | PASS |
| **Storage Engine Factory** | `internal/storage` | **100.0%** | PASS |
| **Localization & RTL (i18n)** | `internal/i18n` | **100.0%** | PASS |
| **Telegram & JWT Auth** | `internal/auth` | **100.0%** | PASS |
| **Analytics Engine** | `internal/analytics` | **100.0%** | PASS |
| **Configuration Loader** | `config` | **100.0%** | PASS |

### Running the Test Suite

```bash
# Run all unit tests
go test -v ./...

# Run tests with statement coverage summary
go test -cover ./internal/... ./config/...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./internal/api/... ./internal/services/... ./internal/storage/...
go tool cover -html=coverage.out -o coverage.html
```

---

## 🚀 Quick Start

### ☁️ Option 1: Deploy on VynTech Cloud (~$1 USD / month)

Deploy TelegramPublisher on **[VynTech Cloud](https://vyntech.cloud)** starting from as low as **~$1 USD per month**:
- 🚀 **High Performance NVMe Storage**: Dedicated high-speed persistence for media metadata and database storage.
- 🔒 **Automated SSL & Ingress**: Pre-configured TLS certificates, DDoS mitigation, and custom domain routing.
- ⚡ **Global Edge Network**: Ultra low-latency content delivery for Telegram Mini Apps and APIs worldwide.
- 📦 **Production Ready**: One-click deploy using the included cloud manifest [`deploy/vyntech-cloud.yaml`](deploy/vyntech-cloud.yaml) or connect your repository via the VynTech Cloud Console.

### 💻 Option 2: Run Locally

#### Prerequisites
- **Go**: Version 1.25 or higher
- **Node.js**: Version 20 or higher
- **Telegram Bot Token**: Obtained from [@BotFather](https://t.me/botfather)

```bash
# 1. Clone repository
git clone https://github.com/vyntechau/TelegramPublisher.git
cd TelegramPublisher

# 2. Build the React frontend
cd web
npm install
npm run build
cd ..

# 3. Configure environment
cp .env.example .env

# Edit .env and supply your TELEGRAM_BOT_TOKEN and credentials
# 4. Launch backend and bot
go run ./cmd/bot
```

### 🐳 Option 3: Run with Docker Compose

```bash
# Start with default SQLite storage (zero external dependencies)
docker compose -f deploy/docker-compose.yml up --build

# Or start with PostgreSQL
docker compose -f deploy/docker-compose.yml --profile postgres up --build

# Or start with MySQL
docker compose -f deploy/docker-compose.yml --profile mysql up --build
```

---

## 📚 API & Documentation

Once the server is running, the following endpoints are available:

| Service | Endpoint | Description |
|---|---|---|
| **Web App & Mini App** | `http://localhost:8080/` | Unified Role-Based React Single Page Application |
| **Scalar Interactive Docs** | `http://localhost:8080/docs` | Interactive OpenAPI 3.1 Documentation & Request Runner |
| **GraphQL Playground** | `http://localhost:8080/graphql` | GraphiQL Interactive Query & Mutation IDE |
| **OpenAPI Spec (JSON)** | `http://localhost:8080/docs/openapi.json` | Raw OpenAPI 3.1 JSON Schema specification |
| **Health Check** | `http://localhost:8080/api/v1/setup/status` | Node status, database readiness, and onboarding state |
| **Payment Sandbox** | `http://localhost:8080/api/v1/payments/mock-checkout` | Simulated sandbox checkout for local testing |

---

## 🗄️ Multi-Database Support

The platform utilizes a clean repository pattern. Choose your target database by setting `DB_TYPE`:

| Engine | `DB_TYPE` | Configuration Example |
|---|---|---|
| **SQLite** | `sqlite` | `DB_FILE_PATH=data/publisher.db` *(Default, CGO-free pure Go)* |
| **PostgreSQL** | `postgres` | `DB_DSN=postgres://user:password@localhost:5432/dbname?sslmode=disable` |
| **MySQL / MariaDB** | `mysql` | `DB_DSN=user:password@tcp(localhost:3306)/dbname?parseTime=true` |

---

## 🤖 Telegram Bot Commands

| Command | Allowed Roles | Description |
|---|---|---|
| `/start` | All Users | Launch bot, select language, and open Telegram Mini App |
| `/start <slug>` | All Users | Retrieve protected media link with auto-delete timer |
| `/subscribe` | All Users | Purchase VIP subscription pass via crypto checkout |
| `/mystatus` | All Users | View active subscription tier and account status |
| `/language` | All Users | Switch preferred language (En, Fa, Ar, Ru, Es, De, Zh) |
| `/help` | All Users | Display user command guide |
| *Send Media File* | Author / Admin | Inspect Telegram `file_id` & generate shareable link |
| `/stats` | Admin / Owner | Real-time analytics, daily active users, revenue, view counts |
| `/users` | Admin / Owner | Browse registered bot user profiles |
| `/ban <id>` / `/unban <id>` | Admin / Owner | Ban or restore user access |
| `/promote <id> [role]` | Admin / Owner | Change user role (`author`, `admin`) |
| `/addchannel <id> <link>` | Admin / Owner | Register mandatory Force-Subscription channel |
| `/reports` | Admin / Owner | View and resolve broken media reports |
| `/broadcast <text>` | Admin / Owner | Dispatch safe broadcast message to active users |
| `/setttl <seconds>` | Admin / Owner | Update default auto-delete countdown timer |

---

## 🤝 Contributing

We love contributions from the open-source community! Here is how to get started:

1. Read our **[Contributing Guidelines](CONTRIBUTING.md)** and **[Code of Conduct](CODE_OF_CONDUCT.md)**.
2. Fork the repository and create your branch: `git checkout -b feature/amazing-feature`.
3. Commit your changes: `git commit -m "feat: add amazing feature"`.
4. Ensure all tests pass: `go test ./...` and `npm run test` (or `npm run build`).
5. Open a Pull Request against the `main` branch.

Please review our **[Security Policy](SECURITY.md)** for reporting vulnerabilities.

---

## 🔒 Security & Vulnerability Reporting

We take the security of **TelegramPublisher** and the safety of our users very seriously.

If you identify a security vulnerability, bug, or potential exploit, please **do NOT report it via public GitHub issues**.

Instead, please report it privately and responsibly to:
📧 **`security@vyntech.com.au`**

- Please provide detailed reproduction steps, sample payloads, and affected versions.
- All valid security reports will be acknowledged within 24 hours, followed by prompt investigation and patched releases.
- For complete policy details and supported versions, please consult **[SECURITY.md](SECURITY.md)**.

---

## ⚠️ Disclaimer

This software is published as an educational open-source project and reference implementation for microservices and cloud infrastructure. Please refer to **[DISCLAIMER.md](DISCLAIMER.md)** for detailed legal and compliance information.

---

## 📄 License

This project is licensed under the **[MIT License](LICENSE)**.
