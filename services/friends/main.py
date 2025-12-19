from __future__ import annotations

import logging
import os
from typing import Annotated
from uuid import UUID

import httpx
import jwt
from fastapi import Depends, FastAPI, Header, HTTPException, Query, status
from sqlalchemy import and_, or_, select
from sqlalchemy.ext.asyncio import AsyncSession

from core.config import get_settings
from core.database import get_db
from models import Friend, FriendRequest
from schemas import (
    FriendOut,
    FriendRequestAction,
    FriendRequestCreate,
    FriendRequestOut,
    UserSearchResult,
)


logger = logging.getLogger(__name__)
settings = get_settings()

app = FastAPI(
    title="Friends Service",
    version=settings.app_version,
    description="Friends & social graph for Music Social",
)

# CORS обрабатывается на уровне gateway, не нужно дублировать здесь


async def get_current_user_id(
    authorization: Annotated[str | None, Header(alias="Authorization")] = None,
) -> UUID:
    if not authorization:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Missing Authorization header",
        )
    if not authorization.startswith("Bearer "):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid Authorization header format",
        )
    token = authorization.removeprefix("Bearer ").strip()
    try:
        payload = jwt.decode(
            token,
            settings.jwt_secret,
            algorithms=[settings.jwt_algorithm],
        )
        user_id = payload.get("user_id")
        if not user_id:
            raise ValueError("user_id missing in token")
        return UUID(user_id)
    except Exception as exc:
        logger.warning("Token validation failed: %s", exc)
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid token",
        ) from exc


@app.get("/health")
async def health() -> dict[str, str]:
    return {"status": "ok"}


@app.post(
    "/requests",
    response_model=FriendRequestOut,
    status_code=status.HTTP_201_CREATED,
)
async def create_friend_request(
    payload: FriendRequestCreate,
    current_user_id: UUID = Depends(get_current_user_id),
    db: AsyncSession = Depends(get_db),
) -> FriendRequestOut:
    """Отправить запрос в друзья."""
    if payload.to_user_id == current_user_id:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Cannot send friend request to yourself",
        )

    # Проверяем, нет ли уже дружбы
    res_f = await db.execute(
        select(Friend).where(
            or_(
                and_(
                    Friend.user_id == current_user_id,
                    Friend.friend_id == payload.to_user_id,
                ),
                and_(
                    Friend.user_id == payload.to_user_id,
                    Friend.friend_id == current_user_id,
                ),
            )
        )
    )
    if res_f.scalar_one_or_none():
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Users are already friends",
        )

    # Проверяем существующий запрос
    res_req = await db.execute(
        select(FriendRequest).where(
            FriendRequest.from_user_id == current_user_id,
            FriendRequest.to_user_id == payload.to_user_id,
        )
    )
    existing = res_req.scalar_one_or_none()
    if existing and existing.status == "pending":
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Friend request already sent",
        )

    request = FriendRequest(
        from_user_id=current_user_id,
        to_user_id=payload.to_user_id,
        status="pending",
    )
    db.add(request)
    await db.commit()
    await db.refresh(request)

    return FriendRequestOut.model_validate(request)


@app.get(
    "/requests/incoming",
    response_model=list[FriendRequestOut],
)
async def list_incoming_requests(
    current_user_id: UUID = Depends(get_current_user_id),
    db: AsyncSession = Depends(get_db),
) -> list[FriendRequestOut]:
    res = await db.execute(
        select(FriendRequest).where(
            FriendRequest.to_user_id == current_user_id,
            FriendRequest.status == "pending",
        )
    )
    requests = list(res.scalars())
    return [FriendRequestOut.model_validate(r) for r in requests]


@app.post(
    "/requests/{request_id}/respond",
)
async def respond_to_request(
    request_id: UUID,
    payload: FriendRequestAction,
    current_user_id: UUID = Depends(get_current_user_id),
    db: AsyncSession = Depends(get_db),
) -> dict[str, bool]:
    res = await db.execute(
        select(FriendRequest).where(
            FriendRequest.id == request_id,
            FriendRequest.to_user_id == current_user_id,
        )
    )
    fr = res.scalar_one_or_none()
    if fr is None or fr.status != "pending":
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Friend request not found",
        )

    if payload.accept:
        fr.status = "accepted"
        # создаем дружбу в обе стороны
        db.add_all(
            [
                Friend(user_id=current_user_id, friend_id=fr.from_user_id),
                Friend(user_id=fr.from_user_id, friend_id=current_user_id),
            ]
        )
        await db.commit()
        
        # Автоматически создаем диалог между друзьями
        # Диалог будет создан автоматически при первой отправке сообщения
        # Или можно вызвать endpoint создания диалога, но для этого нужен токен
        logger.info(f"Friendship accepted between {current_user_id} and {fr.from_user_id}")
        db.add_all(
            [
                Friend(user_id=current_user_id, friend_id=fr.from_user_id),
                Friend(user_id=fr.from_user_id, friend_id=current_user_id),
            ]
        )
    else:
        fr.status = "declined"

    await db.commit()
    return {"success": True}


async def get_user_info(friend_id: UUID, auth_token: str) -> dict | None:
    """Get user information from gateway."""
    try:
        async with httpx.AsyncClient() as client:
            response = await client.get(
                f"{settings.gateway_url}/api/v1/users/{friend_id}",
                headers={"Authorization": f"Bearer {auth_token}"},
                timeout=5.0,
            )
            if response.status_code == 200:
                data = response.json()
                return data.get("user")
    except Exception as e:
        logger.warning(f"Failed to fetch user info for {friend_id}: {e}")
    return None


@app.get(
    "/",
    response_model=list[FriendOut],
)
async def list_friends(
    current_user_id: UUID = Depends(get_current_user_id),
    authorization: Annotated[str | None, Header(alias="Authorization")] = None,
    db: AsyncSession = Depends(get_db),
) -> list[FriendOut]:
    res = await db.execute(
        select(Friend).where(Friend.user_id == current_user_id)
    )
    friends = list(res.scalars())
    
    # Fetch user info for each friend
    result = []
    auth_token = authorization.replace("Bearer ", "") if authorization else None
    
    for friend in friends:
        friend_data = FriendOut.model_validate(friend)
        # Get friend's user info
        if auth_token:
            user_info = await get_user_info(friend.friend_id, auth_token)
            if user_info:
                friend_data.friend_username = user_info.get("username")
                friend_data.friend_avatar_url = user_info.get("avatar_url")
        result.append(friend_data)
    
    return result




if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "main:app",
        host=settings.host,
        port=settings.port,
        reload=False,
    )


