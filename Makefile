include .env
export

DB_DSN_DEV = "mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)?parseTime=true"
DB_DSN_TEST = "mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/restapi_test?parseTime=true"

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

.PHONY: test
test:
	go test ./... -v -race

.PHONY: test-cover
test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

.PHONY: smoke
smoke:
	./scripts/smoke.sh

.PHONY: test-auth
test-auth:
	go test ./internal/app/apiserver/... -v -run "TestAuthenticateUser|TestHandleWhoami"

.PHONY: test-integrations
test-integrations:
	go test ./internal/app/store/... -v -race

.PHONY: test-units
test-units:
	go test ./internal/app/model/... -v -race

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

.DEFAULT_GOAL := build