from datetime import datetime
from typing import Optional
from uuid import UUID

from pydantic import BaseModel, Field

from .point import RoutePointResponse, RoutePointCreate


class RouteBase(BaseModel):
    """Base route schema."""

    title: str = Field(..., max_length=200, description="Route title")
    description: Optional[str] = Field(None, description="Route description")
    city: Optional[str] = Field(None, max_length=100, description="City name")
    country: Optional[str] = Field(None, max_length=100, description="Country name")
    is_public: bool = Field(True, description="Is route public")
    is_linear: bool = Field(True, description="Is route linear (points must be visited in order)")
    cover_image_url: Optional[str] = Field(None, max_length=500, description="Cover image URL")


class RouteCreate(RouteBase):
    """Schema for creating a route."""

    tags: list[str] = Field(default_factory=list, description="Route tags")
    points: Optional[list[RoutePointCreate]] = Field(default_factory=list, description="Route points")


class RouteUpdate(BaseModel):
    """Schema for updating a route."""

    title: Optional[str] = Field(None, max_length=200)
    description: Optional[str] = None
    city: Optional[str] = Field(None, max_length=100)
    country: Optional[str] = Field(None, max_length=100)
    is_public: Optional[bool] = None
    is_linear: Optional[bool] = None
    cover_image_url: Optional[str] = Field(None, max_length=500)
    tags: Optional[list[str]] = None


class RouteResponse(RouteBase):
    """Schema for route response."""

    id: UUID
    user_id: UUID
    total_distance_km: Optional[float] = None
    estimated_minutes: Optional[int] = None
    tags: list[str] = Field(default_factory=list)
    points: list[RoutePointResponse] = Field(default_factory=list)
    created_at: datetime
    updated_at: datetime

    model_config = {"from_attributes": True}


class RouteListResponse(BaseModel):
    """Schema for route list response."""

    routes: list[RouteResponse]
    total: int
    limit: int
    offset: int

