# Quick Start Guide - Music Social API Gateway

## Быстрый запуск

### 1. Убедитесь что Docker запущен
```bash
docker --version
docker compose version
```

### 2. Запустите все сервисы
```bash
cd /path/to/MucisSocial/services

# Запуск всех сервисов
docker compose up -d

# Проверка статуса
docker compose ps
```

### 3. Ожидайте готовности всех сервисов
```bash
# Проверяйте каждые 10-15 секунд пока все не станут healthy
docker compose ps

# Должно показать:
# postgres        Up (healthy)
# users-service   Up (healthy)  
# api-gateway     Up (healthy)
```

### 4. Тестируйте API

#### Health Check
```bash
curl http://localhost:8080/__health
# Ответ: {"status":"ok"}
```

#### Регистрация пользователя
```bash
curl -X POST http://localhost:8080/api/v1/auth/sign-up \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123", 
    "username": "myuser",
    "displayName": "My User"
  }'
```

#### Вход в систему
```bash
curl -X POST http://localhost:8080/api/v1/auth/sign-in \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

Сохраните `access_token` из ответа для дальнейших запросов.

#### Получение профиля (с JWT токеном)
```bash
curl -X GET http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN_HERE"
```

## Доступные эндпоинты

### Публичные (без авторизации)
- `GET /__health` - Health check
- `POST /api/v1/auth/sign-up` - Регистрация
- `POST /api/v1/auth/sign-in` - Вход  
- `POST /api/v1/auth/refresh` - Обновление токена

### Защищенные (требуют JWT токен)
- `GET /api/v1/me` - Получить профиль
- `PUT /api/v1/me` - Обновить профиль
- `GET /api/v1/me/search-history` - История поиска
- `POST /api/v1/me/search-history` - Добавить в историю поиска
- `DELETE /api/v1/me/search-history` - Очистить историю поиска

## Порты

- **API Gateway**: http://localhost:8080
- **PostgreSQL**: localhost:5432
- **Users Service gRPC**: localhost:50051

## Полная документация

Для детальной информации о добавлении новых сервисов см. [README.md](./README.md)

## Troubleshooting

### Если сервисы не стартуют
```bash
# Остановка всех контейнеров
docker compose down

# Пересборка образов
docker compose build

# Запуск заново
docker compose up -d

# Просмотр логов
docker compose logs -f
```

### Если API Gateway показывает unhealthy
```bash
# Проверьте логи
docker compose logs api-gateway

# Перезапустите gateway
docker compose restart api-gateway
```

### Если база данных недоступна
```bash
# Проверьте что PostgreSQL healthy
docker compose ps postgres

# Посмотрите логи
docker compose logs postgres

# При необходимости пересоздайте
docker compose down postgres
docker compose up -d postgres
```