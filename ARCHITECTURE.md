# Архитектура Music Social - Сервис стриминга музыки

## Обзор системы

Music Social - это микросервисная платформа для стриминга музыки с поддержкой совместного прослушивания, управления плейлистами, артистами и треками. Система построена на принципах микросервисной архитектуры с использованием различных технологий для разных сервисов.

## Общая архитектура

```
┌─────────────────────────────────────────────────────────────────┐
│                         Клиентские приложения                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Web Client  │  │    Player    │  │  Mobile App   │          │
│  │  (React)    │  │  (HTML5)     │  │  (Future)     │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
└─────────┼─────────────────┼─────────────────┼──────────────────┘
          │                 │                 │
          └─────────────────┼─────────────────┘
                            │
          ┌─────────────────▼─────────────────┐
          │      Nginx Reverse Proxy          │
          │  (Port 80, 81, 8000)              │
          └─────────────────┬─────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
┌───────▼────────┐  ┌───────▼────────┐  ┌───────▼────────┐
│  API Gateway   │  │  WebSocket     │  │  CDN Service   │
│  (Go, Port 8080)│  │  Gateway      │  │  (Python)      │
│                │  │  (Port 8001)   │  │  (Port 8000)   │
└───────┬────────┘  └───────┬────────┘  └───────┬────────┘
        │                   │                   │
        │                   │                   │
┌───────▼───────────────────▼───────────────────▼────────┐
│              Микросервисы (gRPC/HTTP)                   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐│
│  │  Users   │  │ Artists  │  │  Tracks  │  │ Playlist ││
│  │ (gRPC)   │  │ (gRPC)   │  │(gRPC/HTTP)│  │ (HTTP)   ││
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘│
│       │            │             │             │        │
│  ┌────▼─────┐  ┌───▼─────┐  ┌───▼─────┐  ┌───▼─────┐  │
│  │PostgreSQL│  │PostgreSQL│  │PostgreSQL│  │PostgreSQL│ │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘ │
└──────────────────────────────────────────────────────────┘
        │
┌───────▼──────────────────────────────────────────────────┐
│         Сервисы обработки и стриминга                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │  Upload  │  │Transcoder│  │ Streaming │  │  Session │ │
│  │ (gRPC)   │  │(Consumer)│  │ (Python)  │  │ (Kotlin) │ │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘ │
└───────┼─────────────┼─────────────┼─────────────┼───────┘
        │             │             │             │
┌───────▼─────────────▼─────────────▼─────────────▼───────┐
│              Инфраструктура                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │  MinIO   │  │ Redpanda │  │  Redis   │  │ MongoDB  │ │
│  │ (S3)     │  │ (Kafka)  │  │          │  │          │ │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘ │
└──────────────────────────────────────────────────────────┘
```

## Компоненты системы

### 1. Клиентские приложения

#### Web Client (React + TypeScript + Vite)
- **Расположение**: `apps/web/`
- **Технологии**: React, TypeScript, Vite, Tailwind CSS
- **Порт**: 80 (через Nginx)
- **Функции**:
  - Пользовательский интерфейс для управления музыкой
  - Просмотр треков, артистов, плейлистов
  - Управление профилем пользователя
  - Интеграция с API Gateway

#### Player (HTML5)
- **Расположение**: `apps/player/`
- **Технологии**: HTML5, JavaScript
- **Порт**: 81 (через Nginx)
- **Функции**:
  - Тестовый плеер для проверки стриминга
  - Воспроизведение HLS потоков

### 2. API Gateway

- **Расположение**: `services/gateway/`
- **Технологии**: Go, gRPC, JWT
- **Порт**: 8080
- **Функции**:
  - Единая точка входа для всех HTTP запросов
  - Трансляция HTTP REST в gRPC вызовы
  - JWT аутентификация и авторизация
  - CORS поддержка
  - Централизованная обработка ошибок

