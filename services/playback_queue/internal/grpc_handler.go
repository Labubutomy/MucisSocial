package internal

import (
	"context"
	"strings"

	queuepb "github.com/Labubutomy/MucisSocial/services/playback_queue/api"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	queuepb.UnimplementedPlaybackQueueServiceServer
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) CreateQueue(ctx context.Context, req *queuepb.CreateQueueRequest) (*queuepb.CreateQueueResponse, error) {
	ref, err := s.svc.CreateQueue(ctx, req.GetContextType())
	if err != nil {
		return nil, mapError(err)
	}
	return &queuepb.CreateQueueResponse{Context: toProtoContext(ref)}, nil
}

func (s *GRPCServer) EnqueueTrack(ctx context.Context, req *queuepb.EnqueueTrackRequest) (*queuepb.EnqueueTrackResponse, error) {
	ref, err := parseContext(req.GetContext())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	trackID, err := uuid.Parse(req.GetTrackId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid track_id")
	}
	item, err := s.svc.EnqueueTrack(ctx, ref, trackID)
	if err != nil {
		return nil, mapError(err)
	}
	return &queuepb.EnqueueTrackResponse{Item: toProtoQueueItem(item)}, nil
}

func (s *GRPCServer) ListQueue(ctx context.Context, req *queuepb.ListQueueRequest) (*queuepb.ListQueueResponse, error) {
	ref, err := parseContext(req.GetContext())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	items, err := s.svc.ListQueue(ctx, ref, int(req.GetLimit()))
	if err != nil {
		return nil, mapError(err)
	}
	resp := &queuepb.ListQueueResponse{Items: make([]*queuepb.QueueItem, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, toProtoQueueItem(item))
	}
	return resp, nil
}

func (s *GRPCServer) GetPrevTrack(ctx context.Context, req *queuepb.GetPrevTrackRequest) (*queuepb.GetPrevTrackResponse, error) {
	ref, err := parseContext(req.GetContext())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	prev, err := s.svc.GetPrevTrack(ctx, ref)
	if err != nil {
		return nil, mapError(err)
	}
	resp := &queuepb.GetPrevTrackResponse{}
	if prev != nil {
		resp.Previous = toProtoQueueItem(prev)
	}
	return resp, nil
}

func (s *GRPCServer) GetNextTrack(ctx context.Context, req *queuepb.GetNextTrackRequest) (*queuepb.GetNextTrackResponse, error) {
	ref, err := parseContext(req.GetContext())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	next, err := s.svc.GetNextTrack(ctx, ref)
	if err != nil {
		return nil, mapError(err)
	}
	if next != nil {
		return &queuepb.GetNextTrackResponse{Next: toProtoQueueItem(next)}, nil
	}
	return &queuepb.GetNextTrackResponse{}, nil
}

func (s *GRPCServer) ListPlayedTracks(ctx context.Context, req *queuepb.ListPlayedTracksRequest) (*queuepb.ListPlayedTracksResponse, error) {
	ref, err := parseContext(req.GetContext())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	history, err := s.svc.ListHistory(ctx, ref, int(req.GetLimit()))
	if err != nil {
		return nil, mapError(err)
	}
	resp := &queuepb.ListPlayedTracksResponse{Items: make([]*queuepb.QueueItem, 0, len(history))}
	for _, item := range history {
		resp.Items = append(resp.Items, toProtoQueueItem(item))
	}
	return resp, nil
}

func (s *GRPCServer) ClearQueue(ctx context.Context, req *queuepb.ClearQueueRequest) (*queuepb.ClearQueueResponse, error) {
	ref, err := parseContext(req.GetContext())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.svc.ClearQueue(ctx, ref); err != nil {
		return nil, mapError(err)
	}
	return &queuepb.ClearQueueResponse{Success: true}, nil
}

func (s *GRPCServer) RemoveTrack(ctx context.Context, req *queuepb.RemoveTrackRequest) (*queuepb.RemoveTrackResponse, error) {
	ref, err := parseContext(req.GetContext())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	trackID, err := uuid.Parse(req.GetTrackId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid track_id")
	}
	removed, err := s.svc.RemoveTrack(ctx, ref, trackID)
	if err != nil {
		return nil, mapError(err)
	}
	return &queuepb.RemoveTrackResponse{Success: removed}, nil
}

func (s *GRPCServer) GetCurrentTrack(ctx context.Context, req *queuepb.GetCurrentTrackRequest) (*queuepb.GetCurrentTrackResponse, error) {
	ref, err := parseContext(req.GetContext())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	item, err := s.svc.GetCurrentTrack(ctx, ref)
	if err != nil {
		return nil, mapError(err)
	}
	return &queuepb.GetCurrentTrackResponse{Current: toProtoQueueItem(item)}, nil
}

func parseContext(ref *queuepb.ContextRef) (ContextRef, error) {
	if ref == nil || ref.GetContextType() == "" || ref.GetContextId() == "" {
		return ContextRef{}, ErrBadRequest
	}
	
	var id uuid.UUID
	var err error
	
	// For session context, roomId can be any string, so we use deterministic UUID generation
	// For other contexts (user, group), context_id should be a valid UUID
	if strings.EqualFold(ref.GetContextType(), "session") {
		// Generate deterministic UUID from roomId string using SHA-1 hash
		// This allows any string to be used as roomId while maintaining UUID format
		namespace := uuid.NameSpaceURL
		id = uuid.NewSHA1(namespace, []byte(ref.GetContextId()))
	} else {
		// For user and group contexts, require valid UUID
		id, err = uuid.Parse(ref.GetContextId())
		if err != nil {
			return ContextRef{}, ErrBadRequest
		}
	}
	
	return ContextRef{ContextType: ref.GetContextType(), ContextID: id}, nil
}

func toProtoContext(ref ContextRef) *queuepb.ContextRef {
	if !ref.Valid() {
		return nil
	}
	return &queuepb.ContextRef{ContextType: ref.ContextType, ContextId: ref.ContextID.String()}
}

func toProtoQueueItem(item *QueueItem) *queuepb.QueueItem {
	if item == nil {
		return nil
	}
	return &queuepb.QueueItem{
		TrackId: item.TrackID.String(),
	}
}

func mapError(err error) error {
	switch err {
	case nil:
		return nil
	case ErrBadRequest:
		return status.Error(codes.InvalidArgument, err.Error())
	case ErrEmptyQueue:
		return status.Error(codes.NotFound, err.Error())
	case ErrQueueNotFound:
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
