.DEFAULT_GOAL := help
GO ?= go
# Go 1.26 is shared with the dev container and lint toolchain.
export GOTOOLCHAIN ?= go1.26.8
GOLANGCI_LINT ?= golangci-lint
GOLANGCI_VERSION := v2.13.2
AIR_VERSION := v1.64.5
COMPOSE ?= docker compose

# GNU make on Windows uses cmd.exe by default, but this Makefile contains
# POSIX recipes. Use Git Bash when available and avoid id(1), which is not
# available in cmd.exe. Docker only needs a stable non-root uid/gid for the
# mounted migration directory on Windows.
ifeq ($(OS),Windows_NT)
SHELL := C:/Program Files/Git/bin/bash.exe
BASH := C:/Program Files/Git/bin/bash.exe
GOARCH ?= amd64
export LOCAL_UID ?= 1000
export LOCAL_GID ?= 1000
DBMATE = "$(BASH)" scripts/with-env.sh "$(BASH)" scripts/dbmate.sh
else
GOARCH ?= $(shell $(GO) env GOARCH)
export LOCAL_UID ?= $(shell id -u)
export LOCAL_GID ?= $(shell id -g)
DBMATE = sh scripts/with-env.sh sh scripts/dbmate.sh
endif
PROD_COMPOSE = $(COMPOSE) -f docker-compose.prod.yml
export DBMATE_TARGET ?= local
export DBMATE_VERSION ?= 2.16.0
export DBMATE_NETWORK ?= --network=host

.PHONY: help run dev build tidy fmt fmt-check vet lint lint-fix check test tests test-integration tools-install docker-build compose-local compose-up compose-down compose-logs compose-prod compose-prod-down compose-supabase compose-supabase-down dbmate migrate clean

help: ## Show available commands
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "%-22s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run API locally, loading .env
	@sh scripts/with-env.sh $(GO) run ./cmd/bakery

dev: ## Run locally with Air hot reload, loading .env
	@sh scripts/with-env.sh air -c .air.toml

build: ## Build static Linux binary for the production image
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) $(GO) build -trimpath -buildvcs=false -o build/bakery ./cmd/bakery

tidy: ## Update module metadata
	$(GO) mod tidy

fmt: ## Format Go sources
	$(GOLANGCI_LINT) fmt

fmt-check: ## Check formatting without changing files
	@format_diff=$$(mktemp); trap 'rm -f "$$format_diff"' EXIT; \
	$(GOLANGCI_LINT) fmt --diff > "$$format_diff" || exit $$?; \
	if [ -s "$$format_diff" ]; then cat "$$format_diff"; exit 1; fi

vet: ## Run go vet
	$(GO) vet ./...

lint: fmt-check ## Run linters and architecture dependency checks
	$(GOLANGCI_LINT) run

lint-fix: ## Apply supported formatting and lint fixes
	$(GOLANGCI_LINT) fmt
	$(GOLANGCI_LINT) run --fix

check: lint test build ## Run lint, unit tests and build

test: ## Run tests without connecting to an external database
	TEST_DATABASE_URL= $(GO) test ./...

tests: test ## Alias matching shipment-service

test-integration: ## Run DB tests using TEST_DATABASE_URL (test database only)
	@sh scripts/with-env.sh sh -eu -c ': "$${TEST_DATABASE_URL:?Set TEST_DATABASE_URL to a migrated test database}"; exec $(GO) test ./internal/adapters/outbound/db/product -count=1 -v'

tools-install: ## Install pinned lint and hot reload tools in GOPATH/bin
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
	$(GO) install github.com/air-verse/air@$(AIR_VERSION)

docker-build: build ## Package production image, like shipment-service
	docker build -t bakery:local .

compose-local: ## Start development container with hot reload (foreground)
	$(COMPOSE) up --build

compose-up: ## Start development container in background
	$(COMPOSE) up --build -d

compose-down: ## Stop development containers, preserving database and caches
	$(COMPOSE) down

compose-logs: ## Follow API logs
	$(COMPOSE) logs -f bakery

compose-prod: build ## Build and start production container in background
	$(PROD_COMPOSE) up --build -d

compose-prod-down: ## Stop production container
	$(PROD_COMPOSE) down

compose-supabase: ## Start development API connected to Supabase via .env
	$(COMPOSE) -f docker-compose.supabase.yml up --build -d

compose-supabase-down: ## Stop Supabase development API
	$(COMPOSE) -f docker-compose.supabase.yml down

ifeq ($(OS),Windows_NT)
dbmate: ## Run dbmate in Docker, e.g. make dbmate ARGS="up"
	@if [ -z "$(ARGS)" ]; then \
		echo 'Missing argument: make dbmate ARGS="up" or ARGS="new <migration_name>"'; \
		exit 1; \
	fi
	@$(DBMATE) $(ARGS)
else
dbmate: ## Run dbmate in Docker, e.g. make dbmate ARGS="up"
	@if [ -z "$(ARGS)" ]; then \
		echo 'Missing argument: make dbmate ARGS="up" or ARGS="new <migration_name>"'; \
		exit 1; \
	fi
	@$(DBMATE) $(ARGS)
endif

migrate: ## Alias for make dbmate ARGS="up"
	@$(MAKE) --no-print-directory dbmate ARGS="up"

clean: ## Remove generated binaries and test outputs
	rm -rf build tmp coverage
