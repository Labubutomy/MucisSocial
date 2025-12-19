package generators

import (
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/store"
)

// FallbackGenerator provides fallback candidates when other generators produce too few
type FallbackGenerator struct {
	trackStore store.TrackStore
	statsStore store.GlobalStatsStore
	limit      int
}

// NewFallbackGenerator creates a new fallback candidate generator
func NewFallbackGenerator(trackStore store.TrackStore, statsStore store.GlobalStatsStore, limit int) *FallbackGenerator {
	return &FallbackGenerator{
		trackStore: trackStore,
		statsStore: statsStore,
		limit:      limit,
	}
}

func (g *FallbackGenerator) Name() string {
	return "fallback_global"
}

func (g *FallbackGenerator) Generate(user *models.UserProfile) []models.TrackID {
	// Try to get top tracks from global stats first
	topTracks := g.statsStore.GetTopTracks(g.limit)
	
	// If no stats available, fall back to all tracks from track store
	if len(topTracks) == 0 {
		allTracks := g.trackStore.GetAll()
		result := make([]models.TrackID, 0, g.limit)
		for i, track := range allTracks {
			if i >= g.limit {
				break
			}
			result = append(result, track.TrackID)
		}
		return result
	}
	
	return topTracks
}
