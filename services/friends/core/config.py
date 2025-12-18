from __future__ import annotations

from functools import lru_cache

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Friends service configuration."""

    model_config = SettingsConfigDict(
        env_prefix="FRIENDS_",
        case_sensitive=False,
        env_file=".env",
        extra="ignore",
    )

    app_name: str = Field(default="Friends Service", description="Application name")
    app_version: str = Field(default="0.1.0", description="Application version")

    host: str = Field(default="0.0.0.0", description="HTTP host")
    port: int = Field(default=8080, ge=1, le=65535, description="HTTP port")

    database_url: str = Field(
        default="postgres://postgres:password@postgres-friends:5432/music_social_friends?sslmode=disable",
        description="SQLAlchemy-compatible PostgreSQL URL",
    )

    jwt_secret: str = Field(
        default="your-super-secret-access-key-change-in-production",
        description="JWT secret key (same as API gateway)",
    )
    jwt_algorithm: str = Field(default="HS256", description="JWT algorithm")
    
    gateway_url: str = Field(
        default="http://api-gateway:8080",
        description="API Gateway URL for fetching user information",
    )


@lru_cache()
def get_settings() -> Settings:
    return Settings()


