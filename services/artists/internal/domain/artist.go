package domain

import (
	"time"

	"github.com/google/uuid"
)

type Artist struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	AvatarURL *string   `json:"avatar_url" db:"avatar_url"`
	Genres    []string  `json:"genres" db:"genres"`
	Followers int64     `json:"followers" db:"followers"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type TopTrack struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	CoverURL *string `json:"cover_url"`
}

type ArtistResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	AvatarURL *string    `json:"avatarUrl"`
	Genres    []string   `json:"genres"`
	Followers int64      `json:"followers"`
	TopTracks []TopTrack `json:"topTracks,omitempty"`
}

type TrendingArtist struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	AvatarURL *string  `json:"avatarUrl"`
	Genres    []string `json:"genres"`
}

type CreateArtistRequest struct {
	Name      string   `json:"name" validate:"required,min=1,max=100"`
	AvatarURL *string  `json:"avatar_url"`
	Genres    []string `json:"genres"`
}

type UpdateArtistRequest struct {
	Name      *string  `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	AvatarURL *string  `json:"avatar_url,omitempty"`
	Genres    []string `json:"genres,omitempty"`
}

func NewArtist(req CreateArtistRequest) *Artist {
	return &Artist{
		ID:        uuid.New().String(),
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
		Genres:    req.Genres,
		Followers: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (a *Artist) ToResponse() *ArtistResponse {
	return &ArtistResponse{
		ID:        a.ID,
		Name:      a.Name,
		AvatarURL: a.AvatarURL,
		Genres:    a.Genres,
		Followers: a.Followers,
	}
}

func (a *Artist) ToTrendingArtist() *TrendingArtist {
	return &TrendingArtist{
		ID:        a.ID,
		Name:      a.Name,
		AvatarURL: a.AvatarURL,
		Genres:    a.Genres,
	}
}

func (a *Artist) Update(req UpdateArtistRequest) {
	if req.Name != nil {
		a.Name = *req.Name
	}
	if req.AvatarURL != nil {
		a.AvatarURL = req.AvatarURL
	}
	if req.Genres != nil {
		a.Genres = req.Genres
	}
	a.UpdatedAt = time.Now()
}
