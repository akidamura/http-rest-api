include .env
export

DB_DSN = "mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)?parseTime=true"

.PHONY: build
build:
	go build -v ./cmd/apiserver

.PHONY: test
test:
	go test -v -race -timeout 30s ./...

.PHONY: migrate-up
migrate-up:
	migrate -path migrations -database $(DB_DSN) up

.PHONY: migrate-down
migrate-down:
	migrate -path migrations -database $(DB_DSN) down

.PHONY: migrate-create
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

.DEFAULT_GOAL := build