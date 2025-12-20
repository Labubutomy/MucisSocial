package models

// TrackEvent represents a track creation or update event
type TrackEvent struct {
	EventType  string   `json:"event_type"`
	TrackID    string   `json:"track_id"`
	ArtistID   string   `json:"artist_id"`
	Genres     []string `json:"genres"`
	ReleaseTS  int64    `json:"release_ts"`
	IsExplicit bool     `json:"is_explicit"`
}

// ListeningEvent represents a user listening to a track
type ListeningEvent struct {
	EventType       string   `json:"event_type"`
	UserID          string   `json:"user_id"`
	TrackID         string   `json:"track_id"`
	ListenedSeconds int      `json:"listened_seconds"`
	Timestamp       int64    `json:"ts"`
	Lat             *float64 `json:"lat,omitempty"` // Optional latitude
	Lon             *float64 `json:"lon,omitempty"` // Optional longitude
}
