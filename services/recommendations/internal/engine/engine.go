package engine

import (
	"sort"
	"time"

	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/store"
)

// RecommendationEngine orchestrates the recommendation pipeline
type RecommendationEngine struct {
	Generators []CandidateGenerator
	Filters    []Filter
	Scorer     Scorer
	ReRanker   ReRanker

	trackStore store.TrackStore
}

// NewRecommendationEngine creates a new recommendation engine
func NewRecommendationEngine(
	generators []CandidateGenerator,
	scorer Scorer,
	trackStore store.TrackStore,
) *RecommendationEngine {
	filters := []Filter{
		NewAlreadyListenedFilter(),
		NewExplicitFilter(trackStore),
		NewGenreFilter(trackStore),
		NewFreshnessFilter(trackStore),
	}

	return &RecommendationEngine{
		Generators: generators,
		Filters:    filters,
		Scorer:     scorer,
		ReRanker:   nil,
		trackStore: trackStore,
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
