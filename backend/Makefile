
.PHONY: run build lint fmt test 

BINARY_NAME := api
MAIN_PATH := ./cmd/api/main.go

-include .env
export

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


