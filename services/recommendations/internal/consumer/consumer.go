package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/store"
	"github.com/segmentio/kafka-go"
)

// EventConsumer consumes events from Kafka/Redpanda and updates stores
type EventConsumer struct {
	trackReader     *kafka.Reader
	listeningReader *kafka.Reader

	trackStore       store.TrackStore
	userProfileStore store.UserProfileStore
	globalStatsStore store.GlobalStatsStore
}

// NewEventConsumer creates a new event consumer
func NewEventConsumer(
	brokers []string,
	trackEventsTopic string,
	listeningEventsTopic string,
	trackStore store.TrackStore,
	userProfileStore store.UserProfileStore,
	globalStatsStore store.GlobalStatsStore,
) *EventConsumer {
	trackReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          trackEventsTopic,
		GroupID:        "recommendations-track-consumer",
		MinBytes:       1,
		MaxBytes:       10e6,
		MaxWait:        1 * time.Second,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 1 * time.Second,
	})

	listeningReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          listeningEventsTopic,
		GroupID:        "recommendations-listening-consumer",
		MinBytes:       1,
		MaxBytes:       10e6,
		MaxWait:        1 * time.Second,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 1 * time.Second,
	})

	return &EventConsumer{
		trackReader:      trackReader,
		listeningReader:  listeningReader,
		trackStore:       trackStore,
		userProfileStore: userProfileStore,
		globalStatsStore: globalStatsStore,
	}
}

// Start begins consuming events from both topics
func (c *EventConsumer) Start(ctx context.Context) error {
	errChan := make(chan error, 2)

	go func() {
		errChan <- c.consumeTrackEvents(ctx)
	}()

	go func() {
		errChan <- c.consumeListeningEvents(ctx)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

// Close closes all readers
func (c *EventConsumer) Close() {
	if c.trackReader != nil {
		c.trackReader.Close()
	}
	if c.listeningReader != nil {
		c.listeningReader.Close()
	}
}

func (c *EventConsumer) consumeTrackEvents(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := c.trackReader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("Error reading track event: %v", err)
			continue
		}

		c.handleTrackEvent(msg.Value)
	}
}

func (c *EventConsumer) consumeListeningEvents(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := c.listeningReader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("Error reading listening event: %v", err)
			continue
		}

		c.handleListeningEvent(msg.Value)
	}
}

func (c *EventConsumer) handleTrackEvent(data []byte) {
	var event models.TrackEvent
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Error unmarshaling track event: %v", err)
		return
	}

	if event.EventType != "track_created" && event.EventType != "track_updated" {
		log.Printf("Unknown track event type: %s", event.EventType)
		return
	}

	track := &models.Track{
		TrackID:    event.TrackID,
		ArtistID:   event.ArtistID,
		Genres:     event.Genres,
		ReleaseTS:  event.ReleaseTS,
		IsExplicit: event.IsExplicit,
	}

	c.trackStore.Upsert(track)
	log.Printf("Processed track event: %s for track %s", event.EventType, event.TrackID)
}

func (c *EventConsumer) handleListeningEvent(data []byte) {
	var event models.ListeningEvent
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Error unmarshaling listening event: %v", err)
		return
	}

	if event.EventType != "track_listened" {
		log.Printf("Unknown listening event type: %s", event.EventType)
		return
	}

	profile := c.userProfileStore.GetOrCreate(event.UserID)

	track, ok := c.trackStore.Get(event.TrackID)
	if ok {
		for _, genre := range track.Genres {
			profile.GenreListenCount[genre]++
		}
		profile.ArtistListenCount[track.ArtistID]++
	}

	profile.ListenedTracks[event.TrackID] = struct{}{}
	c.userProfileStore.Update(profile)
	c.globalStatsStore.IncrementPlayCount(event.TrackID)

	log.Printf("Processed listening event: user %s listened to track %s", event.UserID, event.TrackID)
}
