# Реализация сервиса совместного прослушивания музыки

## Обзор

Реализованы два основных сервиса для системы совместного прослушивания музыки:

1. **WebSocket Gateway Service** (Python FastAPI)
2. **Session Service** (Kotlin Spring Boot)

## Архитектура

```
Клиенты (Web/Mobile)
    ↓ (WebSocket)
WebSocket Gateway (Python FastAPI)
    ↓ (Kafka Message Bus)
Session Service (Kotlin Spring)
    ↓ (HTTP/REST)
CDN Service
```

## 1. WebSocket Gateway Service

### Расположение
`services/ws_gateway/`

### Основные компоненты

- **main.py** - точка входа FastAPI приложения
- **api/websocket.py** - WebSocket endpoint и управление соединениями
- **services/redis_service.py** - управление соединениями в Redis
- **services/kafka_service.py** - интеграция с Kafka
- **services/jwt_service.py** - валидация JWT токенов
- **models/messages.py** - модели сообщений
- **core/config.py** - конфигурация

### Функциональность

✅ WebSocket endpoint (`/ws`) с JWT аутентификацией  
✅ Управление соединениями через Redis  
✅ Отправка событий клиентов в Kafka (топик `music-session-events`)  
✅ Получение синхронизаций из Kafka (топик `music-session-sync`)  
✅ Рассылка синхронизаций всем клиентам комнаты  
✅ Пинг-понг для поддержания соединений  
✅ Health check endpoint (`/health`)

### Запуск

```bash
cd services/ws_gateway
poetry install
poetry run python main.py
```

Или через Docker:
```bash
docker-compose up ws-gateway
```

## 2. Session Service

### Расположение
`services/session/`

### Основные компоненты

- **SessionApplication.kt** - точка входа Spring Boot приложения
- **domain/model/** - модели данных (RoomState, ClientEvent, SyncMessage)
- **repository/RoomRepository.kt** - MongoDB репозиторий
- **service/RoomService.kt** - бизнес-логика управления комнатами
- **service/EventProcessor.kt** - обработка событий от клиентов
- **service/SyncService.kt** - отправка синхронизаций в Kafka
- **consumer/EventConsumer.kt** - Kafka consumer для событий
- **controller/RoomController.kt** - REST API для управления комнатами
- **config/KafkaConfig.kt** - конфигурация Kafka

### Функциональность

✅ Управление комнатами (создание/удаление)  
✅ Добавление/удаление участников  
✅ Обработка событий воспроизведения (play, pause, seek, change_track)  
✅ Хранение состояний в MongoDB  
✅ Отправка синхронизаций в Kafka  
✅ REST API для управления комнатами  
✅ Health check endpoint (`/api/rooms/health`)

### Запуск

```bash
cd services/session
./gradlew bootRun
```

Или через Docker:
```bash
docker-compose up session-service
```

## 3. Инфраструктура

### Обновленные файлы

1. **services/docker-compose.yml**
   - Добавлен сервис `redis` для WebSocket Gateway
   - Добавлен сервис `mongodb` для Session Service
   - Добавлен сервис `ws-gateway`
   - Добавлен сервис `session-service`

2. **infrastructure/redpanda/init-topics.sh**
   - Добавлены топики:
     - `music-session-events` (10 partitions)
     - `music-session-sync` (10 partitions)

## 4. Протокол работы

### Подключение к комнате

1. Клиент подключается к WebSocket: `ws://localhost:8001/ws?token=<JWT>&room_id=<ROOM_ID>`
2. Gateway валидирует JWT токен
3. Gateway добавляет соединение в Redis
4. Gateway отправляет событие `join_room` в Kafka (опционально)
5. Session Service обрабатывает событие и обновляет состояние комнаты
6. Session Service отправляет синхронизацию в Kafka
7. Gateway получает синхронизацию и рассылает всем клиентам комнаты

### Действие воспроизведения

1. Клиент отправляет сообщение через WebSocket:
```json
{
  "type": "player_action",
  "room_id": "room_123",
  "user_id": "user_456",
  "action": "play",
  "payload": {}
}
```

2. Gateway валидирует сообщение и отправляет в Kafka (топик `music-session-events`)
3. Session Service получает событие, обрабатывает и обновляет состояние комнаты
4. Session Service отправляет синхронизацию в Kafka (топик `music-session-sync`)
5. Gateway получает синхронизацию и рассылает всем клиентам комнаты:
```json
{
  "type": "sync_state",
  "room_id": "room_123",
  "state": {
    "current_track": null,
    "position": 0.0,
    "is_playing": true,
    "last_action_by": "user_456"
  }
}
```

## 5. Конфигурация

### WebSocket Gateway

Переменные окружения (префикс `WS_GATEWAY_`):
- `HOST`, `PORT` - хост и порт сервера
- `JWT_SECRET` - секретный ключ для JWT
- `REDIS_HOST`, `REDIS_PORT` - настройки Redis
- `KAFKA_BROKERS` - брокеры Kafka
- `KAFKA_EVENTS_TOPIC`, `KAFKA_SYNC_TOPIC` - топики Kafka

### Session Service

Переменные окружения:
- `SERVER_PORT` - порт сервера
- `MONGODB_URI` - URI подключения к MongoDB
- `REDIS_HOST`, `REDIS_PORT` - настройки Redis
- `KAFKA_BROKERS` - брокеры Kafka
- `KAFKA_EVENTS_TOPIC`, `KAFKA_SYNC_TOPIC` - топики Kafka

## 6. Следующие шаги

Для полной функциональности рекомендуется добавить:

1. **Валидация треков через CDN API** в EventProcessor при смене трека
2. **Rate limiting** на действия пользователя
3. **Конфликт-разрешение** при конкурентных событиях
4. **Кворум действий** для критических операций
5. **Ведение истории действий** в комнате
6. **Мониторинг и логирование** событий
7. **Тесты** для обоих сервисов

## 7. Запуск всей системы

```bash
# Запуск всех сервисов
docker-compose -f services/docker-compose.yml up -d

# Проверка статуса
docker-compose -f services/docker-compose.yml ps

# Просмотр логов
docker-compose -f services/docker-compose.yml logs -f ws-gateway
docker-compose -f services/docker-compose.yml logs -f session-service
```

## 8. Тестирование

### WebSocket Gateway

```bash
# Health check
curl http://localhost:8001/health

# WebSocket подключение (используйте wscat или другой WebSocket клиент)
wscat -c "ws://localhost:8001/ws?token=<JWT_TOKEN>&room_id=test_room"
```

### Session Service

```bash
# Health check
curl http://localhost:8081/api/rooms/health

# Создать комнату
curl -X POST http://localhost:8081/api/rooms/test_room

# Получить состояние комнаты
curl http://localhost:8081/api/rooms/test_room

# Добавить участника
curl -X POST "http://localhost:8081/api/rooms/test_room/participants?userId=user1&username=User1"
```

