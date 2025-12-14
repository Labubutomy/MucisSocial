from __future__ import annotations

import asyncio
import base64
import logging
import re
from typing import Any, TYPE_CHECKING

import httpx
from fastapi import Request, Response, status

from core.cache import Cache, CacheEntry
from core.config import Settings

if TYPE_CHECKING:
    from core.event_producer import ListeningEventProducer

logger = logging.getLogger(__name__)

# Pattern to extract track_id from resource path
# Expected format: /tracks/{track_id}/... or /{track_id}/...
TRACK_ID_PATTERN = re.compile(r"/(?:tracks/)?([a-f0-9-]{36})/", re.IGNORECASE)


class CDNService:
    """CDN service that caches content from origin and proxies requests."""

    def __init__(self, settings: Settings, cache: Cache):
        self._settings = settings
        self._cache = cache
        self._client = httpx.AsyncClient(
            timeout=httpx.Timeout(30.0, connect=10.0),
            follow_redirects=True,
        )
        self._event_producer: ListeningEventProducer | None = None

    def set_event_producer(self, producer: ListeningEventProducer) -> None:
        """Set the Kafka event producer for listening events."""
        self._event_producer = producer
        logger.info("Event producer set on CDN service")

    def _get_cache_ttl(self, resource_path: str) -> float:
        """Determine cache TTL based on resource type."""
        if resource_path.endswith(".m3u8"):
            return float(self._settings.cache_playlist_ttl)
        elif resource_path.endswith((".m4s", ".mp4")):
            return float(self._settings.cache_segment_ttl)
        else:
            # Static metadata and images (tech_meta.json, covers, etc.)
            return float(self._settings.cache_static_ttl)

    def _get_content_type(self, resource_path: str) -> str:
        """Determine content type based on file extension."""
        if resource_path.endswith(".m3u8"):
            return "application/vnd.apple.mpegurl"
        elif resource_path.endswith(".m4s"):
            return "video/iso.segment"
        elif resource_path.endswith(".mp4"):
            return "video/mp4"
        elif resource_path.endswith(".json"):
            return "application/json"
        else:
            return "application/octet-stream"

    def _get_resource_category(self, resource_path: str) -> str:
        """Classify resource for analytics headers."""
        lowered = resource_path.lower()
        if lowered.endswith("master.m3u8"):
            return "master_playlist"
        if lowered.endswith(".m3u8"):
            return "variant_playlist"
        if lowered.endswith("init.mp4"):
            return "init_segment"
        if lowered.endswith(".m4s"):
            return "media_segment"
        if lowered.endswith(".json"):
            return "static_asset"
        return "other"

    async def _forward_api_request(
        self,
        method: str,
        url: str,
        *,
        params: dict[str, Any] | None = None,
        json: Any = None,
    ) -> Response:
        try:
            response = await self._client.request(method, url, params=params, json=json)
        except httpx.RequestError as exc:
            logger.error("Error proxying API request to origin: %s", exc)
            return Response(
                content=b'{"detail":"CDN Error: failed to reach streaming API"}',
                status_code=status.HTTP_502_BAD_GATEWAY,
                media_type="application/json",
            )

        # Preserve JSON payload and status
        content = response.content
        media_type = response.headers.get("content-type", "application/json")
        return Response(
            content=content, status_code=response.status_code, media_type=media_type
        )

    def _extract_track_id(self, resource_path: str) -> str | None:
        """Extract track_id from resource path using regex."""
        match = TRACK_ID_PATTERN.search(resource_path)
        return match.group(1) if match else None

    def _extract_user_id(self, request: Request) -> str | None:
        """Extract user_id from request headers or query params."""
        # Check X-User-ID header first
        user_id = request.headers.get("X-User-ID")
        if user_id:
            return user_id
        # Check query params
        user_id = request.query_params.get("user_id")
        return user_id

    def _is_media_segment(self, resource_path: str) -> bool:
        """Check if resource is a media segment (triggers listening event)."""
        return resource_path.endswith(".m4s") and "init" not in resource_path.lower()

    def _extract_quality_from_path(self, resource_path: str) -> str | None:
        """Extract quality from path (e.g., /tracks/{id}/128kbps/...)."""
        quality_match = re.search(r"/(\d+k(?:bps)?)/", resource_path, re.IGNORECASE)
        if quality_match:
            return quality_match.group(1).lower()
        return None

    async def _publish_listening_event(
        self, track_id: str, user_id: str, resource_path: str
    ) -> None:
        """Publish track_listened event to Kafka (async, non-blocking)."""
        if not self._event_producer or not self._event_producer.is_ready:
            return

        quality = self._extract_quality_from_path(resource_path)

        # Fire and forget - don't block response
        asyncio.create_task(
            self._event_producer.publish_track_listened(
                user_id=user_id,
                track_id=track_id,
                quality=quality,
                source="cdn",
            )
        )

    async def serve_resource(self, resource_path: str, request: Request) -> Response:
        """
        Serve resource from cache or proxy to origin.
        Preserves original signed URL when proxying to origin.
        Publishes listening events for media segments.
        """
        # Ensure resource_path starts with /
        if not resource_path.startswith("/"):
            resource_path = f"/{resource_path}"

        # Build full URL with original query parameters (including sig and exp)
        original_url = str(request.url)
        cache_key_url = original_url

        # Extract track_id and user_id for potential listening event
        track_id = self._extract_track_id(resource_path)
        user_id = self._extract_user_id(request)
        should_publish_event = (
            self._is_media_segment(resource_path)
            and track_id is not None
            and user_id is not None
        )

        # Try to get from cache
        cached_entry = self._cache.get(cache_key_url)
        if cached_entry:
            if self._settings.log_requests:
                logger.info(f"Cache HIT: {resource_path}")
            metadata = cached_entry.as_metadata()
            ttl_remaining = max(0, int(metadata["ttl_remaining"]))

            # Publish listening event for media segments
            if should_publish_event:
                await self._publish_listening_event(track_id, user_id, resource_path)

            return Response(
                content=cached_entry.content,
                media_type=cached_entry.content_type,
                headers={
                    "Cache-Control": f"public, max-age={max(ttl_remaining, 0)}",
                    "X-CDN-Cache": "HIT",
                    "X-CDN-TTL-Remaining": str(ttl_remaining),
                    "X-CDN-Resource": metadata["resource"],
                    "X-CDN-Resource-Type": metadata["resource_type"],
                    "X-CDN-Hit-Count": str(metadata["hit_count"]),
                },
            )

        # Cache miss - proxy to origin
        if self._settings.log_requests:
            logger.info(f"Cache MISS: {resource_path}")

        # Build origin URL with original signed parameters
        origin_path = request.url.path
        origin_url = f"{self._settings.origin_base_url.rstrip('/')}{origin_path}"
        # Preserve original query parameters (sig, exp, etc.)
        if request.url.query:
            origin_url = f"{origin_url}?{request.url.query}"

        try:
            # Forward request to origin with original signed URL
            async with self._client.stream("GET", origin_url) as response:
                if response.status_code != status.HTTP_200_OK:
                    # Don't cache errors
                    content = await response.aread()
                    return Response(
                        content=content,
                        status_code=response.status_code,
                        headers=dict(response.headers),
                    )

                # Read response content
                content = await response.aread()
                content_type = response.headers.get(
                    "content-type", self._get_content_type(resource_path)
                )

                # Cache the response
                ttl = self._get_cache_ttl(resource_path)
                self._cache.set(cache_key_url, content, content_type, ttl)

                resource_category = self._get_resource_category(resource_path)
                ttl_int = int(ttl)
                resource_descriptor = origin_path

                # Publish listening event for media segments
                if should_publish_event:
                    await self._publish_listening_event(
                        track_id, user_id, resource_path
                    )

                return Response(
                    content=content,
                    media_type=content_type,
                    headers={
                        "Cache-Control": f"public, max-age={max(ttl_int, 0)}",
                        "X-CDN-Cache": "MISS",
                        "X-CDN-TTL": str(ttl_int),
                        "X-CDN-Resource": resource_descriptor,
                        "X-CDN-Resource-Type": resource_category,
                    },
                )

        except httpx.RequestError as e:
            logger.error(f"Error proxying to origin: {e}")
            return Response(
                content=f"CDN Error: Failed to fetch from origin: {str(e)}".encode(),
                status_code=status.HTTP_502_BAD_GATEWAY,
                media_type="text/plain",
            )

    async def close(self) -> None:
        """Close HTTP client."""
        await self._client.aclose()

    def list_cache_entries(self) -> list[dict[str, Any]]:
        """List metadata for cached entries."""
        return self._cache.list_entries()

    def get_cache_entry(
        self, cache_id: str, include_content: bool = False
    ) -> dict[str, Any] | None:
        """Return metadata (and optional preview) for a specific cache entry."""
        entry = self._cache.get_entry(cache_id)
        if entry is None:
            return None

        metadata = entry.as_metadata()
        if include_content:
            preview_length = min(512, len(entry.content))
            metadata.update(
                {
                    "content_preview_base64": base64.b64encode(
                        entry.content[:preview_length]
                    ).decode("utf-8"),
                    "content_preview_bytes": preview_length,
                    "content_total_bytes": len(entry.content),
                }
            )
        return metadata

    def get_cache_summary(self) -> dict[str, Any]:
        """Aggregate cache metadata for analytics dashboards."""
        entries = self._cache.list_entries()
        by_type: dict[str, dict[str, Any]] = {}
        total_bytes = 0
        for entry in entries:
            resource_type = entry["resource_type"]
            entry_bytes = entry["size_bytes"]
            total_bytes += entry_bytes
            type_stats = by_type.setdefault(
                resource_type,
                {"count": 0, "bytes": 0, "mb": 0.0, "avg_ttl_remaining": 0.0},
            )
            type_stats["count"] += 1
            type_stats["bytes"] += entry_bytes
            type_stats["mb"] = round(type_stats["bytes"] / (1024 * 1024), 2)
            type_stats["avg_ttl_remaining"] += entry["ttl_remaining"]

        for stats in by_type.values():
            if stats["count"] > 0:
                stats["avg_ttl_remaining"] = round(
                    stats["avg_ttl_remaining"] / stats["count"], 2
                )

        return {
            "total_entries": len(entries),
            "total_bytes": total_bytes,
            "total_mb": round(total_bytes / (1024 * 1024), 2),
            "by_type": by_type,
        }

    async def fetch_stream_metadata(
        self, track_id: str, query_params: dict[str, Any]
    ) -> Response:
        url = f"{self._settings.origin_api_base_url.rstrip('/')}/api/stream/{track_id}"
        return await self._forward_api_request("GET", url, params=query_params)

    async def refresh_stream_metadata(self, payload: dict[str, Any]) -> Response:
        url = f"{self._settings.origin_api_base_url.rstrip('/')}/api/stream/refresh"
        return await self._forward_api_request("POST", url, json=payload)
