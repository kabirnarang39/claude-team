DB_URL ?= postgres://localhost:5432/stripe_billing?sslmode=disable

migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down 1

.PHONY: migrate-up migrate-down
