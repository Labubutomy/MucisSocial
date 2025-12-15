from contextlib import asynccontextmanager

import asyncpg
from asyncpg import Pool, Connection

from .config import get_settings

settings = get_settings()

# Global connection pool
_pool: Pool | None = None


async def init_db() -> Pool:
    """Initialize database connection pool."""
    global _pool
    if _pool is None:
        _pool = await asyncpg.create_pool(
            settings.database_url,
            min_size=5,
            max_size=20,
            command_timeout=60,
        )
    return _pool


async def close_db() -> None:
    """Close database connection pool."""
    global _pool
    if _pool is not None:
        await _pool.close()
        _pool = None


@asynccontextmanager
async def get_connection() -> Connection:
    """Get database connection from pool."""
    if _pool is None:
        await init_db()
    async with _pool.acquire() as connection:
        yield connection


async def get_pool() -> Pool:
    """Get database connection pool."""
    if _pool is None:
        await init_db()
    return _pool

