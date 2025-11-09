package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/MucisSocial/user-service/internal/domain"
)

type refreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) domain.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query, token.ID, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt)
	return err
}

func (r *refreshTokenRepository) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	refreshToken := &domain.RefreshToken{}
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = $1 AND expires_at > NOW()`

	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&refreshToken.ID, &refreshToken.UserID, &refreshToken.Token,
		&refreshToken.ExpiresAt, &refreshToken.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return refreshToken, nil
}

func (r *refreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *refreshTokenRepository) DeleteByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at <= $1`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}
