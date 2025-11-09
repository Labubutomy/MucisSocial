package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MusicSocial/upload/internal/config"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	transcoderWriter *kafka.Writer
}

type TranscoderTask struct {
	TrackID  string `json:"track_id"`
	ArtistID string `json:"artist_id"`
	TrackURL string `json:"track_url"`
}

func NewProducer(cfg *config.RedpandaConfig) (*Producer, error) {
	transcoderWriter := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.TranscoderTopic,
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		transcoderWriter: transcoderWriter,
	}, nil
}

func (p *Producer) SendTranscoderTask(ctx context.Context, task TranscoderTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal transcoder task: %w", err)
	}

	err = p.transcoderWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(task.TrackID),
		Value: data,
	})
	if err != nil {
		return fmt.Errorf("failed to send transcoder task: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	if err := p.transcoderWriter.Close(); err != nil {
		return err
	}
	return nil
}
