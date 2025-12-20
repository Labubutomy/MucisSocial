package scorers

import (
	"math"
	"time"

	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/store"
)

// GroupScorer implements scoring for group recommendations
// Balances preferences across multiple users in a room
type GroupScorer struct {
	trackStore store.TrackStore
	statsStore store.GlobalStatsStore
}

// NewGroupScorer creates a new group scorer
func NewGroupScorer(trackStore store.TrackStore, statsStore store.GlobalStatsStore) *GroupScorer {
	return &GroupScorer{
		trackStore: trackStore,
		statsStore: statsStore,
	}
}

func (s *GroupScorer) Name() string {
	return "group_v1"
}

// ScoreForGroup calculates scores for tracks based on group preferences
// Prioritizes tracks that appeal to the majority while avoiding niche preferences
func (s *GroupScorer) ScoreForGroup(group *models.GroupProfile, tracks []models.TrackID) map[models.TrackID]float64 {
	scores := make(map[models.TrackID]float64, len(tracks))
	now := time.Now().Unix()

	for _, trackID := range tracks {
		track, ok := s.trackStore.Get(trackID)
		if !ok {
			scores[trackID] = 0
			continue
		}

		score := 0.0

		// Genre affinity - weighted by how many users like this genre
		score += s.calcGroupGenreScore(group, track) * 3.0

		// Artist affinity - weighted by group popularity
		score += s.calcGroupArtistScore(group, track) * 2.5

		// Global popularity - helps find broadly appealing tracks
		score += s.calcPopularityScore(trackID) * 1.5

		// Freshness bonus - new tracks can be exciting for groups
		score += s.calcFreshnessBonus(track, now) * 0.5

		// Diversity bonus - reward tracks that are fresh for the group
		score += s.calcDiversityBonus(group, trackID) * 2.0

		scores[trackID] = score
	}

	return scores
}

// calcGroupGenreScore calculates genre affinity for the entire group
// Normalizes by total users to avoid bias towards larger groups
func (s *GroupScorer) calcGroupGenreScore(group *models.GroupProfile, track *models.Track) float64 {
	if group.TotalUsers == 0 || len(track.Genres) == 0 {
		return 0
	}

	score := 0.0
	for _, genre := range track.Genres {
		if count, ok := group.GenreListenCount[genre]; ok {
			// Normalize by total users and number of genres
			score += float64(count) / float64(group.TotalUsers)
		}
	}

	// Average across all genres in the track
	return score / float64(len(track.Genres))
}

// calcGroupArtistScore calculates artist affinity for the group
func (s *GroupScorer) calcGroupArtistScore(group *models.GroupProfile, track *models.Track) float64 {
	if group.TotalUsers == 0 {
		return 0
	}

	if count, ok := group.ArtistListenCount[track.ArtistID]; ok {
		// Normalize by total users
		return float64(count) / float64(group.TotalUsers)
	}
	return 0
}

// calcPopularityScore uses global stats to find broadly appealing tracks
func (s *GroupScorer) calcPopularityScore(trackID string) float64 {
	playCount := s.statsStore.GetPlayCount(trackID)
	// Log scale to prevent extremely popular tracks from dominating
	return math.Log(float64(playCount) + 1)
}

// calcFreshnessBonus rewards newer tracks
func (s *GroupScorer) calcFreshnessBonus(track *models.Track, now int64) float64 {
	daysSinceRelease := float64(now-track.ReleaseTS) / (24 * 60 * 60)
	if daysSinceRelease < 0 {
		daysSinceRelease = 0
	}
	// Decay over 30 days
	return 1.0 / (daysSinceRelease/30.0 + 1)
}

// calcDiversityBonus rewards tracks that fewer members have heard
// This ensures variety and discovery for the group
func (s *GroupScorer) calcDiversityBonus(group *models.GroupProfile, trackID string) float64 {
	if group.TotalUsers == 0 {
		return 0
	}

	listenCount, exists := group.ListenedTracks[trackID]
	if !exists {
		listenCount = 0
	}

	// Percentage of group that hasn't heard this track
	unheardPercentage := float64(group.TotalUsers-listenCount) / float64(group.TotalUsers)

	// Bonus increases as more people haven't heard it
	// But we still want some familiarity, so we use a moderate bonus
	return unheardPercentage * 2.0
}
