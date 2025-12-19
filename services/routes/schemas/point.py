from datetime import datetime
from typing import Optional
from uuid import UUID

from pydantic import BaseModel, Field


class RoutePointBase(BaseModel):
    """Base route point schema."""

    latitude: float = Field(..., ge=-90, le=90, description="Latitude")
    longitude: float = Field(..., ge=-180, le=180, description="Longitude")
    radius_meters: int = Field(50, ge=10, le=1000, description="Activation radius in meters")
    track_id: UUID = Field(..., description="Track ID")
    track_start_offset_sec: int = Field(0, ge=0, description="Track start offset in seconds")
    title: Optional[str] = Field(None, max_length=100, description="Point title")
    description: Optional[str] = Field(None, description="Point description")
    image_url: Optional[str] = Field(None, max_length=500, description="Point image URL")


class RoutePointCreate(RoutePointBase):
    """Schema for creating a route point."""

    order_index: Optional[int] = Field(None, ge=0, description="Point order index")


class RoutePointUpdate(BaseModel):
    """Schema for updating a route point."""

    latitude: Optional[float] = Field(None, ge=-90, le=90)
    longitude: Optional[float] = Field(None, ge=-180, le=180)
    radius_meters: Optional[int] = Field(None, ge=10, le=1000)
    track_id: Optional[UUID] = None
    track_start_offset_sec: Optional[int] = Field(None, ge=0)
    title: Optional[str] = Field(None, max_length=100)
    description: Optional[str] = None
    image_url: Optional[str] = Field(None, max_length=500)
    order_index: Optional[int] = Field(None, ge=0)


class RoutePointResponse(RoutePointBase):
    """Schema for route point response."""

    id: UUID
    route_id: UUID
    geohash: str
    order_index: int
    created_at: datetime

    model_config = {"from_attributes": True}


class RoutePointReorderRequest(BaseModel):
    """Schema for reordering route points."""

    point_ids: list[UUID] = Field(..., description="List of point IDs in new order")

