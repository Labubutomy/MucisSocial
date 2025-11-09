# Makefile для API Gateway

.PHONY: help install docs clean run test build

help: ## Показать справку
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install: ## Установить зависимости
	go mod tidy
	go install github.com/swaggo/swag/cmd/swag@latest

docs: ## Сгенерировать Swagger документацию
	~/go/bin/swag init

docs-clean: ## Очистить сгенерированную документацию
	rm -rf docs/

clean: docs-clean ## Очистить все сгенерированные файлы
	rm -f api-gateway

build: docs ## Скомпилировать приложение
	go build -o api-gateway .

run: docs ## Запустить приложение
	go run main.go

test: ## Запустить тесты
	go test ./...

dev: ## Запустить в режиме разработки с автогенерацией документации
	make docs && make run

fmt: ## Форматировать код
	go fmt ./...

lint: ## Проверить код линтером
	golangci-lint run

docker-build: ## Собрать Docker образ
	docker build -t musicssocial-gateway .

docker-run: ## Запустить в Docker
	docker run -p 8080:8080 musicssocial-gateway

swagger-ui: ## Открыть Swagger UI в браузере
	open http://localhost:8080/swagger/

check-docs: ## Проверить актуальность документации
	~/go/bin/swag init --parseInternal --parseDependency
	@if [ -n "$$(git status --porcelain docs/)" ]; then \
		echo "❌ Swagger документация не актуальна. Запустите: make docs"; \
		exit 1; \
	else \
		echo "✅ Swagger документация актуальна"; \
	fi