# Функционал заказа музыки у артиста

## Описание фичи

Реализована система заказа треков у артистов через QR-коды с использованием виртуальных монет.

### Основные возможности:

1. **Виртуальные монеты**
   - Каждый пользователь получает 100 монет при регистрации
   - Заказ трека стоит 1 монету
   - При принятии заказа монета переходит артисту
   - При отклонении заказа монета возвращается пользователю

2. **QR-коды для заказов**
   - Артист создает сессию для приема заказов
   - Генерируется уникальный QR-код
   - QR-код ведет на страницу заказа трека
   - Можно деактивировать сессию в любой момент

3. **Процесс заказа**
   - Пользователь сканирует QR-код
   - Выбирает трек для заказа
   - Тратит 1 монету
   - Артист видит заказ в своем списке
   - Артист может принять или отклонить заказ

4. **Обработка заказов**
   - **Принятие**: трек добавляется в очередь воспроизведения, монета переходит артисту
   - **Отклонение**: монета возвращается пользователю

## Реализованные изменения

### 1. Users Service (Go)

#### Миграции базы данных
- `services/users/migrations/002_add_coins.up.sql`
  - Добавлен столбец `coins` (по умолчанию 100)
  - Добавлен индекс для быстрой сортировки
  - Добавлено ограничение на отрицательные значения

#### Код
- **domain/user.go**: добавлено поле `Coins int` в структуры `User` и `PublicUser`
- **domain/interfaces.go**: добавлен метод `UpdateCoins` в интерфейс `UserRepository`
- **repository/user.go**: реализован метод `UpdateCoins` с проверкой на отрицательный баланс
- **service/user.go**: добавлен метод `UpdateCoins` в `UserService`
- **handler/grpc.go**: добавлен gRPC handler `UpdateCoins`
- **handler/converters.go**: обновлены конвертеры для включения поля `coins`
- **proto/users/v1/user_service.proto**: 
  - Добавлены сообщения `UpdateCoinsRequest` и `UpdateCoinsResponse`
  - Добавлен RPC метод `UpdateCoins`
  - Добавлено поле `coins` в `User` и `PublicUser`

### 2. Music Requests Service (Python/FastAPI)

Новый микросервис: `services/music_requests/`

#### Структура
```
services/music_requests/
├── main.py                 # Основной файл с API endpoints
├── models.py               # SQLAlchemy модели
├── schemas.py              # Pydantic схемы
├── pyproject.toml          # Зависимости (poetry)
├── Dockerfile              # Docker образ
├── env.example             # Пример переменных окружения
├── README.md               # Документация сервиса
├── core/
│   ├── config.py          # Конфигурация
│   ├── database.py        # Подключение к БД
│   └── __init__.py
└── sql/
    ├── 001_initial_schema.up.sql    # Миграция создания таблиц
    └── 001_initial_schema.down.sql  # Откат миграции
```

#### Таблицы БД
1. **request_sessions** - сессии для приема заказов
2. **track_requests** - заказы треков
3. **coin_transactions** - история транзакций монет

#### API Endpoints

**Сессии заказов:**
- `POST /api/v1/sessions` - создать новую сессию
- `GET /api/v1/sessions/my` - получить активную сессию
- `DELETE /api/v1/sessions/{session_id}` - деактивировать сессию
- `GET /api/v1/sessions/{session_code}/qr` - получить QR-код (PNG)

**Заказы треков:**
- `POST /api/v1/requests` - создать заказ (стоит 1 монету)
- `GET /api/v1/requests/incoming` - входящие заказы (для артиста)
- `GET /api/v1/requests/outgoing` - исходящие заказы (от пользователя)
- `PATCH /api/v1/requests/{request_id}` - принять/отклонить заказ

**Монеты:**
- `GET /api/v1/users/{user_id}/coins` - получить баланс

### 3. Docker Compose

Добавлено в `services/docker-compose.yml`:
- **postgres-music-requests** - база данных для music_requests
- **music-requests-service** - сам сервис
  - Порт: 8087:8080
  - База данных: postgres-music-requests:5432
  - Интеграция с: users service (через gateway), playback_queue

## Как использовать

### 1. Запуск сервисов

```bash
cd services

# Собрать protobuf для users service
cd users
make proto
cd ..

# Запустить все сервисы
docker-compose up -d users-service postgres-music-requests music-requests-service
```

### 2. Workflow для артиста

1. **Создать сессию для приема заказов**
```bash
curl -X POST http://localhost:8087/api/v1/sessions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"
```

