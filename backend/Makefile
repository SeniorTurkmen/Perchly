include .env
export

MIGRATE := go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1
SWAG := go run github.com/swaggo/swag/cmd/swag@v1.16.6

.PHONY: dev run build migrate-up migrate-down migrate-force docker-up docker-down swagger admin-create

dev:
	./scripts/dev.sh

run:
	go run ./cmd/api

# Provision or reset an admin dashboard account, e.g.:
#   make admin-create email=you@perchly.app password='a strong password'
admin-create:
	go run ./cmd/createadmin -email="$(email)" -password="$(password)"

# Regenerate docs/ from the @Summary/@Router annotations on handlers —
# run this after adding or changing any endpoint, then restart the
# server; the UI is served at /swagger/index.html.
swagger:
	$(SWAG) init -g cmd/api/main.go -o docs --parseInternal --parseDependency

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
