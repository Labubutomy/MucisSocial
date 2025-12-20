package store

import (
	"sort"
	"sync"
	"time"

	"github.com/mmcloughlin/geohash"
)

// TrackCount represents a track and its play count
type TrackCount struct {
	TrackID string `json:"track_id"`
	Count   int64  `json:"count"`
}

// GeoTopStore manages geolocation-based track popularity aggregates
type GeoTopStore interface {
	// Incr increments the play count for a track in a specific geohash cell
	Incr(geohash string, trackID string, delta int64)

	// GetTopForGeohashes aggregates and returns top tracks across multiple geohash cells
	GetTopForGeohashes(geohashes []string, limit int) []TrackCount

	// Snapshot creates a serializable snapshot of the store
	Snapshot() *GeoTopSnapshot

	// Restore restores the store from a snapshot
	Restore(snapshot *GeoTopSnapshot)

	// GetAllGeohashes returns all geohash cells that have data
	GetAllGeohashes() []string
}

// GeoTopSnapshot represents a serializable snapshot of geo-based aggregates
type GeoTopSnapshot struct {
	Counts    map[string]map[string]int64 `json:"counts"` // map[geohash]map[trackID]count
	Timestamp time.Time                   `json:"timestamp"`
}

// InMemoryGeoTopStore is an in-memory implementation of GeoTopStore
type InMemoryGeoTopStore struct {
	mu           sync.RWMutex
	counts       map[string]map[string]int64 // map[geohash]map[trackID]count
	maxTopPerGeo int                         // Maximum number of top tracks to keep per geohash

	// Cache for query results
	queryCache map[string]*cachedTopResult
	cacheTTL   time.Duration
	cacheMu    sync.RWMutex
}

// cachedTopResult represents a cached query result
type cachedTopResult struct {
	tracks    []TrackCount
	expiresAt time.Time
}

// NewInMemoryGeoTopStore creates a new in-memory geo top store
func NewInMemoryGeoTopStore(maxTopPerGeo int, cacheTTL time.Duration) *InMemoryGeoTopStore {
	return &InMemoryGeoTopStore{
		counts:       make(map[string]map[string]int64),
		maxTopPerGeo: maxTopPerGeo,
		queryCache:   make(map[string]*cachedTopResult),
		cacheTTL:     cacheTTL,
	}
}

// Incr increments the play count for a track in a specific geohash cell
func (s *InMemoryGeoTopStore) Incr(ghash string, trackID string, delta int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Initialize geohash map if needed
	if s.counts[ghash] == nil {
		s.counts[ghash] = make(map[string]int64)
	}

	// Increment count
	s.counts[ghash][trackID] += delta

	// Invalidate cache for this geohash
	s.invalidateCacheForGeohash(ghash)
}

// GetTopForGeohashes aggregates and returns top tracks across multiple geohash cells
func (s *InMemoryGeoTopStore) GetTopForGeohashes(geohashes []string, limit int) []TrackCount {
	// Try cache first
	cacheKey := s.buildCacheKey(geohashes, limit)
	if cached := s.getCached(cacheKey); cached != nil {
		return cached
	}

	s.mu.RLock()

	// Aggregate counts across all geohashes
	aggregated := make(map[string]int64)
	for _, ghash := range geohashes {
		if trackCounts, exists := s.counts[ghash]; exists {
			for trackID, count := range trackCounts {
				aggregated[trackID] += count
			}
		}
	}
	s.mu.RUnlock()

	// Convert to slice and sort
	result := make([]TrackCount, 0, len(aggregated))
	for trackID, count := range aggregated {
		result = append(result, TrackCount{
			TrackID: trackID,
			Count:   count,
		})
	}

	// Sort by count descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	// Limit results
	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}

	// Cache the result
	s.cacheResult(cacheKey, result)

	return result
}

// Snapshot creates a serializable snapshot of the store
func (s *InMemoryGeoTopStore) Snapshot() *GeoTopSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Deep copy the counts map
	countsCopy := make(map[string]map[string]int64)
	for ghash, trackCounts := range s.counts {
		countsCopy[ghash] = make(map[string]int64)
		for trackID, count := range trackCounts {
			countsCopy[ghash][trackID] = count
		}
	}

	return &GeoTopSnapshot{
		Counts:    countsCopy,
		Timestamp: time.Now(),
	}
}

// Restore restores the store from a snapshot
func (s *InMemoryGeoTopStore) Restore(snapshot *GeoTopSnapshot) {
	if snapshot == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Deep copy from snapshot
	s.counts = make(map[string]map[string]int64)
	for ghash, trackCounts := range snapshot.Counts {
		s.counts[ghash] = make(map[string]int64)
		for trackID, count := range trackCounts {
			s.counts[ghash][trackID] = count
		}
	}

	// Clear cache after restore
	s.cacheMu.Lock()
	s.queryCache = make(map[string]*cachedTopResult)
	s.cacheMu.Unlock()
}

