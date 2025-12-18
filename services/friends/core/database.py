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

# Убираем sslmode из URL, так как asyncpg не поддерживает этот параметр
db_url = settings.database_url.replace("postgres://", "postgresql+asyncpg://")
# Удаляем ?sslmode=disable если присутствует
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


