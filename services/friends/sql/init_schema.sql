CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Friend requests between users
CREATE TABLE IF NOT EXISTS friend_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_user_id UUID NOT NULL,
    to_user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending | accepted | declined | cancelled
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_friend_request_from_to UNIQUE (from_user_id, to_user_id)
);

CREATE INDEX IF NOT EXISTS idx_friend_requests_to_user
    ON friend_requests(to_user_id);

CREATE INDEX IF NOT EXISTS idx_friend_requests_from_user
    ON friend_requests(from_user_id);

-- Friends relation (undirected, stored as two directed rows for simplicity)
CREATE TABLE IF NOT EXISTS friends (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    friend_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_friend_pair UNIQUE (user_id, friend_id)
);

CREATE INDEX IF NOT EXISTS idx_friends_user
    ON friends(user_id);


