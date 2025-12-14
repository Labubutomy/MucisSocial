CREATE TABLE IF NOT EXISTS playback_queue_state (
    context_type TEXT NOT NULL,
    context_id UUID NOT NULL,
    context_key TEXT PRIMARY KEY,
    current_position BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS playback_queue_items (
    context_key TEXT NOT NULL REFERENCES playback_queue_state(context_key) ON DELETE CASCADE,
    track_id UUID NOT NULL,
    position BIGINT NOT NULL,
    PRIMARY KEY (context_key, position)
);