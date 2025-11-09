# API Gateway - Swagger Documentation

## Обзор

Этот проект содержит API Gateway для сервиса MucissSocial с интегрированной Swagger документацией.

## Установка и запуск

### 1. Установка зависимостей

```bash
go mod tidy
```

### 2. Генерация Swagger документации

```bash
~/go/bin/swag init
```

Эта команда сгенерирует файлы документации в папке `docs/`:
- `docs.go` - Go код с метаданными
- `swagger.json` - JSON документация
- `swagger.yaml` - YAML документация

### 3. Запуск сервера

```bash
go run main.go
```

## Доступ к документации

После запуска сервера Swagger UI будет доступен по адресу:

```
http://localhost:8080/swagger/
```

## API Endpoints

### Аутентификация
- `POST /api/v1/auth/sign-up` - Регистрация нового пользователя
- `POST /api/v1/auth/sign-in` - Вход в систему
- `POST /api/v1/auth/refresh` - Обновление токена

### Профиль пользователя (требует JWT)
- `GET /api/v1/me` - Получение профиля текущего пользователя
- `PUT /api/v1/me` - Обновление профиля пользователя

### История поиска (требует JWT)
- `GET /api/v1/me/search-history` - Получение истории поиска
- `POST /api/v1/me/search-history` - Добавление записи в историю поиска
- `DELETE /api/v1/me/search-history` - Очистка истории поиска

### Служебные endpoints
- `GET /__health` - Проверка состояния сервиса

## Аутентификация

Большинство endpoints требуют JWT токен в заголовке Authorization:

```
Authorization: Bearer <your-jwt-token>
```

Токен можно получить через endpoints `/api/v1/auth/sign-in` или `/api/v1/auth/sign-up`.

## Структура ответов

### Успешный ответ аутентификации
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "user_123456",
    "username": "musiclover",
    "email": "user@example.com",
    "avatar_url": "https://example.com/avatar.jpg",
    "music_taste_summary": {
      "top_genres": ["Rock", "Pop", "Jazz"],
      "top_artists": ["The Beatles", "Queen", "Pink Floyd"]
    },
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

### Ответ с ошибкой
```json
{
  "error": "Invalid request",
  "code": 400,
  "message": "The request contains invalid data"
}
```

## Разработка

### Обновление документации

После изменения комментариев Swagger в коде, запустите:

```bash
~/go/bin/swag init
```

### Добавление новых endpoints

1. Добавьте Go комментарии с Swagger аннотациями над функцией-обработчиком
2. Используйте формат:
   ```go
   // handlerName godoc
   //
   //	@Summary		Краткое описание
   //	@Description	Подробное описание
   //	@Tags			Категория
   //	@Accept			json
   //	@Produce		json
   //	@Security		BearerAuth  # Если требуется аутентификация
   //	@Param			name	body		Type	true	"Описание параметра"
   //	@Success		200		{object}	ResponseType
   //	@Failure		400		{object}	ErrorResponse
   //	@Router			/path [method]
   ```
3. Регенерируйте документацию командой `swag init`

## Примеры использования

### Регистрация пользователя
```bash
curl -X POST http://localhost:8080/api/v1/auth/sign-up \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "username": "musiclover"
  }'
```

### Получение профиля
```bash
curl -X GET http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer <your-access-token>"
```

### Добавление в историю поиска
```bash
curl -X POST http://localhost:8080/api/v1/me/search-history \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-access-token>" \
  -d '{
    "query": "The Beatles"
  }'
```

## Файлы проекта

- `main.go` - Основной файл с API Gateway и Swagger аннотациями
- `swagger.yaml` - Ручная YAML документация (опциональная)
- `docs/` - Сгенерированная Swagger документация
- `go.mod` - Go модули и зависимости

## Зависимости

- `github.com/gorilla/mux` - HTTP router
- `github.com/golang-jwt/jwt/v5` - JWT library
- `github.com/swaggo/http-swagger` - Swagger UI middleware
- `github.com/swaggo/swag` - Swagger generator
- `google.golang.org/grpc` - gRPC client