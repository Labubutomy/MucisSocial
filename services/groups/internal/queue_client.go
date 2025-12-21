package internal

import (
	"context"
	"fmt"

	queuepb "github.com/Labubutomy/MucisSocial/services/groups/proto/queue"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type QueueClient interface {
	CreateQueue(ctx context.Context, contextType string) (uuid.UUID, error)
	EnqueueTrack(ctx context.Context, queueID uuid.UUID, trackID uuid.UUID) error
	ListQueue(ctx context.Context, queueID uuid.UUID, limit int) ([]uuid.UUID, error)
}

type playbackQueueClient struct {
	client queuepb.PlaybackQueueServiceClient
	conn   *grpc.ClientConn
}

func NewPlaybackQueueClient(addr string) (QueueClient, func(), error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("dial playback queue: %w", err)
	}
	c := &playbackQueueClient{client: queuepb.NewPlaybackQueueServiceClient(conn), conn: conn}
	cleanup := func() {
		_ = conn.Close()
	}
	return c, cleanup, nil
}

func (c *playbackQueueClient) CreateQueue(ctx context.Context, contextType string) (uuid.UUID, error) {
	resp, err := c.client.CreateQueue(ctx, &queuepb.CreateQueueRequest{ContextType: contextType})
	if err != nil {
		return uuid.Nil, err
	}
	id, err := uuid.Parse(resp.GetContext().GetContextId())
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (c *playbackQueueClient) EnqueueTrack(ctx context.Context, queueID uuid.UUID, trackID uuid.UUID) error {
	_, err := c.client.EnqueueTrack(ctx, &queuepb.EnqueueTrackRequest{
		Context: &queuepb.ContextRef{ContextType: "groups", ContextId: queueID.String()},
		TrackId: trackID.String(),
	})
	return err
}

func (c *playbackQueueClient) ListQueue(ctx context.Context, queueID uuid.UUID, limit int) ([]uuid.UUID, error) {
	resp, err := c.client.ListQueue(ctx, &queuepb.ListQueueRequest{
		Context: &queuepb.ContextRef{ContextType: "groups", ContextId: queueID.String()},
		Limit:   int32(limit),
	})
	if err != nil {
		return nil, err
	}
	items := make([]uuid.UUID, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		id, parseErr := uuid.Parse(item.GetTrackId())
		if parseErr != nil {
			return nil, parseErr
		}
		items = append(items, id)
	}
	return items, nil
}
