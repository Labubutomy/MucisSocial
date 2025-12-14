package internal

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// TrackEvent represents a track event for recommendations service
type TrackEvent struct {
	EventType  string   `json:"event_type"`
	TrackID    string   `json:"track_id"`
	ArtistID   string   `json:"artist_id"` // Primary artist
	Genres     []string `json:"genres"`
	ReleaseTS  int64    `json:"release_ts"`
	IsExplicit bool     `json:"is_explicit"`
}

// EventProducer handles publishing events to Kafka
type EventProducer struct {
	writer *kafka.Writer
}

// NewEventProducer creates a new event producer
func NewEventProducer(brokers []string, topic string) *EventProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        true, // Non-blocking writes
	}

	return &EventProducer{writer: writer}
}

// PublishTrackCreated publishes a track_created event
func (p *EventProducer) PublishTrackCreated(ctx context.Context, track *Track) error {
	return p.publishTrackEvent(ctx, "track_created", track)
}

// PublishTrackUpdated publishes a track_updated event
func (p *EventProducer) PublishTrackUpdated(ctx context.Context, track *Track) error {
	return p.publishTrackEvent(ctx, "track_updated", track)
}

func (p *EventProducer) publishTrackEvent(ctx context.Context, eventType string, track *Track) error {
	// Get primary artist (first one)
	artistID := ""
	if len(track.ArtistIDs) > 0 {
		artistID = track.ArtistIDs[0].String()
	}

	event := TrackEvent{
		EventType:  eventType,
		TrackID:    track.ID.String(),
		ArtistID:   artistID,
		Genres:     []string{track.Genre}, // Single genre for now
		ReleaseTS:  track.CreatedAt.Unix(),
		IsExplicit: false, // Not tracked yet
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal track event: %v", err)
		return err
	}

	msg := kafka.Message{
		Key:   []byte(track.ID.String()),
		Value: data,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Printf("Failed to publish track event: %v", err)
		return err
	}

	log.Printf("Published %s event for track %s", eventType, track.ID.String())
	return nil
}

// Close closes the producer
func (p *EventProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
