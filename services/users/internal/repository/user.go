package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/MucisSocial/user-service/internal/domain"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.AvatarURL, user.CreatedAt, user.UpdatedAt)

	return err
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user := &domain.User{}
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.avatar_url, u.created_at, u.updated_at,
		       COALESCE(array_agg(DISTINCT mtg.genre) FILTER (WHERE mtg.genre IS NOT NULL), '{}') as top_genres,
		       COALESCE(array_agg(DISTINCT mta.artist) FILTER (WHERE mta.artist IS NOT NULL), '{}') as top_artists
		FROM users u
		LEFT JOIN music_taste_genres mtg ON u.id = mtg.user_id
		LEFT JOIN music_taste_artists mta ON u.id = mta.user_id
		WHERE u.id = $1
		GROUP BY u.id, u.username, u.email, u.password_hash, u.avatar_url, u.created_at, u.updated_at`

	var topGenres, topArtists pq.StringArray
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
		&topGenres, &topArtists)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if len(topGenres) > 0 || len(topArtists) > 0 {
		user.MusicTasteSummary = &domain.MusicTasteSummary{
			TopGenres:  []string(topGenres),
			TopArtists: []string(topArtists),
		}
	}

	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.avatar_url, u.created_at, u.updated_at,
		       COALESCE(array_agg(DISTINCT mtg.genre) FILTER (WHERE mtg.genre IS NOT NULL), '{}') as top_genres,
		       COALESCE(array_agg(DISTINCT mta.artist) FILTER (WHERE mta.artist IS NOT NULL), '{}') as top_artists
		FROM users u
		LEFT JOIN music_taste_genres mtg ON u.id = mtg.user_id
		LEFT JOIN music_taste_artists mta ON u.id = mta.user_id
		WHERE u.email = $1
		GROUP BY u.id, u.username, u.email, u.password_hash, u.avatar_url, u.created_at, u.updated_at`

	var topGenres, topArtists pq.StringArray
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
		&topGenres, &topArtists)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if len(topGenres) > 0 || len(topArtists) > 0 {
		user.MusicTasteSummary = &domain.MusicTasteSummary{
			TopGenres:  []string(topGenres),
			TopArtists: []string(topArtists),
		}
	}

	return user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	user := &domain.User{}
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.avatar_url, u.created_at, u.updated_at,
		       COALESCE(array_agg(DISTINCT mtg.genre) FILTER (WHERE mtg.genre IS NOT NULL), '{}') as top_genres,
		       COALESCE(array_agg(DISTINCT mta.artist) FILTER (WHERE mta.artist IS NOT NULL), '{}') as top_artists
		FROM users u
		LEFT JOIN music_taste_genres mtg ON u.id = mtg.user_id
		LEFT JOIN music_taste_artists mta ON u.id = mta.user_id
		WHERE u.username = $1
		GROUP BY u.id, u.username, u.email, u.password_hash, u.avatar_url, u.created_at, u.updated_at`

	var topGenres, topArtists pq.StringArray
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
		&topGenres, &topArtists)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if len(topGenres) > 0 || len(topArtists) > 0 {
		user.MusicTasteSummary = &domain.MusicTasteSummary{
			TopGenres:  []string(topGenres),
			TopArtists: []string(topArtists),
		}
	}

	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()
	query := `
		UPDATE users 
		SET username = $2, email = $3, avatar_url = $4, updated_at = $5
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.AvatarURL, user.UpdatedAt)

	return err
}

func (r *userRepository) UpdateMusicTaste(ctx context.Context, userID string, summary *domain.MusicTasteSummary) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing music taste data
	_, err = tx.ExecContext(ctx, "DELETE FROM music_taste_genres WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM music_taste_artists WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	// Insert new genres
	for _, genre := range summary.TopGenres {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO music_taste_genres (user_id, genre) VALUES ($1, $2)",
			userID, genre)
		if err != nil {
			return err
		}
	}

	// Insert new artists
	for _, artist := range summary.TopArtists {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO music_taste_artists (user_id, artist) VALUES ($1, $2)",
			userID, artist)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = $1`
	err := r.db.QueryRowContext(ctx, query, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE username = $1`
	err := r.db.QueryRowContext(ctx, query, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
