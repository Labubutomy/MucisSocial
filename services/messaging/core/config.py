from __future__ import annotations

from functools import lru_cache

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Messaging service configuration."""

    model_config = SettingsConfigDict(
        env_prefix="MESSAGING_",
        case_sensitive=False,
        env_file=".env",
        extra="ignore",
    )

    # Service
    app_name: str = Field(default="Messaging Service", description="Application name")
    app_version: str = Field(default="0.1.0", description="Application version")

    host: str = Field(default="0.0.0.0", description="HTTP host")
    port: int = Field(default=8080, ge=1, le=65535, description="HTTP port")

    # Database
    database_url: str = Field(
        default="postgres://postgres:password@postgres-messaging:5432/music_social_messaging?sslmode=disable",
        description="SQLAlchemy-compatible PostgreSQL URL",
    )

    # JWT
    jwt_secret: str = Field(
        default="your-super-secret-access-key-change-in-production",
        description="JWT secret key (must be same as API gateway / ws gateway)",
    )
    jwt_algorithm: str = Field(default="HS256", description="JWT algorithm")

    # Kafka
    kafka_brokers: str = Field(
        default="redpanda:9092",
        description="Kafka brokers (comma-separated, usually Redpanda)",
    )
    kafka_events_topic: str = Field(
        default="messaging-events",
        description="Topic for messaging domain events (message_sent, message_read, etc.)",
    )


@lru_cache()
def get_settings() -> Settings:
    return Settings()


