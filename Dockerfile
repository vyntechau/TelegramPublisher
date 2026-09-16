# ==========================================
# Stage 1: Build Frontend Web Assets
# ==========================================
FROM node:20-alpine AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# ==========================================
# Stage 2: Build Go Backend Binary
# ==========================================
FROM golang:1.25-alpine AS go-builder
WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/publisher ./cmd/bot

# ==========================================
# Stage 3: Minimal Production Image
# ==========================================
FROM alpine:3.21
WORKDIR /app

# OCI Image Annotations (Links container image to GitHub Repository & Packages)
LABEL org.opencontainers.image.title="TelegramPublisher" \
      org.opencontainers.image.description="Enterprise-grade Open Source Telegram media delivery, copyright timer protection, and content publisher platform." \
      org.opencontainers.image.url="https://github.com/vyntechau/TelegramPublisher" \
      org.opencontainers.image.source="https://github.com/vyntechau/TelegramPublisher" \
      org.opencontainers.image.documentation="https://github.com/vyntechau/TelegramPublisher#readme" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="VynTech AU"

RUN apk add --no-cache ca-certificates tzdata

# Copy binary & static assets
COPY --from=go-builder /app/bin/publisher /app/publisher
COPY --from=web-builder /app/web/dist /app/web/dist
COPY config/config.example.yaml /app/config/config.example.yaml
COPY docs/openapi.json /app/docs/openapi.json

# Create persistent data directory
RUN mkdir -p /app/data

EXPOSE 8080

ENTRYPOINT ["/app/publisher"]
