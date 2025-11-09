package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/MucisSocial/artist-service/internal/domain"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type artistRepository struct {
	db *sql.DB
}

func NewArtistRepository(db *sql.DB) domain.ArtistRepository {
	return &artistRepository{db: db}
}

func (r *artistRepository) Create(ctx context.Context, artist *domain.Artist) error {
	query := `
		INSERT INTO artists (id, name, avatar_url, genres, followers, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		artist.ID, artist.Name, artist.AvatarURL, pq.Array(artist.Genres),
		artist.Followers, artist.CreatedAt, artist.UpdatedAt)

	return err
}

func (r *artistRepository) GetByID(ctx context.Context, id string) (*domain.Artist, error) {
	artist := &domain.Artist{}
	query := `
		SELECT id, name, avatar_url, genres, followers, created_at, updated_at
		FROM artists
		WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&artist.ID,
		&artist.Name,
		&artist.AvatarURL,
		pq.Array(&artist.Genres),
		&artist.Followers,
		&artist.CreatedAt,
		&artist.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("artist with id %s not found", id)
		}
		return nil, err
	}

	return artist, nil
}

func (r *artistRepository) GetByName(ctx context.Context, name string) (*domain.Artist, error) {
	artist := &domain.Artist{}
	query := `
		SELECT id, name, avatar_url, genres, followers, created_at, updated_at
		FROM artists
		WHERE name = $1`

	row := r.db.QueryRowContext(ctx, query, name)
	err := row.Scan(
		&artist.ID,
		&artist.Name,
		&artist.AvatarURL,
		pq.Array(&artist.Genres),
		&artist.Followers,
		&artist.CreatedAt,
		&artist.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("artist with name %s not found", name)
		}
		return nil, err
	}

	return artist, nil
}

func (r *artistRepository) Update(ctx context.Context, artist *domain.Artist) error {
	query := `
		UPDATE artists 
		SET name = $2, avatar_url = $3, genres = $4, followers = $5, updated_at = $6
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		artist.ID, artist.Name, artist.AvatarURL, pq.Array(artist.Genres),
		artist.Followers, artist.UpdatedAt)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("artist with id %s not found", artist.ID)
	}

	return nil
}

func (r *artistRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM artists WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("artist with id %s not found", id)
	}

	return nil
}

func (r *artistRepository) List(ctx context.Context, limit, offset int) ([]*domain.Artist, error) {
	query := `
		SELECT id, name, avatar_url, genres, followers, created_at, updated_at
		FROM artists
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artists []*domain.Artist
	for rows.Next() {
		artist := &domain.Artist{}
		err := rows.Scan(
			&artist.ID,
			&artist.Name,
			&artist.AvatarURL,
			pq.Array(&artist.Genres),
			&artist.Followers,
			&artist.CreatedAt,
			&artist.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		artists = append(artists, artist)
	}

	return artists, nil
}

func (r *artistRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Artist, error) {
	searchQuery := `
		SELECT id, name, avatar_url, genres, followers, created_at, updated_at
		FROM artists
		WHERE name ILIKE $1 OR array_to_string(genres, ',') ILIKE $1
		ORDER BY 
			CASE WHEN name ILIKE $1 THEN 1 ELSE 2 END,
			followers DESC
		LIMIT $2`

	searchPattern := "%" + strings.ToLower(query) + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artists []*domain.Artist
	for rows.Next() {
		artist := &domain.Artist{}
		err := rows.Scan(
			&artist.ID,
			&artist.Name,
			&artist.AvatarURL,
			pq.Array(&artist.Genres),
			&artist.Followers,
			&artist.CreatedAt,
			&artist.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		artists = append(artists, artist)
	}

	return artists, nil
}

func (r *artistRepository) GetTrending(ctx context.Context, limit int) ([]*domain.Artist, error) {
	query := `
		SELECT id, name, avatar_url, genres, followers, created_at, updated_at
		FROM artists
		ORDER BY followers DESC, created_at DESC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artists []*domain.Artist
	for rows.Next() {
		artist := &domain.Artist{}
		err := rows.Scan(
			&artist.ID,
			&artist.Name,
			&artist.AvatarURL,
			pq.Array(&artist.Genres),
			&artist.Followers,
			&artist.CreatedAt,
			&artist.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		artists = append(artists, artist)
	}

	return artists, nil
}

func (r *artistRepository) NameExists(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM artists WHERE name = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
