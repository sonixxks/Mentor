
.PHONY: run build lint fmt test db-up db-down db-logs db-reset migrate-up migrate-down migrate-status migrate-create

BINARY_NAME := api
MAIN_PATH := ./cmd/api/main.go

GOOSE := go run github.com/pressly/goose/v3/cmd/goose@v3.26.0
MIGRATIONS_DIR := ./migrations

-include .env
export

GOOSE_DBSTRING = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

run:
	go run $(MAIN_PATH)

build:
	go build -o ./bin/$(BINARY_NAME) $(MAIN_PATH)

lint:
	golangci-lint run 

fmt:
	golangci-lint fmt 

test:
	go test -v -race ./...

#Docker
db-up:
	docker-compose up -d
db-down:
	docker-compose down
db-logs:
	docker compose logs postgres -f
db-reset:
	docker compose down -v
	docker compose up -d

#Migrations
migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(GOOSE_DBSTRING)" up
migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(GOOSE_DBSTRING)" down
migrate-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(GOOSE_DBSTRING)" status
migrate-create:
	$(GOOSE) -dir $(MIGRATIONS_DIR) -s create $(name) sql
