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

// GroupProfile represents aggregated preferences for a group of users
// Used for group recommendations (e.g., for users listening together in a room)
type GroupProfile struct {
	UserIDs           []string        `json:"user_ids"`
	GenreListenCount  map[string]int  `json:"genre_listen_count"`  // Aggregated genre preferences
	ArtistListenCount map[string]int  `json:"artist_listen_count"` // Aggregated artist preferences
	ListenedTracks    map[string]int  `json:"-"`                   // Track ID -> count of users who listened
	TotalUsers        int             `json:"total_users"`
}

// NewGroupProfile creates a new empty group profile
func NewGroupProfile(userIDs []string) *GroupProfile {
	return &GroupProfile{
		UserIDs:           userIDs,
		GenreListenCount:  make(map[string]int),
		ArtistListenCount: make(map[string]int),
		ListenedTracks:    make(map[string]int),
		TotalUsers:        len(userIDs),
	}
}

// AggregateProfiles combines multiple user profiles into a single group profile
// Weights are distributed equally across all users
func AggregateProfiles(profiles []*UserProfile) *GroupProfile {
	if len(profiles) == 0 {
		return NewGroupProfile([]string{})
	}

	userIDs := make([]string, len(profiles))
	for i, p := range profiles {
		userIDs[i] = p.UserID
	}

	group := NewGroupProfile(userIDs)

	// Aggregate genre preferences
	for _, profile := range profiles {
		for genre, count := range profile.GenreListenCount {
			group.GenreListenCount[genre] += count
		}
	}

	// Aggregate artist preferences
	for _, profile := range profiles {
		for artist, count := range profile.ArtistListenCount {
			group.ArtistListenCount[artist] += count
		}
	}

	// Aggregate listened tracks - track how many users listened to each track
	for _, profile := range profiles {
		for trackID := range profile.ListenedTracks {
			group.ListenedTracks[trackID]++
		}
	}

	return group
}

// HasListenedByMajority checks if majority of users have listened to a track
// Excludes tracks that more than 50% of users have already heard
func (g *GroupProfile) HasListenedByMajority(trackID string) bool {
	if g.TotalUsers == 0 {
		return false
	}
	listenCount, exists := g.ListenedTracks[trackID]
	if !exists {
		return false
	}
	// If more than half the group has listened, exclude it
	return listenCount > (g.TotalUsers / 2)
}

// GetTopGenres returns the most popular genres in the group
func (g *GroupProfile) GetTopGenres(limit int) []string {
	type genreCount struct {
		genre string
		count int
	}

	genres := make([]genreCount, 0, len(g.GenreListenCount))
	for genre, count := range g.GenreListenCount {
		genres = append(genres, genreCount{genre, count})
	}

	// Sort by count descending
	for i := 0; i < len(genres); i++ {
		for j := i + 1; j < len(genres); j++ {
			if genres[j].count > genres[i].count {
				genres[i], genres[j] = genres[j], genres[i]
			}
		}
	}

	result := make([]string, 0, limit)
	for i := 0; i < len(genres) && i < limit; i++ {
		result = append(result, genres[i].genre)
	}

	return result
}

// GetTopArtists returns the most popular artists in the group
func (g *GroupProfile) GetTopArtists(limit int) []string {
	type artistCount struct {
		artist string
		count  int
	}

	artists := make([]artistCount, 0, len(g.ArtistListenCount))
	for artist, count := range g.ArtistListenCount {
		artists = append(artists, artistCount{artist, count})
	}

	// Sort by count descending
	for i := 0; i < len(artists); i++ {
		for j := i + 1; j < len(artists); j++ {
			if artists[j].count > artists[i].count {
				artists[i], artists[j] = artists[j], artists[i]
			}
		}
	}

	result := make([]string, 0, limit)
	for i := 0; i < len(artists) && i < limit; i++ {
		result = append(result, artists[i].artist)
	}

	return result
}
