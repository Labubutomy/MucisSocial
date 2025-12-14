package generators

import (
	"sort"

	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/store"
)

// TopArtistsGenerator generates candidates based on user's top listened artists
type TopArtistsGenerator struct {
	trackStore      store.TrackStore
	statsStore      store.GlobalStatsStore
	topN            int
	tracksPerArtist int
}

// NewTopArtistsGenerator creates a new top artists candidate generator
func NewTopArtistsGenerator(
	trackStore store.TrackStore,
	statsStore store.GlobalStatsStore,
	topN int,
	tracksPerArtist int,
) *TopArtistsGenerator {
	return &TopArtistsGenerator{
		trackStore:      trackStore,
		statsStore:      statsStore,
		topN:            topN,
		tracksPerArtist: tracksPerArtist,
	}
}

func (g *TopArtistsGenerator) Name() string {
	return "top_artists"
}

func (g *TopArtistsGenerator) Generate(user *models.UserProfile) []models.TrackID {
	topArtists := g.getTopArtists(user)

	candidates := make([]models.TrackID, 0)

	for _, artistID := range topArtists {
		artistTracks := g.trackStore.GetByArtist(artistID)
		topTracks := g.statsStore.GetTopTracksInSet(artistTracks, g.tracksPerArtist)
		candidates = append(candidates, topTracks...)
	}

	return candidates
}

func (g *TopArtistsGenerator) getTopArtists(user *models.UserProfile) []string {
	type artistCount struct {
		artistID string
		count    int
	}

	counts := make([]artistCount, 0, len(user.ArtistListenCount))
	for artistID, count := range user.ArtistListenCount {
		counts = append(counts, artistCount{artistID, count})
	}

	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})

	result := make([]string, 0, g.topN)
	for i := 0; i < len(counts) && i < g.topN; i++ {
		result = append(result, counts[i].artistID)
	}

	return result
}
