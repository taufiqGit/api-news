# ============ STAGE 1: BUILD ============
FROM golang:1.26-alpine AS builder

# Tooling yang dibutuhkan untuk build
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache dependency terlebih dahulu (optimasi layer Docker)
COPY go.mod go.sum ./
RUN go mod download

# Copy source & build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/news-api ./cmd/server

# ============ STAGE 2: RUNTIME ============
FROM alpine:3.20

# CA certificates untuk HTTPS outbound & user non-root
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S app -G app

WORKDIR /app

# Copy binary dari builder
COPY --from=builder /app/bin/news-api /app/news-api

# Pastikan user non-root
USER app

EXPOSE 8080

ENTRYPOINT ["/app/news-api"]
