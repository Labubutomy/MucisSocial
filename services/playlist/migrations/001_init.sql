-- Таблица плейлистов
CREATE TABLE IF NOT EXISTS playlists (
    id UUID PRIMARY KEY,
    author_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Связующая таблица плейлист -> треки
-- Примечание: track_id ссылается на tracks.id из другого сервиса, поэтому FOREIGN KEY не используется
CREATE TABLE IF NOT EXISTS playlist_tracks (
    playlist_id UUID NOT NULL,
    track_id UUID NOT NULL,
    position INTEGER DEFAULT 0,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (playlist_id, track_id),
    FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE
);

-- Связующая таблица пользователь -> плейлист (подписки/лайки)
CREATE TABLE IF NOT EXISTS playlist_users (
    user_id UUID NOT NULL,
    playlist_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, playlist_id),
    FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE
);

-- Индексы
CREATE INDEX idx_playlists_author_id ON playlists(author_id);
CREATE INDEX idx_playlists_created_at ON playlists(created_at DESC);
CREATE INDEX idx_playlist_tracks_playlist_id ON playlist_tracks(playlist_id);
CREATE INDEX idx_playlist_tracks_track_id ON playlist_tracks(track_id);
CREATE INDEX idx_playlist_users_user_id ON playlist_users(user_id);
CREATE INDEX idx_playlist_users_playlist_id ON playlist_users(playlist_id);

