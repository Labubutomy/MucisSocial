package internal

import (
	"context"

	"github.com/google/uuid"
)

// Service aggregates queue operations.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateQueue(ctx context.Context, contextType string) (ContextRef, error) {
	if contextType == "" {
		return ContextRef{}, ErrBadRequest
	}
	return s.repo.CreateQueue(ctx, contextType)
}

func (s *Service) EnqueueTrack(ctx context.Context, ref ContextRef, trackID uuid.UUID) (*QueueItem, error) {
	if !ref.Valid() || trackID == uuid.Nil {
		return nil, ErrBadRequest
	}
	item := &QueueItem{
		Context: ref,
		TrackID: trackID,
	}
	if err := s.repo.Enqueue(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) ListQueue(ctx context.Context, ref ContextRef, limit int) ([]*QueueItem, error) {
	if !ref.Valid() {
		return nil, ErrBadRequest
	}
	return s.repo.ListFuture(ctx, ref, limit)
}

func (s *Service) ListHistory(ctx context.Context, ref ContextRef, limit int) ([]*QueueItem, error) {
	if !ref.Valid() {
		return nil, ErrBadRequest
	}
	return s.repo.ListHistory(ctx, ref, limit)
}

func (s *Service) GetNextTrack(ctx context.Context, ref ContextRef, requestedBy uuid.UUID) (*QueueItem, error) {
	if !ref.Valid() || requestedBy == uuid.Nil {
		return nil, ErrBadRequest
	}
	return s.repo.StepNext(ctx, ref)
}

func (s *Service) GetPrevTrack(ctx context.Context, ref ContextRef, requestedBy uuid.UUID) (*QueueItem, error) {
	if !ref.Valid() || requestedBy == uuid.Nil {
		return nil, ErrBadRequest
	}
	return s.repo.StepPrev(ctx, ref)
}

func (s *Service) GetCurrentTrack(ctx context.Context, ref ContextRef) (*QueueItem, error) {
	if !ref.Valid() {
		return nil, ErrBadRequest
	}
	return s.repo.CurrentTrack(ctx, ref)
}

func (s *Service) ClearQueue(ctx context.Context, ref ContextRef) error {
	if !ref.Valid() {
		return ErrBadRequest
	}
	return s.repo.ClearQueue(ctx, ref)
}

func (s *Service) RemoveTrack(ctx context.Context, ref ContextRef, trackID uuid.UUID) (bool, error) {
	if !ref.Valid() || trackID == uuid.Nil {
		return false, ErrBadRequest
	}
	return s.repo.RemoveTrack(ctx, ref, trackID)
}
