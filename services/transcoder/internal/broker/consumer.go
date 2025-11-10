package broker

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/MusicSocial/transcoder/internal/config"
	"github.com/MusicSocial/transcoder/internal/transcoder"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader     *kafka.Reader
	transcoder transcoder.Transcoder
	logger     *log.Logger
}

func NewConsumer(cfg config.KafkaConfig, worker transcoder.Transcoder, logger *log.Logger) (*Consumer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, errors.New("kafka brokers not configured")
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       cfg.MinBytes,
		MaxBytes:       cfg.MaxBytes,
		CommitInterval: cfg.CommitInterval,
	})

	return &Consumer{
		reader:     reader,
		transcoder: worker,
		logger:     logger,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			c.logger.Printf("failed to fetch message: %v", err)
			continue
		}

		var task transcoder.Task
		if err := json.Unmarshal(msg.Value, &task); err != nil {
			c.logger.Printf("failed to decode task: %v", err)
			if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
				c.logger.Printf("failed to commit poison pill: %v", commitErr)
			}
			continue
		}

		if err := c.transcoder.Transcode(ctx, task); err != nil {
			c.logger.Printf("transcode failed for track_id=%s: %v", task.TrackID, err)
			// do not commit to retry later
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Printf("failed to commit message: %v", err)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
