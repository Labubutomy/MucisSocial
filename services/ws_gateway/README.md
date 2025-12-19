# WebSocket Gateway Service

WebSocket Gateway для сервиса совместного прослушивания музыки.

## Описание

Сервис управляет WebSocket соединениями клиентов, маршрутизирует сообщения между клиентами и Session Service через Message Bus (Kafka).

## Функциональность

- WebSocket endpoint для подключения клиентов
- JWT аутентификация
- Управление соединениями через Redis
- Интеграция с Kafka для отправки событий и получения синхронизаций
- Пинг-понг для поддержания соединений

## Конфигурация

Переменные окружения (префикс `WS_GATEWAY_`):

- `HOST` - хост сервера (по умолчанию: 0.0.0.0)
- `PORT` - порт сервера (по умолчанию: 8001)
- `JWT_SECRET` - секретный ключ для JWT
- `JWT_ALGORITHM` - алгоритм JWT (по умолчанию: HS256)
- `REDIS_HOST` - хост Redis (по умолчанию: redis)
- `REDIS_PORT` - порт Redis (по умолчанию: 6379)
- `REDIS_PASSWORD` - пароль Redis (опционально)
- `KAFKA_BROKERS` - брокеры Kafka (по умолчанию: redpanda:9092)
- `KAFKA_EVENTS_TOPIC` - топик для событий (по умолчанию: music-session-events)
- `KAFKA_SYNC_TOPIC` - топик для синхронизаций (по умолчанию: music-session-sync)
- `PING_INTERVAL` - интервал пинга в секундах (по умолчанию: 30)
- `LOG_LEVEL` - уровень логирования (по умолчанию: INFO)

## API

### WebSocket Endpoint

```
ws://localhost:8001/ws?token=<JWT_TOKEN>&room_id=<ROOM_ID>
```

### Health Check

```
GET /health
```

## Форматы сообщений

### Входящие сообщения (Client → Gateway)

```json
{
  "type": "player_action",
  "room_id": "room_123",
  "user_id": "user_456",
  "action": "play|pause|seek|change_track",
  "payload": {
    "track_id": "track_789",
    "position": 125.5,
    "timestamp": 1633046400
  }
}
```

### Исходящие сообщения (Gateway → Client)

```json
{
  "type": "sync_state",
  "room_id": "room_123",
  "state": {
    "current_track": "track_789",
    "position": 130.5,
    "is_playing": true,
    "last_action_by": "user_456",
    "timestamp": 1633046400
  }
}
```

