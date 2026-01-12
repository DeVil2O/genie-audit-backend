SHELL := /bin/bash

ENV_FILE ?= .env
GOOSE ?= goose
MIGRATIONS_DIR := ./migrations

# Resolve goose binary: prefer PATH, otherwise GOPATH/bin/goose.
GOOSE_BIN := $(or $(shell command -v $(GOOSE) 2>/dev/null),$(wildcard $(shell go env GOPATH 2>/dev/null)/bin/goose))

# Load variables from .env if present (simple KEY=VALUE, no spaces around =).
ifneq (,$(wildcard $(ENV_FILE)))
include $(ENV_FILE)
export $(shell sed -n 's/^[A-Za-z_][A-Za-z0-9_]*\\s*=.*/\\1/p' $(ENV_FILE))
endif

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  env-check        - ensure .env exists (copies .env.example if missing)"
	@echo "  migrate-up       - apply all migrations"
	@echo "  migrate-down     - rollback all migrations"
	@echo "  api              - run the API server"

.PHONY: env-check
env-check:
	@test -f .env || (cp .env.example .env && echo "Created .env from .env.example")

.PHONY: goose-check
goose-check:
	@if [ -z "$(GOOSE_BIN)" ]; then \
		echo "goose not found. Install via:"; \
		echo "  go install github.com/pressly/goose/v3/cmd/goose@latest"; \
		echo "and ensure $$GOPATH/bin (or ~/go/bin) is on your PATH"; \
		exit 127; \
	fi

.PHONY: migrate-up
migrate-up: env-check goose-check createdb-if-missing
	@. $(ENV_FILE); \
	if [ -z "$$DATABASE_URL" ]; then echo "DATABASE_URL not set"; exit 1; fi; \
	"$(GOOSE_BIN)" -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" up

.PHONY: migrate-down
migrate-down: env-check goose-check
	@. $(ENV_FILE); \
	if [ -z "$$DATABASE_URL" ]; then echo "DATABASE_URL not set"; exit 1; fi; \
	"$(GOOSE_BIN)" -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" down

.PHONY: api
api: env-check
	@. $(ENV_FILE); \
	go run ./cmd/api
