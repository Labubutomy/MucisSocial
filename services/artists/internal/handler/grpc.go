package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/MucisSocial/artist-service/internal/domain"
	pb "github.com/MucisSocial/artist-service/internal/pb/artists/v1"
)

type ArtistServiceHandler struct {
	pb.UnimplementedArtistServiceServer
	artistService domain.ArtistService
}

func NewArtistServiceHandler(artistService domain.ArtistService) *ArtistServiceHandler {
	return &ArtistServiceHandler{
		artistService: artistService,
	}
}

func (h *ArtistServiceHandler) CreateArtist(ctx context.Context, req *pb.CreateArtistRequest) (*pb.CreateArtistResponse, error) {
	createReq := domain.CreateArtistRequest{
		Name:   req.Name,
		Genres: req.Genres,
	}

	if req.AvatarUrl != "" {
		createReq.AvatarURL = &req.AvatarUrl
	}

	artist, err := h.artistService.CreateArtist(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create artist: %v", err)
	}

	return &pb.CreateArtistResponse{
		Artist: h.convertToProtoArtistResponse(artist),
	}, nil
}

func (h *ArtistServiceHandler) GetArtistById(ctx context.Context, req *pb.GetArtistByIdRequest) (*pb.GetArtistByIdResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "artist id is required")
	}

	artist, err := h.artistService.GetArtistByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "artist not found: %v", err)
	}

	return &pb.GetArtistByIdResponse{
		Artist: h.convertToProtoArtistResponse(artist),
	}, nil
}

func (h *ArtistServiceHandler) UpdateArtist(ctx context.Context, req *pb.UpdateArtistRequest) (*pb.UpdateArtistResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "artist id is required")
	}

	updateReq := domain.UpdateArtistRequest{
		Genres: req.Genres,
	}

	if req.Name != nil {
		updateReq.Name = req.Name
	}
	if req.AvatarUrl != nil {
		updateReq.AvatarURL = req.AvatarUrl
	}

	artist, err := h.artistService.UpdateArtist(ctx, req.Id, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update artist: %v", err)
	}

	return &pb.UpdateArtistResponse{
		Artist: h.convertToProtoArtistResponse(artist),
	}, nil
}

func (h *ArtistServiceHandler) DeleteArtist(ctx context.Context, req *pb.DeleteArtistRequest) (*pb.DeleteArtistResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "artist id is required")
	}

	err := h.artistService.DeleteArtist(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete artist: %v", err)
	}

	return &pb.DeleteArtistResponse{}, nil
}

func (h *ArtistServiceHandler) ListArtists(ctx context.Context, req *pb.ListArtistsRequest) (*pb.ListArtistsResponse, error) {
	limit := int(req.Limit)
	offset := int(req.Offset)

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	artists, err := h.artistService.ListArtists(ctx, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list artists: %v", err)
	}

	var protoArtists []*pb.ArtistResponse
	for _, artist := range artists {
		protoArtists = append(protoArtists, h.convertToProtoArtistResponse(artist))
	}

	return &pb.ListArtistsResponse{
		Artists: protoArtists,
		Total:   int32(len(protoArtists)), // TODO: implement actual total count
	}, nil
}

func (h *ArtistServiceHandler) SearchArtists(ctx context.Context, req *pb.SearchArtistsRequest) (*pb.SearchArtistsResponse, error) {
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "search query is required")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	artists, err := h.artistService.SearchArtists(ctx, req.Query, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search artists: %v", err)
	}

	var protoArtists []*pb.ArtistResponse
	for _, artist := range artists {
		protoArtists = append(protoArtists, h.convertToProtoArtistResponse(artist))
	}

	return &pb.SearchArtistsResponse{
		Artists: protoArtists,
	}, nil
}

func (h *ArtistServiceHandler) GetTrendingArtists(ctx context.Context, req *pb.GetTrendingArtistsRequest) (*pb.GetTrendingArtistsResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	artists, err := h.artistService.GetTrendingArtists(ctx, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get trending artists: %v", err)
	}

	var protoArtists []*pb.TrendingArtist
	for _, artist := range artists {
		protoArtists = append(protoArtists, h.convertToProtoTrendingArtist(artist))
	}

	return &pb.GetTrendingArtistsResponse{
		Artists: protoArtists,
	}, nil
}

func (h *ArtistServiceHandler) convertToProtoArtistResponse(artist *domain.ArtistResponse) *pb.ArtistResponse {
	protoArtist := &pb.ArtistResponse{
		Id:        artist.ID,
		Name:      artist.Name,
		Genres:    artist.Genres,
		Followers: artist.Followers,
	}

	if artist.AvatarURL != nil {
		protoArtist.AvatarUrl = *artist.AvatarURL
	}

	for _, track := range artist.TopTracks {
		protoTrack := &pb.TopTrack{
			Id:    track.ID,
			Title: track.Title,
		}
		if track.CoverURL != nil {
			protoTrack.CoverUrl = *track.CoverURL
		}
		protoArtist.TopTracks = append(protoArtist.TopTracks, protoTrack)
	}

	return protoArtist
}

func (h *ArtistServiceHandler) convertToProtoTrendingArtist(artist *domain.TrendingArtist) *pb.TrendingArtist {
	protoArtist := &pb.TrendingArtist{
		Id:     artist.ID,
		Name:   artist.Name,
		Genres: artist.Genres,
	}

	if artist.AvatarURL != nil {
		protoArtist.AvatarUrl = *artist.AvatarURL
	}

	return protoArtist
}
