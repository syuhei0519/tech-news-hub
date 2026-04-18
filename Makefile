-include .env

COMPOSE := docker compose
DEV_COMPOSE := docker compose -f docker-compose.yml -f docker-compose.dev.yml

.PHONY: up down reset logs ps dev-up dev-down dev-logs dev-ps dev-config build test test-article-integration verify e2e-up e2e-seed test-e2e e2e-down

up:
	$(COMPOSE) up --build

down:
	$(COMPOSE) down

reset:
	$(COMPOSE) down -v

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps

dev-up:
	$(DEV_COMPOSE) up --build

dev-down:
	$(DEV_COMPOSE) down

dev-logs:
	$(DEV_COMPOSE) logs -f

dev-ps:
	$(DEV_COMPOSE) ps

dev-config:
	$(DEV_COMPOSE) config -q

build:
	cd frontend && npm run build

test:
	cd services/article-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
	cd services/api-gateway && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
	cd services/collector-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
	cd services/notification-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...

test-article-integration:
	cd services/article-service && ARTICLE_SERVICE_RUN_INTEGRATION=1 GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...

verify: test build
	$(COMPOSE) config -q

e2e-up:
	$(COMPOSE) up -d --build mysql rabbitmq article-service notification-service api-gateway frontend
	./scripts/e2e/wait-for-stack.sh

e2e-seed:
	./scripts/e2e/seed.sh

test-e2e:
	cd frontend && npm run test:e2e

e2e-down:
	$(COMPOSE) down -v
