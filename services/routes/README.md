# Routes Service

Сервис для управления музыкальными маршрутами в Music Social.

## Функциональность

- CRUD операции для маршрутов
- Управление точками маршрута
- Поиск маршрутов по геолокации
- Публичные/приватные маршруты
- Теги и категории

## Технологии

- Python 3.12
- FastAPI
- PostgreSQL (asyncpg)
- Redis (для кеширования)

## Запуск

### Локально

```bash
poetry install
poetry run uvicorn main:app --reload --host 0.0.0.0 --port 8080
```

### Docker

```bash
docker-compose up routes-service
```

## API Endpoints

- `POST /api/v1/routes` - создание маршрута
- `GET /api/v1/routes/{route_id}` - получение маршрута
- `PUT /api/v1/routes/{route_id}` - обновление маршрута
- `DELETE /api/v1/routes/{route_id}` - удаление маршрута
- `GET /api/v1/routes` - список маршрутов
- `GET /api/v1/routes/nearby` - маршруты поблизости
- `GET /api/v1/routes/search` - поиск маршрутов

## Конфигурация

См. `env.example` для списка переменных окружения.

