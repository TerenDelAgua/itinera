.PHONY: up down migrate run test clean

up:
	docker compose up -d

down:
	docker compose down -v

migrate:
	docker compose exec postgres psql -U ${DB_USER} -d ${DB_NAME} -f /docker-entrypoint-initdb.d/001_init.sql

run:
	go run cmd/api/main.go

test:
	go test ./... -v

clean:
	go clean -cache -modcache -testcache