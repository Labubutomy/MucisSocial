from functools import lru_cache

from fastapi import Depends
from minio import Minio

from core.config import Settings, get_settings
from services.signing import URLSigner
from services.storage import MinioStorageService
from services.streaming import StreamingService


@lru_cache
def get_minio_client() -> Minio:
    settings = get_settings()
    return Minio(
        endpoint=settings.minio_endpoint,
        access_key=settings.minio_access_key,
        secret_key=settings.minio_secret_key,
        secure=settings.minio_secure,
        region=settings.minio_region,
    )


def get_signer(settings: Settings = Depends(get_settings)) -> URLSigner:
    return URLSigner(settings.signing_secret)


def get_storage(settings: Settings = Depends(get_settings)) -> MinioStorageService:
    client = get_minio_client()
    return MinioStorageService(client, settings.minio_bucket)


def get_streaming_service(
    settings: Settings = Depends(get_settings),
    signer: URLSigner = Depends(get_signer),
) -> StreamingService:
    return StreamingService(settings, signer)

