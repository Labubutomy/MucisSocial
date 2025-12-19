from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application configuration loaded from environment variables."""

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
    )

    app_name: str = "Routes Service"
    app_version: str = "0.1.0"

    # Database
    database_url: str = "postgres://postgres:password@localhost:5436/music_social_routes?sslmode=disable"

    # Redis
    redis_host: str = "redis"
    redis_port: int = 6379
    redis_password: str | None = None

    # JWT
    jwt_secret: str = "your-super-secret-access-key-change-in-production"

    # Server
    host: str = "0.0.0.0"
    port: int = 8080


@lru_cache
def get_settings() -> Settings:
    """Cached settings instance for dependency injection."""
    return Settings()