**Интегрированные сервисы**:
- Users Service (gRPC)
- Artists Service (gRPC)
- Tracks Service (HTTP)
- Playlist Service (HTTP)
- Upload Service (gRPC)

### 3. Микросервисы

#### Users Service
- **Технологии**: Go, gRPC, PostgreSQL
- **Порт**: 50051 (gRPC)
- **База данных**: PostgreSQL (порт 5432)
- **Функции**:
  - Регистрация и аутентификация пользователей
  - Управление профилями
  - История поиска
  - JWT токены (access + refresh)

#### Artists Service
- **Технологии**: Go, gRPC, PostgreSQL
- **Порт**: 50052 (gRPC)
- **База данных**: PostgreSQL (порт 5433)
- **Функции**:
  - CRUD операции для артистов
  - Поиск артистов
  - Получение трендовых артистов

#### Tracks Service
- **Технологии**: Go, gRPC, HTTP REST, PostgreSQL
- **Порты**: 50053 (gRPC), 8080 (HTTP)
- **База данных**: PostgreSQL (порт 5434)
- **Функции**:
  - Управление треками
  - Связь треков с артистами (many-to-many)
  - Метаданные треков
  - Статусы обработки (uploaded, processing, ready, failed)

#### Playlist Service
- **Технологии**: Go, HTTP REST, PostgreSQL, Gin
- **Порт**: 50054 (gRPC)
- **База данных**: PostgreSQL (порт 5435)
- **Функции**:
  - Создание и управление плейлистами
  - Управление треками в плейлистах
  - Подписки пользователей на плейлисты

### 4. Сервисы обработки

#### Upload Service
- **Технологии**: Go, gRPC
- **Порт**: 50055 (gRPC)
- **Функции**:
  - Прием аудиофайлов через gRPC streaming
  - Сохранение в MinIO
  - Создание записи в Tracks Service
  - Отправка задачи в очередь транскодера (Redpanda)

**Процесс загрузки**:
1. Клиент отправляет метаданные (artist_ids, track_name, genre)
2. Клиент отправляет бинарные чанки файла
3. Сервис создает трек в Tracks Service
4. Файл сохраняется в MinIO: `{artist_id}/{track_id}/original/`
5. Задача отправляется в Redpanda (топик `transcoder-tasks`)

#### Transcoder Service
- **Технологии**: Go, FFmpeg
- **Функции**:
  - Потребление задач из Redpanda
  - Транскодирование аудио в несколько битрейтов (256kbps, 160kbps, 96kbps)
  - Генерация HLS плейлистов и fMP4 сегментов
  - Сохранение результатов в MinIO

**Структура в MinIO после транскодирования**:
```
{artist_id}/{track_id}/
  ├── original/
  │   └── original.wav
  ├── metadata/
  │   ├── tech_meta.json
  │   └── loudness.json
  └── transcoded/
      ├── master.m3u8
      ├── aac_256/
      │   ├── index.m3u8
      │   ├── init.mp4
      │   └── chunk_*.m4s
      ├── aac_160/
      │   └── ...
      └── aac_96/
          └── ...
```

#### Streaming Service
- **Технологии**: Python, FastAPI
- **Порт**: 8000
- **Функции**:
  - Генерация подписанных URL для плейлистов и сегментов
  - HMAC-SHA256 подпись с TTL
  - Origin endpoint для отдачи контента из MinIO
  - Динамическая перезапись HLS плейлистов с новыми подписями
  - Adaptive Bitrate Streaming (HLS/fMP4)

**API Endpoints**:
- `GET /api/stream/{track_id}` - получение подписанных URL
- `GET /origin/{resource_path}?exp=...&sig=...` - отдача контента

#### CDN Service
- **Технологии**: Python, FastAPI
- **Порт**: 8000 (через Nginx)
- **Функции**:
  - Кеширование плейлистов и сегментов
  - Проксирование запросов к Streaming Service
  - Настраиваемые TTL для разных типов контента

