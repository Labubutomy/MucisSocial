# Transcoder Service

Сервис принимает задачи из `transcoder-tasks`, скачивает оригинал из MinIO, извлекает технические метаданные, считает громкость и подготавливает HLS-пачку в трёх битрейтах. После успешной обработки обновляет Track Service по gRPC (`UpdateTrackInfo`), передавая URL `master.m3u8`.

## Поток обработки

1. **Очередь Redpanda**

   - Consumer читает сообщения, подготовленные Upload Service (`track_id`, `artist_id`, `track_url`).
   - Повторные попытки обеспечиваются за счёт некоммитнутых сообщений.

2. **MinIO**

   - Оригинал скачивается в рабочую директорию.
   - После обработки обратно выгружаются:
     - `artist_id/track_id/metadata/tech_meta.json`
     - `artist_id/track_id/metadata/loudness.json`
     - `artist_id/track_id/transcoded/master.m3u8` и подпапки `aac_256`, `aac_160`, `aac_96` с fMP4 сегментами.

3. **Track Service**
   - Через gRPC вызывается `UpdateTrackInfo`, предоставляя:
     - `track_id`
     - `audio_url` (путь к `master.m3u8`)
     - `duration` (в секундах; берётся из ffprobe)
   - `cover_url` пока не заполняется (резерв под будущий функционал).

## Завершение работы

Контейнер ловит сигналы `SIGINT/SIGTERM`, делает graceful shutdown: consumer, MinIO и gRPC подключение Track Service закрываются корректно, незавершённые задачи останутся в очереди для повторной обработки.
