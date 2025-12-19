package generators

import (
	"sort"

	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/store"
)

// TopGenresGenerator generates candidates based on user's top listened genres
type TopGenresGenerator struct {
	trackStore     store.TrackStore
	statsStore     store.GlobalStatsStore
	topN           int
	tracksPerGenre int
}

// NewTopGenresGenerator creates a new top genres candidate generator
func NewTopGenresGenerator(
	trackStore store.TrackStore,
	statsStore store.GlobalStatsStore,
	topN int,
	tracksPerGenre int,
) *TopGenresGenerator {
	return &TopGenresGenerator{
		trackStore:     trackStore,
		statsStore:     statsStore,
		topN:           topN,
		tracksPerGenre: tracksPerGenre,
	}
}

func (g *TopGenresGenerator) Name() string {
	return "top_genres"
}

func (g *TopGenresGenerator) Generate(user *models.UserProfile) []models.TrackID {
	topGenres := g.getTopGenres(user)

	candidates := make([]models.TrackID, 0)

	for _, genre := range topGenres {
		genreTracks := g.trackStore.GetByGenre(genre)
		topTracks := g.statsStore.GetTopTracksInSet(genreTracks, g.tracksPerGenre)
		candidates = append(candidates, topTracks...)
	}

	return candidates
}

func (g *TopGenresGenerator) getTopGenres(user *models.UserProfile) []string {
	type genreCount struct {
		genre string
		count int
	}

	counts := make([]genreCount, 0, len(user.GenreListenCount))
	for genre, count := range user.GenreListenCount {
		counts = append(counts, genreCount{genre, count})
	}

	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})

	result := make([]string, 0, g.topN)
	for i := 0; i < len(counts) && i < g.topN; i++ {
		result = append(result, counts[i].genre)
	}

	return result
}
