from __future__ import annotations

from datetime import datetime
from enum import Enum
from typing import Any

from pydantic import BaseModel, Field


class ClientMessageType(str, Enum):
    """Types of client messages."""

    PLAYER_ACTION = "player_action"


class PlayerAction(str, Enum):
    """Player action types."""

    PLAY = "play"
    PAUSE = "pause"
    SEEK = "seek"
    CHANGE_TRACK = "change_track"


class ClientMessage(BaseModel):
    """Incoming message from client."""

    type: ClientMessageType
    room_id: str = Field(..., alias="room_id")
    user_id: str = Field(..., alias="user_id")
    action: PlayerAction
    payload: dict[str, Any] = Field(default_factory=dict)
    timestamp: float | None = None

    class Config:
        populate_by_name = True


class ServerMessageType(str, Enum):
    """Types of server messages."""

    SYNC_STATE = "sync_state"
    MESSAGE = "message"
    CONVERSATION_READ = "conversation_read"
    ERROR = "error"
    PONG = "pong"
    # Music request events
    MUSIC_REQUEST_NEW = "music_request_new"
    MUSIC_REQUEST_ACCEPTED = "music_request_accepted"
    MUSIC_REQUEST_DECLINED = "music_request_declined"


class ServerMessage(BaseModel):
    """Outgoing message to client."""

    type: ServerMessageType
    room_id: str | None = None
    state: dict[str, Any] | None = None
    message: dict[str, Any] | None = None
    conversation_id: str | None = None
    error: str | None = None
    timestamp: float | None = None
    # Music request data
    music_request: dict[str, Any] | None = None

    class Config:
        populate_by_name = True
