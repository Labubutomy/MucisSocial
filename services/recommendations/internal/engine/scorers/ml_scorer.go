package scorers

import (
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
)

// MLScorer is a placeholder for future ML-based scoring
// INTEGRATION POINTS:
// 1. Replace HeuristicScorer with MLScorer in main.go
// 2. MLScorer loads a trained model on startup
// 3. Score() method calls model inference
type MLScorer struct {
	modelPath string
}

// NewMLScorer creates a new ML scorer
func NewMLScorer(modelPath string) *MLScorer {
	return &MLScorer{modelPath: modelPath}
}

func (s *MLScorer) Name() string {
	return "ml_ranker_v1"
}

// Score calculates ML-based scores for each track
func (s *MLScorer) Score(user *models.UserProfile, tracks []models.TrackID) map[models.TrackID]float64 {
	scores := make(map[models.TrackID]float64, len(tracks))
	for _, trackID := range tracks {
		scores[trackID] = 0
	}
	return scores
}
