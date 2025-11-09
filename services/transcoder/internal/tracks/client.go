package tracks

import (
	"context"
	"fmt"

	"github.com/MusicSocial/transcoder/internal/config"
	"github.com/MusicSocial/transcoder/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client interface {
	UpdateTrackInfo(ctx context.Context, trackID, audioURL string, duration int32, coverURL string) error
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

func (c *GRPCClient) UpdateTrackInfo(ctx context.Context, trackID, audioURL string, duration int32, coverURL string) error {
	if trackID == "" {
		return fmt.Errorf("trackID is required")
	}
	req := &proto.UpdateTrackInfoRequest{
		TrackId:  trackID,
		AudioUrl: audioURL,
		CoverUrl: coverURL,
		Duration: duration,
	}
	_, err := c.client.UpdateTrackInfo(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update track info: %w", err)
	}
	return nil
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}
