package engine

import (
	"sort"
	"time"

	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/engine/scorers"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/store"
)

// RecommendationEngine orchestrates the recommendation pipeline
type RecommendationEngine struct {
	Generators  []CandidateGenerator
	Filters     []Filter
	Scorer      Scorer
	ReRanker    ReRanker
	GroupScorer *scorers.GroupScorer

	trackStore store.TrackStore
}

// NewRecommendationEngine creates a new recommendation engine
func NewRecommendationEngine(
	generators []CandidateGenerator,
	scorer Scorer,
	trackStore store.TrackStore,
	statsStore store.GlobalStatsStore,
) *RecommendationEngine {
	filters := []Filter{
		NewAlreadyListenedFilter(),
		NewExplicitFilter(trackStore),
		NewGenreFilter(trackStore),
		NewFreshnessFilter(trackStore),
	}

	groupScorer := scorers.NewGroupScorer(trackStore, statsStore)

	return &RecommendationEngine{
		Generators:  generators,
		Filters:     filters,
		Scorer:      scorer,
		ReRanker:    nil,
		GroupScorer: groupScorer,
		trackStore:  trackStore,
	}
}

// RecommendRequest represents an API recommendation request
type RecommendRequest struct {
	UserID  string
	Limit   int
	Filters *FilterOptions
}

// Recommend executes the full recommendation pipeline
func (e *RecommendationEngine) Recommend(user *models.UserProfile, req *RecommendRequest) []models.TrackID {
	// Step 1: Candidate Generation
	candidateSet := make(map[models.TrackID]struct{})
	for _, gen := range e.Generators {
		candidates := gen.Generate(user)
		for _, trackID := range candidates {
			candidateSet[trackID] = struct{}{}
		}
	}

	candidates := make([]models.TrackID, 0, len(candidateSet))
	for trackID := range candidateSet {
		candidates = append(candidates, trackID)
	}

	// Step 2: Hard Filtering
	for _, filter := range e.Filters {
		candidates = filter.Apply(user, candidates, req.Filters)
	}

	if len(candidates) == 0 {
		return []models.TrackID{}
	}

	// Step 3: Scoring
	scores := e.Scorer.Score(user, candidates)

	// Step 4: Ranking
	var ranked []models.TrackID
	if e.ReRanker != nil {
		ranked = e.ReRanker.ReRank(user, scores)
	} else {
		ranked = sortByScore(scores)
	}

	// Step 5: Apply limit
	if len(ranked) > req.Limit {
		ranked = ranked[:req.Limit]
	}

	return ranked
}

// GroupRecommendRequest represents a group recommendation request
type GroupRecommendRequest struct {
	UserIDs []string
	Limit   int
	Filters *FilterOptions
}

// RecommendForGroup generates recommendations for a group of users
// Creates a merged profile and generates recommendations that appeal to the majority
func (e *RecommendationEngine) RecommendForGroup(
	profiles []*models.UserProfile,
	req *GroupRecommendRequest,
) []models.TrackID {
	if len(profiles) == 0 {
		return []models.TrackID{}
	}

	// Step 1: Aggregate user profiles into a group profile
	groupProfile := models.AggregateProfiles(profiles)

	// Step 2: Generate candidates using the group's aggregated preferences
	// We'll use a synthetic user profile based on the group's top preferences
	syntheticProfile := e.createSyntheticProfileFromGroup(groupProfile)

	candidateSet := make(map[models.TrackID]struct{})
	for _, gen := range e.Generators {
		candidates := gen.Generate(syntheticProfile)
		for _, trackID := range candidates {
			candidateSet[trackID] = struct{}{}
		}
	}

	candidates := make([]models.TrackID, 0, len(candidateSet))
	for trackID := range candidateSet {
		candidates = append(candidates, trackID)
	}

	// Step 3: Filter out tracks that majority has already heard
	candidates = e.filterGroupListenedTracks(groupProfile, candidates)

	// Apply other filters (explicit, genre, freshness)
	for _, filter := range e.Filters {
		// Skip the already listened filter for groups
		if filter.Name() == "already_listened" {
			continue
		}
		candidates = filter.Apply(syntheticProfile, candidates, req.Filters)
	}

	if len(candidates) == 0 {
		return []models.TrackID{}
	}

	// Step 4: Score using group scorer
	scores := e.GroupScorer.ScoreForGroup(groupProfile, candidates)

	// Step 5: Rank by score
	ranked := sortByScore(scores)

	// Step 6: Apply limit
	if len(ranked) > req.Limit {
		ranked = ranked[:req.Limit]
	}

	return ranked
}

