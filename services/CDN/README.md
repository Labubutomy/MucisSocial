# CDN Service

Кеширующий слой (CDN) между Streaming Gateway и плеером. CDN выдаёт HLS/fMP4-плейлисты и сегменты из локального кеша, снимая нагрузку со Streaming Gateway и MinIO, и при этом не нарушает модель безопасности с подписанными URL.

## Что реализовано

1. **Кеширование контента**
   - master playlist (`master.m3u8`)
   - variant playlists (`index.m3u8` для каждого битрейта)
   - init-сегменты (`init.mp4`)
   - media-сегменты (`chunk_*.m4s`)
   - статические артефакты (`tech_meta.json`, обложки и т.д.)
   - CDN игнорирует query-параметры `exp` и `sig`, поэтому одинаковый ресурс, но с разными подписями, кешируется как один объект
   - отдельные TTL для плейлистов, сегментов и статических файлов

2. **Работа с подписанными URL**
   - CDN *не* проверяет подпись и не знает секретов
   - при cache miss весь исходный URL (с `sig` и `exp`) проксируется на origin (`Streaming Gateway`)
   - Origin валидирует подпись, отдаёт контент, CDN кеширует его и при следующем запросе возвращает уже из памяти

3. **Поддержка HLS/fMP4**
   - полноценный HLS с `#EXT-X-MAP` и тремя битрейтами (256, 160, 96 kbps)
   - плейлисты могут обновляться каждые 4–8 секунд — CDN позволяет переиспользовать immutable сегменты и контролируемо проксировать обновления плейлистов

4. **Мониторинг и аналитика**
   - health-check `/health`
   - `/stats` — агрегированные показатели (hit/miss, объём кеша)
   - `/cache/entries`, `/cache/entries/{id}`, `/cache/summary` — детальная аналитика по объектам кеша (тип ресурса, TTL, размер, hit-count, превью контента)
   - периодическое логирование статистики кеша (каждые 5 минут, если включено `CDN_LOG_CACHE_STATS`)

5. **Масштабирование и отказоустойчивость**
   - LRU-кеш в памяти, который легко масштабировать горизонтально добавлением POP-нод
   - готовность к замене in-memory кеша на Redis для shared state (см. раздел «Будущие улучшения»)
   - Health-check командой `curl -f http://<cdn>/health` позволяет оркестратору исключить «упавшие» инстансы

6. **Безопасность**
   - CDN не хранит и не логирует подписи
   - в аналитических ручках показываются только очищенные от `sig/exp` пути

## Как это работает

```
Плеер → GET /origin/tracks/.../master.m3u8?exp=...&sig=...
       ↓
     CDN проверяет кеш (exp/sig игнорируются)
       ↓
   [Cache HIT]  → моментально отдаёт контент из памяти
   [Cache MISS] → проксирует исходный запрос на Streaming Gateway
                     ↓
                 Origin валидирует подпись,
                 читает из MinIO и отвечает CDN
                     ↓
                 CDN кеширует ответ и отдаёт пользователю
```

## Быстрый старт

### Установка для локальной разработки

```bash
cd services/CDN
poetry install
poetry run uvicorn main:app --reload --host 0.0.0.0 --port 8080
```

### Docker

```bash
# из корня репозитория
docker compose up streaming cdn
```

> Streaming Gateway автоматически получает `CDN_BASE_URL=http://cdn:8080` из `services/docker-compose.yml`, поэтому выдаваемые им ссылки сразу указывают на CDN.

### Проверка связки Streaming Gateway + CDN

1. Запустите `docker compose up streaming cdn minio` (и необходимые зависимости).
2. Вызовите Streaming API, чтобы получить master playlist:
   ```bash
   curl "http://localhost:8000/api/stream/<track_id>?artist_id=<artist_id>"
   ```
   В ответе URL будут указывать на `http://localhost:8080/origin/...`.
3. Скачайте плейлист через CDN:
   ```bash
   curl -i "http://localhost:8080/origin/tracks/.../master.m3u8?exp=...&sig=..."
   ```
   - Первый запрос вернёт заголовок `X-CDN-Cache: MISS`, контент попадёт в кеш.
   - Повторный запрос покажет `X-CDN-Cache: HIT` и уменьшенный `X-CDN-TTL-Remaining`.

4. Посмотрите список объектов в кеше:
   ```bash
   curl http://localhost:8080/cache/entries | jq .
   ```

## Конфигурация

