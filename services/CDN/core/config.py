from __future__ import annotations

from functools import lru_cache
from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """CDN service configuration."""

    model_config = SettingsConfigDict(
        env_prefix="CDN_",
        case_sensitive=False,
        env_file=".env",
        extra="ignore",
    )

    # Service info
    app_name: str = Field(default="CDN Service", description="Application name")
    app_version: str = Field(default="0.1.0", description="Application version")

    # Origin configuration
    origin_base_url: str = Field(
        default="http://streaming:8000",
        description="Base URL of the Streaming Gateway origin server",
    )
    origin_api_base_url: str = Field(
        default="http://streaming:8000",
        description="Base URL used for proxying Streaming Gateway API requests",
    )

    # Cache configuration
    cache_playlist_ttl: int = Field(
        default=60,
        ge=10,
        le=300,
        description="TTL for cached playlists in seconds (default: 60)",
    )
    cache_segment_ttl: int = Field(
        default=3600,
        ge=300,
        le=86400,
        description="TTL for cached segments in seconds (default: 3600, 1 hour)",
    )
    cache_static_ttl: int = Field(
        default=86400,
        ge=300,
        le=604800,
        description="TTL for cached static files in seconds (default: 86400, 24 hours)",
    )
    cache_max_size: int = Field(
        default=1000,
        ge=100,
        description="Maximum number of items in cache (default: 1000)",
    )

    # Server configuration
    host: str = Field(default="0.0.0.0", description="Server host")
    port: int = Field(default=8080, ge=1, le=65535, description="Server port")

    # Logging
    log_level: str = Field(default="INFO", description="Logging level")
    log_requests: bool = Field(default=True, description="Log all requests")
    log_cache_stats: bool = Field(
        default=True, description="Log cache statistics (hit/miss rate)"
    )

    # Health check
    health_check_interval: int = Field(
        default=30, ge=5, description="Health check interval in seconds"
    )


@lru_cache()
def get_settings() -> Settings:
    """Get cached settings instance."""
    return Settings()

