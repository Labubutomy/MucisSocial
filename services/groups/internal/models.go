package internal

import (
	"time"

	"github.com/google/uuid"
)

const (
	SuggestionStatusPending  = "pending"
	SuggestionStatusAccepted = "accepted"
	SuggestionStatusRejected = "rejected"
)

type Group struct {
	ID         uuid.UUID  `json:"id"`
	OwnerID    uuid.UUID  `json:"owner_id"`
	Name       string     `json:"name"`
	InviteCode string     `json:"invite_code"`
	QueueID    *uuid.UUID `json:"queue_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type Membership struct {
	GroupID uuid.UUID `json:"group_id"`
	UserID  uuid.UUID `json:"user_id"`
	Role    string    `json:"role"`
	Joined  time.Time `json:"joined_at"`
}

type TrackSuggestion struct {
	ID             uuid.UUID  `json:"id"`
	GroupID        uuid.UUID  `json:"group_id"`
	TrackID        uuid.UUID  `json:"track_id"`
	SuggestedBy    uuid.UUID  `json:"suggested_by"`
	Status         string     `json:"status"`
	DecisionBy     *uuid.UUID `json:"decision_by,omitempty"`
	DecisionReason *string    `json:"decision_reason,omitempty"`
	CooldownUntil  *time.Time `json:"cooldown_until,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type QueueEntry struct {
	ID           uuid.UUID `json:"id"`
	GroupID      uuid.UUID `json:"group_id"`
	SuggestionID uuid.UUID `json:"suggestion_id"`
	TrackID      uuid.UUID `json:"track_id"`
	AddedBy      uuid.UUID `json:"added_by"`
	Position     int64     `json:"position"`
	AddedAt      time.Time `json:"added_at"`
}
