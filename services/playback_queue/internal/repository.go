package internal

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateQueue(ctx context.Context, contextType string) (ContextRef, error) {
	if contextType == "" {
		return ContextRef{}, ErrBadRequest
	}
	id := uuid.New()
	ref := ContextRef{ContextType: contextType, ContextID: id}
	_, err := r.db.ExecContext(ctx, `INSERT INTO playback_queue_state (context_type, context_id, context_key, current_position)
	VALUES ($1,$2,$3,0)
	ON CONFLICT (context_key) DO NOTHING`, contextType, id, ref.Key())
	if err != nil {
		return ContextRef{}, err
	}
	return ref, nil
}

func (r *Repository) Enqueue(ctx context.Context, item *QueueItem) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = r.createStateForUserIfNotExists(ctx, tx, item.Context); err != nil {
		return err
	}

	var nextPosition int64
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(position), 0) + 1 FROM playback_queue_items WHERE context_key = $1`, item.Context.Key()).Scan(&nextPosition)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO playback_queue_items (context_key, track_id, position)
	VALUES ($1,$2,$3)`,
		item.Context.Key(),
		item.TrackID,
		nextPosition,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) ListFuture(ctx context.Context, ref ContextRef, limit int) ([]*QueueItem, error) {
	// Для пользователей создаем состояние, если его нет
	if strings.EqualFold(ref.ContextType, "user") {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		err = r.createStateIfNotExists(ctx, tx, ref)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		_ = tx.Commit()
	}

	current, err := r.currentPosition(ctx, ref)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT track_id FROM playback_queue_items
	WHERE context_key = $1 AND position > $2 ORDER BY position ASC LIMIT $3`, ref.Key(), current, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*QueueItem, 0)
	for rows.Next() {
		item := &QueueItem{Context: ref}
		if err := rows.Scan(&item.TrackID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListHistory(ctx context.Context, ref ContextRef, limit int) ([]*QueueItem, error) {
	current, err := r.currentPosition(ctx, ref)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT track_id FROM playback_queue_items
	WHERE context_key = $1 AND position < $2 ORDER BY position DESC LIMIT $3`, ref.Key(), current, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]*QueueItem, 0)
	for rows.Next() {
		rec := &QueueItem{Context: ref}
		if err := rows.Scan(&rec.TrackID); err != nil {
			return nil, err
		}
		history = append(history, rec)
	}
	return history, rows.Err()
}

func (r *Repository) StepNext(ctx context.Context, ref ContextRef) (currentItem *QueueItem, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = r.createStateForUserIfNotExists(ctx, tx, ref); err != nil {
		return nil, err
	}
	current, err := r.currentPositionTx(ctx, tx, ref)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(ctx, `SELECT track_id, position FROM playback_queue_items
	WHERE context_key = $1 AND position > $2 ORDER BY position ASC LIMIT 1`, ref.Key(), current)
	currentItem = &QueueItem{Context: ref}
	var position int64
	err = row.Scan(&currentItem.TrackID, &position)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEmptyQueue
	}
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `UPDATE playback_queue_state SET current_position = $1 WHERE context_key = $2`, position, ref.Key())
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return currentItem, nil
}

func (r *Repository) StepPrev(ctx context.Context, ref ContextRef) (prevItem *QueueItem, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = r.createStateForUserIfNotExists(ctx, tx, ref); err != nil {
		return nil, err
	}
	current, err := r.currentPositionTx(ctx, tx, ref)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(ctx, `SELECT track_id, position FROM playback_queue_items
	WHERE context_key = $1 AND position < $2 ORDER BY position DESC LIMIT 1`, ref.Key(), current)
	prevItem = &QueueItem{Context: ref}
	var position int64
	err = row.Scan(&prevItem.TrackID, &position)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEmptyQueue
	}
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `UPDATE playback_queue_state SET current_position = $1 WHERE context_key = $2`, position, ref.Key())
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return prevItem, nil
}

