APP_NAME = avito-app
IMAGE = $(APP_NAME):latest
COMPOSE = docker compose -f deployments/docker-compose.yml

.PHONY: build up down restart logs test clean

build:
	docker build -t $(IMAGE) .

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

restart: build down up

logs:
	$(COMPOSE) logs -f app

db-logs:
	$(COMPOSE) logs -f db