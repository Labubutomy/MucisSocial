# Music Social - Полная инструкция по запуску User Service

## Быстрый старт

### 1. Создайте Docker сеть
```bash
docker network create music-network
```

### 2. Запустите сервисы
```bash
cd services
docker-compose up -d
```

### 3. Проверьте статус сервисов
```bash
docker-compose ps
```

Вы должны увидеть:
- `music-social-postgres` - База данных PostgreSQL
- `music-social-users` - User Service (gRPC)
- `music-social-gateway` - API Gateway (KrakenD)

### 4. Проверьте health статус
```bash
# Проверка API Gateway
curl http://localhost:8080/__health

# Проверка базы данных
docker-compose exec postgres pg_isready -U postgres

# Проверка пользовательского сервиса (требует grpcurl)
# grpcurl -plaintext localhost:50051 list
```

### 5. Запустите тестирование API
```bash
./test-api.sh
```

## Доступные endpoints

API Gateway доступен по адресу: **http://localhost:8080**

### Аутентификация
- `POST /api/v1/auth/sign-up` - Регистрация пользователя
- `POST /api/v1/auth/sign-in` - Вход в систему  
- `POST /api/v1/auth/refresh` - Обновление токена

### Профиль пользователя (требует авторизации)
- `GET /api/v1/me` - Получить мой профиль
- `PUT /api/v1/me` - Обновить профиль

### Плейлисты (требует авторизации)
- `GET /api/v1/me/playlists` - Мои плейлисты
- `GET /api/v1/users/{userId}/playlists` - Плейлисты пользователя

### История поиска (требует авторизации)
- `GET /api/v1/me/search-history` - Получить историю поиска
- `POST /api/v1/me/search-history` - Добавить в историю
- `DELETE /api/v1/me/search-history` - Очистить историю

## Примеры запросов

### Регистрация
```bash
curl -X POST http://localhost:8080/api/v1/auth/sign-up \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com", 
    "password": "password123"
  }'
```

### Вход в систему
```bash
curl -X POST http://localhost:8080/api/v1/auth/sign-in \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### Получение профиля (с токеном)
```bash
curl -X GET http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Логи и отладка

### Просмотр логов всех сервисов
```bash
docker-compose logs -f
```

### Просмотр логов конкретного сервиса
```bash
# User Service
docker-compose logs -f users-service

# API Gateway
docker-compose logs -f api-gateway

# PostgreSQL
docker-compose logs -f postgres
```

### Подключение к базе данных
```bash
docker-compose exec postgres psql -U postgres -d music_social_users
```

### Вход в контейнер User Service
```bash
docker-compose exec users-service sh
```

## Остановка и очистка

### Остановка сервисов
```bash
docker-compose down
```

### Полная очистка (включая volumes)
```bash
docker-compose down -v
docker system prune -f
```

## Устранение проблем

### Проблема: Сеть не создана
```bash
docker network create music-network
```

### Проблема: Порты заняты
Проверьте, что порты 8080, 50051, 5432 свободны:
```bash
lsof -i :8080
lsof -i :50051  
lsof -i :5432
```

### Проблема: User Service не может подключиться к БД
Проверьте статус PostgreSQL:
```bash
docker-compose logs postgres
```

### Проблема: API Gateway не может достучаться до User Service
Проверьте, что сервисы в одной сети:
```bash
docker network ls
docker network inspect music-network
```

## Архитектура

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client App    │───▶│  API Gateway    │───▶│  User Service   │
│  (Frontend)     │    │   (KrakenD)     │    │   (Go + gRPC)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                        │
                              │                        ▼
                              │                ┌─────────────────┐
                              │                │   PostgreSQL    │
                              │                │   (Database)    │
                              │                └─────────────────┘
                              ▼
                       ┌─────────────────┐
                       │  Other Services │
                       │ (Tracks, CDN,   │
                       │  Playlists...)  │
                       └─────────────────┘
```

## Production Ready Features

✅ **Безопасность**
- JWT аутентификация с refresh токенами
- Bcrypt хэширование паролей
- CORS конфигурация
- Валидация входных данных

✅ **Надежность** 
- Health checks для всех сервисов
- Graceful shutdown
- Database connection pooling
- Автоматические миграции БД

✅ **Мониторинг**
- Структурированное логирование
- HTTP status codes
- gRPC статусы

✅ **Масштабируемость**
- Микросервисная архитектура
- gRPC для межсервисного общения
- API Gateway для единой точки входа
- Clean Architecture

## Следующие шаги для production

1. **Мониторинг**: Добавить Prometheus + Grafana
2. **Трассировка**: Настроить Jaeger/OpenTelemetry  
3. **Rate Limiting**: Настроить в KrakenD
4. **Secrets Management**: Kubernetes secrets или Vault
5. **CI/CD**: GitHub Actions или GitLab CI
6. **Тестирование**: Unit/Integration тесты
7. **Documentation**: OpenAPI/Swagger спецификация