.PHONY: dev build test migrate-up migrate-down sqlc-gen docker-up docker-down lint

dev:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

test:
	go test ./... -v -race -count=1

migrate-up:
	migrate -path sql/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path sql/migrations -database "$(DATABASE_URL)" down

sqlc-gen:
	sqlc generate

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

lint:
	golangci-lint run ./...
