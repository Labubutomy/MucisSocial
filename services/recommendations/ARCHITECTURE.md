# Архитектура сервиса рекомендаций

> 📊 **Визуальные диаграммы:** См. [ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md) для схем архитектуры в формате ASCII.

## Обзор

Сервис рекомендаций — это ML-ready микросервис для генерации персонализированных музыкальных рекомендаций. Он использует событийно-ориентированную архитектуру и модульный pipeline для генерации рекомендаций.

## Компоненты системы

### 1. Точка входа (`cmd/main.go`)

Инициализирует все компоненты системы:
- In-memory хранилища данных
- MinIO backup для персистентности
- Kafka/Redpanda consumers для обработки событий
- Bootstrap для начальной загрузки данных
- HTTP API сервер
- Recommendation Engine

### 2. Хранилища данных (`internal/store/`)

#### 2.1. TrackStore (`track_store.go`)
**Назначение:** Хранение метаданных треков в памяти

**Структура:**
- `tracks map[string]*Track` — основное хранилище треков
- `genreIndex map[string]map[string]struct{}` — индекс по жанрам для быстрого поиска
- `artistIndex map[string]map[string]struct{}` — индекс по артистам

**Методы:**
- `Get(trackID)` — получить трек по ID
- `Upsert(track)` — создать/обновить трек
- `GetByGenre(genre)` — получить треки по жанру
- `GetByArtist(artistID)` — получить треки по артисту
- `GetAll()` — получить все треки
- `GetNewReleases(limit, days)` — получить новые релизы за последние N дней

**Особенности:**
- Thread-safe (использует `sync.RWMutex`)
- Автоматическое обновление индексов при изменении данных

#### 2.2. UserProfileStore (`user_profile_store.go`)
**Назначение:** Хранение профилей пользователей с историей прослушиваний

**Структура:**
- `profiles map[string]*UserProfile` — профили пользователей

**UserProfile содержит:**
- `UserID` — идентификатор пользователя
- `GenreListenCount map[string]int` — количество прослушиваний по жанрам
- `ArtistListenCount map[string]int` — количество прослушиваний по артистам
- `ListenedTracks map[string]struct{}` — множество прослушанных треков

**Методы:**
- `Get(userID)` — получить профиль пользователя
- `GetOrCreate(userID)` — получить или создать профиль
- `Update(profile)` — обновить профиль

**Особенности:**
- Автоматическая инициализация maps при загрузке из backup (защита от nil map panic)
- Профили обновляются только через события, не через API

#### 2.3. GlobalStatsStore (`global_stats_store.go`)
**Назначение:** Хранение глобальной статистики прослушиваний

**Структура:**
- `playCounts map[string]int` — общее количество прослушиваний (all-time)
- `monthlyCounts map[string]int` — количество прослушиваний за текущий месяц
- `monthlyCache *monthlyChartCache` — кеш топ-треков месяца

**Методы:**
- `IncrementPlayCount(trackID)` — увеличить счетчик прослушиваний
- `GetPlayCount(trackID)` — получить общее количество прослушиваний
- `GetTopTracks(limit)` — получить топ-треки (all-time)
- `GetTopTracksForMonth(limit)` — получить топ-треки месяца (с кешированием на 1 минуту)
- `GetTopTracksInSet(trackIDs, limit)` — получить топ-треки из заданного множества

**Особенности:**
- Кеширование месячных чартов с TTL 1 минута
- Автоматическая инвалидация кеша при изменении статистики
- Отдельный учет месячной статистики для актуальных чартов

#### 2.4. MinIOBackupStore (`minio_backup_store.go`)
**Назначение:** Персистентность данных через MinIO (S3-совместимое хранилище)

**Функциональность:**
- Периодическое сохранение всех хранилищ в MinIO (по умолчанию каждые 5 минут)
- Загрузка данных при старте сервиса
- Graceful shutdown с финальным сохранением

**BackupData структура:**
```go
{
  Tracks: map[string]*Track,
  UserProfiles: map[string]*UserProfile,
  GlobalStats: map[string]int,      // all-time
  MonthlyStats: map[string]int,     // текущий месяц
  Timestamp: time.Time
}
```

### 3. Recommendation Engine (`internal/engine/`)

#### 3.1. Pipeline (`engine.go`)

5-этапный pipeline генерации рекомендаций:

```
1. Candidate Generation → 2. Filtering → 3. Scoring → 4. Ranking → 5. Limiting
```

**Этап 1: Candidate Generation**
- Несколько генераторов создают кандидатов параллельно
- Результаты объединяются в один set (дедупликация)

**Этап 2: Filtering (Hard Filters)**
- `AlreadyListenedFilter` — исключает уже прослушанные треки
- `ExplicitFilter` — исключает explicit контент (если запрошено)
- `GenreFilter` — фильтрует по жанрам (если указаны)
- `FreshnessFilter` — фильтрует по дате релиза (если указано)

