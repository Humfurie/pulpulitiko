DATABASE_URL ?= postgres://politics:localdev@localhost:5432/politics_db?sslmode=disable

.PHONY: migrate-up migrate-down migrate-create migrate-version seed build

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-version:
	migrate -path migrations -database "$(DATABASE_URL)" version

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

build:
	go build -o bin/server ./cmd/server
	go build -o bin/seed ./cmd/seed

seed:
	@if [ -f ../.env ]; then \
		set -a && . ../.env && set +a && ./bin/seed; \
	else \
		echo "Error: .env file not found. Copy .env.example to .env and configure it."; \
		exit 1; \
	fi
