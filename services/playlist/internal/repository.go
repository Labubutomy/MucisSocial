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
        SELECT id, author_id, name, created_at, updated_at
        FROM playlists WHERE id = $1
    `
	playlist := &Playlist{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&playlist.ID, &playlist.AuthorID, &playlist.Name,
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
        SELECT id, author_id, name, created_at, updated_at
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
			&playlist.ID, &playlist.AuthorID, &playlist.Name,
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
        INSERT INTO playlists (id, author_id, name, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.ExecContext(ctx, query,
		playlist.ID, playlist.AuthorID, playlist.Name,
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

// загружает базовую информацию о треках из playlist_tracks и tracks
func (r *Repository) loadTracksBase(ctx context.Context, playlistID uuid.UUID) ([]PlaylistTrack, error) {
	query := `
        SELECT 
            pt.track_id,
            pt.position,
            t.id,
            t.title,
            t.genre,
            t.audio_url,
            t.cover_url,
            t.duration_seconds,
            t.status,
            t.created_at,
            t.updated_at
        FROM playlist_tracks pt
        INNER JOIN tracks t ON pt.track_id = t.id
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
		var trackID uuid.UUID

		if err := rows.Scan(
			&trackID,               // pt.track_id
			&track.Position,        // pt.position
			&track.Track.ID,        // t.id
			&track.Track.Title,     // t.title
			&track.Track.Genre,     // t.genre
			&track.Track.AudioURL,  // t.audio_url
			&track.Track.CoverURL,  // t.cover_url
			&track.Track.Duration,  // t.duration_seconds
			&track.Track.Status,    // t.status
			&track.Track.CreatedAt, // t.created_at
			&track.Track.UpdatedAt, // t.updated_at
		); err != nil {
			return nil, err
		}

		// Инициализируем массивы
		track.Track.Artists = make([]Artist, 0)
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// загружает артистов для треков (с именами из таблицы artists)
func (r *Repository) loadTracksArtists(ctx context.Context, tracks []PlaylistTrack) error {
	if len(tracks) == 0 {
		return nil
	}

	// Создаем map для быстрого доступа к трекам по ID
	trackMap := make(map[uuid.UUID]*PlaylistTrack)
	trackIDs := make([]uuid.UUID, 0, len(tracks))
	for i := range tracks {
		trackMap[tracks[i].Track.ID] = &tracks[i]
		trackIDs = append(trackIDs, tracks[i].Track.ID)
	}

	// Загружаем артистов с именами из таблицы artists через JOIN с track_artists
	artistsQuery := `
        SELECT 
            ta.track_id,
            a.id,
            a.name
        FROM track_artists ta
        INNER JOIN artists a ON ta.artist_id = a.id
        WHERE ta.track_id IN (`
	args := make([]interface{}, len(trackIDs))
	for i, id := range trackIDs {
		if i > 0 {
			artistsQuery += ", "
		}
		artistsQuery += fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	artistsQuery += ") ORDER BY ta.track_id"

	artistRows, err := r.db.QueryContext(ctx, artistsQuery, args...)
	if err != nil {
		return err
	}
	defer artistRows.Close()

	// Заполняем массив Artists для каждого трека
	for artistRows.Next() {
		var trackID uuid.UUID
		var artist Artist
		if err := artistRows.Scan(&trackID, &artist.ID, &artist.Name); err != nil {
			return err
		}
		// Находим трек и добавляем артиста с именем
		if track, ok := trackMap[trackID]; ok {
			track.Track.Artists = append(track.Track.Artists, artist)
		}
	}

	return nil
}

// loadTracksArtistsByPointers загружает артистов для треков через указатели (с именами из таблицы artists)
func (r *Repository) loadTracksArtistsByPointers(ctx context.Context, tracks []*PlaylistTrack) error {
	if len(tracks) == 0 {
		return nil
	}

	// Создаем map для быстрого доступа к трекам по ID
	trackMap := make(map[uuid.UUID]*PlaylistTrack)
	trackIDs := make([]uuid.UUID, 0, len(tracks))
	for _, track := range tracks {
		trackMap[track.Track.ID] = track
		trackIDs = append(trackIDs, track.Track.ID)
	}

	// Загружаем артистов с именами из таблицы artists через JOIN с track_artists
	artistsQuery := `
        SELECT 
            ta.track_id,
            a.id,
            a.name
        FROM track_artists ta
        INNER JOIN artists a ON ta.artist_id = a.id
        WHERE ta.track_id IN (`
	args := make([]interface{}, len(trackIDs))
	for i, id := range trackIDs {
		if i > 0 {
			artistsQuery += ", "
		}
		artistsQuery += fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	artistsQuery += ") ORDER BY ta.track_id"

	artistRows, err := r.db.QueryContext(ctx, artistsQuery, args...)
	if err != nil {
		return err
	}
	defer artistRows.Close()

	// Заполняем массив Artists для каждого трека с именами артистов
	for artistRows.Next() {
		var trackID uuid.UUID
		var artist Artist
		// Сканируем track_id, artist.id и artist.name из таблицы artists
		if err := artistRows.Scan(&trackID, &artist.ID, &artist.Name); err != nil {
			return err
		}
		// Находим трек и добавляем артиста с ID и именем
		if track, ok := trackMap[trackID]; ok {
			track.Track.Artists = append(track.Track.Artists, artist)
		}
	}

	return nil
}

// получить треки для нескольких плейлистов с полной информацией
func (r *Repository) GetPlaylistsTracks(ctx context.Context, playlistIDs []uuid.UUID) (map[uuid.UUID][]PlaylistTrack, error) {
	if len(playlistIDs) == 0 {
		return make(map[uuid.UUID][]PlaylistTrack), nil
	}

	// Загружаем базовую информацию о треках
	query := `
        SELECT 
            pt.playlist_id,
            pt.track_id,
            pt.position,
            t.id,
            t.title,
            t.genre,
            t.audio_url,
            t.cover_url,
            t.duration_seconds,
            t.status,
            t.created_at,
            t.updated_at
        FROM playlist_tracks pt
        INNER JOIN tracks t ON pt.track_id = t.id
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
		var trackID uuid.UUID

		if err := rows.Scan(
			&playlistID,            // pt.playlist_id
			&trackID,               // pt.track_id
			&track.Position,        // pt.position
			&track.Track.ID,        // t.id
			&track.Track.Title,     // t.title
			&track.Track.Genre,     // t.genre
			&track.Track.AudioURL,  // t.audio_url
			&track.Track.CoverURL,  // t.cover_url
			&track.Track.Duration,  // t.duration_seconds
			&track.Track.Status,    // t.status
			&track.Track.CreatedAt, // t.created_at
			&track.Track.UpdatedAt, // t.updated_at
		); err != nil {
			return nil, err
		}

		track.Track.Artists = make([]Artist, 0)
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

// получить плейлисты пользователя (подписки)
func (r *Repository) GetUserPlaylists(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Playlist, error) {
	query := `
        SELECT p.id, p.author_id, p.name, p.created_at, p.updated_at
        FROM playlists p
        INNER JOIN playlist_users pu ON p.id = pu.playlist_id
        WHERE pu.user_id = $1
        ORDER BY pu.created_at DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []*Playlist
	for rows.Next() {
		playlist := &Playlist{}
		err := rows.Scan(
			&playlist.ID, &playlist.AuthorID, &playlist.Name,
			&playlist.CreatedAt, &playlist.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, playlist)
	}
	return playlists, nil
}
