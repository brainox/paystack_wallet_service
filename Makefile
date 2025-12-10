.PHONY: help build run test migrate-up migrate-down migrate-create clean

help:
	@echo "Available commands:"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application"
	@echo "  make test          - Run tests"
	@echo "  make migrate-up    - Run database migrations"
	@echo "  make migrate-down  - Rollback database migrations"
	@echo "  make migrate-create NAME=<migration_name> - Create a new migration"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make deps          - Download dependencies"

# Database configuration
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= wallet_service
DB_URL = postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

deps:
	go mod download
	go mod tidy

build: deps
	go build -o bin/wallet-service main.go

run: deps
	go run main.go

test:
	go test -v ./...

migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=create_users_table"; \
		exit 1; \
	fi
	migrate create -ext sql -dir db/migrations -seq $(NAME)

clean:
	rm -rf bin/
	rm -rf tmp/

# Development helpers
dev: deps
	air

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f
