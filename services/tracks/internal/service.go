package internal

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo          *Repository
	eventProducer *EventProducer
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// SetEventProducer sets the event producer for publishing track events
func (s *Service) SetEventProducer(producer *EventProducer) {
	s.eventProducer = producer
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
	track := &Track{
		ID:        uuid.New(),
		Title:     title,
		ArtistIDs: artistIDs,
		Genre:     genre,
		Status:    StatusUploaded,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, track); err != nil {
		return nil, err
	}
	// Publish event for recommendations service
	s.publishTrackCreated(ctx, track)
	return track, nil
}

// CreateTrackGRPC создать трек через gRPC (принимает массив artist_ids)
func (s *Service) CreateTrackGRPC(ctx context.Context, title string, artistIDs []uuid.UUID, genre string) (*Track, error) {
	track := &Track{
		ID:        uuid.New(),
		Title:     title,
		ArtistIDs: artistIDs,
		Genre:     genre,
		Status:    StatusUploaded,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, track); err != nil {
		return nil, err
	}
	// Publish event for recommendations service
	s.publishTrackCreated(ctx, track)
	return track, nil
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
		track.ArtistIDs = artistIDs
	}
	if genre != "" {
		track.Genre = genre
	}
	track.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, track); err != nil {
		return err
	}
	// Publish event for recommendations service
	s.publishTrackUpdated(ctx, track)
	return nil
}

// DeleteTrack удалить трек (admin)
func (s *Service) DeleteTrack(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// UpdateTrackURLs обновить URLs трека (cover_url, audio_url, duration_sec)
func (s *Service) UpdateTrackURLsAndDuration(ctx context.Context, trackID uuid.UUID, coverURL, audioURL string, durationSec int) error {
	err := s.repo.UpdateURLsAndDuration(ctx, trackID, coverURL, audioURL, durationSec)
	if err != nil {
		return err
	}
	// Get updated track and publish event
	track, err := s.repo.GetByID(ctx, trackID)
	if err == nil {
		s.publishTrackUpdated(ctx, track)
	}
	return nil
}

// publishTrackCreated publishes track_created event if producer is set
func (s *Service) publishTrackCreated(ctx context.Context, track *Track) {
	if s.eventProducer != nil {
		if err := s.eventProducer.PublishTrackCreated(ctx, track); err != nil {
			log.Printf("Warning: failed to publish track_created event: %v", err)
		}
	}
}

// publishTrackUpdated publishes track_updated event if producer is set
func (s *Service) publishTrackUpdated(ctx context.Context, track *Track) {
	if s.eventProducer != nil {
		if err := s.eventProducer.PublishTrackUpdated(ctx, track); err != nil {
			log.Printf("Warning: failed to publish track_updated event: %v", err)
		}
	}
}
