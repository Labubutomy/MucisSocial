-- Drop triggers
DROP TRIGGER IF EXISTS update_track_requests_updated_at ON track_requests;
DROP TRIGGER IF EXISTS update_request_sessions_updated_at ON request_sessions;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS coin_transactions;
DROP TABLE IF EXISTS track_requests;
DROP TABLE IF EXISTS request_sessions;
