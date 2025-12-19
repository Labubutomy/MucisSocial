package internal

import (
	"context"
	"time"

	groupspb "github.com/Labubutomy/MucisSocial/services/groups/api"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	groupspb.UnimplementedGroupsServiceServer
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) CreateGroup(ctx context.Context, req *groupspb.CreateGroupRequest) (*groupspb.CreateGroupResponse, error) {
	ownerID, err := parseUUID(req.GetOwnerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
	}
	group, err := s.svc.CreateGroup(ctx, ownerID, req.GetName())
	if err != nil {
		return nil, mapError(err)
	}
	return &groupspb.CreateGroupResponse{
		Group:      toProtoGroup(group),
		InviteLink: s.svc.BuildInviteLink(group.InviteCode),
	}, nil
}

func (s *GRPCServer) JoinGroup(ctx context.Context, req *groupspb.JoinGroupRequest) (*groupspb.JoinGroupResponse, error) {
	userID, err := parseUUID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	group, err := s.svc.JoinByInviteCode(ctx, userID, req.GetInviteCode())
	if err != nil {
		return nil, mapError(err)
	}
	return &groupspb.JoinGroupResponse{
		Group:      toProtoGroup(group),
		InviteLink: s.svc.BuildInviteLink(group.InviteCode),
	}, nil
}

func (s *GRPCServer) LeaveGroup(ctx context.Context, req *groupspb.LeaveGroupRequest) (*groupspb.LeaveGroupResponse, error) {
	groupID, err := parseUUID(req.GetGroupId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid group_id")
	}
	userID, err := parseUUID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	if err := s.svc.LeaveGroup(ctx, groupID, userID); err != nil {
		return nil, mapError(err)
	}
	return &groupspb.LeaveGroupResponse{Success: true}, nil
}

func (s *GRPCServer) DeleteGroup(ctx context.Context, req *groupspb.DeleteGroupRequest) (*groupspb.DeleteGroupResponse, error) {
	groupID, err := parseUUID(req.GetGroupId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid group_id")
	}
	ownerID, err := parseUUID(req.GetOwnerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
	}
	if err := s.svc.DeleteGroup(ctx, groupID, ownerID); err != nil {
		return nil, mapError(err)
	}
	return &groupspb.DeleteGroupResponse{Success: true}, nil
}

func (s *GRPCServer) GetGroup(ctx context.Context, req *groupspb.GetGroupRequest) (*groupspb.GetGroupResponse, error) {
	groupID, err := parseUUID(req.GetGroupId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid group_id")
	}
	group, err := s.svc.GetGroup(ctx, groupID)
	if err != nil {
		return nil, mapError(err)
	}
	return &groupspb.GetGroupResponse{
		Group:      toProtoGroup(group),
		InviteLink: s.svc.BuildInviteLink(group.InviteCode),
	}, nil
}

func (s *GRPCServer) SubmitSuggestion(ctx context.Context, req *groupspb.SubmitSuggestionRequest) (*groupspb.SubmitSuggestionResponse, error) {
	groupID, err := parseUUID(req.GetGroupId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid group_id")
	}
	userID, err := parseUUID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	trackID, err := parseUUID(req.GetTrackId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid track_id")
	}
	suggestion, err := s.svc.SubmitSuggestion(ctx, groupID, userID, trackID)
	if err != nil {
		return nil, mapError(err)
	}
	return &groupspb.SubmitSuggestionResponse{Suggestion: toProtoSuggestion(suggestion)}, nil
}

func (s *GRPCServer) ListSuggestions(ctx context.Context, req *groupspb.ListSuggestionsRequest) (*groupspb.ListSuggestionsResponse, error) {
	groupID, err := parseUUID(req.GetGroupId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid group_id")
	}
	ownerID, err := parseUUID(req.GetOwnerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
	}
	if err := s.svc.ensureOwner(ctx, groupID, ownerID); err != nil {
		return nil, mapError(err)
	}
	limit := int(req.GetLimit())
	offset := int(req.GetOffset())
	suggestions, err := s.svc.ListSuggestions(ctx, groupID, req.GetStatus(), limit, offset)
	if err != nil {
		return nil, mapError(err)
	}
	resp := &groupspb.ListSuggestionsResponse{Suggestions: make([]*groupspb.TrackSuggestion, 0, len(suggestions))}
	for _, sgs := range suggestions {
		resp.Suggestions = append(resp.Suggestions, toProtoSuggestion(sgs))
	}
	return resp, nil
}

func (s *GRPCServer) AcceptSuggestion(ctx context.Context, req *groupspb.AcceptSuggestionRequest) (*groupspb.AcceptSuggestionResponse, error) {
	groupID, err := parseUUID(req.GetGroupId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid group_id")
	}
	ownerID, err := parseUUID(req.GetOwnerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
	}
	suggestionID, err := parseUUID(req.GetSuggestionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid suggestion_id")
	}
	entry, err := s.svc.AcceptSuggestion(ctx, groupID, ownerID, suggestionID)
	if err != nil {
		return nil, mapError(err)
	}
	return &groupspb.AcceptSuggestionResponse{Entry: toProtoQueueEntry(entry)}, nil
}

