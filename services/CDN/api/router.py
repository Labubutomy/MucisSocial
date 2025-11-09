from __future__ import annotations

import asyncio
import logging
from contextlib import asynccontextmanager

from typing import Any

from fastapi import APIRouter, Depends, HTTPException, Request, Response
from fastapi.responses import HTMLResponse, JSONResponse

from core.cache import Cache
from core.config import Settings, get_settings
from services.cdn import CDNService

logger = logging.getLogger(__name__)

# Global cache instance
_cache: Cache | None = None
_cdn_service: CDNService | None = None
_stats_task: asyncio.Task | None = None


async def _log_cache_stats_periodically():
    """Periodically log cache statistics if enabled."""
    while True:
        await asyncio.sleep(300)  # Every 5 minutes
        if _cache and _cdn_service:
            settings = get_settings()
            if settings.log_cache_stats:
                stats = _cache.get_stats()
                logger.info(
                    f"Cache stats: {stats['hits']} hits, {stats['misses']} misses, "
                    f"hit rate: {stats['hit_rate_percent']}%, "
                    f"cached items: {stats['cached_items']}, "
                    f"total cached: {stats['total_mb_cached']} MB"
                )


@asynccontextmanager
async def lifespan(app):
    """Lifespan context manager for startup/shutdown."""
    global _cache, _cdn_service, _stats_task
    settings = get_settings()
    _cache = Cache(max_size=settings.cache_max_size)
    _cdn_service = CDNService(settings, _cache)
    logger.info("CDN service started")
    
    # Start periodic stats logging if enabled
    if settings.log_cache_stats:
        _stats_task = asyncio.create_task(_log_cache_stats_periodically())
    
    yield
    
    # Cancel stats task
    if _stats_task:
        _stats_task.cancel()
        try:
            await _stats_task
        except asyncio.CancelledError:
            pass
    
    if _cdn_service:
        await _cdn_service.close()
    logger.info("CDN service stopped")


router = APIRouter()


def get_cdn_service() -> CDNService:
    """Dependency to get CDN service instance."""
    if _cdn_service is None:
        raise RuntimeError("CDN service not initialized")
    return _cdn_service


def get_cache() -> Cache:
    """Dependency to get cache instance."""
    if _cache is None:
        raise RuntimeError("Cache not initialized")
    return _cache


@router.get("/api/stream/{track_id}")
async def proxy_stream_metadata(
    track_id: str,
    request: Request,
    cdn_service: CDNService = Depends(get_cdn_service),
) -> Response:
    """Proxy streaming metadata requests through CDN."""
    return await cdn_service.fetch_stream_metadata(track_id, dict(request.query_params))


@router.post("/api/stream/refresh")
async def proxy_stream_refresh(
    payload: dict[str, Any],
    cdn_service: CDNService = Depends(get_cdn_service),
) -> Response:
    """Proxy stream refresh requests through CDN."""
    return await cdn_service.refresh_stream_metadata(payload)


@router.get("/origin/{resource_path:path}")
async def proxy_resource(
    resource_path: str,
    request: Request,
    cdn_service: CDNService = Depends(get_cdn_service),
) -> Response:
    """
    Proxy resource request to origin or serve from cache.
    This endpoint handles all requests to CDN, preserving signed URLs.
    """
    return await cdn_service.serve_resource(resource_path, request)


@router.get("/health")
async def health_check(settings: Settings = Depends(get_settings)) -> JSONResponse:
    """Health check endpoint."""
    return JSONResponse(
        content={
            "status": "healthy",
            "service": settings.app_name,
            "version": settings.app_version,
        }
    )


@router.get("/stats")
async def get_stats(
    cache: Cache = Depends(get_cache),
    settings: Settings = Depends(get_settings),
) -> JSONResponse:
    """Get CDN statistics (cache hit/miss rate, bandwidth, etc.)."""
    stats = cache.get_stats()
    return JSONResponse(
        content={
            "cache": stats,
            "service": settings.app_name,
            "version": settings.app_version,
        }
    )


