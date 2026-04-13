-include .env

.PHONY: up down logs ps build test verify

up:
	docker compose up --build

down:
	docker compose down -v

logs:
	docker compose logs -f

ps:
	docker compose ps

build:
	cd frontend && npm run build

test:
	cd services/article-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
	cd services/api-gateway && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
	cd services/collector-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
	cd services/notification-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...

verify: test build
	docker compose config -q
