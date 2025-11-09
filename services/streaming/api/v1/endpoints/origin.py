from __future__ import annotations

import mimetypes
from fastapi import APIRouter, Depends, HTTPException, Request, Response, status
from starlette.responses import PlainTextResponse, StreamingResponse

from core.config import Settings, get_settings
from core.dependencies import get_signer, get_storage
from services.exceptions import ObjectNotFound
from services.signing import URLSigner
from services.storage import MinioStorageService
from utils.playlists import rewrite_master_playlist, rewrite_variant_playlist

router = APIRouter(tags=["origin"])


@router.get("/origin/{resource_path:path}")
async def serve_resource(
    resource_path: str,
    request: Request,
    signer: URLSigner = Depends(get_signer),
    storage: MinioStorageService = Depends(get_storage),
    settings: Settings = Depends(get_settings),
) -> Response:
    signature = request.query_params.get("sig")
    expires_at = request.query_params.get("exp")

    if not signature or not expires_at:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Missing signature parameters")

    try:
        expires_int = int(expires_at)
    except ValueError:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Invalid expiration value")

    signed_path = f"/{resource_path}"
    if not signer.verify(signed_path, expires_int, signature):
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Signature verification failed")

    object_key = resource_path

    try:
        if resource_path.endswith(".m3u8"):
            return await _serve_playlist(resource_path, object_key, signer, storage, settings)
        return await _serve_binary(resource_path, object_key, storage)
    except ObjectNotFound as exc:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(exc))


async def _serve_playlist(
    resource_path: str,
    object_key: str,
    signer: URLSigner,
    storage: MinioStorageService,
    settings: Settings,
) -> PlainTextResponse:
    body = await storage.read_text(object_key)
    full_resource_path = f"/{resource_path}"

    if resource_path.endswith("master.m3u8"):
        rewritten = rewrite_master_playlist(body, full_resource_path, signer, settings)
    else:
        rewritten = rewrite_variant_playlist(body, full_resource_path, signer, settings)

    return PlainTextResponse(rewritten, media_type="application/vnd.apple.mpegurl")


async def _serve_binary(resource_path: str, object_key: str, storage: MinioStorageService) -> StreamingResponse:
    media_type = _guess_media_type(resource_path)
    stream_ctx = storage.stream(object_key)
    iterator = await stream_ctx.__aenter__()

    async def streaming_iterator():
        try:
            async for chunk in iterator:
                yield chunk
        finally:
            await stream_ctx.__aexit__(None, None, None)

    return StreamingResponse(streaming_iterator(), media_type=media_type)


def _guess_media_type(path: str) -> str:
    if path.endswith(".m4s"):
        return "video/iso.segment"
    if path.endswith(".mp4"):
        return "video/mp4"
    if path.endswith(".json"):
        return "application/json"
    if path.endswith(".m3u8"):
        return "application/vnd.apple.mpegurl"
    guess, _ = mimetypes.guess_type(path)
    return guess or "application/octet-stream"