func (r *Repository) ClearQueue(ctx context.Context, ref ContextRef) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM playback_queue_items WHERE context_key = $1`, ref.Key())
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `UPDATE playback_queue_state SET current_position = 0 WHERE context_key = $1`, ref.Key())
	return err
}

func (r *Repository) CurrentTrack(ctx context.Context, ref ContextRef) (*QueueItem, error) {
	current, err := r.currentPosition(ctx, ref)
	if err != nil {
		return nil, err
	}
	if current == 0 {
		return nil, ErrEmptyQueue
	}
	row := r.db.QueryRowContext(ctx, `SELECT track_id FROM playback_queue_items WHERE context_key = $1 AND position = $2`, ref.Key(), current)
	item := &QueueItem{Context: ref}
	if err := row.Scan(&item.TrackID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEmptyQueue
		}
		return nil, err
	}
	return item, nil
}

func (r *Repository) RemoveTrack(ctx context.Context, ref ContextRef, trackID uuid.UUID) (bool, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM playback_queue_items WHERE context_key = $1 AND track_id = $2`, ref.Key(), trackID)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if count == 0 {
		_ = tx.Rollback()
		return false, nil
	}

	// Получаем текущую позицию
	current, err := r.currentPositionTx(ctx, tx, ref)
	if err != nil {
		return false, err
	}

	// Если удаляемый трек был текущим или уже проигран, обновляем current_position
	if position <= current {
		// Находим следующий трек после удаленного
		var nextPosition int64
		err = tx.QueryRowContext(ctx, `SELECT MIN(position) FROM playback_queue_items WHERE context_key = $1 AND position > $2`, ref.Key(), position).Scan(&nextPosition)
		if errors.Is(err, sql.ErrNoRows) || nextPosition == 0 {
			// Нет следующего трека, сдвигаем current_position на предыдущий
			if position == current {
				// Если удалили текущий трек, ищем предыдущий
				err = tx.QueryRowContext(ctx, `SELECT MAX(position) FROM playback_queue_items WHERE context_key = $1 AND position < $2`, ref.Key(), position).Scan(&nextPosition)
				if errors.Is(err, sql.ErrNoRows) || nextPosition == 0 {
					nextPosition = 0 // Очередь пуста
				}
			} else {
				nextPosition = current // Оставляем текущую позицию
			}
		}
		// Обновляем current_position
		_, err = tx.ExecContext(ctx, `UPDATE playback_queue_state SET current_position = $1 WHERE context_key = $2`, nextPosition, ref.Key())
		if err != nil {
			return false, err
		}
	}

	// Сдвигаем позиции всех треков после удаленного
	_, err = tx.ExecContext(ctx, `UPDATE playback_queue_items SET position = position - 1 WHERE context_key = $1 AND position > $2`, ref.Key(), position)
	if err != nil {
		return false, err
	}

	// Если current_position указывал на трек после удаленного, нужно его скорректировать
	// (но только если мы не обновили current_position выше)
	if position < current {
		// Удаляемый трек был в прошлом, current_position нужно уменьшить на 1
		_, err = tx.ExecContext(ctx, `UPDATE playback_queue_state SET current_position = current_position - 1 WHERE context_key = $1 AND current_position > 0`, ref.Key())
		if err != nil {
			return false, err
		}
	} else if position > current {
		// Удаляемый трек был в будущем, current_position не менялся
		// После сдвига позиций current_position остается корректным
		// Ничего не делаем
	}

	if err = tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func (r *Repository) currentPosition(ctx context.Context, ref ContextRef) (int64, error) {
	var pos int64
	err := r.db.QueryRowContext(ctx, `SELECT current_position FROM playback_queue_state WHERE context_key = $1`, ref.Key()).Scan(&pos)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return pos, err
}

func (r *Repository) currentPositionTx(ctx context.Context, tx *sql.Tx, ref ContextRef) (int64, error) {
	var pos int64
	err := tx.QueryRowContext(ctx, `SELECT current_position FROM playback_queue_state WHERE context_key = $1 FOR UPDATE`, ref.Key()).Scan(&pos)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return pos, err
}

func (r *Repository) createStateForUserIfNotExists(ctx context.Context, tx *sql.Tx, ref ContextRef) error {
	if strings.EqualFold(ref.ContextType, "user") {
		_, err := tx.ExecContext(ctx, `INSERT INTO playback_queue_state (context_type, context_id, context_key, current_position)
	VALUES ($1,$2,$3,0)
	ON CONFLICT (context_key) DO NOTHING`,
			ref.ContextType,
			ref.ContextID,
			ref.Key(),
		)
		return err
	}
	var exists int
	err := tx.QueryRowContext(ctx, `SELECT 1 FROM playback_queue_state WHERE context_key = $1`, ref.Key()).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrQueueNotFound
	}
	return err
}
