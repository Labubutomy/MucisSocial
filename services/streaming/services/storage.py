from __future__ import annotations

from contextlib import asynccontextmanager
from typing import AsyncIterator, Callable

import anyio
from minio import Minio
from minio.error import S3Error

from services.exceptions import ObjectNotFound, StorageError


class MinioStorageService:
    """Thin asynchronous wrapper around the MinIO Python client."""

    def __init__(self, client: Minio, bucket: str) -> None:
        self._client = client
        self._bucket = bucket

    async def _wrap_minio_call(self, func: Callable[..., object], *args, **kwargs):
        try:
            return await anyio.to_thread.run_sync(func, *args, **kwargs)
        except S3Error as exc:  # pragma: no cover - relies on Minio exceptions
            if exc.code == "NoSuchKey":
                raise ObjectNotFound(exc.message) from exc
            raise StorageError(exc.message) from exc

    async def read_text(self, key: str, encoding: str = "utf-8") -> str:
        obj = await self._wrap_minio_call(self._client.get_object, self._bucket, key)
        try:
            data: bytes = await anyio.to_thread.run_sync(obj.read)
            return data.decode(encoding)
        finally:
            await anyio.to_thread.run_sync(obj.close)
            await anyio.to_thread.run_sync(obj.release_conn)

    @asynccontextmanager
    async def stream(self, key: str, chunk_size: int = 1 << 20) -> AsyncIterator[AsyncIterator[bytes]]:
        obj = await self._wrap_minio_call(self._client.get_object, self._bucket, key)

        async def iterator() -> AsyncIterator[bytes]:
            try:
                while True:
                    chunk = await anyio.to_thread.run_sync(obj.read, chunk_size)
                    if not chunk:
                        break
                    yield chunk
            finally:
                await anyio.to_thread.run_sync(obj.close)
                await anyio.to_thread.run_sync(obj.release_conn)

        try:
            yield iterator()
        except S3Error as exc:  # pragma: no cover
            if exc.code == "NoSuchKey":
                raise ObjectNotFound(exc.message) from exc
            raise StorageError(exc.message) from exc