Все переменные конфигурации начинаются с префикса `CDN_` (см. `core/config.py`):

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `CDN_ORIGIN_BASE_URL` | `http://streaming:8000` | Базовый URL Streaming Gateway (`/origin/...`) |
| `CDN_CACHE_PLAYLIST_TTL` | `60` сек (мин 10, макс 300) | TTL для `.m3u8` |
| `CDN_CACHE_SEGMENT_TTL` | `3600` сек (мин 300, макс 86400) | TTL для `.m4s` и `.mp4` |
| `CDN_CACHE_STATIC_TTL` | `86400` сек (мин 300, макс 604800) | TTL для статических артефактов |
| `CDN_CACHE_MAX_SIZE` | `1000` | Количество элементов в in-memory-кеше |
| `CDN_HOST`, `CDN_PORT` | `0.0.0.0:8080` | Сетевые параметры приложения |
| `CDN_LOG_LEVEL` | `INFO` | Уровень логирования |
| `CDN_LOG_REQUESTS` | `true` | Логирование HIT/MISS |
| `CDN_LOG_CACHE_STATS` | `true` | Периодическое логирование статистики кеша |

Поддерживается `.env` в каталоге `services/CDN`.

## API Endpoints

### `GET /origin/{resource_path}`
Проксирует запрос к origin или возвращает содержимое из кеша. Заголовки ответа:
- `X-CDN-Cache: HIT|MISS`
- `X-CDN-Resource`: очищенный путь (`/origin/...`)
- `X-CDN-Resource-Type`: тип ресурса (`master_playlist`, `variant_playlist`, `init_segment`, `media_segment`, `static_asset`, `other`)
- `X-CDN-TTL` / `X-CDN-TTL-Remaining`: TTL для MISS и остаток TTL для HIT
- `X-CDN-Hit-Count`: число обращений (для HIT)

### `GET /health`
Простой health-check:
```json
{
  "status": "healthy",
  "service": "CDN Service",
  "version": "0.1.0"
}
```

### `GET /stats`
Сводные показатели кеша (hit/miss, объём, количество объектов).

### `GET /cache/entries`
Полный список объектов в кеше:
```json
{
  "total": 42,
  "entries": [
    {
      "cache_id": "...",
      "resource": "/origin/tracks/.../master.m3u8",
      "resource_type": "master_playlist",
      "size_bytes": 1024,
      "stored_at": 1731080100.12,
      "ttl_remaining": 27.5,
      "hit_count": 12,
      "...": "..."
    }
  ]
}
```

### `GET /cache/entries/{cache_id}?include_content=true`
Детальная информация по одному объекту. При `include_content=true` возвращается base64-превью первых 512 байт (для быстрой диагностики).

### `GET /cache/summary`
Агрегированная аналитика по типам ресурсов:
```json
{
  "total_entries": 42,
  "total_bytes": 7340032,
  "total_mb": 7.0,
  "by_type": {
    "media_segment": { "count": 30, "bytes": 6291456, "mb": 6.0, "avg_ttl_remaining": 3580.2 },
    "master_playlist": { "count": 1, "bytes": 2048, "mb": 0.0, "avg_ttl_remaining": 42.1 }
  }
}
```

## Интеграция со Streaming Gateway

1. В `services/docker-compose.yml` Streaming Gateway запускается с переменными:
   ```yaml
   environment:
     - BASE_URL=http://streaming:8000
     - CDN_BASE_URL=http://cdn:8080
   ```
2. Streaming Gateway формирует подписанные URL вида:
   ```
   http://cdn:8080/origin/tracks/{artist}/{track}/transcoded/aac_256/index.m3u8?exp=...&sig=...
   ```
3. CDN прозрачно обслуживает эти URL — плееру ничего менять не нужно.

## Структура проекта

```
services/CDN/
├── main.py             # Точка входа FastAPI
├── api/
│   └── router.py       # Роутер, health/stats, аналитика кеша и прокси endpoint
├── core/
│   ├── cache.py        # LRU in-memory кеш с аналитикой
│   └── config.py       # Pydantic Settings
├── services/
│   └── cdn.py          # Бизнес-логика CDN (проксирование, TTL, аналитика)
├── Dockerfile
├── pyproject.toml
└── README.md
```

## Масштабирование

- Несколько экземпляров CDN можно поставить за L4/L7-балансировщик, каждый будет держать собственный кеш.
- Для POP в разных регионах достаточно задублировать сервис с разными `CDN_ORIGIN_BASE_URL`.
- Для shared state можно вынести кеш в Redis или Memcached (планируется отдельно).

## Будущие улучшения

- [ ] Внешний кеш (Redis) для шаринга между POP
- [ ] Экспорт метрик Prometheus
- [ ] Rate limiting / защита от бурстов
- [ ] Geo-IP routing для выбора ближайшего POP
- [ ] HTTP/2 и HTTP/3
- [ ] Сжатие плейлистов (gzip/brotli)

