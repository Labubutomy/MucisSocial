# Tracks Service

Микросервис для управления треками в музыкальной социальной сети. Поддерживает создание треков, управление артистами и хранение метаданных о треках.

## 🏗️ Архитектура

Сервис реализует два интерфейса:

- **HTTP API** (порт 8080) - для API Gateway, предоставляет REST API для получения информации о треках
- **gRPC API** (порт 50051) - для межсервисного взаимодействия, используется для создания треков и обновления метаданных

### Стек технологий

- **Go 1.23** - основной язык
- **PostgreSQL** - база данных
- **gRPC** - межсервисная коммуникация
- **Protocol Buffers** - сериализация данных

## 📋 Требования

- Go 1.23+
- PostgreSQL 15+
- protoc (Protocol Buffers Compiler)
- Docker и Docker Compose (опционально)

## 🚀 Быстрый старт

### Локальная разработка

1. **Установите зависимости:**

```bash
# Установите protoc (macOS)
brew install protobuf

# Установите Go плагины для protoc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

2. **Настройте базу данных:**

```bash
# Запустите PostgreSQL через docker-compose
docker-compose up -d postgres

# Или используйте существующую БД
export DATABASE_URL=postgres://user:password@localhost:5432/tracks_db?sslmode=disable
```

3. **Сгенерируйте код из proto файлов:**

```bash
make generate
```

4. **Запустите сервис:**

```bash
# Через Makefile (автоматически генерирует proto)
make run

# Или напрямую
go run cmd/main.go
```

### Docker

```bash
# Собрать и запустить все сервисы
make docker-up

# Или вручную
docker-compose up --build
```

Сервис будет доступен:
- HTTP API: http://localhost:8080
- gRPC API: localhost:50051

## 📁 Структура проекта

```
services/tracks/
├── api/
│   ├── tracks.proto         # gRPC API определение
│   ├── tracks.pb.go         # Сгенерированные структуры
│   └── tracks_grpc.pb.go    # Сгенерированный gRPC код
├── cmd/
│   └── main.go              # Точка входа приложения
├── internal/
│   ├── models.go            # Модели данных
│   ├── repository.go        # Слой работы с БД
│   ├── service.go           # Бизнес-логика
│   ├── grpc_handler.go      # gRPC обработчики
│   ├── hadlers.go           # HTTP обработчики
│   └── utils.go             # Утилиты
├── migrations/
│   └── 001_init.sql         # Миграции БД
├── pkg/
│   └── utils.go             # Общие утилиты
├── Dockerfile               # Docker образ
├── docker-compose.yml       # Docker Compose конфигурация
├── Makefile                 # Автоматизация задач
└── README.md                # Документация
```

## 🗄️ База данных

### Схема БД

```sql
-- Таблица треков
tracks (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    genre VARCHAR(100),
    audio_url TEXT,
    cover_url TEXT,
    duration_seconds INTEGER,
    status VARCHAR(20),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
)

-- Таблица артистов
artists (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL
)

