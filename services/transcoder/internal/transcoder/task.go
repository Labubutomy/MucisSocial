package transcoder

import "context"

type Task struct {
	TrackID  string `json:"track_id"`
	ArtistID string `json:"artist_id"`
	TrackURL string `json:"track_url"`
}

type Transcoder interface {
	Transcode(ctx context.Context, task Task) error
}