**Этап 3: Scoring**
- Каждому кандидату присваивается числовой score
- Текущая реализация: `HeuristicScorer` (правила)
- **ML Integration Point:** можно заменить на `MLScorer`

**Этап 4: Ranking**
- Сортировка по score (по убыванию)
- Опциональный `ReRanker` для дополнительной пересортировки (diversity, etc.)

**Этап 5: Limiting**
- Применение лимита на количество результатов

#### 3.2. Candidate Generators (`generators/`)

**TopGenresGenerator**
- Анализирует топ-N жанров пользователя
- Для каждого жанра берет топ-треки по популярности
- Параметры: `topN` (количество жанров), `tracksPerGenre` (треков на жанр)

**TopArtistsGenerator**
- Анализирует топ-N артистов пользователя
- Для каждого артиста берет топ-треки по популярности
- Параметры: `topN` (количество артистов), `tracksPerArtist` (треков на артиста)

**FallbackGenerator**
- Используется, если у пользователя нет истории прослушиваний
- Берет топ-треки из глобальной статистики
- Если глобальная статистика пуста, использует все треки из TrackStore

#### 3.3. Scorers (`scorers/`)

**HeuristicScorer** (текущая реализация)
Формула score:
```
score = genreAffinity + artistAffinity + popularity + freshness
```

Где:
- `genreAffinity` — среднее количество прослушиваний жанров трека пользователем
- `artistAffinity` — количество прослушиваний артиста пользователем
- `popularity` — log(playCount + 1) для сглаживания
- `freshness` — 1 / (daysSinceRelease + 1) — бонус за новизну

**MLScorer** (заглушка для будущей ML интеграции)
- Интерфейс готов для замены
- Можно использовать обученную модель для scoring

### 4. Event Consumer (`internal/consumer/`)

#### 4.1. Kafka/Redpanda Integration

**Топики:**
- `track-events` — события создания/обновления треков
- `listening-events` — события прослушивания треков

**Обработка событий:**

**Track Events:**
```json
{
  "event_type": "track_created" | "track_updated",
  "track_id": "...",
  "artist_id": "...",
  "genres": ["rock", "indie"],
  "release_ts": 1734134400,
  "is_explicit": false
}
```
→ Обновляет `TrackStore`

**Listening Events:**
```json
{
  "event_type": "track_listened",
  "user_id": "...",
  "track_id": "...",
  "listened_seconds": 180,
  "ts": 1734134400
}
```
→ Обновляет:
- `UserProfileStore` (жанры, артисты, прослушанные треки)
- `GlobalStatsStore` (глобальная и месячная статистика)

**Fallback механизм:**
- Если трек не найден в `TrackStore` при обработке listening event
- Пытается загрузить трек из tracks-service через HTTP API
- Сохраняет трек в `TrackStore` для будущих запросов

### 5. Bootstrap (`internal/bootstrap/`)

**Назначение:** Начальная загрузка треков из tracks-service при старте

**Процесс:**
1. Запрашивает треки пагинацией (limit=100)
2. Фильтрует только треки со статусом "ready" или "active"
3. Заполняет `TrackStore` метаданными треков
4. Запускается в фоне, не блокирует старт сервиса

**Конфигурация:**
- `BOOTSTRAP_ENABLED=true/false` — включить/выключить
- `TRACKS_SERVICE_URL` — URL tracks-service

### 6. HTTP API (`internal/api/server.go`)

#### 6.1. Публичные endpoints

**POST /recommendations**
- Генерирует персонализированные рекомендации для пользователя
- Request:
```json
{
  "user_id": "user123",
  "limit": 20,
  "filters": {
    "exclude_explicit": true,
    "genres": ["rock"],
    "fresh_days": 30
  }
}
```
- Response:
```json
{
  "tracks": ["track_id_1", "track_id_2", ...]
}
```

**GET /charts/top?limit=24**
- Возвращает топ-треки текущего месяца
- Использует кешированный результат (TTL 1 минута)
- Response:
```json
{
  "tracks": [
    {
      "track_id": "...",
      "position": 1,
      "play_count": 150
    },
    ...
  ]
}
```

**GET /tracks/new?limit=24&days=30**
- Возвращает новые релизы за последние N дней
- Сортировка по дате релиза (новые первыми)
- Response:
```json
{
  "tracks": ["track_id_1", "track_id_2", ...]
}
```

**GET /users/:user_id/taste**
- Возвращает топ-10 жанров и артистов пользователя
- Response:
```json
{
  "user_id": "...",
  "top_genres": [
    {"genre": "rock", "count": 45},
    ...
  ],
  "top_artists": [
    {"artist_id": "...", "count": 30},
    ...
  ]
}
```

#### 6.2. Debug endpoints

**GET /debug/tracks**
- Возвращает все треки из TrackStore

**GET /debug/users/:user_id**
- Возвращает полный профиль пользователя

#### 6.3. Ingest endpoints (для тестирования без Kafka)

**POST /ingest/track**
- Ручное добавление трека (имитация track event)

**POST /ingest/listening**
- Ручное добавление listening event