@router.get("/analytics", response_class=HTMLResponse)
async def cache_analytics(
    cdn_service: CDNService = Depends(get_cdn_service),
    settings: Settings = Depends(get_settings),
) -> HTMLResponse:
    """Render a simple HTML dashboard with cache analytics."""
    summary = cdn_service.get_cache_summary()
    entries = cdn_service.list_cache_entries()

    rows = []
    for entry in entries:
        rows.append(
            f"""
            <tr>
                <td>{entry['resource_type']}</td>
                <td>{entry['resource']}</td>
                <td>{entry['size_kb']} KB</td>
                <td>{entry['hit_count']}</td>
                <td>{int(entry['ttl_remaining'])}s</td>
            </tr>
            """
        )

    rows_html = "\n".join(rows) if rows else "<tr><td colspan='5'>Кеш пуст</td></tr>"

    summary_rows = []
    for resource_type, stats in summary["by_type"].items():
        summary_rows.append(
            f"""
            <tr>
                <td>{resource_type}</td>
                <td>{stats['count']}</td>
                <td>{round(stats['mb'], 2)} MB</td>
                <td>{stats['avg_ttl_remaining']}s</td>
            </tr>
            """
        )

    summary_html = "\n".join(summary_rows) if summary_rows else "<tr><td colspan='4'>Нет данных</td></tr>"

    html = f"""
    <!DOCTYPE html>
    <html lang="ru">
    <head>
        <meta charset="utf-8">
        <title>CDN Cache Analytics</title>
        <style>
            body {{
                font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
                background: #f3f4f6;
                margin: 0;
                padding: 24px;
                color: #111827;
            }}
            h1, h2 {{
                margin: 0 0 16px 0;
            }}
            .card {{
                background: #ffffff;
                border-radius: 16px;
                padding: 24px;
                box-shadow: 0 10px 30px rgba(15, 23, 42, 0.08);
                margin-bottom: 32px;
            }}
            table {{
                width: 100%;
                border-collapse: collapse;
            }}
            th, td {{
                padding: 12px 16px;
                border-bottom: 1px solid #e5e7eb;
                text-align: left;
            }}
            th {{
                background: #f9fafb;
                font-weight: 600;
            }}
            tr:hover td {{
                background: #f3f4f6;
            }}
            .meta {{
                color: #6b7280;
                font-size: 14px;
                margin-bottom: 24px;
            }}
        </style>
    </head>
    <body>
        <div class="card">
            <h1>CDN Cache Analytics</h1>
            <div class="meta">
                Сервис: {settings.app_name} · Версия: {settings.app_version}<br>
                Всего объектов: {summary['total_entries']} · Объём: {summary['total_mb']} MB
            </div>
            <h2>Сводка по типам ресурсов</h2>
            <table>
                <thead>
                    <tr>
                        <th>Тип</th>
                        <th>Количество</th>
                        <th>Объём</th>
                        <th>Средний TTL</th>
                    </tr>
                </thead>
                <tbody>
                    {summary_html}
                </tbody>
            </table>
        </div>
        <div class="card">
            <h2>Текущие объекты в кеше</h2>
            <table>
                <thead>
                    <tr>
                        <th>Тип</th>
                        <th>Ресурс</th>
                        <th>Размер</th>
                        <th>Попаданий</th>
                        <th>TTL остаток</th>
                    </tr>
                </thead>
                <tbody>
                    {rows_html}
                </tbody>
            </table>
        </div>
    </body>
    </html>
    """

    return HTMLResponse(content=html)


@router.get("/cache/entries")
async def list_cache_entries(
    cdn_service: CDNService = Depends(get_cdn_service),
) -> JSONResponse:
    """List cached resources for analytical dashboards."""
    entries = cdn_service.list_cache_entries()
    return JSONResponse(
        content={
            "total": len(entries),
            "entries": entries,
        }
    )


@router.get("/cache/entries/{cache_id}")
async def get_cache_entry(
    cache_id: str,
    include_content: bool = False,
    cdn_service: CDNService = Depends(get_cdn_service),
) -> JSONResponse:
    """Get detailed metadata for a cache entry."""
    entry = cdn_service.get_cache_entry(cache_id, include_content=include_content)
    if entry is None:
        raise HTTPException(status_code=404, detail="Cache entry not found")
    return JSONResponse(content=entry)


@router.get("/cache/summary")
async def cache_summary(
    cdn_service: CDNService = Depends(get_cdn_service),
) -> JSONResponse:
    """Aggregate cache stats by resource type."""
    summary = cdn_service.get_cache_summary()
    return JSONResponse(content=summary)

