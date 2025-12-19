from contextlib import asynccontextmanager

from fastapi import FastAPI

from api.v1.router import api_router
from core.config import get_settings
from core.database import close_db, init_db
from services.kafka_service import KafkaService

settings = get_settings()
kafka_service = KafkaService()


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifespan context manager for startup/shutdown."""
    # Startup
    await init_db()
    await kafka_service.connect_producer()
    yield
    # Shutdown
    await kafka_service.disconnect_producer()
    await close_db()


def create_app() -> FastAPI:
    """Create FastAPI application."""
    app = FastAPI(
        title=settings.app_name,
        version=settings.app_version,
        description="Route Session Service for Music Social",
        lifespan=lifespan,
    )

    # CORS is handled by API Gateway, no need for CORS middleware here
    # Adding CORS here causes duplicate headers which breaks CORS

    app.include_router(api_router, prefix="/api/v1")

    @app.get("/health")
    async def health_check():
        """Health check endpoint."""
        return {
            "status": "healthy",
            "service": settings.app_name,
            "version": settings.app_version,
        }

    return app


app = create_app()

