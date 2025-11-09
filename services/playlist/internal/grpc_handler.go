package internal

import (
	"context"

	pb "github.com/Labubutomy/MucisSocial/services/playlist/api"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	service *Service
	pb.UnimplementedPlaylistServiceServer
}

func NewGRPCHandler(service *Service) *GRPCHandler {
	return &GRPCHandler{
		service: service,
	}
}

// CreatePlaylist создает новый плейлист
func (h *GRPCHandler) CreatePlaylist(ctx context.Context, req *pb.CreatePlaylistRequest) (*pb.CreatePlaylistResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	// Пока создаем плейлист без треков
	playlist, err := h.service.CreatePlaylist(ctx, userID, req.Name, []uuid.UUID{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create playlist: %v", err)
	}

	return &pb.CreatePlaylistResponse{
		PlaylistId: playlist.ID.String(),
	}, nil
}

// GetPlaylist получает плейлист по ID
func (h *GRPCHandler) GetPlaylist(ctx context.Context, req *pb.GetPlaylistRequest) (*pb.GetPlaylistResponse, error) {
	// TODO: Нужно добавить метод GetPlaylistByID в Service
	return nil, status.Errorf(codes.Unimplemented, "method not implemented yet")
}

// UpdatePlaylist обновляет плейлист
func (h *GRPCHandler) UpdatePlaylist(ctx context.Context, req *pb.UpdatePlaylistRequest) (*pb.UpdatePlaylistResponse, error) {
	// TODO: Нужно добавить метод UpdatePlaylist в Service
	return nil, status.Errorf(codes.Unimplemented, "method not implemented yet")
}

// DeletePlaylist удаляет плейлист
func (h *GRPCHandler) DeletePlaylist(ctx context.Context, req *pb.DeletePlaylistRequest) (*pb.DeletePlaylistResponse, error) {
	playlistID, err := uuid.Parse(req.PlaylistId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid playlist ID: %v", err)
	}

	err = h.service.DeletePlaylist(ctx, playlistID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete playlist: %v", err)
	}

	return &pb.DeletePlaylistResponse{
		Success: true,
	}, nil
}

// GetUserPlaylists получает плейлисты пользователя
func (h *GRPCHandler) GetUserPlaylists(ctx context.Context, req *pb.GetUserPlaylistsRequest) (*pb.GetUserPlaylistsResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	playlists, err := h.service.GetUserPlaylists(ctx, userID, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user playlists: %v", err)
	}

	var apiPlaylists []*pb.Playlist
	for _, playlist := range playlists {
		apiPlaylists = append(apiPlaylists, &pb.Playlist{
			Id:          playlist.ID.String(),
			UserId:      playlist.AuthorID.String(),
			Name:        playlist.Name,
			Description: "",    // TODO: добавить Description в модель
			IsPrivate:   false, // TODO: добавить IsPrivate в модель
			CreatedAt:   playlist.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   playlist.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			TracksCount: int32(len(playlist.Tracks)),
		})
	}

	return &pb.GetUserPlaylistsResponse{
		Playlists: apiPlaylists,
		Total:     int32(len(apiPlaylists)),
	}, nil
}

// AddTrackToPlaylist добавляет трек в плейлист
func (h *GRPCHandler) AddTrackToPlaylist(ctx context.Context, req *pb.AddTrackToPlaylistRequest) (*pb.AddTrackToPlaylistResponse, error) {
	// TODO: Нужно добавить метод AddTrackToPlaylist в Service
	return nil, status.Errorf(codes.Unimplemented, "method not implemented yet")
}

// RemoveTrackFromPlaylist удаляет трек из плейлиста
func (h *GRPCHandler) RemoveTrackFromPlaylist(ctx context.Context, req *pb.RemoveTrackFromPlaylistRequest) (*pb.RemoveTrackFromPlaylistResponse, error) {
	playlistID, err := uuid.Parse(req.PlaylistId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid playlist ID: %v", err)
	}

	trackID, err := uuid.Parse(req.TrackId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid track ID: %v", err)
	}

	err = h.service.RemoveTrackFromPlaylist(ctx, playlistID, trackID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove track from playlist: %v", err)
	}

	return &pb.RemoveTrackFromPlaylistResponse{
		Success: true,
	}, nil
}

// GetPlaylistTracks получает треки плейлиста
func (h *GRPCHandler) GetPlaylistTracks(ctx context.Context, req *pb.GetPlaylistTracksRequest) (*pb.GetPlaylistTracksResponse, error) {
	playlistID, err := uuid.Parse(req.PlaylistId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid playlist ID: %v", err)
	}

	tracks, err := h.service.GetPlaylistTracks(ctx, playlistID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get playlist tracks: %v", err)
	}

	var apiTracks []*pb.PlaylistTrack
	for _, track := range tracks {
		apiTracks = append(apiTracks, &pb.PlaylistTrack{
			TrackId:    track.Track.ID.String(),
			PlaylistId: req.PlaylistId,
			AddedAt:    "2023-01-01T00:00:00Z07:00", // TODO: получать реальную дату
			Position:   int32(track.Position),
		})
	}

	return &pb.GetPlaylistTracksResponse{
		Tracks: apiTracks,
		Total:  int32(len(apiTracks)),
	}, nil
}
