package store

import (
	"sort"
	"sync"
)

// GlobalStatsStore defines the interface for global track statistics
type GlobalStatsStore interface {
	IncrementPlayCount(trackID string)
	GetPlayCount(trackID string) int
	GetTopTracks(limit int) []string
	GetTopTracksInSet(trackIDs []string, limit int) []string
	GetAllPlayCounts() map[string]int // For backup purposes
}

// InMemoryGlobalStatsStore is an in-memory implementation of GlobalStatsStore
type InMemoryGlobalStatsStore struct {
	mu         sync.RWMutex
	playCounts map[string]int
}

// NewInMemoryGlobalStatsStore creates a new in-memory global stats store
func NewInMemoryGlobalStatsStore() *InMemoryGlobalStatsStore {
	return &InMemoryGlobalStatsStore{
		playCounts: make(map[string]int),
	}
}

func (s *InMemoryGlobalStatsStore) IncrementPlayCount(trackID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playCounts[trackID]++
}

func (s *InMemoryGlobalStatsStore) GetPlayCount(trackID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.playCounts[trackID]
}

func (s *InMemoryGlobalStatsStore) GetTopTracks(limit int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type trackCount struct {
		trackID string
		count   int
	}

	counts := make([]trackCount, 0, len(s.playCounts))
	for trackID, count := range s.playCounts {
		counts = append(counts, trackCount{trackID, count})
	}

	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})

	result := make([]string, 0, limit)
	for i := 0; i < len(counts) && i < limit; i++ {
		result = append(result, counts[i].trackID)
	}

	return result
}

func (s *InMemoryGlobalStatsStore) GetTopTracksInSet(trackIDs []string, limit int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type trackCount struct {
		trackID string
		count   int
	}

	counts := make([]trackCount, 0, len(trackIDs))
	for _, trackID := range trackIDs {
		counts = append(counts, trackCount{trackID, s.playCounts[trackID]})
	}

	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})

	result := make([]string, 0, limit)
	for i := 0; i < len(counts) && i < limit; i++ {
		result = append(result, counts[i].trackID)
	}

	return result
}

// GetAllPlayCounts returns all track play counts for backup purposes
func (s *InMemoryGlobalStatsStore) GetAllPlayCounts() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resultCounts := make(map[string]int)
	for trackID, count := range s.playCounts {
		resultCounts[trackID] = count
	}
	return resultCounts
}
