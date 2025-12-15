package store

import (
	"sync"

	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
)

// UserProfileStore defines the interface for user profile storage
type UserProfileStore interface {
	Get(userID string) (*models.UserProfile, bool)
	GetOrCreate(userID string) *models.UserProfile
	Update(profile *models.UserProfile)
	GetAll() map[string]*models.UserProfile // For backup purposes
}

// InMemoryUserProfileStore is an in-memory implementation of UserProfileStore
type InMemoryUserProfileStore struct {
	mu       sync.RWMutex
	profiles map[string]*models.UserProfile
}

// NewInMemoryUserProfileStore creates a new in-memory user profile store
func NewInMemoryUserProfileStore() *InMemoryUserProfileStore {
	return &InMemoryUserProfileStore{
		profiles: make(map[string]*models.UserProfile),
	}
}

func (s *InMemoryUserProfileStore) Get(userID string) (*models.UserProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	profile, ok := s.profiles[userID]
	if ok && profile != nil {
		// Ensure all maps are initialized (in case they were nil from backup)
		if profile.GenreListenCount == nil {
			profile.GenreListenCount = make(map[string]int)
		}
		if profile.ArtistListenCount == nil {
			profile.ArtistListenCount = make(map[string]int)
		}
		if profile.ListenedTracks == nil {
			profile.ListenedTracks = make(map[string]struct{})
		}
	}
	return profile, ok
}

func (s *InMemoryUserProfileStore) GetOrCreate(userID string) *models.UserProfile {
	s.mu.Lock()
	defer s.mu.Unlock()

	if profile, ok := s.profiles[userID]; ok {
		// Ensure all maps are initialized (in case they were nil from backup)
		if profile.GenreListenCount == nil {
			profile.GenreListenCount = make(map[string]int)
		}
		if profile.ArtistListenCount == nil {
			profile.ArtistListenCount = make(map[string]int)
		}
		if profile.ListenedTracks == nil {
			profile.ListenedTracks = make(map[string]struct{})
		}
		return profile
	}

	profile := models.NewUserProfile(userID)
	s.profiles[userID] = profile
	return profile
}

func (s *InMemoryUserProfileStore) Update(profile *models.UserProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Ensure all maps are initialized (in case they were nil from backup)
	if profile.GenreListenCount == nil {
		profile.GenreListenCount = make(map[string]int)
	}
	if profile.ArtistListenCount == nil {
		profile.ArtistListenCount = make(map[string]int)
	}
	if profile.ListenedTracks == nil {
		profile.ListenedTracks = make(map[string]struct{})
	}
	
	s.profiles[profile.UserID] = profile
}

func (s *InMemoryUserProfileStore) GetAll() map[string]*models.UserProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*models.UserProfile)
	for userID, profile := range s.profiles {
		result[userID] = profile
	}
	return result
}
