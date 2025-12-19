from __future__ import annotations

from functools import lru_cache
from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """WebSocket Gateway service configuration."""

    model_config = SettingsConfigDict(
        env_prefix="WS_GATEWAY_",
        case_sensitive=False,
        env_file=".env",
        extra="ignore",
    )

    # Service info
    app_name: str = Field(default="WebSocket Gateway", description="Application name")
    app_version: str = Field(default="0.1.0", description="Application version")

    # Server configuration
    host: str = Field(default="0.0.0.0", description="Server host")
    port: int = Field(default=8001, ge=1, le=65535, description="Server port")

    # JWT configuration
    # Supports both JWT_SECRET (from gateway) and WS_GATEWAY_JWT_SECRET
    jwt_secret: str = Field(
        default_factory=lambda: __import__("os").getenv(
            "JWT_SECRET",
            __import__("os").getenv(
                "WS_GATEWAY_JWT_SECRET",
                "your-super-secret-access-key-change-in-production"
            )
        ),
        description="JWT secret key for token validation (uses JWT_SECRET or WS_GATEWAY_JWT_SECRET)",
    )
    jwt_algorithm: str = Field(default="HS256", description="JWT algorithm")

    # Redis configuration
    redis_host: str = Field(default="redis", description="Redis host")
    redis_port: int = Field(default=6379, ge=1, le=65535, description="Redis port")
    redis_db: int = Field(default=0, ge=0, description="Redis database number")
    redis_password: str | None = Field(default=None, description="Redis password")

    # Kafka configuration
    kafka_brokers: str = Field(
        default="redpanda:9092", description="Kafka brokers (comma-separated)"
    )
    kafka_events_topic: str = Field(
        default="music-session-events", description="Topic for client events"
    )
    kafka_sync_topic: str = Field(
        default="music-session-sync", description="Topic for sync messages"
    )
    kafka_messaging_topic: str = Field(
        default="messaging-events", description="Topic for messaging events"
    )

    # WebSocket configuration
    ping_interval: int = Field(
        default=30, ge=5, description="WebSocket ping interval in seconds"
    )
    ping_timeout: int = Field(
        default=10, ge=1, description="WebSocket ping timeout in seconds"
    )
    connection_timeout: int = Field(
        default=300, ge=60, description="Connection timeout in seconds"
    )

    # Logging
    log_level: str = Field(default="INFO", description="Logging level")


@lru_cache()
def get_settings() -> Settings:
    """Get cached settings instance."""
    return Settings()

