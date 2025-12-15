from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, status

from schemas.session import RouteSessionStart, RouteSessionUpdate, RouteSessionResponse
from services.session_service import RouteSessionService
from services.kafka_service import KafkaService

router = APIRouter(prefix="/sessions", tags=["sessions"])

session_service = RouteSessionService()
kafka_service = KafkaService()


# Temporary dependency for user_id
async def get_current_user_id() -> UUID:
    """Get current user ID from JWT token."""
    return UUID("00000000-0000-0000-0000-000000000001")


@router.post("/start", response_model=RouteSessionResponse, status_code=status.HTTP_201_CREATED)
async def start_session(
    session_data: RouteSessionStart,
    user_id: UUID = Depends(get_current_user_id),
):
    """Start a new route session."""
    if session_data.user_id != user_id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="User ID mismatch",
        )

    session = await session_service.start_session(user_id, session_data.route_id)

    # Send event to Kafka
    await kafka_service.send_event(
        "route-session-events",
        {
            "type": "route_session_started",
            "session_id": str(session["id"]),
            "user_id": str(user_id),
            "route_id": str(session_data.route_id),
        },
    )

    return RouteSessionResponse(**session)


@router.get("/{session_id}", response_model=RouteSessionResponse)
async def get_session(session_id: UUID):
    """Get session by ID."""
    session = await session_service.get_session(session_id)
    if not session:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Session not found",
        )
    return RouteSessionResponse(**session)


@router.post("/{session_id}/location", status_code=status.HTTP_200_OK)
async def update_location(
    session_id: UUID,
    location: RouteSessionUpdate,
):
    """Update session location and check for triggered points."""
    result = await session_service.update_location(
        session_id,
        location.latitude,
        location.longitude,
    )

    if result and "triggered_point" in result:
        # Send event to Kafka
        await kafka_service.send_event(
            "route-session-events",
            {
                "type": "route_point_triggered",
                "session_id": str(session_id),
                "point": result["triggered_point"],
            },
        )

        return {
            "session": RouteSessionResponse(**{k: v for k, v in result.items() if k != "triggered_point"}),
            "triggered_point": result["triggered_point"],
        }

    return {"session": None, "triggered_point": None}


@router.post("/{session_id}/pause", response_model=RouteSessionResponse)
async def pause_session(
    session_id: UUID,
    user_id: UUID = Depends(get_current_user_id),
):
    """Pause a route session."""
    session = await session_service.pause_session(session_id, user_id)
    if not session:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Session not found",
        )

    # Send event to Kafka
    await kafka_service.send_event(
        "route-session-events",
        {
            "type": "route_session_paused",
            "session_id": str(session_id),
        },
    )

    return RouteSessionResponse(**session)


@router.post("/{session_id}/resume", response_model=RouteSessionResponse)
async def resume_session(
    session_id: UUID,
    user_id: UUID = Depends(get_current_user_id),
):
    """Resume a route session."""
    session = await session_service.resume_session(session_id, user_id)
    if not session:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Session not found",
        )

    # Send event to Kafka
    await kafka_service.send_event(
        "route-session-events",
        {
            "type": "route_session_resumed",
            "session_id": str(session_id),
        },
    )

    return RouteSessionResponse(**session)


@router.post("/{session_id}/end", response_model=RouteSessionResponse)
async def end_session(
    session_id: UUID,
    completed: bool = False,
    user_id: UUID = Depends(get_current_user_id),
):
    """End a route session."""
    session = await session_service.end_session(session_id, user_id, completed)
    if not session:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Session not found",
        )

    # Send event to Kafka
    await kafka_service.send_event(
        "route-session-events",
        {
            "type": "route_session_ended",
            "session_id": str(session_id),
            "completed": completed,
        },
    )

    return RouteSessionResponse(**session)

