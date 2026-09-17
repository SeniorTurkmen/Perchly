include .env
export

MIGRATE := go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

.PHONY: dev run build migrate-up migrate-down migrate-force docker-up docker-down

dev:
	./scripts/dev.sh

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

migrate-up:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" down 1

migrate-force:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" force $(version)

docker-up:
	docker compose up -d

docker-down:
	docker compose down
