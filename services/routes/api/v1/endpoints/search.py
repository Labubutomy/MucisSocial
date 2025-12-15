from fastapi import APIRouter, Query, HTTPException, status

from schemas.route import RouteResponse, RouteListResponse
from services.route_service import RouteService

router = APIRouter(prefix="/routes", tags=["search"])

route_service = RouteService()


@router.get("/nearby", response_model=RouteListResponse)
async def find_nearby_routes(
    latitude: float = Query(..., ge=-90, le=90, description="Latitude"),
    longitude: float = Query(..., ge=-180, le=180, description="Longitude"),
    radius_km: float = Query(5.0, ge=0.1, le=100.0, description="Search radius in kilometers"),
    limit: int = Query(20, ge=1, le=100, description="Limit results"),
    offset: int = Query(0, ge=0, description="Offset results"),
):
    """Find routes near a location."""
    try:
        routes, total = await route_service.find_nearby_routes(
            latitude=latitude,
            longitude=longitude,
            radius_km=radius_km,
            limit=limit,
            offset=offset,
        )
        return RouteListResponse(
            routes=[RouteResponse(**route) for route in routes],
            total=total,
            limit=limit,
            offset=offset,
        )
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e),
        )


@router.get("/search", response_model=RouteListResponse)
async def search_routes(
    q: str = Query(..., min_length=1, description="Search query"),
    city: str | None = Query(None, description="Filter by city"),
    limit: int = Query(20, ge=1, le=100),
    offset: int = Query(0, ge=0),
):
    """Search routes by title, description, or tags."""
    # Simple search implementation - can be enhanced with full-text search
    routes, total = await route_service.list_routes(
        city=city,
        limit=limit,
        offset=offset,
    )

    # Filter by search query
    query_lower = q.lower()
    filtered_routes = [
        route
        for route in routes
        if query_lower in route.get("title", "").lower()
        or query_lower in route.get("description", "").lower()
        or any(query_lower in tag.lower() for tag in route.get("tags", []))
    ]

    return RouteListResponse(
        routes=[RouteResponse(**route) for route in filtered_routes[:limit]],
        total=len(filtered_routes),
        limit=limit,
        offset=offset,
    )

