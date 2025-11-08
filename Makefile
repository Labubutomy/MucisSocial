.PHONY: dev dev-up dev-down infra infra-up infra-down status

dev:
	docker compose up -d --build

dev-up: dev-build
	docker compose up -d

dev-down:
	docker compose -f docker-compose.yml down

infra:
	docker compose -f infrastructure/docker-compose.yml up -d --build

infra-up:
	docker compose -f infrastructure/docker-compose.yml up -d

infra-down:
	docker compose -f infrastructure/docker-compose.yml down

status:
	docker compose -f docker-compose.yml ps
