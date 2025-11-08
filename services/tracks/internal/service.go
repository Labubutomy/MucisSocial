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

// CreateTrack создать трек (admin)
func (s *Service) CreateTrack(ctx context.Context, title, artistName, genre string, artistID uuid.UUID) (*Track, error) {
	track := &Track{
		ID:         uuid.New(),
		Title:      title,
		ArtistID:   artistID,
		ArtistName: artistName,
		Genre:      genre,
		Status:     StatusUploaded,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	return track, s.repo.Create(ctx, track)
}

// UpdateTrack обновить трек (admin)
func (s *Service) UpdateTrack(ctx context.Context, id uuid.UUID, title, artistName, genre string) error {
	track, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if title != "" {
		track.Title = title
	}
	if artistName != "" {
		track.ArtistName = artistName
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

// HandleTrackUploaded обработка события загрузки
func (s *Service) HandleTrackUploaded(ctx context.Context, event TrackUploadedEvent) error {
	trackID, _ := uuid.Parse(event.TrackID)
	artistID, _ := uuid.Parse(event.ArtistID)

	track := &Track{
		ID:         trackID,
		Title:      event.Title,
		ArtistID:   artistID,
		ArtistName: event.ArtistName,
		Genre:      event.Genre,
		Status:     StatusUploaded,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	return s.repo.Create(ctx, track)
}

// HandleTranscodingCompleted обработка завершения транскодирования
func (s *Service) HandleTranscodingCompleted(ctx context.Context, event TranscodingCompletedEvent) error {
	trackID, _ := uuid.Parse(event.TrackID)

	if event.Status == "ready" {
		track, err := s.repo.GetByID(ctx, trackID)
		if err != nil {
			return err
		}

		track.Status = StatusReady
		track.AudioURL = event.AudioURL
		track.CoverURL = event.CoverURL
		track.Duration = event.Duration
		track.UpdatedAt = time.Now()

		return s.repo.Update(ctx, track)
	}

	return s.repo.UpdateStatus(ctx, trackID, StatusFailed)
}
