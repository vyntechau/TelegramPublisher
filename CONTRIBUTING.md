# Contributing to TelegramPublisher

Thank you for your interest in contributing to **TelegramPublisher**! We welcome open-source contributions from developers worldwide.

## Development Setup

1. **Prerequisites**:
   - Go 1.25+
   - Node.js 20+
   - Docker (optional for multi-database testing)

2. **Clone & Install**:
   ```bash
   git clone https://github.com/vyntechau/TelegramPublisher.git
   cd TelegramPublisher
   go mod tidy
   cd web && npm install && npm run build && cd ..
   ```

3. **Running the Application**:
   ```bash
   # Copy sample configuration
   cp config/config.example.yaml config.yaml
   cp .env.example .env

   # Run backend
   go run ./cmd/bot
   ```

4. **Running Unit Tests**:
   ```bash
   go test -v ./...
   ```

## Pull Request Guidelines
- Follow Go clean code conventions and formatting (`go fmt ./...`).
- Ensure all automated unit tests pass.
- Submit PRs against the `main` branch with descriptive summaries.
