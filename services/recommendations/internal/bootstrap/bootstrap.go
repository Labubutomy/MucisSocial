package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/store"
)

// TrackResponse represents track from tracks service API
type TrackResponse struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	ArtistIDs []string `json:"artist_ids"`
	Genre     string   `json:"genre"`
	Duration  int      `json:"duration_seconds"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
}

// TracksListResponse represents paginated response from tracks service
type TracksListResponse struct {
	Tracks []TrackResponse `json:"tracks"`
	Total  int             `json:"total"`
}

// Bootstrapper loads initial data from external services
type Bootstrapper struct {
	tracksServiceURL string
	trackStore       store.TrackStore
	httpClient       *http.Client
}

// NewBootstrapper creates a new bootstrapper
func NewBootstrapper(tracksServiceURL string, trackStore store.TrackStore) *Bootstrapper {
	return &Bootstrapper{
		tracksServiceURL: tracksServiceURL,
		trackStore:       trackStore,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// LoadTracks fetches all tracks from tracks service and populates TrackStore
func (b *Bootstrapper) LoadTracks(ctx context.Context) error {
	if b.tracksServiceURL == "" {
		log.Println("Bootstrap: TRACKS_SERVICE_URL not configured, skipping initial load")
		return nil
	}

	log.Printf("Bootstrap: Loading tracks from %s", b.tracksServiceURL)

	offset := 0
	limit := 100
	totalLoaded := 0

	for {
		url := fmt.Sprintf("%s/tracks?limit=%d&offset=%d", b.tracksServiceURL, limit, offset)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := b.httpClient.Do(req)
		if err != nil {
			log.Printf("Bootstrap: Failed to fetch tracks (offset=%d): %v", offset, err)
			break
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			log.Printf("Bootstrap: Tracks service returned status %d", resp.StatusCode)
			break
		}

		var listResp TracksListResponse
		if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
			resp.Body.Close()
			log.Printf("Bootstrap: Failed to decode response: %v", err)
			break
		}
		resp.Body.Close()

		if len(listResp.Tracks) == 0 {
			break
		}

		for _, tr := range listResp.Tracks {
			if tr.Status != "ready" && tr.Status != "active" {
				continue
			}

			artistID := ""
			if len(tr.ArtistIDs) > 0 {
				artistID = tr.ArtistIDs[0]
			}

			genres := []string{}
			if tr.Genre != "" {
				genres = append(genres, tr.Genre)
			}

			var releaseTS int64
			if t, err := time.Parse(time.RFC3339, tr.CreatedAt); err == nil {
				releaseTS = t.Unix()
			}

			track := &models.Track{
				TrackID:    tr.ID,
				ArtistID:   artistID,
				Genres:     genres,
				ReleaseTS:  releaseTS,
				IsExplicit: false,
			}

			b.trackStore.Upsert(track)
			totalLoaded++
		}

		if len(listResp.Tracks) < limit {
			break
		}

		offset += limit
	}

	log.Printf("Bootstrap: Loaded %d tracks into TrackStore", totalLoaded)
	return nil
}

// Run performs full bootstrap
func (b *Bootstrapper) Run(ctx context.Context) error {
	return b.LoadTracks(ctx)
}
