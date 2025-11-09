
CREATE TABLE IF NOT EXISTS tracks (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    genre VARCHAR(100) DEFAULT '',
    audio_url TEXT DEFAULT '',
    cover_url TEXT DEFAULT '',
    duration_seconds INTEGER DEFAULT 0,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS artists (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL DEFAULT ''
);

-- Связующая таблица для many-to-many связи треков и артистов
CREATE TABLE IF NOT EXISTS track_artists (
    track_id UUID NOT NULL,
    artist_id UUID NOT NULL,
    PRIMARY KEY (track_id, artist_id),
    FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE,
    FOREIGN KEY (artist_id) REFERENCES artists(id) ON DELETE CASCADE
);

-- Индексы
CREATE INDEX idx_track_artists_track_id ON track_artists(track_id);
CREATE INDEX idx_track_artists_artist_id ON track_artists(artist_id);
CREATE INDEX idx_tracks_status ON tracks(status);
CREATE INDEX idx_tracks_created_at ON tracks(created_at DESC);