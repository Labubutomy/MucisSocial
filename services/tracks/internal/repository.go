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
        SELECT id, title, genre, audio_url, cover_url,
               duration_seconds, status, created_at, updated_at
        FROM tracks WHERE id = $1
    `
	track := &Track{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&track.ID, &track.Title, &track.Genre, &track.AudioURL, &track.CoverURL,
		&track.Duration, &track.Status, &track.CreatedAt, &track.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Загружаем ID артистов для трека
	artistIDs, err := r.GetTrackArtistIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	track.ArtistIDs = artistIDs

	return track, nil
}

// List получить список треков
func (r *Repository) List(ctx context.Context, limit, offset int, artistID *uuid.UUID) ([]*Track, error) {
	query := `
        SELECT t.id, t.title, t.genre, t.audio_url, t.cover_url,
               t.duration_seconds, t.status, t.created_at, t.updated_at
        FROM tracks t
    `
	args := []interface{}{StatusReady}
	argPos := 2

	if artistID != nil {
		query += ` INNER JOIN track_artists ta ON t.id = ta.track_id 
		           WHERE t.status = $1 AND ta.artist_id = $2`
		args = append(args, *artistID)
		argPos = 3
	} else {
		query += ` WHERE t.status = $1`
	}

	query += fmt.Sprintf(" ORDER BY t.created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*Track
	var trackIDs []uuid.UUID
	for rows.Next() {
		track := &Track{}
		err := rows.Scan(
			&track.ID, &track.Title, &track.Genre, &track.AudioURL, &track.CoverURL,
			&track.Duration, &track.Status, &track.CreatedAt, &track.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
		trackIDs = append(trackIDs, track.ID)
	}

	// Batch загрузка ID артистов для всех треков
	if len(trackIDs) > 0 {
		artistIDsMap, err := r.GetTracksArtistIDs(ctx, trackIDs)
		if err != nil {
			return nil, err
		}
		// Присваиваем ID артистов к трекам
		for _, track := range tracks {
			track.ArtistIDs = artistIDsMap[track.ID]
		}
	}

	return tracks, nil
}

// Search поиск треков по названию
func (r *Repository) Search(ctx context.Context, query string, limit, offset int) ([]*Track, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	searchQuery := fmt.Sprintf("%%%s%%", query)
	sqlQuery := `
        SELECT t.id, t.title, t.genre, t.audio_url, t.cover_url,
               t.duration_seconds, t.status, t.created_at, t.updated_at
        FROM tracks t
        WHERE t.status = $1 AND t.title ILIKE $2
        ORDER BY t.created_at DESC
        LIMIT $3 OFFSET $4
    `

	rows, err := r.db.QueryContext(ctx, sqlQuery, StatusReady, searchQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*Track
	var trackIDs []uuid.UUID
	for rows.Next() {
		track := &Track{}
		err := rows.Scan(
			&track.ID, &track.Title, &track.Genre, &track.AudioURL, &track.CoverURL,
			&track.Duration, &track.Status, &track.CreatedAt, &track.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
		trackIDs = append(trackIDs, track.ID)
	}

	// Batch загрузка ID артистов для всех треков
	if len(trackIDs) > 0 {
		artistIDsMap, err := r.GetTracksArtistIDs(ctx, trackIDs)
		if err != nil {
			return nil, err
		}
		// Присваиваем ID артистов к трекам
		for _, track := range tracks {
			track.ArtistIDs = artistIDsMap[track.ID]
		}
	}

	return tracks, nil
}

// Create создать трек
func (r *Repository) Create(ctx context.Context, track *Track) error {
	query := `
        INSERT INTO tracks (id, title, genre, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := r.db.ExecContext(ctx, query,
		track.ID, track.Title, track.Genre, track.Status, track.CreatedAt, track.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Создаем связи с артистами
	return r.CreateTrackArtists(ctx, track.ID, track.ArtistIDs)
}

// Update обновить трек
func (r *Repository) Update(ctx context.Context, track *Track) error {
	query := `
        UPDATE tracks SET 
            title = $1, genre = $2,
            audio_url = $3, cover_url = $4, duration_seconds = $5,
            status = $6, updated_at = $7
        WHERE id = $8
    `
	result, err := r.db.ExecContext(ctx, query,
		track.Title, track.Genre,
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

	// Обновляем связи с артистами (удаляем старые, создаем новые)
	if err := r.DeleteTrackArtists(ctx, track.ID); err != nil {
		return err
	}
	return r.CreateTrackArtists(ctx, track.ID, track.ArtistIDs)
}

// UpdateURLs обновить только URLs трека (cover_url, audio_url, duration) без изменения других полей
func (r *Repository) UpdateURLsAndDuration(ctx context.Context, trackID uuid.UUID, coverURL, audioURL string, durationSec int) error {
	// Дефолтная обложка для всех треков
	const defaultCoverURL = "https://mir-s3-cdn-cf.behance.net/projects/202/e2ba0e187042211.Y3JvcCw4MDgsNjMyLDAsMA.png"

	// Строим запрос динамически в зависимости от того, какие поля нужно обновить
	query := `UPDATE tracks SET updated_at = NOW()`
	args := []interface{}{}
	argPos := 1

	if len(audioURL) == 0 {
		return ErrBadRequest
	}

	// Используем дефолтную обложку, если не указана
	if len(coverURL) == 0 {
		coverURL = defaultCoverURL
	}

	query += fmt.Sprintf(", cover_url = $%d", argPos)
	args = append(args, coverURL)
	argPos++

	query += fmt.Sprintf(", audio_url = $%d", argPos)
	args = append(args, audioURL)
	argPos++

	query += fmt.Sprintf(", duration_seconds = $%d", argPos)
	args = append(args, durationSec)
	argPos++

	// Обновляем статус на ready после успешного транскодирования
	query += fmt.Sprintf(", status = $%d", argPos)
	args = append(args, StatusReady)
	argPos++

	// Если ничего не обновляется, возвращаем ошибку
	if len(args) == 0 {
		return ErrBadRequest
	}

	query += fmt.Sprintf(" WHERE id = $%d", argPos)
	args = append(args, trackID)

	result, err := r.db.ExecContext(ctx, query, args...)
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

// GetTracksArtistIDs получить ID артистов для нескольких треков (batch загрузка)
func (r *Repository) GetTracksArtistIDs(ctx context.Context, trackIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	if len(trackIDs) == 0 {
		return make(map[uuid.UUID][]uuid.UUID), nil
	}

	// Строим запрос с IN для каждого track ID
	query := `
        SELECT ta.track_id, ta.artist_id
        FROM track_artists ta
        WHERE ta.track_id IN (`
	args := make([]interface{}, len(trackIDs))
	for i, id := range trackIDs {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	query += ") ORDER BY ta.track_id, ta.artist_id"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	artistIDsMap := make(map[uuid.UUID][]uuid.UUID)
	for rows.Next() {
		var trackID, artistID uuid.UUID
		if err := rows.Scan(&trackID, &artistID); err != nil {
			return nil, err
		}
		artistIDsMap[trackID] = append(artistIDsMap[trackID], artistID)
	}
	return artistIDsMap, nil
}

// GetTrackArtistIDs получить ID артистов для трека
func (r *Repository) GetTrackArtistIDs(ctx context.Context, trackID uuid.UUID) ([]uuid.UUID, error) {
	query := `
        SELECT ta.artist_id
        FROM track_artists ta
        WHERE ta.track_id = $1
        ORDER BY ta.artist_id
    `
	rows, err := r.db.QueryContext(ctx, query, trackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artistIDs []uuid.UUID
	for rows.Next() {
		var artistID uuid.UUID
		if err := rows.Scan(&artistID); err != nil {
			return nil, err
		}
		artistIDs = append(artistIDs, artistID)
	}
	return artistIDs, nil
}

// CreateTrackArtists создать связи между треком и артистами (batch insert)
func (r *Repository) CreateTrackArtists(ctx context.Context, trackID uuid.UUID, artistIDs []uuid.UUID) error {
	if len(artistIDs) == 0 {
		return nil
	}

	// Используем batch insert для оптимизации
	query := `INSERT INTO track_artists (track_id, artist_id) VALUES `
	args := make([]interface{}, 0, len(artistIDs)*2)

	for i, artistID := range artistIDs {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
		args = append(args, trackID, artistID)
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// DeleteTrackArtists удалить все связи трека с артистами
func (r *Repository) DeleteTrackArtists(ctx context.Context, trackID uuid.UUID) error {
	query := `DELETE FROM track_artists WHERE track_id = $1`
	_, err := r.db.ExecContext(ctx, query, trackID)
	return err
}
