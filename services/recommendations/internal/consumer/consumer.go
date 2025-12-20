package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	geoTopStore      store.GeoTopStore
	tracksServiceURL string // For fallback track loading
	enableGeoTop     bool
	geohashPrecision int
}

// NewEventConsumer creates a new event consumer
func NewEventConsumer(
	brokers []string,
	trackEventsTopic string,
	listeningEventsTopic string,
	trackStore store.TrackStore,
	userProfileStore store.UserProfileStore,
	globalStatsStore store.GlobalStatsStore,
	geoTopStore store.GeoTopStore,
	tracksServiceURL string,
	enableGeoTop bool,
	geohashPrecision int,
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
		geoTopStore:      geoTopStore,
		tracksServiceURL: tracksServiceURL,
		enableGeoTop:     enableGeoTop,
		geohashPrecision: geohashPrecision,
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
	if !ok {
		// Try to load track from tracks-service as fallback
		if c.tracksServiceURL != "" {
			log.Printf("Attempting to load track %s from tracks-service (URL: %s)", event.TrackID, c.tracksServiceURL)
			if loadedTrack := c.loadTrackFromService(event.TrackID); loadedTrack != nil {
				track = loadedTrack
				ok = true
				log.Printf("Successfully loaded track %s from tracks-service as fallback (genres: %v, artist: %s)", event.TrackID, track.Genres, track.ArtistID)
			} else {
				log.Printf("Failed to load track %s from tracks-service", event.TrackID)
			}
		} else {
			log.Printf("tracksServiceURL is empty, cannot load track %s from service", event.TrackID)
		}

		if !ok {
			log.Printf("Warning: Track %s not found in TrackStore when processing listening event for user %s. Statistics will not be updated.", event.TrackID, event.UserID)
			// Still mark track as listened even if metadata is missing
			profile.ListenedTracks[event.TrackID] = struct{}{}
			c.userProfileStore.Update(profile)
			c.globalStatsStore.IncrementPlayCount(event.TrackID)
			return
		}
	}

	// Update genre statistics
	if len(track.Genres) > 0 {
		for _, genre := range track.Genres {
			profile.GenreListenCount[genre]++
		}
	} else {
		log.Printf("Warning: Track %s has no genres", event.TrackID)
	}

	// Update artist statistics
	if track.ArtistID != "" {
		profile.ArtistListenCount[track.ArtistID]++
	} else {
		log.Printf("Warning: Track %s has no artist_id", event.TrackID)
	}

	profile.ListenedTracks[event.TrackID] = struct{}{}
	c.userProfileStore.Update(profile)
	c.globalStatsStore.IncrementPlayCount(event.TrackID)

	// Update geo-based aggregates if coordinates are provided and feature is enabled
	if c.enableGeoTop && event.Lat != nil && event.Lon != nil {
		ghash := store.EncodeGeohash(*event.Lat, *event.Lon, c.geohashPrecision)
		c.geoTopStore.Incr(ghash, event.TrackID, 1)
		log.Printf("Updated geo aggregate: geohash=%s, track=%s, lat=%f, lon=%f",
			ghash, event.TrackID, *event.Lat, *event.Lon)
	}

	log.Printf("Processed listening event: user %s listened to track %s (genres: %v, artist: %s)",
		event.UserID, event.TrackID, track.Genres, track.ArtistID)
}

// loadTrackFromService attempts to load a track from tracks-service
func (c *EventConsumer) loadTrackFromService(trackID string) *models.Track {
	if c.tracksServiceURL == "" {
		log.Printf("loadTrackFromService: tracksServiceURL is empty")
		return nil
	}

	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("%s/api/tracks/%s", c.tracksServiceURL, trackID)
	log.Printf("loadTrackFromService: requesting %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("loadTrackFromService: failed to create request: %v", err)
		return nil
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("loadTrackFromService: HTTP request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("loadTrackFromService: tracks-service returned status %d for track %s", resp.StatusCode, trackID)
		return nil
	}

	var trackResp struct {
		ID        string    `json:"id"`
		ArtistIDs []string  `json:"artist_ids"` // UUIDs as strings
		Genre     string    `json:"genre"`
		CreatedAt time.Time `json:"created_at"` // time.Time from Go
	}

	if err := json.NewDecoder(resp.Body).Decode(&trackResp); err != nil {
		log.Printf("loadTrackFromService: failed to decode response: %v", err)
		return nil
	}

	log.Printf("loadTrackFromService: received track data - ID: %s, ArtistIDs: %v, Genre: %s", trackResp.ID, trackResp.ArtistIDs, trackResp.Genre)

	artistID := ""
	if len(trackResp.ArtistIDs) > 0 {
		artistID = trackResp.ArtistIDs[0]
	}

	genres := []string{}
	if trackResp.Genre != "" {
		genres = append(genres, trackResp.Genre)
	}

	releaseTS := trackResp.CreatedAt.Unix()

	track := &models.Track{
		TrackID:    trackID,
		ArtistID:   artistID,
		Genres:     genres,
		ReleaseTS:  releaseTS,
		IsExplicit: false,
	}

	// Store the track for future use
	c.trackStore.Upsert(track)
	return track
}
