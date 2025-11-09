# API Specification

Документ описывает необходимые на текущем этапе эндпоинты и модели данных для клиентского приложения. Сервисы разделены на: `tracks`, `cdn`, `users`, `artists`, `playlists`.

---

## tracks service

**Резюме:** управление музыкальными треками, их подборками и рекомендациями; выдаёт данные для главной ленты, страницы трека и смешанного поиска.

- **Endpoint:** `GET /tracks?filter={trending|popular|new}&limit=24`  
  **Model:**

  ```json
  {
    "filter": "trending",
    "items": [
      {
        "id": "nebula-night",
        "title": "Nebula Night",
        "durationSec": 214,
        "coverUrl": "https://cdn.example.com/images/tracks/nebula-night.jpg",
        "artist": {
          "id": "aviana",
          "name": "Aviana"
        },
        "album": {
          "id": "album-nebula",
          "title": "Starlight Bloom"
        },
        "isLiked": true,
        "stream": {
          "quality": ["aac_128", "aac_256", "flac"],
          "hlsMasterUrl": "https://cdn.example.com/hls/nebula-night/master.m3u8"
        }
      }
    ]
  }
  ```

- **Endpoint:** `GET /tracks/{trackId}`  
  **Model:**

  ```json
  {
    "id": "nebula-night",
    "title": "Nebula Night",
    "durationSec": 214,
    "coverUrl": "https://cdn.example.com/images/tracks/nebula-night.jpg",
    "artist": {
      "id": "aviana",
      "name": "Aviana"
    },
    "album": {
      "id": "album-nebula",
      "title": "Starlight Bloom",
      "coverUrl": "https://cdn.example.com/images/albums/starlight-bloom.jpg",
      "releasedAt": "2024-02-10"
    },
    "lyrics": null,
    "credits": ["Aviana", "Kyro"],
    "bpm": 102,
    "isLiked": true,
    "stream": {
      "quality": ["aac_128", "aac_256", "flac"],
      "hlsMasterUrl": "https://cdn.example.com/hls/nebula-night/master.m3u8"
    }
  }
  ```

- **Endpoint:** `GET /tracks/{trackId}/recommendations?limit=12`  
  **Model:**

  ```json
  {
    "trackId": "nebula-night",
    "items": [
      {
        "id": "horizon-sway",
        "title": "Horizon Sway",
        "durationSec": 205,
        "coverUrl": "https://cdn.example.com/images/tracks/horizon-sway.jpg",
        "artist": {
          "id": "lunaric",
          "name": "Lunaric"
        },
        "isLiked": false
      }
    ]
  }
  ```

- **Endpoint:** `GET /tracks/search?q={query}&limit=20`  
  **Model:**

  ```json
  {
    "query": "синтвейв",
    "items": [
      {
        "type": "track",
        "data": {
          "id": "city-halo",
          "title": "City Halo",
          "durationSec": 205,
          "coverUrl": "https://cdn.example.com/images/tracks/city-halo.jpg",
          "artist": {
            "id": "artist-shea",
            "name": "Shea Monarch"
          },
          "isLiked": false
        }
      },
      {
        "type": "playlist",
        "data": {
          "id": "neon-focus",
          "title": "Neon Focus",
          "coverUrl": "https://cdn.example.com/images/playlists/neon-focus.jpg",
          "itemsCount": 42,
          "description": "Свечение синтвейва для глубокого погружения в работу."
        }
      }
    ]
  }
  ```

- **Endpoint:** `POST /tracks/{trackId}/like`  
  **Model:**

  ```json
  {
    "trackId": "nebula-night",
    "isLiked": true,
    "likedAt": "2025-11-08T19:34:00.000Z"
  }
  ```

- **Endpoint:** `GET /tracks/search/trending`  
  **Model:**
  ```json
  {
    "items": [{ "query": "синтвейв ночь" }, { "query": "неоновый фанк" }, { "query": "dream pop" }]
  }
  ```

