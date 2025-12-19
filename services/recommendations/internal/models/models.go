package models

// TrackID is a type alias for track identifiers
type TrackID = string

// Track represents a music track with its metadata
type Track struct {
	TrackID    string   `json:"track_id"`
	ArtistID   string   `json:"artist_id"`
	Genres     []string `json:"genres"`
	ReleaseTS  int64    `json:"release_ts"`
	IsExplicit bool     `json:"is_explicit"`
}

// UserProfile represents a user's listening history and preferences
// Updated ONLY from events, never from API
type UserProfile struct {
	UserID            string              `json:"user_id"`
	GenreListenCount  map[string]int      `json:"genre_listen_count"`
	ArtistListenCount map[string]int      `json:"artist_listen_count"`
	ListenedTracks    map[string]struct{} `json:"-"`
}

// NewUserProfile creates a new empty user profile
func NewUserProfile(userID string) *UserProfile {
	return &UserProfile{
		UserID:            userID,
		GenreListenCount:  make(map[string]int),
		ArtistListenCount: make(map[string]int),
		ListenedTracks:    make(map[string]struct{}),
	}
}

// HasListened checks if user has listened to a track
func (u *UserProfile) HasListened(trackID string) bool {
	_, ok := u.ListenedTracks[trackID]
	return ok
}

// GlobalTrackStats holds global popularity statistics for a track
type GlobalTrackStats struct {
	TrackID   string `json:"track_id"`
	PlayCount int    `json:"play_count"`
}
