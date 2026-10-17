# ─────────────────────────────────────
# Этап 1: сборка
# ─────────────────────────────────────
FROM golang:1.27-alpine AS builder

WORKDIR /build

# Сначала только модули — для кеширования слоёв
COPY go.mod go.sum ./
RUN go mod download

# Потом исходники
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Сборка статического бинарника
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /build/app \
    ./cmd/apiserver

# ─────────────────────────────────────
# Этап 2: запуск
# ─────────────────────────────────────
FROM alpine:latest

# Сертификаты для HTTPS-запросов
RUN apk --no-cache add ca-certificates

# Не запускать от root
RUN adduser -D -u 1000 appuser
USER appuser

WORKDIR /app

# Копируем только бинарник из builder
COPY --from=builder --chown=appuser:appuser /build/app /app/app

EXPOSE 8080 50051

CMD ["/app/app"]