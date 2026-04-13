include .env

.PHONY: up down logs ps

up:
	docker compose up --build

down:
	docker compose down -v

logs:
	docker compose logs -f

ps:
	docker compose ps
