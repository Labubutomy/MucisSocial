from fastapi import APIRouter

from api.v1.endpoints import stream

api_router = APIRouter(prefix="/api")
api_router.include_router(stream.router)

