# Session Service

Сервис управления состояниями комнат прослушивания музыки.

## Описание

Session Service обрабатывает события от клиентов через Message Bus (Kafka), управляет состояниями комнат, синхронизирует состояния между всеми участниками комнаты.

## Функциональность

- Управление комнатами (создание/удаление)
- Добавление/удаление участников в комнаты
- Обработка событий воспроизведения (play, pause, seek, change_track)
- Синхронизация состояний через Kafka
- Хранение состояний в MongoDB
- Кэширование в Redis

## Технологии

- Kotlin
- Spring Boot 4.0.0
- Spring Data MongoDB
- Spring Data Redis
- Spring Kafka
- Java 21

## Конфигурация

Переменные окружения:

- `SERVER_PORT` - порт сервера (по умолчанию: 8080)
- `MONGODB_HOST` - хост MongoDB (по умолчанию: mongodb)
- `MONGODB_PORT` - порт MongoDB (по умолчанию: 27017)
- `MONGODB_DATABASE` - имя базы данных (по умолчанию: music_sessions)
- `MONGODB_USERNAME` - имя пользователя MongoDB (опционально)
- `MONGODB_PASSWORD` - пароль MongoDB (опционально)
- `REDIS_HOST` - хост Redis (по умолчанию: redis)
- `REDIS_PORT` - порт Redis (по умолчанию: 6379)
- `REDIS_PASSWORD` - пароль Redis (опционально)
- `REDIS_DB` - номер базы данных Redis (по умолчанию: 0)
- `KAFKA_BROKERS` - брокеры Kafka (по умолчанию: redpanda:9092)
- `KAFKA_EVENTS_TOPIC` - топик для событий (по умолчанию: music-session-events)
- `KAFKA_SYNC_TOPIC` - топик для синхронизаций (по умолчанию: music-session-sync)
- `LOG_LEVEL` - уровень логирования (по умолчанию: INFO)

## API

### Health Check

```
GET /api/rooms/health
```

### Получить состояние комнаты

```
GET /api/rooms/{roomId}
```

### Создать комнату

```
POST /api/rooms/{roomId}
```

### Добавить участника

```
POST /api/rooms/{roomId}/participants?userId={userId}&username={username}
```

### Удалить участника

```
DELETE /api/rooms/{roomId}/participants/{userId}
```

### Удалить комнату

```
DELETE /api/rooms/{roomId}
```

## Модели данных

### RoomState

```kotlin
data class RoomState(
    val roomId: String,
    val currentTrack: TrackInfo?,
    val position: Double,
    val isPlaying: Boolean,
    val participants: List<Participant>,
    val queue: List<TrackInfo>,
    val lastAction: Action?,
    val createdAt: Instant,
    val updatedAt: Instant
)
```

### ClientEvent (из Kafka)

```kotlin
data class ClientEvent(
    val eventId: String,
    val type: EventType,
    val roomId: String,
    val userId: String,
    val action: PlayerAction,
    val payload: Map<String, Any>,
    val clientTimestamp: Instant,
    val serverTimestamp: Instant
)
```

### SyncMessage (в Kafka)

```kotlin
data class SyncMessage(
    val syncId: String,
    val roomId: String,
    val state: RoomState,
    val triggeredBy: String,
    val timestamp: Instant
)
```

## Протокол работы

1. Клиент отправляет событие через WebSocket Gateway
2. Gateway отправляет событие в Kafka (топик `music-session-events`)
3. Session Service обрабатывает событие и обновляет состояние комнаты
4. Session Service отправляет синхронизацию в Kafka (топик `music-session-sync`)
5. Gateway получает синхронизацию и рассылает всем клиентам комнаты

