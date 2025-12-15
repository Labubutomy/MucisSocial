from typing import Optional
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Query, status
from pydantic import BaseModel

from schemas.route import RouteCreate, RouteUpdate, RouteResponse, RouteListResponse
from schemas.point import RoutePointResponse
from services.route_service import RouteService

router = APIRouter(prefix="/routes", tags=["routes"])

route_service = RouteService()


# Temporary dependency for user_id (should be extracted from JWT in real implementation)
async def get_current_user_id() -> UUID:
    """Get current user ID from JWT token."""
    # TODO: Extract from JWT token
    # For now, return a default UUID for testing
    return UUID("00000000-0000-0000-0000-000000000001")


@router.post("", response_model=RouteResponse, status_code=status.HTTP_201_CREATED)
async def create_route(
    route_data: RouteCreate,
    user_id: UUID = Depends(get_current_user_id),
):
    """Create a new route."""
    try:
        route = await route_service.create_route(user_id, route_data)
        # Convert points to RoutePointResponse objects if they exist
        if route.get("points"):
            route["points"] = [RoutePointResponse(**point) for point in route["points"]]
        return RouteResponse(**route)
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )


@router.get("/{route_id}", response_model=RouteResponse)
async def get_route(
    route_id: UUID,
    include_points: bool = Query(False, description="Include route points"),
):
    """Get route by ID."""
    route = await route_service.get_route(route_id, include_points=include_points)
    if not route:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Route not found",
        )
    
    # Convert points to RoutePointResponse objects if they exist
    if route.get("points"):
        route["points"] = [RoutePointResponse(**point) for point in route["points"]]
    
    return RouteResponse(**route)


@router.put("/{route_id}", response_model=RouteResponse)
async def update_route(
    route_id: UUID,
    route_data: RouteUpdate,
    user_id: UUID = Depends(get_current_user_id),
):
    """Update route."""
    try:
        route = await route_service.update_route(route_id, user_id, route_data)
        if not route:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Route not found",
            )
        return RouteResponse(**route)
    except PermissionError as e:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail=str(e),
        )


@router.delete("/{route_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_route(
    route_id: UUID,
    user_id: UUID = Depends(get_current_user_id),
):
    """Delete route."""
    try:
        deleted = await route_service.delete_route(route_id, user_id)
        if not deleted:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Route not found",
            )
    except PermissionError as e:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail=str(e),
        )


@router.get("", response_model=RouteListResponse)
async def list_routes(
    user_id: Optional[UUID] = Query(None, description="Filter by user ID"),
    is_public: Optional[bool] = Query(None, description="Filter by public/private"),
    city: Optional[str] = Query(None, description="Filter by city"),
    limit: int = Query(20, ge=1, le=100, description="Limit results"),
    offset: int = Query(0, ge=0, description="Offset results"),
):
    """List routes with filters."""
    routes, total = await route_service.list_routes(
        user_id=user_id,
        is_public=is_public,
        city=city,
        limit=limit,
        offset=offset,
    )
    return RouteListResponse(
        routes=[RouteResponse(**route) for route in routes],
        total=total,
        limit=limit,
        offset=offset,
    )