---

## cdn service

**Резюме:** обслуживание медиаконтента — HLS потоков и статических ассетов (обложки, превью).

- **Endpoint:** `GET https://cdn.example.com/hls/{trackId}/master.m3u8`  
  **Model (пример ответа M3U8):**

  ```
  #EXTM3U
  #EXT-X-VERSION:3
  #EXT-X-STREAM-INF:BANDWIDTH=128000,CODECS="mp4a.40.2"
  https://cdn.example.com/hls/nebula-night/128k/playlist.m3u8
  #EXT-X-STREAM-INF:BANDWIDTH=256000,CODECS="mp4a.40.2"
  https://cdn.example.com/hls/nebula-night/256k/playlist.m3u8
  #EXT-X-STREAM-INF:BANDWIDTH=1411200,CODECS="flac"
  https://cdn.example.com/hls/nebula-night/flac/playlist.m3u8
  ```

- **Endpoint:** `GET https://cdn.example.com/hls/{trackId}/{quality}/segment-{n}.aac`  
  **Model:** бинарные аудиосегменты HLS (пример не представлен).

- **Endpoint:** `GET https://cdn.example.com/images/{entityType}/{entityId}.jpg`  
  **Model:** статический JPEG/WEBP, используется в UI для обложек треков, альбомов, плейлистов и артистов.

---

## users service

**Резюме:** хранение профилей, вкусов пользователя, поиска, аутентификация и пользовательские предпочтения.

- **Endpoint:** `POST /auth/sign-in`  
  **Model:**

  ```json
  {
    "email": "hello@music.social",
    "password": "string"
  }
  ```

  **Response:**

  ```json
  {
    "accessToken": "jwt-token",
    "refreshToken": "refresh-token",
    "user": {
      "id": "user-ava",
      "username": "ava.wave",
      "avatarUrl": "https://cdn.example.com/images/users/user-ava.jpg"
    }
  }
  ```

- **Endpoint:** `POST /auth/sign-up`  
  **Model:** идентичен `sign-in`, ответ содержит нового пользователя и токены.

- **Endpoint:** `GET /me`  
  **Model:**

  ```json
  {
    "id": "user-ava",
    "username": "ava.wave",
    "avatarUrl": "https://cdn.example.com/images/users/user-ava.jpg",
    "musicTasteSummary": {
      "topGenres": ["Синтвейв", "Дрим-поп", "Инди-электроника"],
      "topArtists": ["Aviana", "Kyro", "Luna Wave", "Solaria"]
    }
  }
  ```

- **Endpoint:** `GET /me/playlists?limit=50`  
  **Model:**

  ```json
  {
    "items": [
      {
        "id": "uplink",
        "title": "Uplink Sessions",
        "coverUrl": "https://cdn.example.com/images/playlists/uplink.jpg",
        "itemsCount": 35,
        "description": "Даунтемпо и синтовые текстуры для поздней ночи."
      }
    ]
  }
  ```

- **Endpoint:** `GET /me/search-history`  
  **Model:**

  ```json
  {
    "items": [
      { "id": "history-1", "query": "синтвейв полночь", "createdAt": "2025-11-07T12:00:00.000Z" },
      {
        "id": "history-2",
        "query": "лоуфай для концентрации",
        "createdAt": "2025-11-05T09:30:00.000Z"
      }
    ]
  }
  ```

- **Endpoint:** `POST /me/search-history`  
  **Model:**
  ```json
  {
    "query": "dream pop"
  }
  ```
  **Response:**
  ```json
  {
    "id": "history-3",
    "query": "dream pop",
    "createdAt": "2025-11-08T19:33:00.000Z"
  }
  ```

---

## artists service

**Резюме:** данные об артистах, их профилях и подборках; поддержка поиска и curated-подборок.

