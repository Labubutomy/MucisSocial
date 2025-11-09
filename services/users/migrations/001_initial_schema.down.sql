-- Drop triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS idx_refresh_tokens_token;
DROP INDEX IF EXISTS idx_search_history_created_at;
DROP INDEX IF EXISTS idx_search_history_user_id;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS search_history;
DROP TABLE IF EXISTS music_taste_artists;
DROP TABLE IF EXISTS music_taste_genres;
DROP TABLE IF EXISTS users;

-- Drop extension (optional, might be used by other schemas)
-- DROP EXTENSION IF EXISTS "uuid-ossp";