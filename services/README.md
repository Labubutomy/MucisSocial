# Music Social - User Service

Микросервис управления пользователями для платформы Music Social.

## Архитектура

- **User Service**: gRPC сервис на Go для управления пользователями и аутентификации
- **PostgreSQL**: База данных для хранения пользователских данных
- **KrakenD API Gateway**: API Gateway с поддержкой gRPC и JWT аутентификации

## Функциональность

### Аутентификация
- Регистрация пользователей
- Вход в систему
- Обновление токенов
- JWT токены (access + refresh)

### Управление пользователями
- Получение профиля пользователя
- Обновление профиля
- Музыкальные вкусы пользователя

### История поиска
- Сохранение истории поиска
- Получение истории
- Очистка истории

## API Endpoints

### Аутентификация
- `POST /api/v1/auth/sign-up` - Регистрация
- `POST /api/v1/auth/sign-in` - Вход в систему
- `POST /api/v1/auth/refresh` - Обновление токена

### Профиль пользователя
- `GET /api/v1/me` - Получить мой профиль
- `PUT /api/v1/me` - Обновить профиль

### История поиска
- `GET /api/v1/me/search-history` - Получить историю поиска
- `POST /api/v1/me/search-history` - Добавить в историю
- `DELETE /api/v1/me/search-history` - Очистить историю

## Быстрый старт

### Требования
- Docker и Docker Compose
- Go 1.22+ (для разработки)
- Make (опционально, для удобства)

### Запуск

1. Создайте сеть Docker:
```bash
docker network create music-network
```

2. Запустите сервисы:
```bash
cd services
docker-compose up -d
```

3. Проверьте статус:
```bash
docker-compose ps
```

### API Gateway
API Gateway доступен по адресу: `http://localhost:8080`

### Примеры запросов

#### Регистрация
```bash
curl -X POST http://localhost:8080/api/v1/auth/sign-up \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

#### Вход в систему
```bash
curl -X POST http://localhost:8080/api/v1/auth/sign-in \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

#### Получение профиля (требует токен)
```bash
curl -X GET http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Разработка

### Структура проекта
```
services/users/
├── cmd/                    # Точка входа приложения
├── internal/
│   ├── config/            # Конфигурация
│   ├── domain/            # Доменные модели и интерфейсы
│   ├── handler/           # gRPC хэндлеры
│   ├── repository/        # Репозитории (DB)
│   └── service/           # Бизнес-логика
├── proto/                 # Protobuf определения
├── migrations/            # Миграции БД
├── Dockerfile
├── go.mod
└── Makefile
```

### Команды разработки

```bash
# Генерация protobuf файлов
make proto

# Запуск тестов
make test

# Форматирование кода
make fmt

# Линтинг
make lint

# Сборка образа
make build

# Создание миграции
make migrate-create name=migration_name

# Применение миграций
make migrate

# Просмотр логов
make logs-users

# Подключение к БД
make db-shell
```

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `SERVER_HOST` | Хост сервера | `0.0.0.0` |
| `SERVER_PORT` | Порт сервера | `50051` |
| `DB_HOST` | Хост БД | `localhost` |
| `DB_PORT` | Порт БД | `5432` |
| `DB_USER` | Пользователь БД | `postgres` |
| `DB_PASSWORD` | Пароль БД | `password` |
| `DB_NAME` | Имя БД | `music_social_users` |
| `JWT_ACCESS_SECRET` | Секрет для access токенов | - |
| `JWT_REFRESH_SECRET` | Секрет для refresh токенов | - |
| `JWT_ACCESS_EXPIRATION` | Время жизни access токена | `15m` |
| `JWT_REFRESH_EXPIRATION` | Время жизни refresh токена | `168h` |

## База данных

### Основные таблицы
- `users` - Пользователи
- `music_taste_genres` - Жанры пользователя
- `music_taste_artists` - Артисты пользователя
- `search_history` - История поиска
- `refresh_tokens` - Refresh токены

### Миграции
Миграции находятся в папке `migrations/` и автоматически применяются при запуске сервиса.

## Мониторинг и логирование

### Health checks
- User Service: `nc -z localhost 50051`
- API Gateway: `curl -f http://localhost:8080/__health`
- PostgreSQL: `pg_isready -U postgres`

### Логи
```bash
# Все сервисы
docker-compose logs -f

# Только user service
docker-compose logs -f users-service

# Только API gateway
docker-compose logs -f api-gateway
```

## Production Ready Features

- ✅ JWT аутентификация с refresh токенами
- ✅ Валидация входящих данных
- ✅ Структурированное логирование (zap)
- ✅ Graceful shutdown
- ✅ Health checks
- ✅ Database connection pooling
- ✅ Миграции БД
- ✅ CORS поддержка в API Gateway
- ✅ Rate limiting готов для настройки
- ✅ Clean Architecture
- ✅ gRPC с рефлексией для разработки

## Безопасность

- Пароли хэшируются с использованием bcrypt
- JWT токены с коротким временем жизни
- Refresh токены хранятся в БД
- CORS настроен в API Gateway
- Валидация всех входящих данных
- Prepared statements для защиты от SQL injection

## Следующие шаги

1. Добавить rate limiting в KrakenD
2. Настроить мониторинг (Prometheus/Grafana)
3. Добавить трассировку (Jaeger)
4. Настроить CI/CD пайплайн
5. Добавить unit и integration тесты
6. Настроить production секреты
