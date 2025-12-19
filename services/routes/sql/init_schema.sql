-- Music Social Routes Database Schema
-- This script creates all tables for the routes feature

-- Table: routes
CREATE TABLE IF NOT EXISTS routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    city VARCHAR(100),
    country VARCHAR(100),
    is_public BOOLEAN DEFAULT true,
    is_linear BOOLEAN DEFAULT true,
    total_distance_km DECIMAL(8,2),
    estimated_minutes INT,
    cover_image_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_routes_user_id ON routes(user_id);
CREATE INDEX IF NOT EXISTS idx_routes_city ON routes(city);
CREATE INDEX IF NOT EXISTS idx_routes_public ON routes(is_public) WHERE is_public = true;
CREATE INDEX IF NOT EXISTS idx_routes_created_at ON routes(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_routes_deleted_at ON routes(deleted_at) WHERE deleted_at IS NULL;

-- Table: route_points
CREATE TABLE IF NOT EXISTS route_points (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    geohash CHAR(9) NOT NULL,
    radius_meters INT DEFAULT 50,
    track_id UUID NOT NULL,
    track_start_offset_sec INT DEFAULT 0,
    order_index INT NOT NULL,
    title VARCHAR(100),
    description TEXT,
    image_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT route_points_route_order_unique UNIQUE (route_id, order_index)
);

CREATE INDEX IF NOT EXISTS idx_route_points_route_id ON route_points(route_id);
CREATE INDEX IF NOT EXISTS idx_route_points_geohash ON route_points(geohash);
CREATE INDEX IF NOT EXISTS idx_route_points_location ON route_points(latitude, longitude);

-- Table: route_tags
CREATE TABLE IF NOT EXISTS route_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_route_tags_name ON route_tags(name);

-- Table: route_tag_relations
CREATE TABLE IF NOT EXISTS route_tag_relations (
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES route_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (route_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_route_tag_relations_route ON route_tag_relations(route_id);
CREATE INDEX IF NOT EXISTS idx_route_tag_relations_tag ON route_tag_relations(tag_id);

-- Table: route_sessions
CREATE TABLE IF NOT EXISTS route_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    started_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    current_point_index INT DEFAULT 0,
    visited_points JSONB DEFAULT '[]'::jsonb,
    total_distance_km DECIMAL(8,2),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT route_sessions_status_check CHECK (status IN ('active', 'paused', 'completed', 'abandoned'))
);

CREATE INDEX IF NOT EXISTS idx_route_sessions_user_id ON route_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_route_sessions_route_id ON route_sessions(route_id);
CREATE INDEX IF NOT EXISTS idx_route_sessions_status ON route_sessions(status) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_route_sessions_started_at ON route_sessions(started_at DESC);

-- Table: route_favorites
CREATE TABLE IF NOT EXISTS route_favorites (
    user_id UUID NOT NULL,
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, route_id)
);

CREATE INDEX IF NOT EXISTS idx_route_favorites_user ON route_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_route_favorites_route ON route_favorites(route_id);

-- Table: route_ratings
CREATE TABLE IF NOT EXISTS route_ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    review TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT route_ratings_user_route_unique UNIQUE (user_id, route_id)
);

CREATE INDEX IF NOT EXISTS idx_route_ratings_route_id ON route_ratings(route_id);
CREATE INDEX IF NOT EXISTS idx_route_ratings_rating ON route_ratings(rating);

-- Table: route_stats
CREATE TABLE IF NOT EXISTS route_stats (
    route_id UUID PRIMARY KEY REFERENCES routes(id) ON DELETE CASCADE,
    total_completions INT DEFAULT 0,
    total_starts INT DEFAULT 0,
    average_rating DECIMAL(3,2),
    total_favorites INT DEFAULT 0,
    last_completed_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_route_stats_rating ON route_stats(average_rating DESC);
CREATE INDEX IF NOT EXISTS idx_route_stats_completions ON route_stats(total_completions DESC);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_routes_updated_at BEFORE UPDATE ON routes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_route_sessions_updated_at BEFORE UPDATE ON route_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_route_ratings_updated_at BEFORE UPDATE ON route_ratings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_route_stats_updated_at BEFORE UPDATE ON route_stats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

