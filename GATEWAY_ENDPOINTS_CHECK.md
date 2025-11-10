# Проверка Gateway Endpoints

## ✅ Endpoints, используемые клиентом

### 1. Auth Endpoints
- ✅ `POST /api/v1/auth/sign-in` - Реализован, возвращает `{access_token, refresh_token, user}`
- ✅ `POST /api/v1/auth/sign-up` - Реализован, возвращает `{access_token, refresh_token, user}`
- ✅ `GET /api/v1/me` - Реализован, возвращает `{user: {...}}`

### 2. Tracks Endpoints
- ✅ `GET /api/v1/tracks` - Проксирует к tracks-service, возвращает `{tracks: [...], limit, offset}`
- ✅ `GET /api/v1/tracks/{trackId}` - Проксирует к tracks-service, возвращает трек напрямую

### 3. Playlists Endpoints
- ✅ `GET /api/v1/playlists` - Реализован, возвращает `{playlists: [...], total}`
- ✅ `POST /api/v1/playlists` - Реализован, возвращает `{playlist_id}`
- ✅ `GET /api/v1/playlists/{playlistId}` - Реализован, возвращает плейлист напрямую

### 4. Search Endpoints
- ✅ `GET /api/v1/me/search-history` - Реализован, возвращает `{items: [...]}`
- ✅ `POST /api/v1/me/search-history` - Реализован, возвращает `{item: {...}}`
- ✅ `DELETE /api/v1/me/search-history` - Реализован, возвращает `{success: true}`
- ✅ `GET /api/v1/artists/search` - Реализован, возвращает `{query, items: [...]}`

## ⚠️ Endpoints, которые остаются в mock-api (не реализованы в gateway)

### Tracks
- `GET /api/v1/tracks/search` - Поиск треков
- `GET /api/v1/tracks/{trackId}/recommendations` - Рекомендации
- `POST /api/v1/tracks/{trackId}/like` - Лайк трека

### Playlists
- `GET /api/v1/playlists/{playlistId}/tracks` - Треки плейлиста (gateway возвращает только track_id)
- `POST /api/v1/playlists/{playlistId}/tracks` - Добавить треки (gateway принимает один track_id)

### Search
- `GET /api/v1/tracks/search/trending` - Трендовые запросы
- `GET /api/v1/playlists/search` - Поиск плейлистов

### Users
- `GET /api/v1/users/{userId}/playlists` - Плейлисты другого пользователя

## 🔍 Проверка форматов ответов

### Auth
- ✅ `sign-in` / `sign-up`: `{access_token, refresh_token, user: {id, username, avatar_url}}`
- ✅ `me`: `{user: {id, username, avatar_url, music_taste_summary}}`

### Tracks
- ✅ `GET /tracks`: `{tracks: [{id, title, artists: [{id, name}], duration_seconds, cover_url, ...}], limit, offset}`
- ✅ `GET /tracks/{id}`: `{id, title, artists: [{id, name}], duration_seconds, cover_url, ...}`

### Playlists
- ✅ `GET /playlists`: `{playlists: [{id, user_id, name, description, is_private, tracks_count, ...}], total}`
- ✅ `POST /playlists`: `{playlist_id}`
- ✅ `GET /playlists/{id}`: `{id, user_id, name, description, is_private, tracks_count, created_at, updated_at}`

### Search
- ✅ `GET /me/search-history`: `{items: [{id, query, created_at, ...}]}`
- ✅ `POST /me/search-history`: `{item: {...}}`
- ✅ `GET /artists/search`: `{query, items: [{id, name, avatar_url, genres}]}`

## ✅ Все endpoints, используемые клиентом, правильно реализованы!

