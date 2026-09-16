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

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 10001 -S appgroup && \
    adduser -u 10001 -S appuser -G appgroup

# Copy binary & static assets
COPY --from=go-builder --chown=10001:10001 /app/bin/publisher /app/publisher
COPY --from=web-builder --chown=10001:10001 /app/web/dist /app/web/dist
COPY --chown=10001:10001 config/config.example.yaml /app/config/config.example.yaml
COPY --chown=10001:10001 docs/openapi.json /app/docs/openapi.json

# Create persistent data directory and ensure non-root ownership
RUN mkdir -p /app/data && \
    chown -R 10001:10001 /app

USER 10001:10001

EXPOSE 8080

ENTRYPOINT ["/app/publisher"]
