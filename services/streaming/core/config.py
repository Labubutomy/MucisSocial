from functools import lru_cache

from pydantic import AnyHttpUrl, Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application configuration loaded from environment variables."""

    app_name: str = "Streaming Gateway"
    app_version: str = "0.1.0"

    base_url: AnyHttpUrl = Field(
        default="http://localhost:8000", description="Base URL of this service (used for streaming URLs)"
    )
    cdn_base_url: AnyHttpUrl | None = Field(
        default=None, description="Optional CDN base URL. If not set, uses base_url instead"
    )
    signing_secret: str = Field(default="change-me", min_length=8, description="Secret used for URL signing")

    playlist_ttl_seconds: int = Field(default=300, ge=60, le=3600, description="TTL for playlists")
    segment_ttl_seconds: int = Field(default=60, ge=10, le=600, description="TTL for media segments and init files")

    available_bitrates: tuple[int, ...] = Field(default=(256_000, 160_000, 96_000), description="Supported audio bitrates")

    minio_endpoint: str = Field(default="localhost:9000", description="MinIO/S3 endpoint host:port")
    minio_access_key: str = Field(default="minioadmin", description="MinIO access key")
    minio_secret_key: str = Field(default="minioadmin", description="MinIO secret key")
    minio_bucket: str = Field(default="tracks", description="MinIO bucket with track assets")
    minio_secure: bool = Field(default=False, description="Use HTTPS when talking to MinIO")
    minio_region: str | None = Field(default=None, description="Optional MinIO/S3 region")


@lru_cache
def get_settings() -> Settings:
    """Cached settings instance for dependency injection."""

    return Settings()

