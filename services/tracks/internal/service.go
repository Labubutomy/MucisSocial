package internal

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetTrack получить трек
func (s *Service) GetTrack(ctx context.Context, id uuid.UUID) (*Track, error) {
	return s.repo.GetByID(ctx, id)
}

// ListTracks список треков
func (s *Service) ListTracks(ctx context.Context, limit, offset int, artistID *uuid.UUID) ([]*Track, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset, artistID)
}

// SearchTracks поиск треков по названию
func (s *Service) SearchTracks(ctx context.Context, query string, limit, offset int) ([]*Track, error) {
	if query == "" {
		return s.ListTracks(ctx, limit, offset, nil)
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.Search(ctx, query, limit, offset)
}


// CreateTrack создать трек (admin) - принимает массив artist_ids
func (s *Service) CreateTrack(ctx context.Context, title string, artistIDs []uuid.UUID, genre string) (*Track, error) {
	// Закомментировано: tracks-service не должен хранить артистов локально
	// artists, err := s.validateAndGetArtists(ctx, artistIDs)
	// if err != nil {
	// 	return nil, err
	// }

	track := &Track{
		ID:        uuid.New(),
		Title:     title,
		ArtistIDs: artistIDs, // Сохраняем только ID артистов
		Genre:     genre,
		Status:    StatusUploaded,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return track, s.repo.Create(ctx, track)
}

// CreateTrackGRPC создать трек через gRPC (принимает массив artist_ids)
func (s *Service) CreateTrackGRPC(ctx context.Context, title string, artistIDs []uuid.UUID, genre string) (*Track, error) {
	track := &Track{
		ID:        uuid.New(),
		Title:     title,
		ArtistIDs: artistIDs, // Сохраняем только ID артистов
		Genre:     genre,
		Status:    StatusUploaded,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return track, s.repo.Create(ctx, track)
}

// UpdateTrack обновить трек (admin)
func (s *Service) UpdateTrack(ctx context.Context, id uuid.UUID, title string, artistIDs []uuid.UUID, genre string) error {
	track, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if title != "" {
		track.Title = title
	}
	if len(artistIDs) > 0 {
		track.ArtistIDs = artistIDs // Обновляем только ID артистов
	}
	if genre != "" {
		track.Genre = genre
	}
	track.UpdatedAt = time.Now()

	return s.repo.Update(ctx, track)
}

// DeleteTrack удалить трек (admin)
func (s *Service) DeleteTrack(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// UpdateTrackURLs обновить URLs трека (cover_url, audio_url, duration_sec)
func (s *Service) UpdateTrackURLsAndDuration(ctx context.Context, trackID uuid.UUID, coverURL, audioURL string, durationSec int) error {
	// Используем специальный метод для обновления только URLs
	return s.repo.UpdateURLsAndDuration(ctx, trackID, coverURL, audioURL, durationSec)
}
