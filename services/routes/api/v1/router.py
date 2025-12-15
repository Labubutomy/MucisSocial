from fastapi import APIRouter

from .endpoints import routes, points, search

api_router = APIRouter()

api_router.include_router(routes.router)
api_router.include_router(points.router)
api_router.include_router(search.router)

