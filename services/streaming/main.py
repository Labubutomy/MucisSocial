from __future__ import annotations

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from api.v1.router import api_router
from api.v1.endpoints import origin
from core.config import get_settings


def create_app() -> FastAPI:
    settings = get_settings()
    app = FastAPI(
        title=settings.app_name,
        version=settings.app_version,
        description="Streaming Gateway Service",
    )

    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    app.include_router(api_router)
    app.include_router(origin.router)

    @app.get("/health")
    async def health_check():
        """Health check endpoint for monitoring."""
        return {"status": "healthy", "service": settings.app_name, "version": settings.app_version}

    return app


app = create_app()

