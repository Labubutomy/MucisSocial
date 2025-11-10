package internal

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Track статусы
const (
	StatusUploaded   = "uploaded"
	StatusProcessing = "processing"
	StatusReady      = "ready"
	StatusFailed     = "failed"
)

// Track модель трека
type Track struct {
	ID        uuid.UUID   `json:"id"`
	Title     string      `json:"title"`
	ArtistIDs []uuid.UUID `json:"artist_ids"` // Массив ID артистов (информация об артистах хранится в artists-service)
	Genre     string      `json:"genre,omitempty"`
	AudioURL  string      `json:"audio_url,omitempty"`
	CoverURL  string      `json:"cover_url,omitempty"`
	Duration  int         `json:"duration_seconds"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// Ошибки
var (
	ErrNotFound     = errors.New("track not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrBadRequest   = errors.New("bad request")
)
