include .env
export


DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: migrate-up migrate-down migrate-status migrate-create

migrate-up:
	goose -dir ./migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir ./migrations postgres "$(DB_URL)" down

migrate-status:
	goose -dir ./migrations postgres "$(DB_URL)" status


migrate-create:
	goose -dir ./migrations create $(name) sql