CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Request sessions table
CREATE TABLE IF NOT EXISTS request_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    artist_id UUID NOT NULL,
    session_code VARCHAR(50) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_request_sessions_artist_id ON request_sessions(artist_id);
CREATE INDEX IF NOT EXISTS idx_request_sessions_session_code ON request_sessions(session_code);
CREATE INDEX IF NOT EXISTS idx_request_sessions_is_active ON request_sessions(is_active);

-- Track requests table
CREATE TABLE IF NOT EXISTS track_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES request_sessions(id) ON DELETE CASCADE,
    requester_id UUID NOT NULL,
    artist_id UUID NOT NULL,
    track_id UUID NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' NOT NULL,
    message VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_track_requests_session_id ON track_requests(session_id);
CREATE INDEX IF NOT EXISTS idx_track_requests_requester_id ON track_requests(requester_id);
CREATE INDEX IF NOT EXISTS idx_track_requests_artist_id ON track_requests(artist_id);
CREATE INDEX IF NOT EXISTS idx_track_requests_track_id ON track_requests(track_id);
CREATE INDEX IF NOT EXISTS idx_track_requests_status ON track_requests(status);

-- Coin transactions table
CREATE TABLE IF NOT EXISTS coin_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_user_id UUID NOT NULL,
    to_user_id UUID NOT NULL,
    amount INTEGER NOT NULL,
    request_id UUID UNIQUE NOT NULL,
    transaction_type VARCHAR(20) DEFAULT 'music_request' NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_coin_transactions_from_user_id ON coin_transactions(from_user_id);
CREATE INDEX IF NOT EXISTS idx_coin_transactions_to_user_id ON coin_transactions(to_user_id);
CREATE INDEX IF NOT EXISTS idx_coin_transactions_request_id ON coin_transactions(request_id);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to automatically update updated_at
DROP TRIGGER IF EXISTS update_request_sessions_updated_at ON request_sessions;
CREATE TRIGGER update_request_sessions_updated_at BEFORE UPDATE ON request_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_track_requests_updated_at ON track_requests;
CREATE TRIGGER update_track_requests_updated_at BEFORE UPDATE ON track_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
