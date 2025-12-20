# Recommendation Service

ML-ready recommendation service for music application.

## Architecture

This service follows a modular architecture designed for easy ML integration:

```
┌─────────────────────────────────────────────────────────────────┐
│                    Recommendation Engine                        │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ CandidateGen 1  │  │ CandidateGen 2  │  │ CandidateGen N  │ │
│  │ (TopGenres)     │  │ (TopArtists)    │  │ (Fallback)      │ │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘ │
│           └─────────────┬──────┴─────────────────────┘          │
│                         ▼                                       │
│           ┌─────────────────────────┐                          │
│           │     Candidate Set       │                          │
│           └─────────────┬───────────┘                          │
│                         ▼                                       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   Filter 1      │→ │   Filter 2      │→ │   Filter N      │ │
│  │ (AlreadyListen) │  │   (Explicit)    │  │ (Genre/Fresh)   │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│                         ▼                                       │
│           ┌─────────────────────────┐                          │
│           │        Scorer           │  ← ML REPLACEMENT POINT  │
│           │   (Heuristic / ML)      │                          │
│           └─────────────┬───────────┘                          │
│                         ▼                                       │
│           ┌─────────────────────────┐                          │
│           │    Ranked Results       │                          │
│           └─────────────────────────┘                          │
└─────────────────────────────────────────────────────────────────┘
```

## ML Integration Points

### 1. Scorer (Primary)
Replace `HeuristicScorer` with `MLScorer` in `cmd/main.go`:
```go
// Current (heuristic)
scorer := scorers.NewHeuristicScorer(trackStore, globalStatsStore)

// ML replacement
scorer := scorers.NewMLScorer("path/to/model")
```

### 2. CandidateGenerator (Secondary)
Add embedding-based retrieval as a new generator:
```go
embeddingGen := generators.NewEmbeddingGenerator(embeddingService)
generators := []engine.CandidateGenerator{
    genreGenerator,
    artistGenerator,
    embeddingGen,  // New ML-based generator
    fallbackGenerator,
}
```

### 3. ReRanker (Optional)
Add post-scoring re-ranking for diversity:
```go
recEngine.ReRanker = rerankers.NewDiversityReRanker()
```

## API

### POST /recommendations

Request:
```json
{
  "user_id": "user123",
  "limit": 20,
  "filters": {
    "exclude_explicit": true,
    "genres": ["rock", "indie"],
    "fresh_days": 30
  }
}
```

Response:
```json
{
  "tracks": ["track_id_1", "track_id_2", "..."]
}
```

## Events

### Track Events (Kafka topic: `track-events`)
```json
{
  "event_type": "track_created",
  "track_id": "track123",
  "artist_id": "artist456",
  "genres": ["rock", "indie"],
  "release_ts": 1734134400,
  "is_explicit": false
}
```

### Listening Events (Kafka topic: `listening-events`)
```json
{
  "event_type": "track_listened",
  "user_id": "user123",
  "track_id": "track456",
  "listened_seconds": 180,
  "ts": 1734134400,
  "lat": 59.9311,
  "lon": 30.3609
}
```

**Note**: `lat` and `lon` fields are optional. When provided and `ENABLE_GEO_TOP=true`, the service will aggregate listening statistics by geographic location using geohash.

## Testing without Kafka

Use HTTP ingest endpoints:

```bash
# Add a track
curl -X POST http://localhost:8080/ingest/track \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "track_created",
    "track_id": "track1",
    "artist_id": "artist1",
    "genres": ["rock"],
    "release_ts": 1734134400,
    "is_explicit": false
  }'

# Add a listening event
curl -X POST http://localhost:8080/ingest/listening \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "track_listened",
    "user_id": "user1",
    "track_id": "track1",
    "listened_seconds": 180,
    "ts": 1734134400
  }'

# Get recommendations
curl -X POST http://localhost:8080/recommendations \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user1",
    "limit": 10
  }'

# Get top tracks nearby a location (requires ENABLE_GEO_TOP=true)
curl "http://localhost:8080/charts/top/nearby?lat=59.9311&lon=30.3609&radius_m=2000&limit=20"
```

## Geo-based Top Charts API

When `ENABLE_GEO_TOP=true`, the service provides geographic-based track popularity:

### GET /charts/top/nearby

Query parameters:
- `lat` (required): Latitude (-90 to 90)
- `lon` (required): Longitude (-180 to 180)
- `radius_m` (optional): Search radius in meters (default: 1000)
- `limit` (optional): Max results to return (default: 20)

Example response:
```json
{
  "meta": {
    "precision": 6,
    "radius_m": 2000,
    "geohashes_queried": 9,
    "center_lat": 59.9311,
    "center_lon": 30.3609
  },
  "track_ids": ["track1", "track2", "track3"],
  "tracks": [
    {"track_id": "track1", "count": 123},
    {"track_id": "track2", "count": 100},
    {"track_id": "track3", "count": 87}
  ]
}
```

Example with coordinates in listening event:
```bash
curl -X POST http://localhost:8080/ingest/listening \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "track_listened",
    "user_id": "user1",
    "track_id": "track1",
    "listened_seconds": 180,
    "ts": 1734134400,
    "lat": 59.9311,
    "lon": 30.3609
  }'
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| HTTP_PORT | 8080 | HTTP server port |
| KAFKA_BROKERS | redpanda:9092 | Kafka/Redpanda brokers |
| TRACK_EVENTS_TOPIC | track-events | Topic for track events |
| LISTENING_EVENTS_TOPIC | listening-events | Topic for listening events |
| ENABLE_GEO_TOP | false | Enable geo-based top charts feature |
| GEOHASH_PRECISION | 6 | Geohash precision (1-10, ~1.2km at precision 6) |
| GEO_TOP_CACHE_TTL | 60s | Cache TTL for geo query results |
| GEO_TOP_MAX_PER_GEO | 1000 | Max tracks stored per geohash cell |
| GEO_TOP_DEFAULT_RADIUS_M | 1000 | Default search radius in meters |

## Running

### Docker
```bash
docker build -t recommendation-service .
docker run -p 8080:8080 recommendation-service
```

### Docker Compose
See `services/docker-compose.yml`

### Local Development
```bash
cd services/recommendations
go mod tidy
go run cmd/main.go
```
