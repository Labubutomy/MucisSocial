package store

import (
	"sync"

	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
)

// TrackStore defines the interface for track storage
type TrackStore interface {
	Get(trackID string) (*models.Track, bool)
	Upsert(track *models.Track)
	GetByGenre(genre string) []string
	GetByArtist(artistID string) []string
	GetAll() []*models.Track
}

// InMemoryTrackStore is an in-memory implementation of TrackStore
type InMemoryTrackStore struct {
	mu          sync.RWMutex
	tracks      map[string]*models.Track
	genreIndex  map[string]map[string]struct{}
	artistIndex map[string]map[string]struct{}
}

// NewInMemoryTrackStore creates a new in-memory track store
func NewInMemoryTrackStore() *InMemoryTrackStore {
	return &InMemoryTrackStore{
		tracks:      make(map[string]*models.Track),
		genreIndex:  make(map[string]map[string]struct{}),
		artistIndex: make(map[string]map[string]struct{}),
	}
}

func (s *InMemoryTrackStore) Get(trackID string) (*models.Track, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	track, ok := s.tracks[trackID]
	return track, ok
}

func (s *InMemoryTrackStore) Upsert(track *models.Track) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.tracks[track.TrackID]; ok {
		for _, genre := range existing.Genres {
			if idx, ok := s.genreIndex[genre]; ok {
				delete(idx, track.TrackID)
			}
		}
		if idx, ok := s.artistIndex[existing.ArtistID]; ok {
			delete(idx, track.TrackID)
		}
	}

	s.tracks[track.TrackID] = track

	for _, genre := range track.Genres {
		if s.genreIndex[genre] == nil {
			s.genreIndex[genre] = make(map[string]struct{})
		}
		s.genreIndex[genre][track.TrackID] = struct{}{}
	}

	if s.artistIndex[track.ArtistID] == nil {
		s.artistIndex[track.ArtistID] = make(map[string]struct{})
	}
	s.artistIndex[track.ArtistID][track.TrackID] = struct{}{}
}

func (s *InMemoryTrackStore) GetByGenre(genre string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trackIDs := make([]string, 0)
	if idx, ok := s.genreIndex[genre]; ok {
		for trackID := range idx {
			trackIDs = append(trackIDs, trackID)
		}
	}
	return trackIDs
}

func (s *InMemoryTrackStore) GetByArtist(artistID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trackIDs := make([]string, 0)
	if idx, ok := s.artistIndex[artistID]; ok {
		for trackID := range idx {
			trackIDs = append(trackIDs, trackID)
		}
	}
	return trackIDs
}

func (s *InMemoryTrackStore) GetAll() []*models.Track {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tracks := make([]*models.Track, 0, len(s.tracks))
	for _, track := range s.tracks {
		tracks = append(tracks, track)
	}
	return tracks
}
