package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetByID получить трек по ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Track, error) {
	query := `
        SELECT id, title, artist_id, artist_name, genre, audio_url, cover_url,
               duration_seconds, status, created_at, updated_at
        FROM tracks WHERE id = $1
    `
	track := &Track{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&track.ID, &track.Title, &track.ArtistID, &track.ArtistName,
		&track.Genre, &track.AudioURL, &track.CoverURL, &track.Duration,
		&track.Status, &track.CreatedAt, &track.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return track, err
}

// List получить список треков
func (r *Repository) List(ctx context.Context, limit, offset int, artistID *uuid.UUID) ([]*Track, error) {
	query := `
        SELECT id, title, artist_id, artist_name, genre, audio_url, cover_url,
               duration_seconds, status, created_at, updated_at
        FROM tracks 
        WHERE status = $1
    `
	args := []interface{}{StatusReady}

	if artistID != nil {
		query += " AND artist_id = $2"
		args = append(args, *artistID)
	}

	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1) +
		" OFFSET $" + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*Track
	for rows.Next() {
		track := &Track{}
		err := rows.Scan(
			&track.ID, &track.Title, &track.ArtistID, &track.ArtistName,
			&track.Genre, &track.AudioURL, &track.CoverURL, &track.Duration,
			&track.Status, &track.CreatedAt, &track.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}
	return tracks, nil
}

// Create создать трек
func (r *Repository) Create(ctx context.Context, track *Track) error {
	query := `
        INSERT INTO tracks (id, title, artist_id, artist_name, genre, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err := r.db.ExecContext(ctx, query,
		track.ID, track.Title, track.ArtistID, track.ArtistName,
		track.Genre, track.Status, track.CreatedAt, track.UpdatedAt,
	)
	return err
}

// Update обновить трек
func (r *Repository) Update(ctx context.Context, track *Track) error {
	query := `
        UPDATE tracks SET 
            title = $1, artist_name = $2, genre = $3,
            audio_url = $4, cover_url = $5, duration_seconds = $6,
            status = $7, updated_at = $8
        WHERE id = $9
    `
	result, err := r.db.ExecContext(ctx, query,
		track.Title, track.ArtistName, track.Genre,
		track.AudioURL, track.CoverURL, track.Duration,
		track.Status, track.UpdatedAt, track.ID,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete удалить трек
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tracks WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateStatus обновить статус
func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE tracks SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
