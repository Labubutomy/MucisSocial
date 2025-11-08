package internal

import (
	"errors"
	"github.com/google/uuid"
	"time"
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
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	ArtistID   uuid.UUID `json:"artist_id"`
	ArtistName string    `json:"artist_name"`
	Genre      string    `json:"genre,omitempty"`
	AudioURL   string    `json:"audio_url,omitempty"`
	CoverURL   string    `json:"cover_url,omitempty"`
	Duration   int       `json:"duration_seconds"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Kafka события
type TrackUploadedEvent struct {
	TrackID    string `json:"track_id"`
	Title      string `json:"title"`
	ArtistID   string `json:"artist_id"`
	ArtistName string `json:"artist_name"`
	Genre      string `json:"genre"`
}

type TranscodingCompletedEvent struct {
	TrackID  string `json:"track_id"`
	Status   string `json:"status"`
	AudioURL string `json:"audio_url"`
	CoverURL string `json:"cover_url"`
	Duration int    `json:"duration_seconds"`
}

// Ошибки
var (
	ErrNotFound     = errors.New("track not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrBadRequest   = errors.New("bad request")
)
