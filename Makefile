GO ?= go
MOCKERY ?= mockery
SWAG ?= swag
DOCKER_COMPOSE ?= docker compose
GOMODCACHE ?=
MOCKERY_GOCACHE ?= /tmp/taskflow-mockery-go-build
GOENV :=
GO_FILES := $(shell rg --files -g '*.go' .)
BIN_DIR := $(CURDIR)/bin

.PHONY: setup ensure-env infra-up infra-down migrate run run-worker tidy fmt test test-unit mocks swagger build build-api build-worker build-migrations clean

ifneq ($(strip $(GOMODCACHE)),)
GOENV += GOMODCACHE=$(GOMODCACHE)
endif

setup:
	$(MAKE) clean
	$(MAKE) ensure-env
	$(if $(strip $(GOMODCACHE)),mkdir -p $(GOMODCACHE),true)
	$(GOENV) $(GO) mod download
	$(GOENV) $(GO) install github.com/vektra/mockery/v2@v2.53.6
	$(MAKE) infra-up
	$(MAKE) migrate

ensure-env:
	test -f .env || cp .env.sample .env

infra-up:
	$(DOCKER_COMPOSE) up -d postgres redis

infra-down:
	$(DOCKER_COMPOSE) down

migrate:
	$(GOENV) $(GO) run ./cmd/postgres-migrations -up

run:
	$(GOENV) $(GO) run ./cmd/taskflow-api

run-worker:
	$(GOENV) $(GO) run ./cmd/taskflow-worker

tidy:
	$(GOENV) $(GO) mod tidy

fmt:
	gofmt -w $(GO_FILES)

test:
	$(GOENV) $(GO) test ./...

test-unit:
	$(GOENV) $(GO) test ./internal/service/...

mocks:
	mkdir -p mocks
	GOCACHE=$(MOCKERY_GOCACHE) $(MOCKERY) --config .mockery.yaml --dir internal/service --name TaskRepository --output mocks --outpkg mocks --filename task_repository.go --structname TaskRepository
	GOCACHE=$(MOCKERY_GOCACHE) $(MOCKERY) --config .mockery.yaml --dir internal/service --name TaskCache --output mocks --outpkg mocks --filename task_cache.go --structname TaskCache
	GOCACHE=$(MOCKERY_GOCACHE) $(MOCKERY) --config .mockery.yaml --dir internal/service --name UserRepository --output mocks --outpkg mocks --filename user_repository.go --structname UserRepository

swagger:
	$(SWAG) init -g cmd/taskflow-api/main.go -o docs

build: build-api build-worker build-migrations

build-api:
	mkdir -p $(BIN_DIR)
	$(GOENV) $(GO) build -o $(BIN_DIR)/taskflow-api ./cmd/taskflow-api

build-worker:
	mkdir -p $(BIN_DIR)
	$(GOENV) $(GO) build -o $(BIN_DIR)/taskflow-worker ./cmd/taskflow-worker

build-migrations:
	mkdir -p $(BIN_DIR)
	$(GOENV) $(GO) build -o $(BIN_DIR)/postgres-migrations ./cmd/postgres-migrations

clean:
	rm -rf $(BIN_DIR)
