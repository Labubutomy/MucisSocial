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

// получить плейлист по ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Playlist, error) {
	query := `
        SELECT id, author_id, name, COALESCE(description, ''), COALESCE(is_private, false), created_at, updated_at
        FROM playlists WHERE id = $1
    `
	playlist := &Playlist{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&playlist.ID, &playlist.AuthorID, &playlist.Name, &playlist.Description, &playlist.IsPrivate,
		&playlist.CreatedAt, &playlist.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Загружаем треки плейлиста
	tracks, err := r.GetPlaylistTracks(ctx, id)
	if err != nil {
		return nil, err
	}
	playlist.Tracks = tracks

	return playlist, nil
}

// получить список плейлистов
func (r *Repository) List(ctx context.Context, limit, offset int, authorID *uuid.UUID) ([]*Playlist, error) {
	query := `
        SELECT id, author_id, name, COALESCE(description, ''), COALESCE(is_private, false), created_at, updated_at
        FROM playlists
    `
	args := []interface{}{}
	argPos := 1

	if authorID != nil {
		query += fmt.Sprintf(" WHERE author_id = $%d", argPos)
		args = append(args, *authorID)
		argPos++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []*Playlist
	var playlistIDs []uuid.UUID
	for rows.Next() {
		playlist := &Playlist{}
		err := rows.Scan(
			&playlist.ID, &playlist.AuthorID, &playlist.Name, &playlist.Description, &playlist.IsPrivate,
			&playlist.CreatedAt, &playlist.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, playlist)
		playlistIDs = append(playlistIDs, playlist.ID)
	}

	// Batch загрузка треков для всех плейлистов
	if len(playlistIDs) > 0 {
		tracksMap, err := r.GetPlaylistsTracks(ctx, playlistIDs)
		if err != nil {
			return nil, err
		}
		for _, playlist := range playlists {
			playlist.Tracks = tracksMap[playlist.ID]
		}
	}

	return playlists, nil
}

// создать плейлист
func (r *Repository) Create(ctx context.Context, playlist *Playlist) error {
	query := `
        INSERT INTO playlists (id, author_id, name, description, is_private, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := r.db.ExecContext(ctx, query,
		playlist.ID, playlist.AuthorID, playlist.Name, playlist.Description, playlist.IsPrivate,
		playlist.CreatedAt, playlist.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Создаем связи с треками, если они есть
	if len(playlist.Tracks) > 0 {
		return r.CreatePlaylistTracks(ctx, playlist.ID, playlist.Tracks)
	}

	return nil
}

// обновить плейлист
func (r *Repository) Update(ctx context.Context, playlist *Playlist) error {
	query := `
        UPDATE playlists SET 
            name = $1, updated_at = $2
        WHERE id = $3
    `
	result, err := r.db.ExecContext(ctx, query,
		playlist.Name, playlist.UpdatedAt, playlist.ID,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	// Обновляем треки, если они переданы
	if playlist.Tracks != nil {
		if err := r.DeletePlaylistTracks(ctx, playlist.ID); err != nil {
			return err
		}
		if len(playlist.Tracks) > 0 {
			return r.CreatePlaylistTracks(ctx, playlist.ID, playlist.Tracks)
		}
	}

	return nil
}

// удалить плейлист
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM playlists WHERE id = $1`
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

// получить треки плейлиста с полной информацией из таблицы tracks
func (r *Repository) GetPlaylistTracks(ctx context.Context, playlistID uuid.UUID) ([]PlaylistTrack, error) {
	tracks, err := r.loadTracksBase(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	if len(tracks) == 0 {
		return tracks, nil
	}

	// Загружаем артистов для всех треков
	if err := r.loadTracksArtists(ctx, tracks); err != nil {
		return nil, err
	}

	return tracks, nil
}

// загружает базовую информацию о треках из playlist_tracks
// Примечание: полная информация о треках должна получаться через tracks-service API
func (r *Repository) loadTracksBase(ctx context.Context, playlistID uuid.UUID) ([]PlaylistTrack, error) {
	query := `
        SELECT 
            pt.track_id,
            pt.position
        FROM playlist_tracks pt
        WHERE pt.playlist_id = $1
        ORDER BY pt.position, pt.added_at
    `
	rows, err := r.db.QueryContext(ctx, query, playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []PlaylistTrack

	for rows.Next() {
		var track PlaylistTrack

		if err := rows.Scan(
			&track.Track.ID,  // pt.track_id
			&track.Position,  // pt.position
		); err != nil {
			return nil, err
		}

		// Инициализируем массивы и устанавливаем значения по умолчанию
		track.Track.Artists = make([]Artist, 0)
		track.Track.Status = "active" // Значение по умолчанию
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// загружает артистов для треков
// Примечание: информация об артистах должна получаться через tracks-service API
// Эта функция оставлена для совместимости, но не выполняет запросов к БД
func (r *Repository) loadTracksArtists(ctx context.Context, tracks []PlaylistTrack) error {
	// В микросервисной архитектуре информация об артистах должна получаться
	// через API tracks-service, а не из локальной БД playlist-service
	// Оставляем функцию пустой, так как Artists уже инициализирован как пустой массив
	return nil
}

// loadTracksArtistsByPointers загружает артистов для треков через указатели
// Примечание: информация об артистах должна получаться через tracks-service API
// Эта функция оставлена для совместимости, но не выполняет запросов к БД
func (r *Repository) loadTracksArtistsByPointers(ctx context.Context, tracks []*PlaylistTrack) error {
	// В микросервисной архитектуре информация об артистах должна получаться
	// через API tracks-service, а не из локальной БД playlist-service
	// Оставляем функцию пустой, так как Artists уже инициализирован как пустой массив
	return nil
}

// получить треки для нескольких плейлистов с полной информацией
func (r *Repository) GetPlaylistsTracks(ctx context.Context, playlistIDs []uuid.UUID) (map[uuid.UUID][]PlaylistTrack, error) {
	if len(playlistIDs) == 0 {
		return make(map[uuid.UUID][]PlaylistTrack), nil
	}

	// Загружаем базовую информацию о треках из playlist_tracks
	// Примечание: полная информация о треках должна получаться через tracks-service API
	query := `
        SELECT 
            pt.playlist_id,
            pt.track_id,
            pt.position
        FROM playlist_tracks pt
        WHERE pt.playlist_id IN (`
	args := make([]interface{}, len(playlistIDs))
	for i, id := range playlistIDs {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	query += ") ORDER BY pt.playlist_id, pt.position, pt.added_at"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tracksMap := make(map[uuid.UUID][]*PlaylistTrack)
	var allTracks []*PlaylistTrack

	for rows.Next() {
		var playlistID uuid.UUID
		track := &PlaylistTrack{}

		if err := rows.Scan(
			&playlistID,         // pt.playlist_id
			&track.Track.ID,     // pt.track_id
			&track.Position,     // pt.position
		); err != nil {
			return nil, err
		}

		// Инициализируем массивы и устанавливаем значения по умолчанию
		track.Track.Artists = make([]Artist, 0)
		track.Track.Status = "active" // Значение по умолчанию
		
		if tracksMap[playlistID] == nil {
			tracksMap[playlistID] = make([]*PlaylistTrack, 0)
		}
		tracksMap[playlistID] = append(tracksMap[playlistID], track)
		allTracks = append(allTracks, track)
	}

	// Загружаем артистов для всех треков
	if len(allTracks) > 0 {
		// Загружаем артистов напрямую через указатели
		if err := r.loadTracksArtistsByPointers(ctx, allTracks); err != nil {
			return nil, err
		}
	}

	// Преобразуем map[*PlaylistTrack] в map[[]PlaylistTrack]
	result := make(map[uuid.UUID][]PlaylistTrack)
	for playlistID, tracks := range tracksMap {
		result[playlistID] = make([]PlaylistTrack, len(tracks))
		for i, track := range tracks {
			result[playlistID][i] = *track
		}
	}

	return result, nil
}

// создать связи между плейлистом и треками
func (r *Repository) CreatePlaylistTracks(ctx context.Context, playlistID uuid.UUID, tracks []PlaylistTrack) error {
	if len(tracks) == 0 {
		return nil
	}

	query := `INSERT INTO playlist_tracks (playlist_id, track_id, position) VALUES `
	args := make([]interface{}, 0, len(tracks)*3)

	for i, track := range tracks {
		if i > 0 {
			query += ", "
		}
		position := i
		if track.Position > 0 {
			position = track.Position
		}
		query += fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
		args = append(args, playlistID, track.Track.ID, position)
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// удалить все связи плейлиста с треками
func (r *Repository) DeletePlaylistTracks(ctx context.Context, playlistID uuid.UUID) error {
	query := `DELETE FROM playlist_tracks WHERE playlist_id = $1`
	_, err := r.db.ExecContext(ctx, query, playlistID)
	return err
}

// добавить трек в плейлист
func (r *Repository) AddTrackToPlaylist(ctx context.Context, playlistID, trackID uuid.UUID, position int) error {
	query := `
        INSERT INTO playlist_tracks (playlist_id, track_id, position)
        VALUES ($1, $2, $3)
        ON CONFLICT (playlist_id, track_id) DO UPDATE SET position = $3
    `
	_, err := r.db.ExecContext(ctx, query, playlistID, trackID, position)
	return err
}

// удалить трек из плейлиста
func (r *Repository) RemoveTrackFromPlaylist(ctx context.Context, playlistID, trackID uuid.UUID) error {
	query := `DELETE FROM playlist_tracks WHERE playlist_id = $1 AND track_id = $2`
	result, err := r.db.ExecContext(ctx, query, playlistID, trackID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// подписать пользователя на плейлист
func (r *Repository) SubscribeUser(ctx context.Context, userID, playlistID uuid.UUID) error {
	query := `
        INSERT INTO playlist_users (user_id, playlist_id)
        VALUES ($1, $2)
        ON CONFLICT (user_id, playlist_id) DO NOTHING
    `
	_, err := r.db.ExecContext(ctx, query, userID, playlistID)
	return err
}

// отписать пользователя от плейлиста
func (r *Repository) UnsubscribeUser(ctx context.Context, userID, playlistID uuid.UUID) error {
	query := `DELETE FROM playlist_users WHERE user_id = $1 AND playlist_id = $2`
	result, err := r.db.ExecContext(ctx, query, userID, playlistID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// получить плейлисты пользователя (созданные пользователем + подписки)
func (r *Repository) GetUserPlaylists(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Playlist, error) {
	// Возвращаем плейлисты, которые пользователь создал (author_id = user_id)
	// ИЛИ на которые пользователь подписан (через playlist_users)
	// Также подсчитываем количество треков для каждого плейлиста
	query := `
        SELECT DISTINCT 
            p.id, 
            p.author_id, 
            p.name, 
            COALESCE(p.description, ''), 
            COALESCE(p.is_private, false), 
            p.created_at, 
            p.updated_at,
            COALESCE(COUNT(pt.track_id), 0) as tracks_count
        FROM playlists p
        LEFT JOIN playlist_users pu ON p.id = pu.playlist_id AND pu.user_id = $1
        LEFT JOIN playlist_tracks pt ON p.id = pt.playlist_id
        WHERE p.author_id = $1 OR pu.user_id = $1
        GROUP BY p.id, p.author_id, p.name, p.description, p.is_private, p.created_at, p.updated_at
        ORDER BY p.created_at DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Инициализируем как пустой слайс, чтобы избежать nil
	playlists := make([]*Playlist, 0)
	for rows.Next() {
		playlist := &Playlist{}
		var tracksCount int
		err := rows.Scan(
			&playlist.ID, &playlist.AuthorID, &playlist.Name, &playlist.Description, &playlist.IsPrivate,
			&playlist.CreatedAt, &playlist.UpdatedAt, &tracksCount,
		)
		if err != nil {
			return nil, err
		}
		// Инициализируем Tracks с правильным размером для корректного подсчета
		playlist.Tracks = make([]PlaylistTrack, tracksCount)
		playlists = append(playlists, playlist)
	}
	return playlists, nil
}
