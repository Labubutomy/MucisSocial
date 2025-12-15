package store

import (
	"sort"
	"sync"
	"time"
)

// GlobalStatsStore defines the interface for global track statistics
type GlobalStatsStore interface {
	IncrementPlayCount(trackID string)
	GetPlayCount(trackID string) int
	GetTopTracks(limit int) []string
	GetTopTracksInSet(trackIDs []string, limit int) []string
	GetAllPlayCounts() map[string]int // For backup purposes
	GetTopTracksForMonth(limit int) []string
	GetMonthlyPlayCount(trackID string) int
	GetAllMonthlyPlayCounts() map[string]int // For backup purposes
}

// InMemoryGlobalStatsStore is an in-memory implementation of GlobalStatsStore
type InMemoryGlobalStatsStore struct {
	mu              sync.RWMutex
	playCounts      map[string]int // All time
	monthlyCounts   map[string]int  // Current month
	monthlyCache    *monthlyChartCache
}

// monthlyChartCache caches the top tracks for the current month
type monthlyChartCache struct {
	tracks    []string
	monthKey  string // Format: "2025-12" (YYYY-MM)
	expiresAt time.Time
	mu        sync.RWMutex
}

func newMonthlyChartCache() *monthlyChartCache {
	return &monthlyChartCache{
		tracks:    nil,
		monthKey:  "",
		expiresAt: time.Time{},
	}
}

// getMonthKey returns the current month key in format "YYYY-MM"
func getMonthKey(t time.Time) string {
	return t.Format("2006-01")
}

// getCurrentMonthKey returns the current month key
func getCurrentMonthKey() string {
	return getMonthKey(time.Now())
}

// NewInMemoryGlobalStatsStore creates a new in-memory global stats store
func NewInMemoryGlobalStatsStore() *InMemoryGlobalStatsStore {
	return &InMemoryGlobalStatsStore{
		playCounts:    make(map[string]int),
		monthlyCounts: make(map[string]int),
		monthlyCache:  newMonthlyChartCache(),
	}
}

func (s *InMemoryGlobalStatsStore) IncrementPlayCount(trackID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Increment all-time count
	s.playCounts[trackID]++
	
	// Increment monthly count
	currentMonth := getCurrentMonthKey()
	s.monthlyCounts[trackID]++
	
	// Always invalidate cache when stats change to ensure fresh data
	s.monthlyCache.mu.Lock()
	s.monthlyCache.tracks = nil
	s.monthlyCache.monthKey = currentMonth
	s.monthlyCache.expiresAt = time.Time{}
	s.monthlyCache.mu.Unlock()
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

// GetTopTracksForMonth returns top tracks for the current month
// Results are cached for 1 minute to avoid constant recalculation
func (s *InMemoryGlobalStatsStore) GetTopTracksForMonth(limit int) []string {
	currentMonth := getCurrentMonthKey()
	now := time.Now()
	
	// Check cache first
	s.monthlyCache.mu.RLock()
	if s.monthlyCache.monthKey == currentMonth &&
		s.monthlyCache.tracks != nil &&
		!s.monthlyCache.expiresAt.IsZero() &&
		now.Before(s.monthlyCache.expiresAt) {
		result := make([]string, 0, len(s.monthlyCache.tracks))
		if limit > len(s.monthlyCache.tracks) {
			limit = len(s.monthlyCache.tracks)
		}
		result = append(result, s.monthlyCache.tracks[:limit]...)
		s.monthlyCache.mu.RUnlock()
		return result
	}
	s.monthlyCache.mu.RUnlock()
	
	// Cache miss or expired - recalculate
	s.mu.RLock()
	
	type trackCount struct {
		trackID string
		count   int
	}
	
	counts := make([]trackCount, 0, len(s.monthlyCounts))
	for trackID, count := range s.monthlyCounts {
		counts = append(counts, trackCount{trackID, count})
	}
	
	s.mu.RUnlock()
	
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})
	
	result := make([]string, 0, limit)
	maxLen := limit
	if maxLen > len(counts) {
		maxLen = len(counts)
	}
	for i := 0; i < maxLen; i++ {
		result = append(result, counts[i].trackID)
	}
	
	// Update cache (store full result, limit will be applied on read)
	s.monthlyCache.mu.Lock()
	s.monthlyCache.tracks = make([]string, len(result))
	copy(s.monthlyCache.tracks, result)
	s.monthlyCache.monthKey = currentMonth
	s.monthlyCache.expiresAt = now.Add(1 * time.Minute) // Cache for 1 minute
	s.monthlyCache.mu.Unlock()
	
	return result
}

// GetMonthlyPlayCount returns the play count for a track in the current month
func (s *InMemoryGlobalStatsStore) GetMonthlyPlayCount(trackID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.monthlyCounts[trackID]
}

// GetAllMonthlyPlayCounts returns all monthly play counts for backup purposes
func (s *InMemoryGlobalStatsStore) GetAllMonthlyPlayCounts() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resultCounts := make(map[string]int)
	for trackID, count := range s.monthlyCounts {
		resultCounts[trackID] = count
	}
	return resultCounts
}
