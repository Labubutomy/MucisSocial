CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Artists table
CREATE TABLE IF NOT EXISTS artists (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    avatar_url TEXT,
    genres TEXT[] DEFAULT '{}',
    followers BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for artists table
CREATE INDEX IF NOT EXISTS idx_artists_name ON artists(name);
CREATE INDEX IF NOT EXISTS idx_artists_followers ON artists(followers DESC);
CREATE INDEX IF NOT EXISTS idx_artists_created_at ON artists(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_artists_genres ON artists USING GIN(genres);

-- Create a function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for artists table
DROP TRIGGER IF EXISTS update_artists_updated_at ON artists;
CREATE TRIGGER update_artists_updated_at 
    BEFORE UPDATE ON artists 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Sample data removed - use scripts/faker/create_artists_and_tracks.py to populate data