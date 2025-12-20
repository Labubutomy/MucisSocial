-- Add coins column to users table
ALTER TABLE users ADD COLUMN coins INTEGER DEFAULT 100 NOT NULL;

-- Create index for coins column (for queries like top users by coins)
CREATE INDEX IF NOT EXISTS idx_users_coins ON users(coins DESC);

-- Add constraint to prevent negative coins
ALTER TABLE users ADD CONSTRAINT check_coins_non_negative CHECK (coins >= 0);
