GO ?= go
MOCKERY ?= mockery
GOMODCACHE ?=
GOENV :=
GO_FILES := $(shell rg --files -g '*.go' .)
BIN_DIR := $(CURDIR)/bin

.PHONY: setup tidy fmt test test-unit mocks build build-api build-worker build-migrations clean

ifneq ($(strip $(GOMODCACHE)),)
GOENV += GOMODCACHE=$(GOMODCACHE)
endif

setup:
	$(if $(strip $(GOMODCACHE)),mkdir -p $(GOMODCACHE),true)
	$(GOENV) $(GO) mod download
	$(GOENV) $(GO) install github.com/vektra/mockery/v2@latest

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
	$(MOCKERY) --dir internal/service --name TaskRepository --output mocks --outpkg mocks --filename task_repository.go --structname TaskRepository
	$(MOCKERY) --dir internal/service --name UserRepository --output mocks --outpkg mocks --filename user_repository.go --structname UserRepository

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
