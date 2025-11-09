package main

// CreateTrackRequest represents the request body for creating a track
type CreateTrackRequest struct {
	Title     string   `json:"title" example:"Beautiful Song"`
	ArtistIds []string `json:"artist_ids" example:"['uuid1', 'uuid2']"`
	Genre     string   `json:"genre" example:"Pop"`
}

// CreateTrackResponse represents the response for track creation
type CreateTrackResponse struct {
	TrackId string `json:"track_id" example:"uuid"`
}

// UpdateTrackInfoRequest represents the request body for updating track info
type UpdateTrackInfoRequest struct {
	CoverUrl    string `json:"cover_url" example:"https://example.com/cover.jpg"`
	AudioUrl    string `json:"audio_url" example:"https://example.com/audio.mp3"`
	DurationSec int32  `json:"duration_sec" example:"180"`
}

// CreatePlaylistRequest represents the request body for creating a playlist
type CreatePlaylistRequest struct {
	Name        string `json:"name" example:"My Favorite Songs"`
	Description string `json:"description" example:"Collection of my favorite tracks"`
	IsPrivate   bool   `json:"is_private" example:"false"`
}

// CreatePlaylistResponse represents the response for playlist creation
type CreatePlaylistResponse struct {
	PlaylistId string `json:"playlist_id" example:"uuid"`
}

// UpdatePlaylistRequest represents the request body for updating a playlist
type UpdatePlaylistRequest struct {
	Name        string `json:"name" example:"Updated Playlist Name"`
	Description string `json:"description" example:"Updated description"`
	IsPrivate   bool   `json:"is_private" example:"true"`
}

// GetUserPlaylistsResponse represents the response for getting user playlists
type GetUserPlaylistsResponse struct {
	Playlists []Playlist `json:"playlists"`
	Total     int32      `json:"total" example:"10"`
}

// GetPlaylistResponse represents the response for getting a single playlist
type GetPlaylistResponse struct {
	Playlist Playlist `json:"playlist"`
}

// GetPlaylistTracksResponse represents the response for getting playlist tracks
type GetPlaylistTracksResponse struct {
	Tracks []PlaylistTrack `json:"tracks"`
	Total  int32           `json:"total" example:"5"`
}

// AddTrackToPlaylistRequest represents the request body for adding a track to playlist
type AddTrackToPlaylistRequest struct {
	TrackId string `json:"track_id" example:"uuid"`
}

// Playlist represents a playlist object
type Playlist struct {
	Id          string `json:"id" example:"uuid"`
	UserId      string `json:"user_id" example:"uuid"`
	Name        string `json:"name" example:"My Playlist"`
	Description string `json:"description" example:"Description"`
	IsPrivate   bool   `json:"is_private" example:"false"`
	CreatedAt   string `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt   string `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	TracksCount int32  `json:"tracks_count" example:"10"`
}

// PlaylistTrack represents a track in a playlist
type PlaylistTrack struct {
	TrackId    string `json:"track_id" example:"uuid"`
	PlaylistId string `json:"playlist_id" example:"uuid"`
	AddedAt    string `json:"added_at" example:"2023-01-01T00:00:00Z"`
	Position   int32  `json:"position" example:"1"`
}