### 5. Сервисы синхронизации

#### Session Service
- **Технологии**: Kotlin, Spring Boot, MongoDB, Redis, Kafka
- **Порты**: 8081 (HTTP), 50056 (gRPC)
- **Функции**:
  - Управление комнатами совместного прослушивания
  - Синхронизация состояний воспроизведения
  - Обработка событий (play, pause, seek, change_track)
  - Хранение состояний в MongoDB
  - Кеширование в Redis

**Протокол работы**:
1. Клиент отправляет событие через WebSocket Gateway
2. Gateway отправляет событие в Kafka (топик `music-session-events`)
3. Session Service обрабатывает событие и обновляет состояние
4. Session Service отправляет синхронизацию в Kafka (топик `music-session-sync`)
5. Gateway получает синхронизацию и рассылает всем клиентам комнаты

#### WebSocket Gateway
- **Технологии**: Python, FastAPI, WebSockets
- **Порт**: 8001
- **Функции**:
  - Управление WebSocket соединениями
  - JWT аутентификация
  - Маршрутизация сообщений между клиентами и Session Service
  - Интеграция с Kafka для событий и синхронизаций
  - Управление соединениями через Redis

**WebSocket Endpoint**:
```
ws://localhost:8001/ws?token=<JWT_TOKEN>&room_id=<ROOM_ID>
```

### 6. Инфраструктура

#### MinIO (S3-совместимое хранилище)
- **Порт**: 9000 (API), 9001 (Console)
- **Функции**:
  - Хранение оригинальных аудиофайлов
  - Хранение транскодированных сегментов и плейлистов
  - Хранение метаданных

#### Redpanda (Kafka-совместимый брокер)
- **Порты**: 19092 (Kafka), 18081 (Schema Registry), 18082 (Proxy)
- **Функции**:
  - Очередь задач для транскодера
  - События сессий совместного прослушивания
  - Синхронизация состояний

**Топики**:
- `transcoder-tasks` - задачи транскодирования
- `music-session-events` - события от клиентов
- `music-session-sync` - синхронизация состояний

#### PostgreSQL
- **Экземпляры**:
  - Users DB: порт 5432
  - Artists DB: порт 5433
  - Tracks DB: порт 5434
  - Playlist DB: порт 5435

#### Redis
- **Порт**: 6379
- **Функции**:
  - Кеширование соединений WebSocket Gateway
  - Кеширование состояний Session Service

#### MongoDB
- **Порт**: 27017
- **Функции**:
  - Хранение состояний комнат Session Service

#### Nginx Reverse Proxy
- **Порты**: 80 (Web Client), 81 (Player), 8000 (CDN)
- **Функции**:
  - Маршрутизация запросов к клиентским приложениям
  - Проксирование к CDN

## Потоки данных

### Загрузка и обработка трека

```
1. Клиент → Upload Service (gRPC streaming)
   └─ Метаданные + бинарные чанки

2. Upload Service → Tracks Service (gRPC)
   └─ CreateTrack(track_name, artist_ids, genre)

3. Upload Service → MinIO
   └─ Сохранение: {artist_id}/{track_id}/original/original.wav

4. Upload Service → Redpanda
   └─ Отправка задачи в топик transcoder-tasks

5. Transcoder Service ← Redpanda
   └─ Потребление задачи

6. Transcoder Service → MinIO
   └─ Чтение оригинального файла

7. Transcoder Service → FFmpeg
   └─ Транскодирование в 3 битрейта

8. Transcoder Service → MinIO
   └─ Сохранение плейлистов и сегментов

9. Transcoder Service → Tracks Service (gRPC)
   └─ UpdateTrackInfo(cover_url, audio_url)
```

### Воспроизведение трека

