package generators

import (
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/store"
)

// FallbackGenerator provides fallback candidates when other generators produce too few
type FallbackGenerator struct {
	statsStore store.GlobalStatsStore
	limit      int
}

// NewFallbackGenerator creates a new fallback candidate generator
func NewFallbackGenerator(statsStore store.GlobalStatsStore, limit int) *FallbackGenerator {
	return &FallbackGenerator{
		statsStore: statsStore,
		limit:      limit,
	}
}

func (g *FallbackGenerator) Name() string {
	return "fallback_global"
}

func (g *FallbackGenerator) Generate(user *models.UserProfile) []models.TrackID {
	return g.statsStore.GetTopTracks(g.limit)
}
