-- Remove coins column from users table
DROP INDEX IF EXISTS idx_users_coins;
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_coins_non_negative;
ALTER TABLE users DROP COLUMN IF EXISTS coins;
