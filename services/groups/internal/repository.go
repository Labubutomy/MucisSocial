package internal

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateGroup(ctx context.Context, group *Group) error {
	query := `INSERT INTO groups (id, owner_id, name, invite_code, queue_id, created_at, updated_at)
              VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.db.ExecContext(ctx, query,
		group.ID,
		group.OwnerID,
		group.Name,
		group.InviteCode,
		group.QueueID,
		group.CreatedAt,
		group.UpdatedAt,
	)
	return err
}

func (r *Repository) GetGroupByInviteCode(ctx context.Context, code string) (*Group, error) {
	query := `SELECT id, owner_id, name, invite_code, queue_id, created_at, updated_at
              FROM groups WHERE invite_code = $1`
	group := &Group{}
	var queueID sql.NullString
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&group.ID,
		&group.OwnerID,
		&group.Name,
		&group.InviteCode,
		&queueID,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInviteCodeInvalid
	}
	if queueID.Valid {
		if id, parseErr := uuid.Parse(queueID.String); parseErr == nil {
			group.QueueID = &id
		}
	}
	return group, err
}

func (r *Repository) GetGroupByID(ctx context.Context, id uuid.UUID) (*Group, error) {
	query := `SELECT id, owner_id, name, invite_code, queue_id, created_at, updated_at FROM groups WHERE id = $1`
	group := &Group{}
	var queueID sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&group.ID,
		&group.OwnerID,
		&group.Name,
		&group.InviteCode,
		&queueID,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if queueID.Valid {
		if parsed, parseErr := uuid.Parse(queueID.String); parseErr == nil {
			group.QueueID = &parsed
		}
	}
	return group, err
}

func (r *Repository) AddMembership(ctx context.Context, membership Membership) error {
	query := `INSERT INTO group_memberships (group_id, user_id, role, joined_at)
              VALUES ($1,$2,$3,$4)
              ON CONFLICT (group_id, user_id) DO NOTHING`
	res, err := r.db.ExecContext(ctx, query,
		membership.GroupID,
		membership.UserID,
		membership.Role,
		membership.Joined,
	)
	if err != nil {
		return err
	}
	rows, rowsErr := res.RowsAffected()
	if rowsErr != nil {
		return rowsErr
	}
	if rows == 0 {
		return ErrAlreadyMember
	}
	return nil
}

func (r *Repository) RemoveMembership(ctx context.Context, groupID, userID uuid.UUID) error {
	query := `DELETE FROM group_memberships WHERE group_id = $1 AND user_id = $2`
	res, err := r.db.ExecContext(ctx, query, groupID, userID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotMember
	}
	return nil
}

func (r *Repository) IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	query := `SELECT 1 FROM group_memberships WHERE group_id = $1 AND user_id = $2`
	var exists int
	err := r.db.QueryRowContext(ctx, query, groupID, userID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *Repository) CountPendingSuggestions(ctx context.Context, groupID, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM group_track_suggestions
              WHERE group_id = $1 AND suggested_by = $2 AND status = 'pending'`
	var count int
	err := r.db.QueryRowContext(ctx, query, groupID, userID).Scan(&count)
	return count, err
}

func (r *Repository) HasActiveCooldown(ctx context.Context, groupID, userID, trackID uuid.UUID, now time.Time) (bool, *time.Time, error) {
	query := `SELECT cooldown_until FROM group_track_suggestions
              WHERE group_id = $1 AND suggested_by = $2 AND track_id = $3
                AND status = 'rejected' AND cooldown_until IS NOT NULL
              ORDER BY cooldown_until DESC LIMIT 1`
	var cooldown sql.NullTime
	err := r.db.QueryRowContext(ctx, query, groupID, userID, trackID).Scan(&cooldown)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}
	if cooldown.Valid && cooldown.Time.After(now) {
		return true, &cooldown.Time, nil
	}
	return false, nil, nil
}

