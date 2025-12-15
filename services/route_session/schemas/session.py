from datetime import datetime
from typing import Optional
from uuid import UUID

from pydantic import BaseModel, Field


class RouteSessionBase(BaseModel):
    """Base route session schema."""

    route_id: UUID = Field(..., description="Route ID")
    user_id: UUID = Field(..., description="User ID")


class RouteSessionStart(RouteSessionBase):
    """Schema for starting a route session."""

    pass


class RouteSessionUpdate(BaseModel):
    """Schema for updating route session location."""

    latitude: float = Field(..., ge=-90, le=90)
    longitude: float = Field(..., ge=-180, le=180)
    timestamp: Optional[datetime] = None


class RouteSessionResponse(BaseModel):
    """Schema for route session response."""

    id: UUID
    user_id: UUID
    route_id: UUID
    started_at: datetime
    completed_at: Optional[datetime] = None
    current_point_index: int
    visited_points: list[str] = Field(default_factory=list)
    total_distance_km: Optional[float] = None
    status: str  # active, paused, completed, abandoned

    model_config = {"from_attributes": True}


class RoutePointTriggered(BaseModel):
    """Schema for triggered route point."""

    point_id: UUID
    track_id: UUID
    title: Optional[str] = None
    track_start_offset_sec: int = 0
    order_index: int


class RouteProgressUpdate(BaseModel):
    """Schema for route progress update."""

    current_point: int
    total_points: int
    percentage: float
    visited_points: list[UUID]

