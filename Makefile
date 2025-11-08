.PHONY: dev dev-up dev-down apps apps-up apps-down infra nfra-up infra-down


dev:
	docker compose up -d --build

dev-up: dev-build
	docker compose up -d

dev-down:
	docker compose -f down


apps:
	docker compose -f apps/docker-compose.yml up -d --build

apps-up:
	docker compose -f apps/docker-compose.yml up -d

apps-down:
	docker compose -f apps/docker-compose.yml down


infra:
	docker compose -f infrastructure/docker-compose.yml up -d --build

infra-up:
	docker compose -f infrastructure/docker-compose.yml up -d

infra-down:
	docker compose -f infrastructure/docker-compose.yml down
