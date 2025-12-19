from fastapi import APIRouter

from .endpoints import sessions

api_router = APIRouter()

api_router.include_router(sessions.router)