### 7. Конфигурация (`internal/config/`)

**Переменные окружения:**

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `HTTP_PORT` | 8080 | Порт HTTP сервера |
| `KAFKA_BROKERS` | redpanda:9092 | Брокеры Kafka/Redpanda |
| `TRACK_EVENTS_TOPIC` | track-events | Топик для событий треков |
| `LISTENING_EVENTS_TOPIC` | listening-events | Топик для событий прослушивания |
| `TRACKS_SERVICE_URL` | "" | URL tracks-service (для bootstrap и fallback) |
| `BOOTSTRAP_ENABLED` | true | Включить начальную загрузку |
| `MINIO_ENDPOINT` | minio:9000 | MinIO endpoint |
| `MINIO_ACCESS_KEY` | minioadmin | MinIO access key |
| `MINIO_SECRET_KEY` | minioadmin | MinIO secret key |
| `MINIO_BUCKET_NAME` | recommendations | MinIO bucket |
| `BACKUP_INTERVAL` | 5m | Интервал сохранения backup |

## Потоки данных

### 1. Инициализация сервиса

```
1. Загрузка конфигурации
2. Инициализация in-memory хранилищ
3. Подключение к MinIO и загрузка backup (если есть)
4. Инициализация Recommendation Engine
5. Запуск Kafka consumers (в фоне)
6. Запуск Bootstrap (в фоне, если включен)
7. Запуск HTTP сервера
8. Запуск периодического backup
```

### 2. Обработка listening event

```
1. Kafka consumer получает событие
2. Парсинг JSON → ListeningEvent
3. Получение/создание UserProfile
4. Попытка получить трек из TrackStore
   └─ Если не найден → загрузка из tracks-service (fallback)
5. Обновление статистики:
   - UserProfile.GenreListenCount
   - UserProfile.ArtistListenCount
   - UserProfile.ListenedTracks
   - GlobalStatsStore.playCounts (all-time)
   - GlobalStatsStore.monthlyCounts (текущий месяц)
6. Инвалидация кеша месячных чартов
7. Сохранение обновленного UserProfile
```

### 3. Генерация рекомендаций

```
1. HTTP запрос POST /recommendations
2. Получение UserProfile из UserProfileStore
3. Запуск Recommendation Engine:
   a. Candidate Generation (несколько генераторов)
   b. Filtering (hard filters)
   c. Scoring (heuristic или ML)
   d. Ranking (сортировка по score)
   e. Limiting (применение лимита)
4. Возврат списка track IDs
```

## ML Integration Points

### 1. Scorer (Primary)
**Текущая реализация:** `HeuristicScorer` (правила)
**Замена:** `MLScorer` с обученной моделью

```go
// В cmd/main.go
scorer := scorers.NewMLScorer("path/to/model")
```

### 2. Candidate Generator (Secondary)
**Добавление:** Embedding-based retrieval

```go
embeddingGen := generators.NewEmbeddingGenerator(embeddingService)
generators := []engine.CandidateGenerator{
    genreGenerator,
    artistGenerator,
    embeddingGen,  // Новый ML-based генератор
    fallbackGenerator,
}
```

### 3. ReRanker (Optional)
**Добавление:** Diversity re-ranking

```go
recEngine.ReRanker = rerankers.NewDiversityReRanker()
```

## Особенности реализации

### 1. In-Memory Storage
- Все данные хранятся в памяти для максимальной производительности
- MinIO backup обеспечивает персистентность
- Периодическое сохранение предотвращает потерю данных

### 2. Thread Safety
- Все хранилища используют `sync.RWMutex` для безопасного доступа из горутин
- Kafka consumers работают в отдельных горутинах

### 3. Fallback Mechanisms
- Если трек не найден в TrackStore → загрузка из tracks-service
- Если нет глобальной статистики → использование всех треков из TrackStore
- Если у пользователя нет истории → FallbackGenerator

### 4. Caching
- Месячные чарты кешируются на 1 минуту
- Автоматическая инвалидация при изменении данных
- Проверка смены месяца для обновления кеша

### 5. Event-Driven Updates
- Профили пользователей обновляются только через события
- API не позволяет напрямую изменять профили
- Гарантирует консистентность данных

## Масштабирование

### Текущие ограничения:
- In-memory storage ограничено размером RAM
- Single instance (нет распределения)

### Возможные улучшения:
1. **Sharding:** Разделение пользователей по инстансам
2. **Redis:** Использование Redis вместо in-memory для shared state
3. **Database:** Миграция на PostgreSQL для больших объемов данных
4. **Caching Layer:** Redis для кеширования рекомендаций
5. **Load Balancing:** Несколько инстансов за load balancer

## Мониторинг и отладка

### Debug endpoints:
- `/debug/tracks` — все треки
- `/debug/users/:user_id` — профиль пользователя

### Логирование:
- Обработка событий (успех/ошибки)
- Fallback загрузка треков
- Bootstrap процесс
- Backup операции

### Health check:
- `GET /health` — проверка работоспособности сервиса

