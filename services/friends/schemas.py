from __future__ import annotations

from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, Field


class FriendRequestCreate(BaseModel):
    to_user_id: UUID


class FriendRequestOut(BaseModel):
    id: UUID
    from_user_id: UUID
    to_user_id: UUID
    status: str
    created_at: datetime

    class Config:
        from_attributes = True


class FriendOut(BaseModel):
    user_id: UUID
    friend_id: UUID
    created_at: datetime
    friend_username: str | None = None
    friend_avatar_url: str | None = None

    class Config:
        from_attributes = True


class FriendRequestAction(BaseModel):
    accept: bool = Field(
        default=True,
        description="True to accept, False to decline",
    )


class UserSearchResult(BaseModel):
    id: UUID
    username: str
    avatar_url: str | None = None


