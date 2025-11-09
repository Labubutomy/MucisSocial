# Streaming Gateway Service

FastAPI-based origin service that provides signed playlists and segments for adaptive audio streaming (HLS/fMP4).

## Что реализовано

### Основной функционал

1. **Генерация подписанных URL для плейлистов**
   - Master playlist (master.m3u8) с вариантами качества
   - Variant playlists для каждого битрейта (aac_256/index.m3u8, aac_160/index.m3u8, aac_96/index.m3u8)
   - Подпись на основе HMAC-SHA256 с TTL

2. **Origin endpoint для отдачи контента**
   - Валидация подписей запросов
   - Динамическая перезапись плейлистов с новыми подписями для сегментов
   - Отдача бинарных сегментов (init.mp4, chunk_*.m4s) из MinIO
   - Работает без CDN (отдает контент напрямую) или через CDN (если указан `CDN_BASE_URL`)

3. **Adaptive Bitrate Streaming**
   - Поддержка 3 качеств: 256kbps, 160kbps, 96kbps
   - Сегменты fMP4 длительностью 4 секунды
   - HLS playlists с поддержкой #EXT-X-MAP для init сегментов

4. **Интеграция с MinIO/S3**
   - Чтение плейлистов и сегментов из объектного хранилища
   - Streaming для больших файлов

## Как это работает

### Архитектура запросов

```
Плеер → GET /api/stream/{track_id}
       ↓
   Получает master.m3u8 URL + variant URLs (все подписаны)
       ↓
Плеер → GET /origin/tracks/.../master.m3u8?sig=...&exp=...
       ↓
   Origin проверяет подпись, читает master.m3u8 из MinIO,
   переписывает variant URLs с новыми подписями
       ↓
Плеер → GET /origin/tracks/.../aac_256/index.m3u8?sig=...&exp=...
       ↓
   Origin проверяет подпись, читает variant playlist,
   переписывает сегменты (init.mp4, chunk_*.m4s) с новыми подписями
       ↓
Плеер → GET /origin/tracks/.../aac_256/chunk_001.m4s?sig=...&exp=...
       ↓
   Origin проверяет подпись, отдаёт бинарный сегмент из MinIO
```

### Подпись URL (HMAC-SHA256)

**Формат подписи:**
```
sig = HMAC-SHA256(secret, resource_path + expires_at)
URL = {BASE_URL}/origin/path/to/resource?exp=1234567890&sig=abc123...
```

