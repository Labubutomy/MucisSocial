
CREATE TABLE IF NOT EXISTS tracks (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    artist_id UUID NOT NULL,
    artist_name VARCHAR(255) NOT NULL,
    genre VARCHAR(100) DEFAULT '',
    audio_url TEXT DEFAULT '',
    cover_url TEXT DEFAULT '',
    duration_seconds INTEGER DEFAULT 0,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

-- Индексы
CREATE INDEX idx_tracks_artist_id ON tracks(artist_id);
CREATE INDEX idx_tracks_status ON tracks(status);
CREATE INDEX idx_tracks_created_at ON tracks(created_at DESC);