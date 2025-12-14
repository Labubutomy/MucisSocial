package scorers

import (
	"math"
	"time"

	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/store"
)

// HeuristicScorer implements heuristic-based scoring
// PRIMARY ML REPLACEMENT POINT
type HeuristicScorer struct {
	trackStore store.TrackStore
	statsStore store.GlobalStatsStore
}

// NewHeuristicScorer creates a new heuristic scorer
func NewHeuristicScorer(trackStore store.TrackStore, statsStore store.GlobalStatsStore) *HeuristicScorer {
	return &HeuristicScorer{
		trackStore: trackStore,
		statsStore: statsStore,
	}
}

func (s *HeuristicScorer) Name() string {
	return "heuristic_v1"
}

// Score calculates heuristic scores for each track
func (s *HeuristicScorer) Score(user *models.UserProfile, tracks []models.TrackID) map[models.TrackID]float64 {
	scores := make(map[models.TrackID]float64, len(tracks))
	now := time.Now().Unix()

	for _, trackID := range tracks {
		track, ok := s.trackStore.Get(trackID)
		if !ok {
			scores[trackID] = 0
			continue
		}

		score := 0.0
		score += s.calcGenreAffinityScore(user, track)
		score += s.calcArtistAffinityScore(user, track)
		score += s.calcPopularityScore(trackID)
		score += s.calcFreshnessBonus(track, now)

		scores[trackID] = score
	}

	return scores
}

func (s *HeuristicScorer) calcGenreAffinityScore(user *models.UserProfile, track *models.Track) float64 {
	score := 0.0
	for _, genre := range track.Genres {
		if count, ok := user.GenreListenCount[genre]; ok {
			score += float64(count)
		}
	}
	if len(track.Genres) > 0 {
		score = score / float64(len(track.Genres))
	}
	return score
}

func (s *HeuristicScorer) calcArtistAffinityScore(user *models.UserProfile, track *models.Track) float64 {
	if count, ok := user.ArtistListenCount[track.ArtistID]; ok {
		return float64(count)
	}
	return 0
}

func (s *HeuristicScorer) calcPopularityScore(trackID string) float64 {
	playCount := s.statsStore.GetPlayCount(trackID)
	return math.Log(float64(playCount) + 1)
}

func (s *HeuristicScorer) calcFreshnessBonus(track *models.Track, now int64) float64 {
	daysSinceRelease := float64(now-track.ReleaseTS) / (24 * 60 * 60)
	if daysSinceRelease < 0 {
		daysSinceRelease = 0
	}
	return 1.0 / (daysSinceRelease + 1)
}
