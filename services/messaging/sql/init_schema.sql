CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Conversations table
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_type VARCHAR(20) NOT NULL DEFAULT 'direct', -- 'direct' | 'group'
    title VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_message_at TIMESTAMPTZ,
    last_message_id UUID
);

CREATE INDEX IF NOT EXISTS idx_conversations_updated_at
    ON conversations(updated_at DESC);

-- User <-> Conversation link
CREATE TABLE IF NOT EXISTS user_conversations (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    unread_count INT NOT NULL DEFAULT 0,
    last_read_message_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (conversation_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_conversations_user_id
    ON user_conversations(user_id);

CREATE INDEX IF NOT EXISTS idx_user_conversations_unread
    ON user_conversations(user_id, unread_count)
    WHERE unread_count > 0;

-- Messages
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_created
    ON messages(conversation_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_messages_sender
    ON messages(sender_id);

-- Message <-> Track
CREATE TABLE IF NOT EXISTS message_tracks (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    track_id UUID NOT NULL,
    PRIMARY KEY (message_id, track_id)
);

CREATE INDEX IF NOT EXISTS idx_message_tracks_track_id
    ON message_tracks(track_id);

-- Message contexts (playlist/session/route/track etc.)
CREATE TABLE IF NOT EXISTS message_contexts (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    context_type VARCHAR(50) NOT NULL,
    context_id UUID NOT NULL,
    PRIMARY KEY (message_id, context_type, context_id)
);

CREATE INDEX IF NOT EXISTS idx_message_contexts_context
    ON message_contexts(context_type, context_id);

-- Per-user read status (for fine grained read receipts)
CREATE TABLE IF NOT EXISTS message_read_status (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (message_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_message_read_status_user
    ON message_read_status(user_id);


