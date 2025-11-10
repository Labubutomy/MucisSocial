package internal

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Track модель трека (дублируется из tracks сервиса для избежания зависимости от internal пакета)
type Track struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Artists   []Artist  `json:"artists"` // Массив артистов
	Genre     string    `json:"genre,omitempty"`
	AudioURL  string    `json:"audio_url,omitempty"`
	CoverURL  string    `json:"cover_url,omitempty"`
	Duration  int       `json:"duration_seconds"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Artist модель артиста (дублируется из tracks сервиса)
type Artist struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// PlaylistTrack трек в плейлисте с позицией
type PlaylistTrack struct {
	Track
	Position int `json:"position"`
}

// Playlist модель плейлиста
type Playlist struct {
	ID          uuid.UUID       `json:"id"`
	AuthorID    uuid.UUID       `json:"author_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	IsPrivate   bool            `json:"is_private"`
	Tracks      []PlaylistTrack `json:"tracks,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// PlaylistUser связь пользователя и плейлиста (подписка/лайк)
type PlaylistUser struct {
	UserID     uuid.UUID `json:"user_id"`
	PlaylistID uuid.UUID `json:"playlist_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// Ошибки
var (
	ErrNotFound     = errors.New("playlist not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrBadRequest   = errors.New("bad request")
)
