package repository

import (
	"context"
	"database/sql"

	"github.com/MucisSocial/user-service/internal/domain"
)

type searchHistoryRepository struct {
	db *sql.DB
}

func NewSearchHistoryRepository(db *sql.DB) domain.SearchHistoryRepository {
	return &searchHistoryRepository{db: db}
}

func (r *searchHistoryRepository) GetUserSearchHistory(ctx context.Context, userID string, limit int) ([]*domain.SearchHistoryItem, error) {
	query := `
		SELECT id, user_id, query, created_at
		FROM search_history
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.SearchHistoryItem
	for rows.Next() {
		item := &domain.SearchHistoryItem{}
		err := rows.Scan(&item.ID, &item.UserID, &item.Query, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *searchHistoryRepository) Add(ctx context.Context, item *domain.SearchHistoryItem) error {
	query := `
		INSERT INTO search_history (id, user_id, query, created_at)
		VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query, item.ID, item.UserID, item.Query, item.CreatedAt)
	return err
}

func (r *searchHistoryRepository) ClearUserHistory(ctx context.Context, userID string) error {
	query := `DELETE FROM search_history WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *searchHistoryRepository) DeleteOldEntries(ctx context.Context, userID string, keepLast int) error {
	query := `
		DELETE FROM search_history 
		WHERE user_id = $1 
		AND id NOT IN (
			SELECT id FROM search_history 
			WHERE user_id = $1 
			ORDER BY created_at DESC 
			LIMIT $2
		)`

	_, err := r.db.ExecContext(ctx, query, userID, keepLast)
	return err
}
