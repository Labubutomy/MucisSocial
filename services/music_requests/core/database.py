from __future__ import annotations

from typing import AsyncGenerator

from sqlalchemy.ext.asyncio import (
    AsyncEngine,
    AsyncSession,
    async_sessionmaker,
    create_async_engine,
)

from .config import get_settings


settings = get_settings()

# Convert postgres:// to postgresql+asyncpg://
db_url = settings.database_url.replace("postgres://", "postgresql+asyncpg://")
# Remove ?sslmode=disable if present (asyncpg doesn't support this parameter)
if "?sslmode=" in db_url:
    db_url = db_url.split("?")[0]

engine: AsyncEngine = create_async_engine(
    db_url,
    echo=False,
    future=True,
)

SessionLocal = async_sessionmaker(
    bind=engine,
    class_=AsyncSession,
    expire_on_commit=False,
)


async def get_db() -> AsyncGenerator[AsyncSession, None]:
    async with SessionLocal() as session:
        yield session
