include .env
export

DB_DSN_DEV = "mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)?parseTime=true"
DB_DSN_TEST = "mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/restapi_test?parseTime=true"

# ─────────────────────────────────────
# Сборка и запуск
# ─────────────────────────────────────
.PHONY: build
build:
	go build -v ./cmd/apiserver

.PHONY: start
start:
	./apiserver

.PHONY: run
run: migrate-up build
	./apiserver

.PHONY: clean
clean:
	rm -f apiserver coverage.out cookies.txt .cookies

# ─────────────────────────────────────
# Тестирование
# ─────────────────────────────────────
# Запустить вообще ВСЕ тесты (включая интеграционные с БД)
.PHONY: test
test:
	go test ./... -v -race

# Покрытие (запускает всё)
.PHONY: test-cover
test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# Только Unit-тесты (БЕЗ базы данных) — быстро, можно запускать локально
.PHONY: test-units
test-units:
	go test ./internal/app/model/... ./internal/app/apiserver/... ./internal/app/grpcserver/... -v -race

# Конкретно gRPC тесты
.PHONY: test-grpc
test-grpc:
	go test ./internal/app/grpcserver/... -v -race

# Конкретно HTTP тесты
.PHONY: test-http
test-http:
	go test ./internal/app/apiserver/... -v -race

# Только интеграционные тесты (ТРЕБУЕТ запущенную MySQL!)
.PHONY: test-integrations
test-integrations:
	go test ./internal/app/store/... -v -race

# Smoke-тесты (ТРЕБУЕТ запущенный сервер)
.PHONY: smoke
smoke:
	./scripts/smoke.sh

# ─────────────────────────────────────
# Миграции
# ─────────────────────────────────────
.PHONY: migrate-up
migrate-up:
	migrate -path migrations -database $(DB_DSN_DEV) up

.PHONY: migrate-up-test
migrate-up-test:
	migrate -path migrations -database $(DB_DSN_TEST) up

.PHONY: migrate-down
migrate-down:
	migrate -path migrations -database $(DB_DSN_DEV) down

.PHONY: migrate-down-test
migrate-down-test:
	migrate -path migrations -database $(DB_DSN_TEST) down

.PHONY: migrate-create
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

# ─────────────────────────────────────
# Docker
# ─────────────────────────────────────
.PHONY: docker-up
docker-up:
	docker compose up --build

.PHONY: docker-down
docker-down:
	docker compose down

.DEFAULT_GOAL := build