-- Связующая таблица (many-to-many)
track_artists (
    track_id UUID REFERENCES tracks(id) ON DELETE CASCADE,
    artist_id UUID REFERENCES artists(id) ON DELETE CASCADE,
    PRIMARY KEY (track_id, artist_id)
)
```

### Миграции

Миграции автоматически применяются при первом запуске PostgreSQL через docker-compose.

Для ручного применения:
```bash
psql $DATABASE_URL -f migrations/001_init.sql
```

## 🔌 API

### HTTP API (для API Gateway)

#### Получить список треков
```http
GET /api/tracks?limit=20&offset=0&artist_id={uuid}
```

**Ответ:**
```json
{
  "tracks": [
    {
      "id": "uuid",
      "title": "Song Title",
      "artists": [
        {"id": "uuid", "name": "Artist Name"}
      ],
      "genre": "Pop",
      "audio_url": "https://s3.../audio.mp3",
      "cover_url": "https://s3.../cover.jpg",
      "duration_seconds": 180,
      "status": "ready",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "limit": 20,
  "offset": 0
}
```

#### Получить трек по ID
```http
GET /api/tracks/{id}
```

#### Создать трек (Admin)
```http
POST /api/admin/tracks
Headers: X-User-Role: admin
Body:
{
  "title": "Song Title",
  "artist_ids": ["uuid1", "uuid2"],
  "genre": "Pop"
}
```

#### Обновить трек (Admin)
```http
PUT /api/admin/tracks/{id}
Headers: X-User-Role: admin
Body:
{
  "title": "New Title",
  "artist_ids": ["uuid1"],
  "genre": "Rock"
}
```

#### Удалить трек (Admin)
```http
DELETE /api/admin/tracks/{id}
Headers: X-User-Role: admin
```

#### Health Check
```http
GET /health
```

### gRPC API (для других сервисов)

#### CreateTrack

Создает новый трек с одним или несколькими артистами.

**Запрос:**
```protobuf
message CreateTrackRequest {
  string title = 1;
  repeated string artist_ids = 2;  // Массив UUID артистов
  int32 duration_sec = 3;
  string genre = 4;
}
```

**Ответ:**
```protobuf
message CreateTrackResponse {
  string track_id = 1;  // UUID созданного трека
}
```

**Пример использования:**
```go
req := &tracks.CreateTrackRequest{
    Title:      "My Song",
    ArtistIds:  []string{"artist-uuid-1", "artist-uuid-2"},
    DurationSec: 180,
    Genre:      "Pop",
}
resp, err := client.CreateTrack(ctx, req)
```

#### UpdateTrackInfo

Обновляет URLs трека (cover_url, audio_url). Обновляет только переданные непустые значения.

**Запрос:**
```protobuf
message UpdateTrackInfoRequest {
  string track_id = 1;
  string cover_url = 2;  // Путь до S3/Minio (опционально)
  string audio_url = 3;  // Путь до S3/Minio (опционально)
}
```

**Ответ:**
```protobuf
message UpdateTrackInfoResponse {
  bool success = 1;
}
```

**Пример использования:**
```go
req := &tracks.UpdateTrackInfoRequest{
    TrackId:  "track-uuid",
    CoverUrl: "https://s3.../cover.jpg",
    AudioUrl: "https://s3.../audio.mp3",
}
resp, err := client.UpdateTrackInfo(ctx, req)
```

## 🛠️ Makefile команды

```bash
make generate      # Генерировать код из proto файлов
make build         # Собрать приложение (с авто-генерацией proto)
make run           # Запустить приложение локально
make docker-build  # Собрать Docker образ
make docker-up     # Запустить все сервисы через docker-compose
make docker-down   # Остановить все сервисы
make help          # Показать все доступные команды
```

## 🔧 Переменные окружения

| Переменная | Описание | По умолчанию |
|-----------|----------|--------------|
| `DATABASE_URL` | URL подключения к PostgreSQL | `postgres://postgres:postgres@localhost:5432/tracks_db?sslmode=disable` |
| `PORT` | Порт HTTP сервера | `8080` |
| `GRPC_PORT` | Порт gRPC сервера | `50051` |

## 📊 Статусы треков

- `uploaded` - трек загружен, ожидает обработки
- `processing` - трек обрабатывается
- `ready` - трек готов к использованию
- `failed` - ошибка при обработке

## 🔐 Безопасность

- HTTP API требует заголовок `X-User-Role: admin` для административных операций
- gRPC API предназначен для внутреннего использования между сервисами
- Рекомендуется использовать TLS для gRPC в production


## 📝 Особенности реализации

- **Поддержка нескольких артистов**: Один трек может иметь несколько артистов через связующую таблицу `track_artists`
- **Автоматическая генерация**: Proto код генерируется автоматически при сборке Docker образа

## 🐛 Troubleshooting

### Ошибка: "could not import .../api"

**Решение:** Сгенерируйте код из proto файлов:
```bash
make generate
```

### Ошибка: "protoc-gen-go: program not found"

**Решение:** Убедитесь, что `$GOPATH/bin` в PATH:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Ошибка подключения к БД

**Решение:** Проверьте, что PostgreSQL запущен и `DATABASE_URL` корректна:
```bash
docker-compose up -d postgres
# Или проверьте подключение
psql $DATABASE_URL -c "SELECT 1"
```
