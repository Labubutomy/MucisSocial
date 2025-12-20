package generators

import (
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/store"
)

// LocalTopGenerator provides track recommendations based on geographic location
type LocalTopGenerator struct {
	geoTopStore      store.GeoTopStore
	defaultRadiusM   int
	geohashPrecision int
	limit            int
}

// NewLocalTopGenerator creates a new local top candidate generator
func NewLocalTopGenerator(
	geoTopStore store.GeoTopStore,
	defaultRadiusM int,
	geohashPrecision int,
	limit int,
) *LocalTopGenerator {
	return &LocalTopGenerator{
		geoTopStore:      geoTopStore,
		defaultRadiusM:   defaultRadiusM,
		geohashPrecision: geohashPrecision,
		limit:            limit,
	}
}

func (g *LocalTopGenerator) Name() string {
	return "local_top"
}

// Generate returns top tracks near the user's location
// Note: This generator expects location information to be available in the user profile
// If no location is available, it returns nil (empty candidates)
func (g *LocalTopGenerator) Generate(user *models.UserProfile) []models.TrackID {
	// Check if user has location data
	// For now, we return nil as location is not stored in UserProfile
	// In a real implementation, you would either:
	// 1. Add LastKnownLat/LastKnownLon fields to UserProfile
	// 2. Pass location as part of recommendation request context
	// 3. Use a separate location service/store

	// This generator is primarily intended to be used via the GenerateWithLocation method
	// or when location data is added to the user profile
	return nil
}

// GenerateWithLocation returns top tracks near the specified location
func (g *LocalTopGenerator) GenerateWithLocation(lat, lon float64, radiusM int) []models.TrackID {
	if radiusM <= 0 {
		radiusM = g.defaultRadiusM
	}

	// Determine precision based on radius
	precision := store.GeohashPrecisionForRadius(radiusM)

	// Get geohashes covering the radius
	geohashes := store.ExpandNeighbors(lat, lon, radiusM, precision)

	// Get top tracks for these geohashes
	topTracks := g.geoTopStore.GetTopForGeohashes(geohashes, g.limit)

	// Convert to TrackID slice
	result := make([]models.TrackID, 0, len(topTracks))
	for _, tc := range topTracks {
		result = append(result, tc.TrackID)
	}

	return result
}
