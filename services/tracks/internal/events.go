package internal

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
)

const (
	TrackUploadedEventType             = "track.uploaded"
	TrackTranscodingCompletedEventType = "track.transcoding.completed"
)

type EventConsumer struct {
	service *Service
	reader  *kafka.Reader
}

func NewEventConsumer(service *Service, brokers string) *EventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokers},
		Topic:   "track-events",
		GroupID: "track-service",
	})

	return &EventConsumer{
		service: service,
		reader:  reader,
	}
}

func (c *EventConsumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return c.reader.Close()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				log.Printf("Kafka error: %v", err)
				continue
			}

			if err := c.handleMessage(ctx, msg); err != nil {
				log.Printf("Handle error: %v", err)
			} else {
				c.reader.CommitMessages(ctx, msg)
			}
		}
	}
}

func (c *EventConsumer) handleMessage(ctx context.Context, msg kafka.Message) error {
	eventType := getHeader(msg.Headers, "event-type")

	switch eventType {
	case TrackUploadedEventType:
		var event TrackUploadedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return err
		}
		return c.service.HandleTrackUploaded(ctx, event)

	case TrackTranscodingCompletedEventType:
		var event TranscodingCompletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return err
		}
		return c.service.HandleTranscodingCompleted(ctx, event)
	}

	return nil
}

func getHeader(headers []kafka.Header, key string) string {
	for _, h := range headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}
