package tracks

import (
	"context"
	"fmt"

	"github.com/MucisSocial/upload/internal/config"
	"github.com/MucisSocial/upload/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client interface {
	CreateTrack(ctx context.Context, artistID, name string, duration int64) (string, error)
	UpdateTrackLocation(ctx context.Context, trackID, storageURL string) error
	Close() error
}

type GRPCClient struct {
	conn   *grpc.ClientConn
	client proto.TrackServiceClient
}

func NewTrackClient(cfg *config.TrackServiceConfig) (*GRPCClient, error) {
	conn, err := grpc.NewClient(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to track service: %w", err)
	}

	return &GRPCClient{
		conn:   conn,
		client: proto.NewTrackServiceClient(conn),
	}, nil
}

func (c *GRPCClient) CreateTrack(ctx context.Context, artistID, name string, duration int64) (string, error) {
	resp, err := c.client.CreateTrack(ctx, &proto.CreateTrackRequest{
		ArtistId: artistID,
		Name:     name,
		Duration: duration,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create track in track service: %w", err)
	}

	trackID := resp.GetTrackId()
	if trackID == "" {
		return "", fmt.Errorf("track service returned empty track id")
	}

	return trackID, nil
}

func (c *GRPCClient) UpdateTrackLocation(ctx context.Context, trackID, storageURL string) error {
	if trackID == "" {
		return fmt.Errorf("trackID is required")
	}
	if storageURL == "" {
		return fmt.Errorf("storageURL is required")
	}

	_, err := c.client.UpdateTrackLocation(ctx, &proto.UpdateTrackLocationRequest{
		TrackId:    trackID,
		StorageUrl: storageURL,
	})
	if err != nil {
		return fmt.Errorf("failed to update track location in track service: %w", err)
	}

	return nil
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}
