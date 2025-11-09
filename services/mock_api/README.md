## Mock API Service

Лёгкий FastAPI-сервис, эмулирующий необходимые для веб-клиента эндпоинты (`tracks`, `cdn`, `users`, `artists`, `playlists`). Состояние хранится в оперативной памяти и сбрасывается при перезапуске.

### Запуск

```bash
cd services/mock_api
poetry install
poetry run uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

Или через Docker:

```bash
docker build -t mock-api .
docker run -it --rm -p 8100:8000 mock-api
```

### Эндпоинты

- `/tracks`, `/tracks/{id}`, `/tracks/{id}/recommendations`, `/tracks/search`, `/tracks/{id}/like`, `/tracks/search/trending`
- `/hls/{trackId}/master.m3u8` и связанные плейлисты/сегменты
- `/auth/sign-in`, `/auth/sign-up`, `/me`, `/me/playlists`, `/me/search-history`
- `/artists/{id}`, `/artists/trending`, `/artists/search`
- `/playlists/{id}`, `/playlists/{id}/tracks`, `/playlists`, `/playlists/{id}/tracks` (POST), `/users/{userId}/playlists`
- `/health`

Начальные данные соответствуют треку `1` («Терновый куст») артиста `Vulpes Vult` (`id=1`).

