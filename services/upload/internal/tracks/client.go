package tracks

import (
	"context"
	"fmt"

	"github.com/MusicSocial/upload/internal/config"
	trackspb "github.com/MusicSocial/upload/proto/tracks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client interface {
	CreateTrack(ctx context.Context, name string, artistIDs []string, genre string, duration int32) (string, error)
	Close() error
}

type GRPCClient struct {
	conn   *grpc.ClientConn
	client trackspb.TracksServiceClient
}

func NewTrackClient(cfg *config.TrackServiceConfig) (*GRPCClient, error) {
	conn, err := grpc.NewClient(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to track service: %w", err)
	}

	return &GRPCClient{
		conn:   conn,
		client: trackspb.NewTracksServiceClient(conn),
	}, nil
}

func (c *GRPCClient) CreateTrack(ctx context.Context, name string, artistIDs []string, genre string, duration int32) (string, error) {
	resp, err := c.client.CreateTrack(ctx, &trackspb.CreateTrackRequest{
		Title:     name,
		ArtistIds: artistIDs,
		Genre:     genre,
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

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}
