package service

import (
	"context"
	"fmt"

	"github.com/MucisSocial/artist-service/internal/domain"
)

type artistService struct {
	artistRepo domain.ArtistRepository
}

func NewArtistService(
	artistRepo domain.ArtistRepository,
) domain.ArtistService {
	return &artistService{
		artistRepo: artistRepo,
	}
}

func (s *artistService) CreateArtist(ctx context.Context, req domain.CreateArtistRequest) (*domain.ArtistResponse, error) {
	// Check if artist name already exists
	exists, err := s.artistRepo.NameExists(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check artist name: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("artist with name %s already exists", req.Name)
	}

	artist := domain.NewArtist(req)
	if err := s.artistRepo.Create(ctx, artist); err != nil {
		return nil, fmt.Errorf("failed to create artist: %w", err)
	}

	return artist.ToResponse(), nil
}

func (s *artistService) GetArtistByID(ctx context.Context, id string) (*domain.ArtistResponse, error) {
	artist, err := s.artistRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get artist: %w", err)
	}

	// TODO: Fetch top tracks from tracks service
	response := artist.ToResponse()

	return response, nil
}

func (s *artistService) UpdateArtist(ctx context.Context, id string, req domain.UpdateArtistRequest) (*domain.ArtistResponse, error) {
	artist, err := s.artistRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get artist: %w", err)
	}

	// Check if new name already exists (if name is being changed)
	if req.Name != nil && *req.Name != artist.Name {
		exists, err := s.artistRepo.NameExists(ctx, *req.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check artist name: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("artist with name %s already exists", *req.Name)
		}
	}

	artist.Update(req)
	if err := s.artistRepo.Update(ctx, artist); err != nil {
		return nil, fmt.Errorf("failed to update artist: %w", err)
	}

	return artist.ToResponse(), nil
}

func (s *artistService) DeleteArtist(ctx context.Context, id string) error {
	if err := s.artistRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete artist: %w", err)
	}

	return nil
}

func (s *artistService) ListArtists(ctx context.Context, limit, offset int) ([]*domain.ArtistResponse, error) {
	artists, err := s.artistRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list artists: %w", err)
	}

	var responses []*domain.ArtistResponse
	for _, artist := range artists {
		responses = append(responses, artist.ToResponse())
	}

	return responses, nil
}

func (s *artistService) SearchArtists(ctx context.Context, query string, limit int) ([]*domain.ArtistResponse, error) {
	artists, err := s.artistRepo.Search(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search artists: %w", err)
	}

	var responses []*domain.ArtistResponse
	for _, artist := range artists {
		responses = append(responses, artist.ToResponse())
	}

	return responses, nil
}

func (s *artistService) GetTrendingArtists(ctx context.Context, limit int) ([]*domain.TrendingArtist, error) {
	artists, err := s.artistRepo.GetTrending(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending artists: %w", err)
	}

	var responses []*domain.TrendingArtist
	for _, artist := range artists {
		responses = append(responses, artist.ToTrendingArtist())
	}

	return responses, nil
}
