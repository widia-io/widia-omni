-include .env
export

.PHONY: dev dev-api dev-worker dev-web dev-all \
       stop-api stop-web stop-dev restart-api restart-web restart-dev \
       build build-api build-worker build-cli build-web \
       release-cli release-cli-snapshot \
       test lint \
       migrate-up migrate-down sqlc-gen \
       infra-up infra-down infra-status \
       stripe-listen stripe-webhook-secret \
       obs-up obs-down obs-status \
       install clean check

PORT ?= 8080
WEB_PORT ?= 5173
STRIPE_EVENTS ?= checkout.session.completed,customer.subscription.created,customer.subscription.updated,customer.subscription.deleted
STRIPE_FORWARD_URL ?= http://localhost:$(PORT)/webhooks/stripe
PROJECT_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
WEB_DIR := $(PROJECT_ROOT)/web
GORELEASER := $(shell command -v goreleaser 2>/dev/null || echo "go run github.com/goreleaser/goreleaser/v2@latest")

# ─── Development ─────────────────────────────────────────────

dev: dev-api ## Run API server (default)

dev-api: ## Run API server
	cd $(PROJECT_ROOT) && go run ./cmd/api

dev-worker: ## Run background worker
	cd $(PROJECT_ROOT) && go run ./cmd/worker

dev-web: ## Run frontend dev server
	cd $(WEB_DIR) && npm run dev -- --port $(WEB_PORT)

dev-all: ## Run API + worker + web concurrently
	@trap 'kill 0' EXIT; \
	$(MAKE) dev-api & \
	$(MAKE) dev-worker & \
	$(MAKE) dev-web & \
	wait

stop-api: ## Stop process listening on API port
	@pids=$$(lsof -tiTCP:$(PORT) -sTCP:LISTEN 2>/dev/null || true); \
	if [ -n "$$pids" ]; then \
		echo "Stopping API port $(PORT): $$pids"; \
		kill $$pids 2>/dev/null || true; \
		sleep 1; \
		pids=$$(lsof -tiTCP:$(PORT) -sTCP:LISTEN 2>/dev/null || true); \
		if [ -n "$$pids" ]; then \
			echo "Force stopping API port $(PORT): $$pids"; \
			kill -9 $$pids 2>/dev/null || true; \
		fi; \
	else \
		echo "No process on API port $(PORT)"; \
	fi

stop-web: ## Stop process listening on frontend port
	@pids=$$(lsof -tiTCP:$(WEB_PORT) -sTCP:LISTEN 2>/dev/null || true); \
	if [ -n "$$pids" ]; then \
		echo "Stopping WEB port $(WEB_PORT): $$pids"; \
		kill $$pids 2>/dev/null || true; \
		sleep 1; \
		pids=$$(lsof -tiTCP:$(WEB_PORT) -sTCP:LISTEN 2>/dev/null || true); \
		if [ -n "$$pids" ]; then \
			echo "Force stopping WEB port $(WEB_PORT): $$pids"; \
			kill -9 $$pids 2>/dev/null || true; \
		fi; \
	else \
		echo "No process on WEB port $(WEB_PORT)"; \
	fi

stop-dev: stop-api stop-web ## Stop local API and frontend listeners

restart-api: stop-api dev-api ## Restart API on configured port

restart-web: stop-web dev-web ## Restart frontend on configured port

restart-dev: stop-dev dev-all ## Stop ports and run API + worker + web

# ─── Build ───────────────────────────────────────────────────

build: build-api build-worker build-cli build-web ## Build everything

build-api: ## Build API binary
	go build -o bin/api ./cmd/api

build-worker: ## Build worker binary
	go build -o bin/worker ./cmd/worker

build-cli: ## Build CLI binary
	go build -o bin/meufoco ./cmd/cli

build-web: ## Build frontend
	cd web && npm run build

release-cli: ## Build and publish CLI release artifacts to GitHub (requires tags + token)
	$(GORELEASER) release --clean

release-cli-snapshot: ## Build CLI release assets without publishing
	$(GORELEASER) release --snapshot --clean --skip=publish

# ─── Quality ─────────────────────────────────────────────────

test: ## Run Go tests
	go test ./... -v -race -count=1

lint: ## Lint Go + frontend
	golangci-lint run ./...
	cd web && npm run lint

check: ## Type-check frontend
	cd web && npx tsc --noEmit

# ─── Database ────────────────────────────────────────────────

migrate-up: ## Run migrations up
	migrate -path sql/migrations -database "$(DATABASE_URL)" up

migrate-down: ## Run migrations down
	migrate -path sql/migrations -database "$(DATABASE_URL)" down

sqlc-gen: ## Regenerate sqlc queries
	sqlc generate

# ─── Infrastructure (Supabase + Redis) ───────────────────────

INFRA_DIR := $(HOME)/Developer/infra/supabase

infra-up: ## Start Supabase + Redis containers
	docker compose -f $(INFRA_DIR)/docker-compose.yml up -d

infra-down: ## Stop Supabase + Redis containers
	docker compose -f $(INFRA_DIR)/docker-compose.yml down

infra-status: ## Show infra container status
	@docker compose -f $(INFRA_DIR)/docker-compose.yml ps

# ─── Stripe ───────────────────────────────────────────────

stripe-listen: ## Listen Stripe webhooks and forward to local API
	stripe listen --api-key $(STRIPE_SECRET_KEY) --events $(STRIPE_EVENTS) --forward-to $(STRIPE_FORWARD_URL)

stripe-webhook-secret: ## Print webhook secret for current local forward setup
	stripe listen --api-key $(STRIPE_SECRET_KEY) --events $(STRIPE_EVENTS) --forward-to $(STRIPE_FORWARD_URL) --print-secret

# ─── Observability (Prometheus + Grafana + Loki) ──────────────

obs-up: ## Start observability stack
	docker compose -f $(INFRA_DIR)/docker-compose.yml up -d prometheus loki promtail grafana

obs-down: ## Stop observability stack
	docker compose -f $(INFRA_DIR)/docker-compose.yml stop prometheus loki promtail grafana

obs-status: ## Show observability container status
	@docker compose -f $(INFRA_DIR)/docker-compose.yml ps prometheus loki promtail grafana

# ─── Setup ───────────────────────────────────────────────────

install: ## Install all dependencies
	go mod download
	cd web && npm install

clean: ## Remove build artifacts
	rm -rf bin/ web/dist/ web/tsconfig.tsbuildinfo

# ─── Help ────────────────────────────────────────────────────

help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
