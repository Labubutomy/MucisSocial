-- Drop trigger and function
DROP TRIGGER IF EXISTS update_artists_updated_at ON artists;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_artists_genres;
DROP INDEX IF EXISTS idx_artists_created_at;
DROP INDEX IF EXISTS idx_artists_followers;
DROP INDEX IF EXISTS idx_artists_name;

-- Drop artists table
DROP TABLE IF EXISTS artists;