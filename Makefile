-include .env
export

.PHONY: dev dev-api dev-worker dev-web dev-all \
       build build-api build-worker build-web \
       test lint \
       migrate-up migrate-down sqlc-gen \
       infra-up infra-down infra-status \
       obs-up obs-down obs-status \
       install clean check

# ─── Development ─────────────────────────────────────────────

dev: dev-api ## Run API server (default)

dev-api: ## Run API server
	go run ./cmd/api

dev-worker: ## Run background worker
	go run ./cmd/worker

dev-web: ## Run frontend dev server
	cd web && npm run dev

dev-all: ## Run API + worker + web concurrently
	@trap 'kill 0' EXIT; \
	$(MAKE) dev-api & \
	$(MAKE) dev-worker & \
	$(MAKE) dev-web & \
	wait

# ─── Build ───────────────────────────────────────────────────

build: build-api build-worker build-web ## Build everything

build-api: ## Build API binary
	go build -o bin/api ./cmd/api

build-worker: ## Build worker binary
	go build -o bin/worker ./cmd/worker

build-web: ## Build frontend
	cd web && npm run build

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
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