- **Endpoint:** `GET /artists/{artistId}`  
  **Model:**

  ```json
  {
    "id": "aviana",
    "name": "Aviana",
    "avatarUrl": "https://cdn.example.com/images/artists/aviana.jpg",
    "genres": ["Синтвейв", "Электроника"],
    "followers": 1204300,
    "topTracks": [
      {
        "id": "nebula-night",
        "title": "Nebula Night",
        "coverUrl": "https://cdn.example.com/images/tracks/nebula-night.jpg"
      }
    ]
  }
  ```

- **Endpoint:** `GET /artists/trending`  
  **Model:**

  ```json
  {
    "items": [
      {
        "id": "artist-aleo",
        "name": "Aleo",
        "avatarUrl": "https://cdn.example.com/images/artists/aleo.jpg",
        "genres": ["Гиперпоп", "Инди"]
      }
    ]
  }
  ```

- **Endpoint:** `GET /artists/search?q={query}`  
  **Model:** аналогично `GET /artists/trending`, возвращает массив артистов, совпадающих с поиском.

---

## playlists service

**Резюме:** управление плейлистами — получение, создание, добавление треков, curated подборки и связи с пользователями.

- **Endpoint:** `GET /playlists/{playlistId}`  
  **Model:**

  ```json
  {
    "id": "uplink",
    "title": "Uplink Sessions",
    "description": "Кураторское путешествие по неону даунтемпо и атмосферной электронике.",
    "coverUrl": "https://cdn.example.com/images/playlists/uplink.jpg",
    "itemsCount": 35,
    "owner": {
      "id": "user-ava",
      "name": "ava.wave"
    },
    "type": "playlist",
    "isLiked": false,
    "totalDurationSec": 8420
  }
  ```

- **Endpoint:** `GET /playlists/{playlistId}/tracks`  
  **Model:**

  ```json
  {
    "items": [
      {
        "id": "uplink-1",
        "title": "Luminous Drift",
        "durationSec": 230,
        "coverUrl": "https://cdn.example.com/images/tracks/luminous-drift.jpg",
        "artist": {
          "id": "solaria",
          "name": "Solaria"
        },
        "isLiked": false
      }
    ]
  }
  ```

- **Endpoint:** `POST /playlists`  
  **Model:**

  ```json
  {
    "title": "Ночной драйв",
    "description": "Синтвейв для поздних поездок",
    "isPrivate": false,
    "coverImageId": null,
    "genres": ["Синтвейв", "Электроника"]
  }
  ```

  **Response:**

  ```json
  {
    "id": "playlist-123",
    "title": "Ночной драйв",
    "description": "Синтвейв для поздних поездок",
    "isPrivate": false,
    "coverUrl": "https://cdn.example.com/images/playlists/playlist-123.jpg",
    "itemsCount": 0,
    "owner": { "id": "user-ava", "name": "ava.wave" }
  }
  ```

- **Endpoint:** `POST /playlists/{playlistId}/tracks`  
  **Model:**

  ```json
  {
    "trackIds": ["city-halo", "horizon-sway"]
  }
  ```

  **Response:**

  ```json
  {
    "playlistId": "playlist-123",
    "added": [
      { "trackId": "city-halo", "position": 1 },
      { "trackId": "horizon-sway", "position": 2 }
    ]
  }
  ```

- **Endpoint:** `GET /playlists?filter=curated&limit=24`  
  **Model:**

  ```json
  {
    "filter": "curated",
    "items": [
      {
        "id": "neon-focus",
        "title": "Neon Focus",
        "coverUrl": "https://cdn.example.com/images/playlists/neon-focus.jpg",
        "itemsCount": 42,
        "description": "Свечение синтвейва для глубокого погружения в работу."
      }
    ]
  }
  ```

- **Endpoint:** `GET /users/{userId}/playlists`  
  **Model:** идентична `GET /me/playlists`, возвращает плейлисты указанного пользователя (используется для просмотра чужих профилей).

---
