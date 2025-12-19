package engine

import (
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
)

// CandidateGenerator generates candidate tracks for recommendation
// ML-ready: Can be replaced with embedding-based retrieval
type CandidateGenerator interface {
	Generate(user *models.UserProfile) []models.TrackID
	Name() string
}

// Filter applies hard filtering rules to remove ineligible tracks
// NOT replaced by ML - enforces business rules
type Filter interface {
	Apply(user *models.UserProfile, tracks []models.TrackID, opts *FilterOptions) []models.TrackID
	Name() string
}

// FilterOptions contains filter configuration from the API request
type FilterOptions struct {
	ExcludeExplicit bool
	Genres          []string
	FreshDays       int
}

// Scorer assigns a score to each candidate track
// PRIMARY ML INTEGRATION POINT: Replace with ML ranker
type Scorer interface {
	Score(user *models.UserProfile, tracks []models.TrackID) map[models.TrackID]float64
	Name() string
}

// ReRanker is an optional post-scoring re-ranking step
// Reserved for future ML integration
type ReRanker interface {
	ReRank(user *models.UserProfile, scoredTracks map[models.TrackID]float64) []models.TrackID
	Name() string
}
