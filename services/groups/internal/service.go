package internal

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo   *Repository
	config Config
	queue  QueueClient
}

func NewService(repo *Repository, cfg Config, queue QueueClient) *Service {
	return &Service{repo: repo, config: cfg, queue: queue}
}

func (s *Service) CreateGroup(ctx context.Context, ownerID uuid.UUID, name string) (*Group, error) {
	if name == "" || ownerID == uuid.Nil {
		return nil, ErrBadRequest
	}
	if s.queue == nil {
		return nil, ErrInternal
	}

	queueID, err := s.queue.CreateQueue(ctx, "groups")
	if err != nil {
		return nil, err
	}

	code, err := s.generateInviteCode(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	group := &Group{
		ID:         uuid.New(),
		OwnerID:    ownerID,
		Name:       name,
		InviteCode: code,
		QueueID:    &queueID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.CreateGroup(ctx, group); err != nil {
		return nil, err
	}

	if err := s.repo.AddMembership(ctx, Membership{
		GroupID: group.ID,
		UserID:  ownerID,
		Role:    "owner",
		Joined:  now,
	}); err != nil && !errors.Is(err, ErrAlreadyMember) {
		return nil, err
	}

	return group, nil
}

func (s *Service) JoinByInviteCode(ctx context.Context, userID uuid.UUID, inviteCode string) (*Group, error) {
	if userID == uuid.Nil || inviteCode == "" {
		return nil, ErrBadRequest
	}

	group, err := s.repo.GetGroupByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, err
	}

	joined := Membership{
		GroupID: group.ID,
		UserID:  userID,
		Role:    "member",
		Joined:  time.Now().UTC(),
	}
	if err := s.repo.AddMembership(ctx, joined); err != nil {
		if errors.Is(err, ErrAlreadyMember) {
			return group, nil
		}
		return nil, err
	}

	return group, nil
}

func (s *Service) SubmitSuggestion(ctx context.Context, groupID, userID, trackID uuid.UUID) (*TrackSuggestion, error) {
	if groupID == uuid.Nil || userID == uuid.Nil || trackID == uuid.Nil {
		return nil, ErrBadRequest
	}

	if err := s.ensureMembership(ctx, groupID, userID); err != nil {
		return nil, err
	}

	count, err := s.repo.CountPendingSuggestions(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if count >= s.config.SuggestionLimit {
		return nil, ErrSuggestionLimit
	}

	now := time.Now().UTC()
	hasCooldown, _, err := s.repo.HasActiveCooldown(ctx, groupID, userID, trackID, now)
	if err != nil {
		return nil, err
	}
	if hasCooldown {
		return nil, ErrSuggestionCooldown
	}

	suggestion := &TrackSuggestion{
		ID:          uuid.New(),
		GroupID:     groupID,
		TrackID:     trackID,
		SuggestedBy: userID,
		Status:      SuggestionStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateSuggestion(ctx, suggestion); err != nil {
		return nil, err
	}

	return suggestion, nil
}

func (s *Service) ListSuggestions(ctx context.Context, groupID uuid.UUID, status string, limit, offset int) ([]*TrackSuggestion, error) {
	if groupID == uuid.Nil {
		return nil, ErrBadRequest
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListSuggestions(ctx, groupID, status, limit, offset)
}

func (s *Service) AcceptSuggestion(ctx context.Context, groupID, ownerID, suggestionID uuid.UUID) (*QueueEntry, error) {
	if err := s.ensureOwner(ctx, groupID, ownerID); err != nil {
		return nil, err
	}

	group, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.QueueID == nil || *group.QueueID == uuid.Nil {
		return nil, ErrInternal
	}

	suggestion, err := s.repo.GetSuggestionByID(ctx, suggestionID)
	if err != nil {
		return nil, err
	}
	if suggestion.GroupID != groupID {
		return nil, ErrBadRequest
	}
	if suggestion.Status != SuggestionStatusPending {
		return nil, ErrSuggestionHandled
	}

	if s.queue == nil {
		return nil, ErrInternal
	}

	if err := s.queue.EnqueueTrack(ctx, *group.QueueID, suggestion.TrackID); err != nil {
		return nil, err
	}

	tracks, err := s.queue.ListQueue(ctx, *group.QueueID, 1000)
	if err != nil {
		return nil, err
	}
	position := int64(len(tracks))

	if err := s.repo.UpdateSuggestionStatus(ctx, suggestion.ID, SuggestionStatusAccepted, ownerID, nil, nil); err != nil {
		return nil, err
	}

	entry := &QueueEntry{
		ID:           uuid.New(),
		GroupID:      groupID,
		SuggestionID: suggestion.ID,
		TrackID:      suggestion.TrackID,
		AddedBy:      ownerID,
		Position:     position,
		AddedAt:      time.Now().UTC(),
	}

	return entry, nil
}

func (s *Service) RejectSuggestion(ctx context.Context, groupID, ownerID, suggestionID uuid.UUID, reason string) error {
	if err := s.ensureOwner(ctx, groupID, ownerID); err != nil {
		return err
	}

	suggestion, err := s.repo.GetSuggestionByID(ctx, suggestionID)
	if err != nil {
		return err
	}
	if suggestion.GroupID != groupID {
		return ErrBadRequest
	}
	if suggestion.Status != SuggestionStatusPending {
		return ErrSuggestionHandled
	}

	cooldown := time.Now().UTC().Add(s.config.SuggestionCooldown)
	if reason == "" {
		reason = "Rejected by owner"
	}

	if err := s.repo.UpdateSuggestionStatus(ctx, suggestion.ID, SuggestionStatusRejected, ownerID, &reason, &cooldown); err != nil {
		return err
	}

	return nil
}

func (s *Service) ListQueue(ctx context.Context, groupID uuid.UUID, limit int) ([]*QueueEntry, error) {
	if groupID == uuid.Nil {
		return nil, ErrBadRequest
	}
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	group, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.QueueID == nil || *group.QueueID == uuid.Nil {
		return nil, ErrInternal
	}
	if s.queue == nil {
		return nil, ErrInternal
	}
	tracks, err := s.queue.ListQueue(ctx, *group.QueueID, limit)
	if err != nil {
		return nil, err
	}
	entries := make([]*QueueEntry, 0, len(tracks))
	for idx, trackID := range tracks {
		entries = append(entries, &QueueEntry{
			ID:           uuid.Nil,
			GroupID:      groupID,
			SuggestionID: uuid.Nil,
			TrackID:      trackID,
			AddedBy:      uuid.Nil,
			Position:     int64(idx + 1),
			AddedAt:      time.Time{},
		})
	}
	return entries, nil
}

func (s *Service) GetGroup(ctx context.Context, groupID uuid.UUID) (*Group, error) {
	return s.repo.GetGroupByID(ctx, groupID)
}

func (s *Service) LeaveGroup(ctx context.Context, groupID, userID uuid.UUID) error {
	if groupID == uuid.Nil || userID == uuid.Nil {
		return ErrBadRequest
	}

	if err := s.ensureMembership(ctx, groupID, userID); err != nil {
		return err
	}

	ownerID, err := s.repo.GroupOwnerID(ctx, groupID)
	if err != nil {
		return err
	}
	if ownerID == userID {
		return ErrOwnerCannotLeave
	}

	return s.repo.RemoveMembership(ctx, groupID, userID)
}

func (s *Service) DeleteGroup(ctx context.Context, groupID, ownerID uuid.UUID) error {
	if groupID == uuid.Nil || ownerID == uuid.Nil {
		return ErrBadRequest
	}
	if err := s.ensureOwner(ctx, groupID, ownerID); err != nil {
		return err
	}
	return s.repo.DeleteGroup(ctx, groupID)
}

func (s *Service) ensureMembership(ctx context.Context, groupID, userID uuid.UUID) error {
	member, err := s.repo.IsMember(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if !member {
		return ErrNotMember
	}
	return nil
}

func (s *Service) ensureOwner(ctx context.Context, groupID, userID uuid.UUID) error {
	owner, err := s.repo.GroupOwnerID(ctx, groupID)
	if err != nil {
		return err
	}
	if owner != userID {
		return ErrNotOwner
	}
	return nil
}

func (s *Service) generateInviteCode(ctx context.Context) (string, error) {
	for i := 0; i < 5; i++ {
		code := randomSlug()
		exists, err := s.repo.InviteCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", ErrBadRequest
}

func randomSlug() string {
	buf := make([]byte, 9)
	if _, err := rand.Read(buf); err != nil {
		return uuid.New().String()
	}
	encoded := base64.RawURLEncoding.EncodeToString(buf)
	encoded = strings.TrimRight(encoded, "=")
	if len(encoded) > 12 {
		encoded = encoded[:12]
	}
	return encoded
}

func (s *Service) BuildInviteLink(code string) string {
	base := strings.TrimRight(s.config.LinkBaseURL, "/")
	if base == "" {
		base = "https://music.local/groups"
	}
	return base + "/" + code
}
