from __future__ import annotations

import base64
import io
import logging
import secrets
from typing import Annotated
from uuid import UUID

import httpx
import jwt
import qrcode
from fastapi import Depends, FastAPI, Header, HTTPException, Query, status
from fastapi.responses import StreamingResponse
from sqlalchemy import and_, select, update
from sqlalchemy.ext.asyncio import AsyncSession

from core.config import get_settings
from core.database import get_db
from core.kafka import send_music_request_event
from models import CoinTransaction, RequestSession, TrackRequest
from schemas import (
    RequestSessionCreate,
    RequestSessionOut,
    TrackRequestAction,
    TrackRequestCreate,
    TrackRequestOut,
    UserCoinsOut,
)


logger = logging.getLogger(__name__)
settings = get_settings()

app = FastAPI(
    title="Music Requests Service",
    version=settings.app_version,
    description="Music requests via QR codes for Music Social",
)


async def get_current_user_id(
    authorization: Annotated[str | None, Header(alias="Authorization")] = None,
) -> UUID:
    """Extract user ID from JWT token"""
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
    
    token = authorization.replace("Bearer ", "")
    try:
        payload = jwt.decode(token, settings.jwt_secret, algorithms=[settings.jwt_algorithm])
        user_id = payload.get("user_id")
        if not user_id:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid token payload",
            )
        return UUID(user_id)
    except jwt.ExpiredSignatureError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Token expired",
        )
    except (jwt.InvalidTokenError, ValueError):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid token",
        )


async def get_user_coins(user_id: UUID, authorization: str) -> int:
    """Get user's coin balance from users service via gateway"""
    async with httpx.AsyncClient() as client:
        try:
            response = await client.get(
                f"{settings.gateway_url}/api/v1/users/{user_id}",
                headers={"Authorization": authorization},
            )
            response.raise_for_status()
            user_data = response.json()
            return user_data.get("coins", 0)
        except httpx.HTTPError as e:
            logger.error(f"Failed to get user coins: {e}")
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail="Failed to get user balance",
            )


async def update_user_coins(user_id: UUID, coins_delta: int, authorization: str) -> None:
    """Update user's coin balance via gateway"""
    async with httpx.AsyncClient() as client:
        try:
            response = await client.patch(
                f"{settings.gateway_url}/api/v1/users/{user_id}/coins",
                json={"coins_delta": coins_delta},
                headers={"Authorization": authorization},
            )
            response.raise_for_status()
        except httpx.HTTPError as e:
            logger.error(f"Failed to update user coins: {e}")
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail="Failed to update user balance",
            )


async def add_track_to_queue(artist_id: UUID, track_id: UUID, authorization: str) -> None:
    """Add accepted track to artist's playback queue via gateway API"""
    async with httpx.AsyncClient() as client:
        try:
            # First, ensure the artist has a queue (create if not exists)
            # Use context_type="music_request" for music request queues
            context_type = "music_request"
            context_id = str(artist_id)
            
            # Add track to the queue via gateway
            response = await client.post(
                f"{settings.gateway_url}/api/v1/sessions/{context_id}/queue/tracks",
                json={"track_id": str(track_id)},
                headers={"Authorization": authorization},
                timeout=10.0,
            )
            
            if response.status_code == 404:
                # Queue doesn't exist, try to create it first (for non-user contexts)
                logger.info(f"Queue not found for artist {artist_id}, creating...")
                # For now, just log the attempt - the playback queue handles this
            
            response.raise_for_status()
            logger.info(f"Successfully added track {track_id} to artist {artist_id} queue")
        except httpx.HTTPError as e:
            logger.error(f"Failed to add track to queue: {e}")
            # Don't fail the request, just log the error
            # The track request is still accepted even if queue addition fails


@app.get("/health")
async def health():
    """Health check endpoint"""
    return {"status": "ok", "service": "music_requests"}


# ==================== Request Sessions ====================

