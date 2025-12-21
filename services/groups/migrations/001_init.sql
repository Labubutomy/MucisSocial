CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL,
    name VARCHAR(120) NOT NULL,
    invite_code VARCHAR(32) NOT NULL UNIQUE,
    queue_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS group_memberships (
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role VARCHAR(24) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

CREATE TABLE IF NOT EXISTS group_track_suggestions (
    id UUID PRIMARY KEY,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    track_id UUID NOT NULL,
    suggested_by UUID NOT NULL,
    decision_by UUID,
    decision_reason TEXT,
    status VARCHAR(24) NOT NULL DEFAULT 'pending',
    cooldown_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_group_track_pending
    ON group_track_suggestions (group_id, suggested_by)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_group_track_cooldown
    ON group_track_suggestions (group_id, track_id, suggested_by)
    WHERE status = 'rejected';
