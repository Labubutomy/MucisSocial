# Использование Mock API в клиенте

Этот документ описывает все места, где клиент использует mock-api, и статус замены на gateway.

## ✅ Заменено на Gateway

### Плейлисты
- ✅ `addTracksToPlaylist` - теперь использует gateway (POST /api/v1/playlists/{playlistId}/tracks)
- ✅ `fetchPlaylistTracks` - теперь использует gateway (GET /api/v1/playlists/{playlistId}/tracks)
- ✅ `fetchPlaylistDetail` - уже использовал gateway
- ✅ `createPlaylist` - уже использовал gateway
- ✅ `fetchPlaylists` - уже использовал gateway

### Треки
- ✅ `searchTracks` - теперь использует gateway (GET /api/v1/tracks/search)
- ✅ `fetchTrackRecommendations` - временно возвращает список всех треков через gateway (GET /api/v1/tracks)

## ❌ Остается на Mock API (не реализовано в gateway)

### Треки
- ❌ `toggleTrackLike` - нет в gateway

### Плейлисты
- ❌ Поиск плейлистов - нет в gateway

### Пользователи
- ❌ `fetchUserPlaylists` (плейлисты других пользователей) - нет в gateway

### Поиск
- ❌ `fetchTrendingQueries` - нет в gateway
- ❌ Поиск плейлистов - нет в gateway

## Файлы с использованием Mock API

1. `apps/web/src/entities/track/api.ts`
   - `toggleTrackLike` - mock-api

2. `apps/web/src/entities/playlist/api.ts`
   - ✅ Все методы теперь используют gateway

3. `apps/web/src/entities/user/api.ts`
   - `fetchUserPlaylists` - mock-api (плейлисты других пользователей)

4. `apps/web/src/pages/search/model/api.ts`
   - `fetchTrendingQueries` - mock-api
   - `fetchSearchResults` - использует gateway для треков и артистов, mock-api для плейлистов

5. `apps/web/src/shared/ui/app-header/model.ts`
   - `fetchSuggestionSeedsFx` - mock-api (trending queries)

## Примечания

- Gateway проксирует запросы к tracks-service для получения списка треков, трека по ID и поиска треков
- Поиск треков реализован в tracks-service (поиск по названию через ILIKE)
- Рекомендации треков временно заменены на список всех треков (через gateway)
- Лайки треков не реализованы
- Поиск плейлистов не реализован

