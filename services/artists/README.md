# Artist Service

Микросервис для управления артистами в системе MucissSocial.

## Функциональность

- Получение информации об артисте по ID
- Поиск артистов
- Получение трендовых артистов
- CRUD операции для артистов (через gRPC)

## API Endpoints (через Gateway)

### Публичные endpoints
- `GET /api/v1/artists/{artistId}` - получить артиста по ID
- `GET /api/v1/artists/trending` - получить трендовых артистов
- `GET /api/v1/artists/search?q={query}` - поиск артистов

## gRPC Service

Сервис предоставляет следующие gRPC методы:

- `CreateArtist` - создание артиста
- `GetArtistById` - получение артиста по ID
- `UpdateArtist` - обновление артиста
- `DeleteArtist` - удаление артиста
- `ListArtists` - список артистов с пагинацией
- `SearchArtists` - поиск артистов
- `GetTrendingArtists` - получение трендовых артистов

## Структура проекта

```
services/artists/
├── cmd/
│   └── main.go              # Точка входа приложения
├── internal/
│   ├── config/
│   │   └── config.go        # Конфигурация
│   ├── domain/
│   │   ├── artist.go        # Модели данных
│   │   └── interfaces.go    # Интерфейсы
│   ├── handler/
│   │   └── grpc.go          # gRPC handlers
│   ├── repository/
│   │   └── artist.go        # Работа с базой данных
│   └── service/
│       └── artist.go        # Бизнес-логика
├── migrations/
│   ├── 001_initial_schema.up.sql
│   └── 001_initial_schema.down.sql
├── proto/
│   └── artists/
│       └── v1/
│           └── artist_service.proto
├── Dockerfile
└── go.mod
```

## Переменные окружения

- `SERVER_HOST` - хост сервера (по умолчанию: 0.0.0.0)
- `SERVER_PORT` - порт сервера (по умолчанию: 50052)
- `DB_HOST` - хост базы данных
- `DB_PORT` - порт базы данных
- `DB_NAME` - имя базы данных
- `DB_USER` - пользователь базы данных
- `DB_PASSWORD` - пароль базы данных
- `DB_SSLMODE` - режим SSL для базы данных
- `LOG_LEVEL` - уровень логирования
- `LOG_FORMAT` - формат логирования

## Запуск

1. Настройте переменные окружения
2. Запустите PostgreSQL
3. Запустите сервис:

```bash
go run cmd/main.go
```

## Docker

```bash
docker build -t artist-service .
docker run -p 50052:50052 artist-service
```