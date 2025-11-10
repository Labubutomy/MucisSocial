package internal

import (
	"context"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	tracks "github.com/Labubutomy/MucisSocial/services/tracks/api"
)

type GRPCHandler struct {
	tracks.UnimplementedTracksServiceServer
	service *Service
}

func NewGRPCHandler(service *Service) *GRPCHandler {
	return &GRPCHandler{
		service: service,
	}
}

// CreateTrack создает новый трек
func (h *GRPCHandler) CreateTrack(ctx context.Context, req *tracks.CreateTrackRequest) (*tracks.CreateTrackResponse, error) {
	// Валидация входных данных
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if len(req.ArtistIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one artist_id is required")
	}

	// Парсим UUID артистов
	artistIDs, err := parseUUIDs(req.ArtistIds)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid artist_id format: "+err.Error())
	}

	// Создаем трек через сервис
	track, err := h.service.CreateTrackGRPC(ctx, req.Title, artistIDs, req.Genre)
	if err != nil {
		// Закомментировано: tracks-service не должен проверять артистов
		// if err == ErrNotFound {
		// 	return nil, status.Error(codes.NotFound, "one or more artists not found")
		// }
		if err == ErrBadRequest {
			return nil, status.Error(codes.InvalidArgument, "invalid request")
		}
		log.Printf("Error creating track: %v", err)
		return nil, status.Error(codes.Internal, "failed to create track")
	}

	return &tracks.CreateTrackResponse{
		TrackId: track.ID.String(),
	}, nil
}

// UpdateTrackInfo обновляет информацию о треке (cover_url, audio_url)
func (h *GRPCHandler) UpdateTrackInfo(ctx context.Context, req *tracks.UpdateTrackInfoRequest) (*tracks.UpdateTrackInfoResponse, error) {
	// Валидация
	if req.TrackId == "" {
		return nil, status.Error(codes.InvalidArgument, "track_id is required")
	}
	// Убираем требование обязательности хотя бы одного URL, так как теперь можно обновлять только один

	// Парсим UUID трека
	trackID, err := uuid.Parse(req.TrackId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid track_id format")
	}

	if req.DurationSec < 0 {
		return nil, status.Error(codes.InvalidArgument, "negative track duration")
	}

	// Обновляем URLs трека
	err = h.service.UpdateTrackURLsAndDuration(ctx, trackID, req.CoverUrl, req.AudioUrl, int(req.DurationSec))
	if err != nil {
		if err == ErrNotFound {
			return nil, status.Error(codes.NotFound, "track not found")
		}
		log.Printf("Error updating track info: %v", err)
		return nil, status.Error(codes.Internal, "failed to update track info")
	}

	return &tracks.UpdateTrackInfoResponse{
		Success: true,
	}, nil
}
