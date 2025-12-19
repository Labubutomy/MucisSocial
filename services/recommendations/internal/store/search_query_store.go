package store

import (
	"sort"
	"strings"
	"sync"
	"time"
)

// SearchQueryStore defines the interface for search query storage
type SearchQueryStore interface {
	Increment(query string)
	GetTrendingQueries(limit int, days int) []TrendingQuery
	GetAll() map[string]SearchQueryData // For backup purposes
	Restore(data map[string]SearchQueryData) // For restore from backup
}

// SearchQueryData represents search query data for backup
type SearchQueryData struct {
	QueryCounts  map[string]int              `json:"query_counts"`
	QueryHistory map[string][]time.Time      `json:"query_history"`
	QueryMap     map[string]string           `json:"query_map"`
}

// TrendingQuery represents a trending search query
type TrendingQuery struct {
	Query string `json:"query"`
	Count int    `json:"count"`
}

// InMemorySearchQueryStore is an in-memory implementation of SearchQueryStore
type InMemorySearchQueryStore struct {
	mu           sync.RWMutex
	queryCounts  map[string]int              // normalized query -> count
	queryHistory map[string][]time.Time      // normalized query -> list of timestamps
	queryMap     map[string]string           // normalized query -> original query (for display)
}

// NewInMemorySearchQueryStore creates a new in-memory search query store
func NewInMemorySearchQueryStore() *InMemorySearchQueryStore {
	return &InMemorySearchQueryStore{
		queryCounts:  make(map[string]int),
		queryHistory: make(map[string][]time.Time),
		queryMap:     make(map[string]string),
	}
}

// Increment increments the count for a search query and records the timestamp
func (s *InMemorySearchQueryStore) Increment(query string) {
	if query == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Normalize query (trim and lowercase for consistency, but preserve all characters including Cyrillic)
	normalized := normalizeQuery(query)
	
	// Store original query for display (use the first one we see, or update if we see a longer version)
	original := strings.TrimSpace(query)
	// Limit to 100 runes (not bytes) to properly handle Unicode characters
	runes := []rune(original)
	if len(runes) > 100 {
		original = string(runes[:100])
	}
	if existing, ok := s.queryMap[normalized]; !ok || len(original) > len(existing) {
		s.queryMap[normalized] = original
	}

	s.queryCounts[normalized]++
	s.queryHistory[normalized] = append(s.queryHistory[normalized], time.Now())

	// Keep only last 1000 timestamps per query to prevent memory bloat
	if len(s.queryHistory[normalized]) > 1000 {
		s.queryHistory[normalized] = s.queryHistory[normalized][len(s.queryHistory[normalized])-1000:]
	}
}

// GetTrendingQueries returns the most popular search queries within the last N days
func (s *InMemorySearchQueryStore) GetTrendingQueries(limit int, days int) []TrendingQuery {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	cutoffTime := now.AddDate(0, 0, -days)

	// Count queries within the time window
	type queryCount struct {
		query string
		count int
	}

	queryCounts := make([]queryCount, 0)

	for query, timestamps := range s.queryHistory {
		count := 0
		for _, ts := range timestamps {
			if ts.After(cutoffTime) {
				count++
			}
		}
		if count > 0 {
			queryCounts = append(queryCounts, queryCount{
				query: query,
				count: count,
			})
		}
	}

	// Sort by count descending
	sort.Slice(queryCounts, func(i, j int) bool {
		return queryCounts[i].count > queryCounts[j].count
	})

	// Return top N with original query text
	result := make([]TrendingQuery, 0, limit)
	maxLen := limit
	if maxLen > len(queryCounts) {
		maxLen = len(queryCounts)
	}
	for i := 0; i < maxLen; i++ {
		normalizedQuery := queryCounts[i].query
		// Get original query for display, fallback to normalized if not found
		originalQuery := normalizedQuery
		if orig, ok := s.queryMap[normalizedQuery]; ok {
			originalQuery = orig
		}
		result = append(result, TrendingQuery{
			Query: originalQuery,
			Count: queryCounts[i].count,
		})
	}

	return result
}

// normalizeQuery normalizes a search query for consistent storage
// Preserves all Unicode characters including Cyrillic, but converts to lowercase
func normalizeQuery(query string) string {
	// Trim whitespace
	normalized := strings.TrimSpace(query)
	
	// Limit length to 100 runes (not bytes) to properly handle Unicode characters
	runes := []rune(normalized)
	if len(runes) > 100 {
		normalized = string(runes[:100])
	}
	
	// Convert to lowercase for case-insensitive matching
	// This works correctly for both Latin and Cyrillic characters
	normalized = strings.ToLower(normalized)
	
	return normalized
}

// GetAll returns all search query data for backup purposes
func (s *InMemorySearchQueryStore) GetAll() map[string]SearchQueryData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a deep copy of the data
	queryCountsCopy := make(map[string]int)
	for k, v := range s.queryCounts {
		queryCountsCopy[k] = v
	}

	queryHistoryCopy := make(map[string][]time.Time)
	for k, v := range s.queryHistory {
		timestamps := make([]time.Time, len(v))
		copy(timestamps, v)
		queryHistoryCopy[k] = timestamps
	}

	queryMapCopy := make(map[string]string)
	for k, v := range s.queryMap {
		queryMapCopy[k] = v
	}

	return map[string]SearchQueryData{
		"data": {
			QueryCounts:  queryCountsCopy,
			QueryHistory: queryHistoryCopy,
			QueryMap:     queryMapCopy,
		},
	}
}

// Restore restores search query data from backup
func (s *InMemorySearchQueryStore) Restore(data map[string]SearchQueryData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if searchData, ok := data["data"]; ok {
		s.queryCounts = searchData.QueryCounts
		s.queryHistory = searchData.QueryHistory
		s.queryMap = searchData.QueryMap
	}
}

