from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, status

from schemas.point import RoutePointCreate, RoutePointUpdate, RoutePointResponse, RoutePointReorderRequest
from services.point_service import PointService

router = APIRouter(prefix="/routes", tags=["points"])

point_service = PointService()


# Temporary dependency for user_id
async def get_current_user_id() -> UUID:
    """Get current user ID from JWT token."""
    # TODO: Extract from JWT token
    return UUID("00000000-0000-0000-0000-000000000001")


@router.post("/{route_id}/points", response_model=RoutePointResponse, status_code=status.HTTP_201_CREATED)
async def add_point(
    route_id: UUID,
    point_data: RoutePointCreate,
    user_id: UUID = Depends(get_current_user_id),
):
    """Add a point to a route."""
    try:
        point = await point_service.add_point(route_id, user_id, point_data)
        return RoutePointResponse(**point)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e),
        )
    except PermissionError as e:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail=str(e),
        )


@router.get("/{route_id}/points", response_model=list[RoutePointResponse])
async def get_route_points(route_id: UUID):
    """Get all points for a route."""
    points = await point_service.get_route_points(route_id)
    return [RoutePointResponse(**point) for point in points]


@router.put("/{route_id}/points/{point_id}", response_model=RoutePointResponse)
async def update_point(
    route_id: UUID,
    point_id: UUID,
    point_data: RoutePointUpdate,
    user_id: UUID = Depends(get_current_user_id),
):
    """Update a route point."""
    try:
        point = await point_service.update_point(point_id, user_id, point_data)
        if not point:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Point not found",
            )
        return RoutePointResponse(**point)
    except PermissionError as e:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail=str(e),
        )


@router.delete("/{route_id}/points/{point_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_point(
    route_id: UUID,
    point_id: UUID,
    user_id: UUID = Depends(get_current_user_id),
):
    """Delete a route point."""
    try:
        deleted = await point_service.delete_point(point_id, user_id)
        if not deleted:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Point not found",
            )
    except PermissionError as e:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail=str(e),
        )


@router.post("/{route_id}/points/reorder", status_code=status.HTTP_200_OK)
async def reorder_points(
    route_id: UUID,
    reorder_data: RoutePointReorderRequest,
    user_id: UUID = Depends(get_current_user_id),
):
    """Reorder route points."""
    try:
        await point_service.reorder_points(route_id, user_id, reorder_data.point_ids)
        return {"status": "ok"}
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )
    except PermissionError as e:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail=str(e),
        )