```
1. Клиент → API Gateway (HTTP)
   └─ GET /api/v1/tracks/{track_id}

2. API Gateway → Tracks Service (HTTP)
   └─ Получение метаданных трека

3. Клиент → Streaming Service (HTTP)
   └─ GET /api/stream/{track_id}?artist_id=...

4. Streaming Service → MinIO
   └─ Проверка наличия плейлистов

5. Streaming Service → Клиент
   └─ Подписанные URL для master.m3u8 и variants

6. Плеер → Streaming Service / CDN
   └─ GET /origin/{path}?exp=...&sig=...
   └─ Получение плейлистов и сегментов

7. Streaming Service / CDN → MinIO
   └─ Чтение и отдача контента
```

### Совместное прослушивание

```
1. Клиент → WebSocket Gateway (WebSocket)
   └─ Подключение: ws://.../ws?token=...&room_id=...

2. Клиент → WebSocket Gateway
   └─ Отправка события: {type: "player_action", action: "play", ...}

3. WebSocket Gateway → Redpanda
   └─ Отправка в топик music-session-events

4. Session Service ← Redpanda
   └─ Потребление события

5. Session Service → MongoDB
   └─ Обновление состояния комнаты

6. Session Service → Redis
   └─ Кеширование состояния

7. Session Service → Redpanda
   └─ Отправка синхронизации в топик music-session-sync

8. WebSocket Gateway ← Redpanda
   └─ Потребление синхронизации

9. WebSocket Gateway → Все клиенты комнаты
   └─ Рассылка обновленного состояния
```

## Технологический стек

### Backend
- **Go 1.23+** - основной язык для микросервисов
- **Python 3** - Streaming, CDN, WebSocket Gateway
- **Kotlin + Spring Boot** - Session Service
- **gRPC** - межсервисное взаимодействие
- **HTTP REST** - публичные API

### Frontend
- **React + TypeScript** - Web Client
- **Vite** - сборщик
- **Tailwind CSS** - стилизация

### Базы данных
- **PostgreSQL 15** - реляционные данные
- **MongoDB 7** - документо-ориентированные данные (сессии)
- **Redis 7** - кеширование

### Message Queue
- **Redpanda** - Kafka-совместимый брокер

### Хранилище
- **MinIO** - S3-совместимое объектное хранилище

### Инфраструктура
- **Docker + Docker Compose** - контейнеризация
- **Nginx** - reverse proxy
- **FFmpeg** - транскодирование аудио

## Безопасность

- **JWT аутентификация** - access и refresh токены
- **HMAC-SHA256 подпись** - для защищенных URL стриминга
- **TTL для подписей** - ограничение времени жизни URL
- **CORS** - настройка в API Gateway
- **Bcrypt** - хэширование паролей
- **Prepared statements** - защита от SQL injection

## Масштабируемость

- **Микросервисная архитектура** - независимое масштабирование сервисов
- **Горизонтальное масштабирование** - каждый сервис может масштабироваться отдельно
- **Асинхронная обработка** - транскодирование через очередь
- **Кеширование** - CDN и Redis для снижения нагрузки
- **Connection pooling** - для баз данных

## Мониторинг и логирование

- **Health checks** - для всех сервисов
- **Структурированное логирование** - zap для Go, стандартное для Python/Kotlin
- **Graceful shutdown** - корректное завершение работы

## Развертывание

Проект использует Docker Compose для оркестрации всех сервисов:

```bash
# Запуск всей инфраструктуры
make infra-up

# Запуск всех сервисов
make dev-up

# Запуск клиентских приложений
make apps-up
```

Все сервисы объединены в Docker сеть `music-network` для внутренней коммуникации.

## Дальнейшее развитие

- Мониторинг: Prometheus + Grafana
- Трассировка: Jaeger/OpenTelemetry
- Rate limiting в API Gateway
- CI/CD пайплайны
- Unit и integration тесты
- Поддержка DASH и Progressive streaming
- Geo-блокировка и проверка прав доступа

