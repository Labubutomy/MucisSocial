from __future__ import annotations

import hashlib
import time
from collections import OrderedDict
from dataclasses import dataclass
from typing import Any
from urllib.parse import parse_qs, urlencode, urlparse, urlunparse


def _categorize_resource(resource: str) -> str:
    resource_lower = resource.lower()
    if resource_lower.endswith("master.m3u8"):
        return "master_playlist"
    if resource_lower.endswith(".m3u8"):
        return "variant_playlist"
    if resource_lower.endswith("init.mp4"):
        return "init_segment"
    if resource_lower.endswith(".m4s"):
        return "media_segment"
    if resource_lower.endswith(".json"):
        return "static_asset"
    return "other"


@dataclass
class CacheEntry:
    """Cache entry with content, content type, and expiration time."""

    cache_id: str
    resource_url: str
    origin_host: str
    content: bytes
    content_type: str
    expires_at: float
    size: int  # Size in bytes for statistics
    stored_at: float
    hit_count: int = 0
    last_accessed_at: float | None = None

    def ttl_remaining(self) -> float:
        return max(0.0, self.expires_at - time.time())

    def as_metadata(self) -> dict[str, Any]:
        return {
            "cache_id": self.cache_id,
            "resource": self.resource_url,
            "resource_type": _categorize_resource(self.resource_url),
            "origin_host": self.origin_host,
            "content_type": self.content_type,
            "size_bytes": self.size,
            "size_kb": round(self.size / 1024, 2),
            "stored_at": self.stored_at,
            "expires_at": self.expires_at,
            "ttl_remaining": self.ttl_remaining(),
            "hit_count": self.hit_count,
            "last_accessed_at": self.last_accessed_at,
        }


class Cache:
    """In-memory cache with LRU eviction policy."""

    def __init__(self, max_size: int = 1000):
        self._cache: OrderedDict[str, CacheEntry] = OrderedDict()
        self._max_size = max_size
        self._hits = 0
        self._misses = 0
        self._total_bytes_cached = 0

    def _normalize_url(self, url: str) -> tuple[str, str, str]:
        """
        Normalize URL by removing exp and sig query parameters.
        This ensures the same resource is cached regardless of signature.
        """
        parsed = urlparse(url)
        query_params = parse_qs(parsed.query, keep_blank_values=True)

        # Remove exp and sig parameters
        query_params.pop("exp", None)
        query_params.pop("sig", None)

        # Rebuild URL without exp/sig
        new_query = urlencode(query_params, doseq=True)
        normalized_full = urlunparse(
            (parsed.scheme, parsed.netloc, parsed.path, parsed.params, new_query, parsed.fragment)
        )

        # Path + sanitized query for analytics
        resource = parsed.path
        if new_query:
            resource = f"{resource}?{new_query}"

        # Use hash for very long URLs to save memory
        if len(normalized_full) > 500:
            cache_id = hashlib.sha256(normalized_full.encode()).hexdigest()
        else:
            cache_id = normalized_full

        return cache_id, resource, parsed.netloc

    def get(self, url: str) -> CacheEntry | None:
        """Get entry from cache if not expired."""
        cache_key, _, _ = self._normalize_url(url)
        entry = self._cache.get(cache_key)

        if entry is None:
            self._misses += 1
            return None

        # Check expiration
        if time.time() > entry.expires_at:
            # Expired, remove from cache
            del self._cache[cache_key]
            self._total_bytes_cached -= entry.size
            self._misses += 1
            return None

        # Move to end (most recently used)
        self._cache.move_to_end(cache_key)
        self._hits += 1
        entry.hit_count += 1
        entry.last_accessed_at = time.time()
        return entry

    def set(
        self, url: str, content: bytes, content_type: str, ttl: float
    ) -> None:
        """Store entry in cache with TTL."""
        cache_key, resource, host = self._normalize_url(url)
        expires_at = time.time() + ttl
        stored_at = time.time()
        size = len(content)

        # Remove old entry if exists
        if cache_key in self._cache:
            old_entry = self._cache[cache_key]
            self._total_bytes_cached -= old_entry.size

        # Evict oldest entries if cache is full
        while len(self._cache) >= self._max_size:
            oldest_key, oldest_entry = self._cache.popitem(last=False)
            self._total_bytes_cached -= oldest_entry.size

        entry = CacheEntry(
            cache_id=cache_key,
            resource_url=resource,
            origin_host=host,
            content=content,
            content_type=content_type,
            expires_at=expires_at,
            size=size,
            stored_at=stored_at,
            last_accessed_at=stored_at,
        )
        self._cache[cache_key] = entry
        self._total_bytes_cached += size

    def get_stats(self) -> dict[str, Any]:
        """Get cache statistics."""
        total_requests = self._hits + self._misses
        hit_rate = (self._hits / total_requests * 100) if total_requests > 0 else 0.0

        return {
            "hits": self._hits,
            "misses": self._misses,
            "total_requests": total_requests,
            "hit_rate_percent": round(hit_rate, 2),
            "cached_items": len(self._cache),
            "total_bytes_cached": self._total_bytes_cached,
            "total_mb_cached": round(self._total_bytes_cached / (1024 * 1024), 2),
        }

    def list_entries(self) -> list[dict[str, Any]]:
        """Return metadata for all cache entries."""
        return [entry.as_metadata() for entry in self._cache.values()]

    def get_entry(self, cache_id: str) -> CacheEntry | None:
        """Return cache entry by its internal cache id."""
        return self._cache.get(cache_id)

    def clear(self) -> None:
        """Clear all cache entries."""
        self._cache.clear()
        self._hits = 0
        self._misses = 0
        self._total_bytes_cached = 0