func (r *Repository) CreateSuggestion(ctx context.Context, suggestion *TrackSuggestion) error {
	query := `INSERT INTO group_track_suggestions
                (id, group_id, track_id, suggested_by, status, created_at, updated_at)
              VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.db.ExecContext(ctx, query,
		suggestion.ID,
		suggestion.GroupID,
		suggestion.TrackID,
		suggestion.SuggestedBy,
		suggestion.Status,
		suggestion.CreatedAt,
		suggestion.UpdatedAt,
	)
	return err
}

func (r *Repository) GetSuggestionByID(ctx context.Context, id uuid.UUID) (*TrackSuggestion, error) {
	query := `SELECT id, group_id, track_id, suggested_by, status, decision_by, decision_reason, cooldown_until, created_at, updated_at
              FROM group_track_suggestions WHERE id = $1`
	suggestion := &TrackSuggestion{}
	var decisionBy sql.NullString
	var decisionReason sql.NullString
	var cooldown sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&suggestion.ID,
		&suggestion.GroupID,
		&suggestion.TrackID,
		&suggestion.SuggestedBy,
		&suggestion.Status,
		&decisionBy,
		&decisionReason,
		&cooldown,
		&suggestion.CreatedAt,
		&suggestion.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if decisionBy.Valid {
		id, parseErr := uuid.Parse(decisionBy.String)
		if parseErr == nil {
			suggestion.DecisionBy = &id
		}
	}
	if decisionReason.Valid {
		reason := decisionReason.String
		suggestion.DecisionReason = &reason
	}
	if cooldown.Valid {
		t := cooldown.Time
		suggestion.CooldownUntil = &t
	}
	return suggestion, nil
}

func (r *Repository) UpdateSuggestionStatus(ctx context.Context, suggestionID uuid.UUID, status string, decisionBy uuid.UUID, decisionReason *string, cooldownUntil *time.Time) error {
	query := `UPDATE group_track_suggestions
              SET status = $1, decision_by = $2, decision_reason = $3, cooldown_until = $4, updated_at = $5
              WHERE id = $6`
	var decisionByPtr interface{}
	if decisionBy == uuid.Nil {
		decisionByPtr = nil
	} else {
		decisionByPtr = decisionBy
	}
	var cooldown interface{}
	if cooldownUntil != nil {
		cooldown = *cooldownUntil
	}
	_, err := r.db.ExecContext(ctx, query,
		status,
		decisionByPtr,
		decisionReason,
		cooldown,
		time.Now().UTC(),
		suggestionID,
	)
	return err
}

func (r *Repository) ListSuggestions(ctx context.Context, groupID uuid.UUID, status string, limit, offset int) ([]*TrackSuggestion, error) {
	query := `SELECT id, group_id, track_id, suggested_by, status, decision_by, decision_reason, cooldown_until, created_at, updated_at
              FROM group_track_suggestions WHERE group_id = $1`
	args := []interface{}{groupID}
	argPos := 2
	if status != "" {
		query += ` AND status = $` + strconv.Itoa(argPos)
		args = append(args, status)
		argPos++
	}
	query += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argPos) + ` OFFSET $` + strconv.Itoa(argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*TrackSuggestion, 0)
	for rows.Next() {
		suggestion := &TrackSuggestion{}
		var decisionBy sql.NullString
		var decisionReason sql.NullString
		var cooldown sql.NullTime
		if err := rows.Scan(
			&suggestion.ID,
			&suggestion.GroupID,
			&suggestion.TrackID,
			&suggestion.SuggestedBy,
			&suggestion.Status,
			&decisionBy,
			&decisionReason,
			&cooldown,
			&suggestion.CreatedAt,
			&suggestion.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if decisionBy.Valid {
			id, parseErr := uuid.Parse(decisionBy.String)
			if parseErr == nil {
				suggestion.DecisionBy = &id
			}
		}
		if decisionReason.Valid {
			reason := decisionReason.String
			suggestion.DecisionReason = &reason
		}
		if cooldown.Valid {
			t := cooldown.Time
			suggestion.CooldownUntil = &t
		}
		results = append(results, suggestion)
	}

	return results, nil
}

func (r *Repository) GroupOwnerID(ctx context.Context, groupID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT owner_id FROM groups WHERE id = $1`
	var owner uuid.UUID
	err := r.db.QueryRowContext(ctx, query, groupID).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return owner, err
}

func (r *Repository) InviteCodeExists(ctx context.Context, code string) (bool, error) {
	query := `SELECT 1 FROM groups WHERE invite_code = $1`
	var exists int
	err := r.db.QueryRowContext(ctx, query, code).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *Repository) DeleteGroup(ctx context.Context, groupID uuid.UUID) error {
	query := `DELETE FROM groups WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, groupID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