Ответ:
```json
{
  "id": "uuid",
  "artist_id": "uuid",
  "session_code": "unique-code",
  "is_active": true,
  "qr_code_url": "/api/v1/sessions/unique-code/qr",
  "created_at": "2025-12-20T..."
}
```

2. **Получить QR-код**
```bash
curl http://localhost:8087/api/v1/sessions/unique-code/qr -o qr-code.png
```

3. **Посмотреть входящие заказы**
```bash
curl http://localhost:8087/api/v1/requests/incoming \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

4. **Принять заказ**
```bash
curl -X PATCH http://localhost:8087/api/v1/requests/{request_id} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"action": "accept"}'
```

5. **Отклонить заказ**
```bash
curl -X PATCH http://localhost:8087/api/v1/requests/{request_id} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"action": "decline"}'
```

### 3. Workflow для пользователя

1. **Проверить баланс монет**
```bash
curl http://localhost:8087/api/v1/users/{user_id}/coins \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

2. **Создать заказ трека**
```bash
curl -X POST http://localhost:8087/api/v1/requests \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_code": "unique-code",
    "track_id": "track-uuid",
    "message": "Пожалуйста, сыграйте эту песню!"
  }'
```

3. **Посмотреть свои заказы**
```bash
curl http://localhost:8087/api/v1/requests/outgoing \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Интеграция с API Gateway

Для полной интеграции необходимо добавить в API Gateway (`services/gateway/main.go`) proxy endpoints для music_requests сервиса:

```go
// Добавить в структуру Gateway
musicRequestsURL string

// Добавить в main()
musicRequestsURL := getEnv("MUSIC_REQUESTS_URL", "http://music-requests-service:8080")

// Добавить роуты
r.HandleFunc("/api/v1/sessions", proxyHandler(gw.musicRequestsURL)).Methods("POST", "OPTIONS")
r.HandleFunc("/api/v1/sessions/my", proxyHandler(gw.musicRequestsURL)).Methods("GET", "OPTIONS")
r.HandleFunc("/api/v1/sessions/{session_id}", proxyHandler(gw.musicRequestsURL)).Methods("DELETE", "OPTIONS")
r.HandleFunc("/api/v1/sessions/{session_code}/qr", proxyHandler(gw.musicRequestsURL)).Methods("GET", "OPTIONS")
r.HandleFunc("/api/v1/requests", proxyHandler(gw.musicRequestsURL)).Methods("POST", "GET", "OPTIONS")
r.HandleFunc("/api/v1/requests/incoming", proxyHandler(gw.musicRequestsURL)).Methods("GET", "OPTIONS")
r.HandleFunc("/api/v1/requests/outgoing", proxyHandler(gw.musicRequestsURL)).Methods("GET", "OPTIONS")
r.HandleFunc("/api/v1/requests/{request_id}", proxyHandler(gw.musicRequestsURL)).Methods("PATCH", "OPTIONS")

// Добавить endpoint для обновления монет (proxy к users service gRPC)
r.HandleFunc("/api/v1/users/{user_id}/coins", updateCoinsHandler).Methods("PATCH", "OPTIONS")
```

## Дальнейшие улучшения

1. **Интеграция с Playback Queue**
   - В `music_requests/main.py` функция `add_track_to_queue` сейчас заглушка
   - Нужно реализовать gRPC вызов к playback_queue сервису

2. **WebSocket уведомления**
   - Отправлять real-time уведомления артисту о новых заказах
   - Уведомлять пользователя о принятии/отклонении заказа

3. **Frontend**
   - Страница генерации QR-кода для артиста
   - Страница создания заказа для пользователя
   - Список заказов для артиста и пользователя
   - Отображение баланса монет

4. **Аналитика**
   - История заказов
   - Статистика по артистам (сколько заказов принято/отклонено)
   - Топ заказываемых треков

5. **Монетизация**
   - Покупка дополнительных монет
   - Бонусные монеты за активность
   - Система наград

## Коммит изменений

```bash
git add .
git commit -m "feat: добавлен функционал заказа музыки через QR-коды

- Добавлена система виртуальных монет в users service
- Создан новый микросервис music_requests для обработки заказов
- Реализована генерация QR-кодов для сессий заказов
- Добавлена логика принятия/отклонения заказов с транзакциями монет
- Обновлен docker-compose с новым сервисом и БД
- Создана документация по использованию новой фичи"

git push origin feature/music-requests
```

## Тестирование

После запуска сервисов можно протестировать через:
- **Swagger UI**: http://localhost:8087/docs (автоматически генерируется FastAPI)
- **curl** команды (примеры выше)
- **Postman** коллекция (можно создать на основе Swagger)