// createSyntheticProfileFromGroup creates a user profile representing group preferences
func (e *RecommendationEngine) createSyntheticProfileFromGroup(group *models.GroupProfile) *models.UserProfile {
	profile := &models.UserProfile{
		UserID:            "group_" + group.UserIDs[0], // synthetic ID
		GenreListenCount:  make(map[string]int),
		ArtistListenCount: make(map[string]int),
		ListenedTracks:    make(map[string]struct{}),
	}

	// Use normalized counts from the group
	for genre, count := range group.GenreListenCount {
		profile.GenreListenCount[genre] = count
	}

	for artist, count := range group.ArtistListenCount {
		profile.ArtistListenCount[artist] = count
	}

	return profile
}

// filterGroupListenedTracks removes tracks that majority of the group has heard
func (e *RecommendationEngine) filterGroupListenedTracks(
	group *models.GroupProfile,
	tracks []models.TrackID,
) []models.TrackID {
	result := make([]models.TrackID, 0, len(tracks))
	for _, trackID := range tracks {
		if !group.HasListenedByMajority(trackID) {
			result = append(result, trackID)
		}
	}
	return result
}

func sortByScore(scores map[models.TrackID]float64) []models.TrackID {
	type trackScore struct {
		trackID models.TrackID
		score   float64
	}

	ts := make([]trackScore, 0, len(scores))
	for trackID, score := range scores {
		ts = append(ts, trackScore{trackID, score})
	}

	sort.Slice(ts, func(i, j int) bool {
		return ts[i].score > ts[j].score
	})

	result := make([]models.TrackID, len(ts))
	for i, t := range ts {
		result[i] = t.trackID
	}

	return result
}

// AlreadyListenedFilter removes tracks the user has already listened to
type AlreadyListenedFilter struct{}

func NewAlreadyListenedFilter() *AlreadyListenedFilter {
	return &AlreadyListenedFilter{}
}

func (f *AlreadyListenedFilter) Name() string {
	return "already_listened"
}

func (f *AlreadyListenedFilter) Apply(user *models.UserProfile, tracks []models.TrackID, opts *FilterOptions) []models.TrackID {
	result := make([]models.TrackID, 0, len(tracks))
	for _, trackID := range tracks {
		if !user.HasListened(trackID) {
			result = append(result, trackID)
		}
	}
	return result
}

// ExplicitFilter removes explicit tracks if filter is enabled
type ExplicitFilter struct {
	trackStore store.TrackStore
}

func NewExplicitFilter(trackStore store.TrackStore) *ExplicitFilter {
	return &ExplicitFilter{trackStore: trackStore}
}

func (f *ExplicitFilter) Name() string {
	return "explicit"
}

func (f *ExplicitFilter) Apply(user *models.UserProfile, tracks []models.TrackID, opts *FilterOptions) []models.TrackID {
	if opts == nil || !opts.ExcludeExplicit {
		return tracks
	}

	result := make([]models.TrackID, 0, len(tracks))
	for _, trackID := range tracks {
		track, ok := f.trackStore.Get(trackID)
		if !ok || !track.IsExplicit {
			result = append(result, trackID)
		}
	}
	return result
}

// GenreFilter keeps only tracks matching specified genres
type GenreFilter struct {
	trackStore store.TrackStore
}

func NewGenreFilter(trackStore store.TrackStore) *GenreFilter {
	return &GenreFilter{trackStore: trackStore}
}

func (f *GenreFilter) Name() string {
	return "genre"
}

func (f *GenreFilter) Apply(user *models.UserProfile, tracks []models.TrackID, opts *FilterOptions) []models.TrackID {
	if opts == nil || len(opts.Genres) == 0 {
		return tracks
	}

	genreSet := make(map[string]struct{})
	for _, g := range opts.Genres {
		genreSet[g] = struct{}{}
	}

	result := make([]models.TrackID, 0, len(tracks))
	for _, trackID := range tracks {
		track, ok := f.trackStore.Get(trackID)
		if !ok {
			continue
		}
		for _, genre := range track.Genres {
			if _, ok := genreSet[genre]; ok {
				result = append(result, trackID)
				break
			}
		}
	}
	return result
}

// FreshnessFilter keeps only tracks released within N days
type FreshnessFilter struct {
	trackStore store.TrackStore
}

func NewFreshnessFilter(trackStore store.TrackStore) *FreshnessFilter {
	return &FreshnessFilter{trackStore: trackStore}
}

func (f *FreshnessFilter) Name() string {
	return "freshness"
}

func (f *FreshnessFilter) Apply(user *models.UserProfile, tracks []models.TrackID, opts *FilterOptions) []models.TrackID {
	if opts == nil || opts.FreshDays <= 0 {
		return tracks
	}

	cutoff := time.Now().Unix() - int64(opts.FreshDays*24*60*60)

	result := make([]models.TrackID, 0, len(tracks))
	for _, trackID := range tracks {
		track, ok := f.trackStore.Get(trackID)
		if !ok {
			continue
		}
		if track.ReleaseTS >= cutoff {
			result = append(result, trackID)
		}
	}
	return result
}
