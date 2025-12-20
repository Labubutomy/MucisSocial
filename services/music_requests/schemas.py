from __future__ import annotations

from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, Field


class RequestSessionCreate(BaseModel):
    """Create a new request session"""
    pass


class RequestSessionOut(BaseModel):
    """Request session output"""
    id: UUID
    artist_id: UUID
    session_code: str
    is_active: bool
    qr_code_url: str | None = None
    created_at: datetime
    updated_at: datetime

    model_config = {"from_attributes": True}


class TrackRequestCreate(BaseModel):
    """Create a new track request"""
    session_code: str
    track_id: UUID
    message: str | None = Field(None, max_length=500)


class TrackRequestOut(BaseModel):
    """Track request output"""
    id: UUID
    session_id: UUID
    requester_id: UUID
    artist_id: UUID
    track_id: UUID
    status: str
    message: str | None
    created_at: datetime
    updated_at: datetime

    model_config = {"from_attributes": True}


class TrackRequestAction(BaseModel):
    """Accept or decline a track request"""
    action: str = Field(..., pattern="^(accept|decline)$")


class CoinTransactionOut(BaseModel):
    """Coin transaction output"""
    id: UUID
    from_user_id: UUID
    to_user_id: UUID
    amount: int
    request_id: UUID
    transaction_type: str
    created_at: datetime

    model_config = {"from_attributes": True}


class UserCoinsOut(BaseModel):
    """User coins balance"""
    user_id: UUID
    coins: int
