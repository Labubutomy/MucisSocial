from __future__ import annotations

import json
from typing import Any

import redis.asyncio as aioredis

from core.config import Settings


class RedisService:
    """Service for managing WebSocket connections in Redis."""

    def __init__(self, settings: Settings):
        self.settings = settings
        self.redis: aioredis.Redis | None = None
        self._connection_string = (
            f"redis://{settings.redis_host}:{settings.redis_port}/{settings.redis_db}"
        )
        if settings.redis_password:
            self._connection_string = (
                f"redis://:{settings.redis_password}@{settings.redis_host}:"
                f"{settings.redis_port}/{settings.redis_db}"
            )

    async def connect(self) -> None:
        """Connect to Redis."""
        self.redis = await aioredis.from_url(
            self._connection_string, decode_responses=True
        )

    async def disconnect(self) -> None:
        """Disconnect from Redis."""
        if self.redis:
            await self.redis.close()

    async def add_connection(
        self, room_id: str, user_id: str, connection_id: str
    ) -> None:
        """Add a connection to a room."""
        if not self.redis:
            raise RuntimeError("Redis not connected")

        key = f"room:{room_id}:connections"
        await self.redis.sadd(key, f"{user_id}:{connection_id}")
        await self.redis.expire(key, self.settings.connection_timeout)

        # Store connection metadata
        import time
        conn_key = f"connection:{connection_id}"
        await self.redis.hset(
            conn_key,
            mapping={
                "room_id": room_id,
                "user_id": user_id,
                "connected_at": str(int(time.time())),
            },
        )
        await self.redis.expire(conn_key, self.settings.connection_timeout)

    async def remove_connection(
        self, room_id: str, user_id: str, connection_id: str
    ) -> None:
        """Remove a connection from a room."""
        if not self.redis:
            raise RuntimeError("Redis not connected")

        key = f"room:{room_id}:connections"
        await self.redis.srem(key, f"{user_id}:{connection_id}")

        # Remove connection metadata
        conn_key = f"connection:{connection_id}"
        await self.redis.delete(conn_key)

    async def get_room_connections(self, room_id: str) -> list[str]:
        """Get all connection IDs for a room."""
        if not self.redis:
            raise RuntimeError("Redis not connected")

        key = f"room:{room_id}:connections"
        connections = await self.redis.smembers(key)
        return list(connections)

    async def get_connection_info(self, connection_id: str) -> dict[str, Any] | None:
        """Get connection metadata."""
        if not self.redis:
            raise RuntimeError("Redis not connected")

        conn_key = f"connection:{connection_id}"
        data = await self.redis.hgetall(conn_key)
        return data if data else None

    async def ping(self) -> bool:
        """Check Redis connection."""
        if not self.redis:
            return False
        try:
            await self.redis.ping()
            return True
        except Exception:
            return False

