# Project variables
APP_NAME := services-api
PKG := ./...
BIN := bin/api
IMAGE := ghcr.io/<your-org>/services-api:$(shell git rev-parse --short HEAD)
PORT ?= 8080

# Tools
GOCMD := go
LINT := golangci-lint
GOOSE := goose
SWAG := swag

.PHONY: all build run dev test lint fmt tidy clean docker docker-run docker-push migrate-up migrate-down seed coverage ci docs

all: build

## Build (local)
build: docs
	$(GOCMD) build -trimpath -o $(BIN) ./cmd/api

## Run (local, requires MySQL running)
run:
	MYSQL_DSN?=app:app@tcp(127.0.0.1:3306)/servicesdb?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_ai_ci
	PORT=$(PORT) ./$(BIN)

## Dev: compose up db + api (builds image)
dev:
	docker compose up --build

## Tests
test:
	$(GOCMD) test -race -count=1 $(PKG)

coverage:
	$(GOCMD) test -race -coverprofile=coverage.out -covermode=atomic $(PKG)
	$(GOCMD) tool cover -func=coverage.out | tail -n 1

## Lint
lint:
	$(LINT) run

## Format & tidy
fmt:
	gofumpt -w .
	$(GOCMD) fmt $(PKG)
	test -f go.mod && $(GOCMD) mod tidy

tidy:
	$(GOCMD) mod tidy

## Clean
clean:
	rm -rf bin coverage.out

## Docker build & run
docker:
	docker build -f build/docker/Dockerfile -t $(IMAGE) .

docker-run:
	docker run --rm -p $(PORT):8080 -e PORT=$(PORT) $(IMAGE)

## Push image (requires GHCR login)
docker-push:
	docker push $(IMAGE)

## DB migrations (using goose; set DB_DSN if different)
DB_DSN ?= app:app@tcp(127.0.0.1:3306)/servicesdb?parseTime=true&multiStatements=true
migrate-up:
	$(GOOSE) -dir ./migrations mysql "$(DB_DSN)" up

migrate-down:
	$(GOOSE) -dir ./migrations mysql "$(DB_DSN)" down

seed:
	mysql --protocol tcp -u app -papp -h 127.0.0.1 -P 3306 servicesdb < migrations/0002_demo_seed.sql

## Generate Swagger documentation
docs:
	$(SWAG) init -g cmd/api/main.go -o docs

## CI entrypoint
ci: fmt lint test coverage docs