func (s *GRPCServer) RejectSuggestion(ctx context.Context, req *groupspb.RejectSuggestionRequest) (*groupspb.RejectSuggestionResponse, error) {
	groupID, err := parseUUID(req.GetGroupId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid group_id")
	}
	ownerID, err := parseUUID(req.GetOwnerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
	}
	suggestionID, err := parseUUID(req.GetSuggestionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid suggestion_id")
	}
	if err := s.svc.RejectSuggestion(ctx, groupID, ownerID, suggestionID, req.GetReason()); err != nil {
		return nil, mapError(err)
	}
	return &groupspb.RejectSuggestionResponse{Success: true}, nil
}

func (s *GRPCServer) ListQueue(ctx context.Context, req *groupspb.ListQueueRequest) (*groupspb.ListQueueResponse, error) {
	groupID, err := parseUUID(req.GetGroupId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid group_id")
	}
	limit := int(req.GetLimit())
	entries, err := s.svc.ListQueue(ctx, groupID, limit)
	if err != nil {
		return nil, mapError(err)
	}
	resp := &groupspb.ListQueueResponse{Queue: make([]*groupspb.QueueEntry, 0, len(entries))}
	for _, entry := range entries {
		resp.Queue = append(resp.Queue, toProtoQueueEntry(entry))
	}
	return resp, nil
}

func toProtoGroup(group *Group) *groupspb.Group {
	if group == nil {
		return nil
	}
	var queueID string
	if group.QueueID != nil {
		queueID = group.QueueID.String()
	}
	return &groupspb.Group{
		Id:         group.ID.String(),
		OwnerId:    group.OwnerID.String(),
		Name:       group.Name,
		InviteCode: group.InviteCode,
		QueueId:    queueID,
		QueueContextType: func() string {
			if group.QueueID != nil {
				return "groups"
			}
			return ""
		}(),
		CreatedAt: formatTime(group.CreatedAt),
		UpdatedAt: formatTime(group.UpdatedAt),
	}
}

func toProtoSuggestion(suggestion *TrackSuggestion) *groupspb.TrackSuggestion {
	if suggestion == nil {
		return nil
	}
	var handledBy string
	if suggestion.DecisionBy != nil {
		handledBy = suggestion.DecisionBy.String()
	}
	var handledAt string
	if suggestion.DecisionBy != nil {
		handledAt = formatTime(suggestion.UpdatedAt)
	}
	var rejection string
	if suggestion.DecisionReason != nil {
		rejection = *suggestion.DecisionReason
	}
	return &groupspb.TrackSuggestion{
		Id:              suggestion.ID.String(),
		GroupId:         suggestion.GroupID.String(),
		TrackId:         suggestion.TrackID.String(),
		SuggestedBy:     suggestion.SuggestedBy.String(),
		Status:          suggestion.Status,
		CreatedAt:       formatTime(suggestion.CreatedAt),
		UpdatedAt:       formatTime(suggestion.UpdatedAt),
		HandledBy:       handledBy,
		HandledAt:       handledAt,
		RejectionReason: rejection,
	}
}

func toProtoQueueEntry(entry *QueueEntry) *groupspb.QueueEntry {
	if entry == nil {
		return nil
	}
	return &groupspb.QueueEntry{
		Id:           entry.ID.String(),
		GroupId:      entry.GroupID.String(),
		SuggestionId: entry.SuggestionID.String(),
		TrackId:      entry.TrackID.String(),
		AddedBy:      entry.AddedBy.String(),
		Position:     int32(entry.Position),
		AddedAt:      formatTime(entry.AddedAt),
	}
}

func parseUUID(val string) (uuid.UUID, error) {
	return uuid.Parse(val)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func mapError(err error) error {
	switch err {
	case nil:
		return nil
	case ErrBadRequest, ErrInviteCodeInvalid:
		return status.Error(codes.InvalidArgument, err.Error())
	case ErrAlreadyMember:
		return status.Error(codes.AlreadyExists, err.Error())
	case ErrNotMember:
		return status.Error(codes.PermissionDenied, err.Error())
	case ErrNotOwner:
		return status.Error(codes.PermissionDenied, err.Error())
	case ErrOwnerCannotLeave:
		return status.Error(codes.FailedPrecondition, err.Error())
	case ErrSuggestionLimit, ErrSuggestionCooldown:
		return status.Error(codes.ResourceExhausted, err.Error())
	case ErrSuggestionHandled:
		return status.Error(codes.FailedPrecondition, err.Error())
	case ErrNotFound:
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
