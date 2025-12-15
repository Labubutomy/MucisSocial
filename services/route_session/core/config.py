from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application configuration loaded from environment variables."""

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
    )

    app_name: str = "Route Session Service"
    app_version: str = "0.1.0"

    # Database
    database_url: str = "postgres://postgres:password@localhost:5436/music_social_routes?sslmode=disable"

    # Redis
    redis_host: str = "redis"
    redis_port: int = 6379
    redis_password: str | None = None

    # Kafka
    kafka_brokers: str = "redpanda:9092"
    kafka_events_topic: str = "route-session-events"
    kafka_sync_topic: str = "route-session-sync"

    # Routes Service
    routes_service_url: str = "http://routes-service:8080"

    # Server
    host: str = "0.0.0.0"
    port: int = 8080


@lru_cache
def get_settings() -> Settings:
    """Cached settings instance for dependency injection."""
    return Settings()

