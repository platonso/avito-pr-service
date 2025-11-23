COMPOSE_FILE=docker-compose.yml

LOCAL_BIN:=$(CURDIR)/bin

build:
	docker-compose -f $(COMPOSE_FILE) build

up:
	docker-compose -f $(COMPOSE_FILE) up -d --build

down:
	docker-compose -f $(COMPOSE_FILE) down

restart: down up

install-deps:
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@latest
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

install-golangci-lint:
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint: install-golangci-lint
	$(LOCAL_BIN)/golangci-lint run ./... --config .golangci.yml

format:
	gofmt -w -s ./internal ./cmd

.PHONY: cover
cover:
	go test -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out

.PHONY: build up down restart install-deps install-golangci-lint lint lint-fix format cover