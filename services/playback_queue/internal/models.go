package internal

import (
	"strings"

	"github.com/google/uuid"
)

type ContextRef struct {
	ContextType string
	ContextID   uuid.UUID
}

func (c ContextRef) Valid() bool {
	return c.ContextType != "" && c.ContextID != uuid.Nil
}

func (c ContextRef) Key() string {
	return strings.ToLower(c.ContextType) + ":" + c.ContextID.String()
}

type QueueItem struct {
	Context ContextRef
	TrackID uuid.UUID
}