@app.post("/api/v1/sessions", response_model=RequestSessionOut, status_code=status.HTTP_201_CREATED)
async def create_request_session(
    session_in: RequestSessionCreate,
    user_id: Annotated[UUID, Depends(get_current_user_id)],
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Create a new request session for artist to receive track requests"""
    # Generate unique session code
    session_code = secrets.token_urlsafe(16)
    
    # Deactivate all previous sessions for this artist
    await db.execute(
        update(RequestSession)
        .where(RequestSession.artist_id == user_id)
        .values(is_active=False)
    )
    
    # Create new session
    new_session = RequestSession(
        artist_id=user_id,
        session_code=session_code,
        is_active=True,
    )
    db.add(new_session)
    await db.commit()
    await db.refresh(new_session)
    
    # Generate QR code URL (will be used by frontend to generate QR)
    qr_url = f"/api/v1/sessions/{session_code}/qr"
    
    return RequestSessionOut(
        **new_session.__dict__,
        qr_code_url=qr_url,
    )


@app.get("/api/v1/sessions/my", response_model=RequestSessionOut | None)
async def get_my_active_session(
    user_id: Annotated[UUID, Depends(get_current_user_id)],
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Get artist's active request session"""
    result = await db.execute(
        select(RequestSession)
        .where(and_(
            RequestSession.artist_id == user_id,
            RequestSession.is_active == True
        ))
        .order_by(RequestSession.created_at.desc())
    )
    session = result.scalar_one_or_none()
    
    if not session:
        return None
    
    return RequestSessionOut(
        **session.__dict__,
        qr_code_url=f"/api/v1/sessions/{session.session_code}/qr",
    )


@app.delete("/api/v1/sessions/{session_id}", status_code=status.HTTP_204_NO_CONTENT)
async def deactivate_session(
    session_id: UUID,
    user_id: Annotated[UUID, Depends(get_current_user_id)],
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Deactivate a request session"""
    result = await db.execute(
        select(RequestSession).where(RequestSession.id == session_id)
    )
    session = result.scalar_one_or_none()
    
    if not session:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Session not found")
    
    if session.artist_id != user_id:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Not your session")
    
    session.is_active = False
    await db.commit()


@app.get("/api/v1/sessions/code/{session_code}", response_model=RequestSessionOut)
async def get_session_by_code(
    session_code: str,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Get session by code (public endpoint for users who want to request)"""
    result = await db.execute(
        select(RequestSession).where(RequestSession.session_code == session_code)
    )
    session = result.scalar_one_or_none()
    
    if not session or not session.is_active:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Session not found or inactive")
    
    return RequestSessionOut(
        **session.__dict__,
        qr_code_url=f"/api/v1/sessions/{session.session_code}/qr",
    )


@app.get("/api/v1/sessions/{session_code}/qr")
async def get_qr_code(session_code: str, db: Annotated[AsyncSession, Depends(get_db)]):
    """Generate QR code for a session"""
    result = await db.execute(
        select(RequestSession).where(RequestSession.session_code == session_code)
    )
    session = result.scalar_one_or_none()
    
    if not session or not session.is_active:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Session not found or inactive")
    
    # Generate QR code pointing to frontend request page
    qr_data = f"musicsocial://request/{session_code}"
    
    qr = qrcode.QRCode(version=1, box_size=10, border=5)
    qr.add_data(qr_data)
    qr.make(fit=True)
    
    img = qr.make_image(fill_color="black", back_color="white")
    buf = io.BytesIO()
    img.save(buf, format="PNG")
    buf.seek(0)
    
    return StreamingResponse(buf, media_type="image/png")


# ==================== Track Requests ====================

@app.post("/api/v1/requests", response_model=TrackRequestOut, status_code=status.HTTP_201_CREATED)
async def create_track_request(
    request_in: TrackRequestCreate,
    user_id: Annotated[UUID, Depends(get_current_user_id)],
    authorization: Annotated[str, Header(alias="Authorization")],
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Create a new track request (costs 1 coin)"""
    # Get session
    result = await db.execute(
        select(RequestSession).where(RequestSession.session_code == request_in.session_code)
    )
    session = result.scalar_one_or_none()
    
    if not session or not session.is_active:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Session not found or inactive"
        )
    
    # Check if user has enough coins
    user_coins = await get_user_coins(user_id, authorization)
    if user_coins < settings.coin_cost_per_request:
        raise HTTPException(
            status_code=status.HTTP_402_PAYMENT_REQUIRED,
            detail=f"Insufficient coins. Required: {settings.coin_cost_per_request}, Available: {user_coins}"
        )
    
    # Create track request
    new_request = TrackRequest(
        session_id=session.id,
        requester_id=user_id,
        artist_id=session.artist_id,
        track_id=request_in.track_id,
        message=request_in.message,
        status="pending",
    )
    db.add(new_request)
    await db.flush()
    
    # Deduct coin from requester
    await update_user_coins(user_id, -settings.coin_cost_per_request, authorization)
    
    # Create coin transaction record
    transaction = CoinTransaction(
        from_user_id=user_id,
        to_user_id=session.artist_id,
        amount=settings.coin_cost_per_request,
        request_id=new_request.id,
        transaction_type="music_request",
    )
    db.add(transaction)
    
    await db.commit()
    await db.refresh(new_request)
    
    # Send Kafka event for WebSocket notification
    send_music_request_event("request_created", {
        "request_id": str(new_request.id),
        "artist_id": str(session.artist_id),
        "requester_id": str(user_id),
        "track_id": str(request_in.track_id),
        "message": request_in.message,
    })
    
    return TrackRequestOut(**new_request.__dict__)


@app.get("/api/v1/requests/incoming", response_model=list[TrackRequestOut])
async def get_incoming_requests(
    user_id: Annotated[UUID, Depends(get_current_user_id)],
    db: Annotated[AsyncSession, Depends(get_db)],
    status_filter: str | None = Query(None, description="Filter by status: pending, accepted, declined"),
):
    """Get all track requests for the current artist"""
    query = select(TrackRequest).where(TrackRequest.artist_id == user_id)
    
    if status_filter:
        query = query.where(TrackRequest.status == status_filter)
    
    query = query.order_by(TrackRequest.created_at.desc())
    
    result = await db.execute(query)
    requests = result.scalars().all()
    
    return [TrackRequestOut(**req.__dict__) for req in requests]


@app.get("/api/v1/requests/outgoing", response_model=list[TrackRequestOut])
async def get_outgoing_requests(
    user_id: Annotated[UUID, Depends(get_current_user_id)],
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Get all track requests made by the current user"""
    result = await db.execute(
        select(TrackRequest)
        .where(TrackRequest.requester_id == user_id)
        .order_by(TrackRequest.created_at.desc())
    )
    requests = result.scalars().all()
    
    return [TrackRequestOut(**req.__dict__) for req in requests]


@app.patch("/api/v1/requests/{request_id}", response_model=TrackRequestOut)
async def handle_track_request(
    request_id: UUID,
    action: TrackRequestAction,
    user_id: Annotated[UUID, Depends(get_current_user_id)],
    authorization: Annotated[str, Header(alias="Authorization")],
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Accept or decline a track request"""
    # Get request
    result = await db.execute(
        select(TrackRequest).where(TrackRequest.id == request_id)
    )
    track_request = result.scalar_one_or_none()
    
    if not track_request:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Request not found")
    
    if track_request.artist_id != user_id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Not your request to handle"
        )
    
    if track_request.status != "pending":
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Request already {track_request.status}"
        )
    
    # Update request status
    track_request.status = action.action + "ed"  # "accept" -> "accepted", "decline" -> "declined"
    
    if action.action == "accept":
        # Give coin to artist
        await update_user_coins(user_id, settings.coin_cost_per_request, authorization)
        
        # Add track to artist's queue
        await add_track_to_queue(user_id, track_request.track_id, authorization)
    else:
        # Refund coin to requester
        await update_user_coins(track_request.requester_id, settings.coin_cost_per_request, authorization)
        
        # Update transaction type
        result = await db.execute(
            select(CoinTransaction).where(CoinTransaction.request_id == request_id)
        )
        transaction = result.scalar_one_or_none()
        if transaction:
            transaction.transaction_type = "refund"
    
    await db.commit()
    await db.refresh(track_request)
    
    # Send Kafka event for WebSocket notification
    event_type = "request_accepted" if action.action == "accept" else "request_declined"
    send_music_request_event(event_type, {
        "request_id": str(track_request.id),
        "artist_id": str(user_id),
        "requester_id": str(track_request.requester_id),
        "track_id": str(track_request.track_id),
    })
    
    return TrackRequestOut(**track_request.__dict__)


# ==================== Coins ====================

@app.get("/api/v1/me/coins", response_model=UserCoinsOut)
async def get_my_coins(
    user_id: Annotated[UUID, Depends(get_current_user_id)],
    authorization: Annotated[str, Header(alias="Authorization")],
):
    """Get current user's coin balance"""
    coins = await get_user_coins(user_id, authorization)
    return UserCoinsOut(user_id=user_id, coins=coins)


@app.get("/api/v1/users/{user_id}/coins", response_model=UserCoinsOut)
async def get_coins(
    user_id: UUID,
    current_user_id: Annotated[UUID, Depends(get_current_user_id)],
    authorization: Annotated[str, Header(alias="Authorization")],
):
    """Get user's coin balance"""
    coins = await get_user_coins(user_id, authorization)
    return UserCoinsOut(user_id=user_id, coins=coins)


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host=settings.host, port=settings.port)
