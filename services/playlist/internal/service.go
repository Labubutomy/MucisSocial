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

// CreatePlaylist создать плейлист
func (s *Service) CreatePlaylist(ctx context.Context, authorID uuid.UUID, name string, trackIDs []uuid.UUID) (*Playlist, error) {
	if name == "" {
		return nil, ErrBadRequest
	}

	playlist := &Playlist{
		ID:        uuid.New(),
		AuthorID:  authorID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Преобразуем trackIDs в PlaylistTrack структуры
	if len(trackIDs) > 0 {
		tracks := make([]PlaylistTrack, len(trackIDs))
		for i, trackID := range trackIDs {
			tracks[i] = PlaylistTrack{
				Track: Track{
					ID: trackID,
				},
				Position: i,
			}
		}
		playlist.Tracks = tracks
	}

	return playlist, s.repo.Create(ctx, playlist)
}

// DeletePlaylist удалить плейлист
func (s *Service) DeletePlaylist(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// RemoveTrackFromPlaylist удалить трек из плейлиста
func (s *Service) RemoveTrackFromPlaylist(ctx context.Context, playlistID, trackID uuid.UUID) error {
	return s.repo.RemoveTrackFromPlaylist(ctx, playlistID, trackID)
}

// SubscribeUser подписать пользователя на плейлист
func (s *Service) SubscribeUser(ctx context.Context, userID, playlistID uuid.UUID) error {
	// Проверяем существование плейлиста
	_, err := s.repo.GetByID(ctx, playlistID)
	if err != nil {
		return err
	}
	return s.repo.SubscribeUser(ctx, userID, playlistID)
}

// UnsubscribeUser отписать пользователя от плейлиста
func (s *Service) UnsubscribeUser(ctx context.Context, userID, playlistID uuid.UUID) error {
	return s.repo.UnsubscribeUser(ctx, userID, playlistID)
}

// GetUserPlaylists получить плейлисты пользователя (подписки)
func (s *Service) GetUserPlaylists(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Playlist, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.GetUserPlaylists(ctx, userID, limit, offset)
}

// получить треки плейлиста
func (s *Service) GetPlaylistTracks(ctx context.Context, playlistID uuid.UUID) ([]PlaylistTrack, error) {
	// Проверяем существование плейлиста
	_, err := s.repo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetPlaylistTracks(ctx, playlistID)
}

// получить автора плейлиста
func (s *Service) GetPlaylistAuthor(ctx context.Context, playlistID uuid.UUID) (uuid.UUID, error) {
	playlist, err := s.repo.GetByID(ctx, playlistID)
	if err != nil {
		return uuid.Nil, err
	}
	return playlist.AuthorID, nil
}