// GetAllGeohashes returns all geohash cells that have data
func (s *InMemoryGeoTopStore) GetAllGeohashes() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	geohashes := make([]string, 0, len(s.counts))
	for ghash := range s.counts {
		geohashes = append(geohashes, ghash)
	}
	return geohashes
}

// buildCacheKey creates a cache key from geohashes and limit
func (s *InMemoryGeoTopStore) buildCacheKey(geohashes []string, limit int) string {
	// Sort geohashes for consistent cache key
	sorted := make([]string, len(geohashes))
	copy(sorted, geohashes)
	sort.Strings(sorted)

	key := ""
	for _, gh := range sorted {
		key += gh + ":"
	}
	return key + string(rune(limit))
}

// getCached retrieves a cached result if it exists and hasn't expired
func (s *InMemoryGeoTopStore) getCached(cacheKey string) []TrackCount {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	cached, exists := s.queryCache[cacheKey]
	if !exists {
		return nil
	}

	if time.Now().After(cached.expiresAt) {
		return nil
	}

	return cached.tracks
}

// cacheResult stores a result in the cache
func (s *InMemoryGeoTopStore) cacheResult(cacheKey string, tracks []TrackCount) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.queryCache[cacheKey] = &cachedTopResult{
		tracks:    tracks,
		expiresAt: time.Now().Add(s.cacheTTL),
	}
}

// invalidateCacheForGeohash invalidates all cache entries that contain the given geohash
func (s *InMemoryGeoTopStore) invalidateCacheForGeohash(ghash string) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	// Simple approach: invalidate all cache entries
	// More sophisticated approach would track which geohashes are in each cache key
	s.queryCache = make(map[string]*cachedTopResult)
}

// GeohashPrecisionForRadius returns the appropriate geohash precision for a given radius in meters
func GeohashPrecisionForRadius(radiusM int) int {
	// Geohash precision table (approximate cell widths):
	// precision 1: ~5000 km
	// precision 2: ~1250 km
	// precision 3: ~156 km
	// precision 4: ~39.1 km
	// precision 5: ~4.89 km
	// precision 6: ~1.22 km
	// precision 7: ~153 m
	// precision 8: ~38.2 m
	// precision 9: ~4.77 m
	// precision 10: ~1.19 m

	// We want cell width to be <= radius * 2 to properly cover the area
	diameter := radiusM * 2

	if diameter >= 5000000 {
		return 1
	} else if diameter >= 1250000 {
		return 2
	} else if diameter >= 156000 {
		return 3
	} else if diameter >= 39100 {
		return 4
	} else if diameter >= 4890 {
		return 5
	} else if diameter >= 1220 {
		return 6
	} else if diameter >= 153 {
		return 7
	} else if diameter >= 38 {
		return 8
	} else if diameter >= 5 {
		return 9
	}
	return 10
}

// ExpandNeighbors returns a list of geohashes that cover a circular area
// centered at (lat, lon) with the given radius in meters
func ExpandNeighbors(lat, lon float64, radiusM int, precision int) []string {
	// Encode center point
	center := geohash.EncodeWithPrecision(lat, lon, uint(precision))

	// Approximate cell width at this precision
	cellWidths := map[int]int{
		1: 5000000, 2: 1250000, 3: 156000, 4: 39100, 5: 4890,
		6: 1220, 7: 153, 8: 38, 9: 5, 10: 1,
	}

	cellWidth := cellWidths[precision]
	if cellWidth == 0 {
		cellWidth = 1000 // default fallback
	}

	// Calculate how many cells we need in each direction
	// We use a square grid that covers the circle
	k := (radiusM / cellWidth) + 1

	// Start with center
	result := make(map[string]bool)
	result[center] = true

	// BFS to expand neighbors
	queue := []string{center}
	visited := make(map[string]bool)
	visited[center] = true

	for depth := 0; depth < k; depth++ {
		levelSize := len(queue)
		for i := 0; i < levelSize; i++ {
			current := queue[0]
			queue = queue[1:]

			// Get all 8 neighbors
			neighbors := geohash.Neighbors(current)
			for _, neighbor := range neighbors {
				if !visited[neighbor] {
					visited[neighbor] = true
					result[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}
	}

	// Convert to slice
	geohashes := make([]string, 0, len(result))
	for gh := range result {
		geohashes = append(geohashes, gh)
	}

	return geohashes
}

// EncodeGeohash encodes latitude and longitude into a geohash string with the given precision
func EncodeGeohash(lat, lon float64, precision int) string {
	return geohash.EncodeWithPrecision(lat, lon, uint(precision))
}