По умолчанию используется `BASE_URL` (http://localhost:8000), если не указан `CDN_BASE_URL`.

**Логика:**
- `expires_at = current_time + TTL`
- Подпись вычисляется от конкатенации пути и времени истечения
- При верификации проверяется: подпись валидна И время не истекло

**TTL:**
- Плейлисты: по умолчанию 300 секунд (5 минут), настраивается через `PLAYLIST_TTL_SECONDS` (минимум 60, максимум 3600 секунд)
- Сегменты: по умолчанию 60 секунд, настраивается через `SEGMENT_TTL_SECONDS` (минимум 10, максимум 600 секунд)

### Перезапись плейлистов

Плейлисты в MinIO хранятся с относительными путями:
```m3u8
#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=256000
aac_256/index.m3u8
```

Origin endpoint динамически переписывает их, добавляя подписи:
```m3u8
#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=256000
aac_256/index.m3u8?exp=1234567890&sig=abc123...
```

Аналогично для variant playlists: переписываются пути к init.mp4 и chunk_*.m4s с новыми подписями.

## Быстрый старт

### Установка

```bash
poetry install
```

### Запуск

```bash
# Через uvicorn (рекомендуется)
poetry run uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

### Конфигурация

Настройка через переменные окружения (префикс `STREAMING_`):

**Обязательные:**
- `SIGNING_SECRET` – секрет для подписи URL (минимум 8 символов, по умолчанию `change-me`)

**URL настройки:**
- `BASE_URL` – базовый URL сервиса (по умолчанию `http://localhost:8000`). Используется для формирования streaming URLs.
- `CDN_BASE_URL` – опциональный URL CDN. Если не указан, используется `BASE_URL` (сервис работает без CDN)

**Опциональные:**
- `PLAYLIST_TTL_SECONDS` – TTL для плейлистов в секундах (по умолчанию `300`, минимум 60, максимум 3600)
- `SEGMENT_TTL_SECONDS` – TTL для сегментов в секундах (по умолчанию `60`, минимум 10, максимум 600)
- `AVAILABLE_BITRATES` – поддерживаемые битрейты (по умолчанию `256000,160000,96000`)

**MinIO/S3:**
- `MINIO_ENDPOINT` – хост:порт MinIO (по умолчанию `localhost:9000`)
- `MINIO_ACCESS_KEY` – ключ доступа (по умолчанию `minioadmin`)
- `MINIO_SECRET_KEY` – секретный ключ (по умолчанию `minioadmin`)
- `MINIO_BUCKET` – имя бакета (по умолчанию `tracks`)
- `MINIO_SECURE` – использовать HTTPS (по умолчанию `false`)
- `MINIO_REGION` – регион (опционально)

Поддерживается файл `.env` в корне проекта.

## Структура хранения в MinIO

### Структура файлов трека

```
tracks/
  {artist_id}/
    {track_id}/
      original/
        original.wav                   # Исходный аудио файл
      metadata/
        tech_meta.json                 # Технические метаданные (duration, sample_rate, channels, codec)
        loudness.json                  # Интегрированная громкость (LUFS, true peak)
      transcoded/
        master.m3u8                    # Master playlist для HLS
        aac_256/
          index.m3u8                   # Variant playlist для 256kbps
          init.mp4                     # Init segment для fMP4
          chunk_001.m4s                # Media segments (fMP4)
          chunk_002.m4s
          ...
        aac_160/
          index.m3u8
          init.mp4
          chunk_001.m4s
          chunk_002.m4s
          ...
        aac_96/
          index.m3u8
          init.mp4
          chunk_001.m4s
          chunk_002.m4s
          ...
```

**Важно:** Плейлисты в MinIO содержат относительные пути. Origin endpoint переписывает их на лету с подписями.

## API Endpoints

### `GET /api/stream/{track_id}`

Возвращает подписанные URL для master и variant плейлистов.

**Query параметры:**
- `artist_id` (обязательный) – идентификатор артиста
- `available_bitrates` (опциональный) – список битрейтов через запятую (например, `256000,160000,96000`). Если не указан, используются значения по умолчанию из настроек.

**Пример запроса:**
```
GET /api/stream/550e8400-e29b-41d4-a716-446655440000?artist_id=660e8400-e29b-41d4-a716-446655440001&available_bitrates=256000,160000,96000
```

**Response:**
```json
{
  "master_url": "http://localhost:8000/origin/tracks/{artist_id}/{track_id}/transcoded/master.m3u8?exp=123&sig=abc",
  "variants": [
    {
      "bitrate": 256000,
      "url": "http://localhost:8000/origin/tracks/{artist_id}/{track_id}/transcoded/aac_256/index.m3u8?exp=123&sig=def"
    },
    {
      "bitrate": 160000,
      "url": "http://localhost:8000/origin/tracks/{artist_id}/{track_id}/transcoded/aac_160/index.m3u8?exp=123&sig=ghi"
    },
    {
      "bitrate": 96000,
      "url": "http://localhost:8000/origin/tracks/{artist_id}/{track_id}/transcoded/aac_96/index.m3u8?exp=123&sig=jkl"
    }
  ],
  "expires_in": 300
}
```

**Примечание:** URL формируются на основе `BASE_URL` (по умолчанию `http://localhost:8000`) или `CDN_BASE_URL` (если указан).

### `POST /api/stream/refresh`

Обновляет подписанные URL для трека (генерирует новые подписи с новым TTL).

**Request:**
```json
{
  "track_id": "550e8400-e29b-41d4-a716-446655440000",
  "artist_id": "660e8400-e29b-41d4-a716-446655440001",
  "available_bitrates": [256000, 160000, 96000]
}
```

**Поля:**
- `track_id` (обязательный) – идентификатор трека
- `artist_id` (обязательный) – идентификатор артиста
- `available_bitrates` (опциональный) – список битрейтов. Если не указан, используются значения по умолчанию из настроек.

**Response:** Аналогичен `GET /api/stream/{track_id}`

### `GET /origin/{resource_path}`

Origin endpoint для отдачи контента. Валидирует подпись и отдаёт ресурс из MinIO.

**Query параметры:**
- `exp` – timestamp истечения подписи
- `sig` – HMAC-SHA256 подпись

**Логика:**
1. Проверяет наличие `exp` и `sig`
2. Валидирует подпись и время истечения
3. Если это `.m3u8` – читает из MinIO и переписывает с новыми подписями
4. Если это бинарный файл (`.m4s`, `.mp4`) – стримит из MinIO

**Ошибки:**
- `403 Forbidden` – отсутствуют параметры подписи, неверная подпись, истёкший TTL
- `404 Not Found` – ресурс не найден в MinIO

### `GET /health`

Health check endpoint для мониторинга состояния сервиса.

**Response:**
```json
{
  "status": "healthy",
  "service": "Streaming Gateway",
  "version": "0.1.0"
}
```

## Структура проекта

```
.
├── main.py                 # Точка входа FastAPI приложения
├── player.html             # Тестовый HTML плеер для проверки стриминга
├── api/
│   └── v1/
│       ├── router.py       # Главный роутер API v1
│       └── endpoints/
│           ├── stream.py   # GET /api/stream/{track_id}, POST /api/stream/refresh
│           └── origin.py  # GET /origin/{path}
├── core/
│   ├── config.py           # Настройки приложения (Pydantic Settings)
│   └── dependencies.py     # Dependency injection для FastAPI
├── services/
│   ├── streaming.py        # Бизнес-логика генерации подписанных URL
│   ├── signing.py          # HMAC подпись и верификация URL
│   ├── storage.py          # Интеграция с MinIO/S3
│   ├── tracks.py           # Сервис получения метаданных треков (не используется, клиент передает данные напрямую)
│   └── exceptions.py       # Кастомные исключения
├── schemas/
│   ├── stream.py           # Pydantic модели для API ответов
│   └── track.py            # Модель метаданных трека
├── utils/
│   └── playlists.py        # Утилиты перезаписи HLS плейлистов
└── tests/
    └── test_streaming_gateway.py  # Интеграционные тесты
```

## Тестирование

```bash
poetry run pytest
```

Тесты покрывают:
- Генерацию подписанных URL
- Валидацию подписей
- Перезапись master и variant плейлистов
- Отдачу сегментов через origin endpoint

## Тестовый HTML плеер

В корне проекта есть файл `player.html` - простой веб-интерфейс для тестирования стриминга.

### Использование

1. Запустите сервис:
   ```bash
   poetry run uvicorn main:app --reload --host 0.0.0.0 --port 8000
   ```

2. Откройте `player.html` в браузере (можно через простой HTTP сервер):
   ```bash
   # Python 3
   python -m http.server 8080
   # Затем откройте http://localhost:8080/player.html
   ```

3. Введите:
   - Track ID: `1`
   - Artist ID: `1`
   - Available Bitrates: `256000,160000,96000` (опционально)

4. Нажмите "Загрузить стрим" для воспроизведения.

**Примечание:** Для работы плеера необходимо:
- Сервис по умолчанию работает без CDN и использует свой `BASE_URL` (по умолчанию `http://localhost:8000`)
- Иметь реальные HLS файлы в MinIO по пути `tracks/{artist_id}/{track_id}/transcoded/...`

## Будущие улучшения

- [ ] Проверка прав доступа (JWT, подписка, geo, ban)
- [ ] Метрики и мониторинг (request rate, cache miss rate, bandwidth)
- [ ] Кеширование метаданных треков
- [ ] Поддержка других форматов (DASH, Progressive)