package domain

import (
	"context"
)

type ArtistRepository interface {
	Create(ctx context.Context, artist *Artist) error
	GetByID(ctx context.Context, id string) (*Artist, error)
	GetByName(ctx context.Context, name string) (*Artist, error)
	Update(ctx context.Context, artist *Artist) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*Artist, error)
	Search(ctx context.Context, query string, limit int) ([]*Artist, error)
	GetTrending(ctx context.Context, limit int) ([]*Artist, error)
	NameExists(ctx context.Context, name string) (bool, error)
}

type ArtistService interface {
	CreateArtist(ctx context.Context, req CreateArtistRequest) (*ArtistResponse, error)
	GetArtistByID(ctx context.Context, id string) (*ArtistResponse, error)
	UpdateArtist(ctx context.Context, id string, req UpdateArtistRequest) (*ArtistResponse, error)
	DeleteArtist(ctx context.Context, id string) error
	ListArtists(ctx context.Context, limit, offset int) ([]*ArtistResponse, error)
	SearchArtists(ctx context.Context, query string, limit int) ([]*ArtistResponse, error)
	GetTrendingArtists(ctx context.Context, limit int) ([]*TrendingArtist, error)
}
