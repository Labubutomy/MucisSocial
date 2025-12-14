from __future__ import annotations

import asyncio
import json
import uuid
from typing import Any

from fastapi import WebSocket, WebSocketDisconnect, Query, status
from fastapi.responses import JSONResponse

from core.config import Settings
from models.messages import ClientMessage, ServerMessage, ServerMessageType
from services.jwt_service import JWTService
from services.kafka_service import KafkaService
from services.redis_service import RedisService


class ConnectionManager:
    """Manages WebSocket connections."""

    def __init__(
        self,
        redis_service: RedisService,
        kafka_service: KafkaService,
        jwt_service: JWTService,
        settings: Settings,
    ):
        self.redis_service = redis_service
        self.kafka_service = kafka_service
        self.jwt_service = jwt_service
        self.settings = settings
        self.active_connections: dict[str, WebSocket] = {}
        self.connection_rooms: dict[str, str] = {}  # connection_id -> room_id
        self.connection_users: dict[str, str] = {}  # connection_id -> user_id

    async def connect(
        self, websocket: WebSocket, token: str, room_id: str
    ) -> tuple[str, str] | None:
        """Connect a WebSocket client."""
        try:
            print(f"[ConnectionManager] Connecting client - room_id: {room_id}")
            
            # Token already validated in endpoint, just extract user_id
            claims = self.jwt_service.validate_token(token)
            if not claims:
                print(f"[ConnectionManager] JWT token validation failed (should not happen)")
                return None
                
            user_id = claims.get("user_id")
            if not user_id:
                print(f"[ConnectionManager] No user_id in JWT claims (should not happen)")
                return None

            print(f"[ConnectionManager] User ID extracted: {user_id}")

            # Accept connection
            await websocket.accept()
            print(f"[ConnectionManager] WebSocket connection accepted")

            # Generate connection ID
            connection_id = str(uuid.uuid4())
            print(f"[ConnectionManager] Generated connection ID: {connection_id}")

            # Store connection
            self.active_connections[connection_id] = websocket
            self.connection_rooms[connection_id] = room_id
            self.connection_users[connection_id] = user_id
            print(f"[ConnectionManager] Connection stored in memory")

            # Add to Redis (non-blocking)
            try:
                await self.redis_service.add_connection(room_id, user_id, connection_id)
                print(f"[ConnectionManager] Connection added to Redis")
            except Exception as e:
                print(f"[ConnectionManager] Warning: Failed to add connection to Redis: {e}")
                import traceback
                traceback.print_exc()
                # Continue anyway - connection is established in memory

            return connection_id, user_id
        except Exception as e:
            print(f"[ConnectionManager] Error in connect: {e}")
            import traceback
            traceback.print_exc()
            try:
                await websocket.close(code=status.WS_1011_INTERNAL_ERROR, reason=str(e))
            except:
                pass
            return None

    async def disconnect(self, connection_id: str) -> None:
        """Disconnect a WebSocket client."""
        if connection_id not in self.active_connections:
            return

        room_id = self.connection_rooms.get(connection_id)
        user_id = self.connection_users.get(connection_id)

        if room_id and user_id:
            await self.redis_service.remove_connection(room_id, user_id, connection_id)

        del self.active_connections[connection_id]
        if connection_id in self.connection_rooms:
            del self.connection_rooms[connection_id]
        if connection_id in self.connection_users:
            del self.connection_users[connection_id]

    async def send_personal_message(
        self, message: dict[str, Any], connection_id: str
    ) -> None:
        """Send message to a specific connection."""
        if connection_id in self.active_connections:
            websocket = self.active_connections[connection_id]
            try:
                await websocket.send_json(message)
            except Exception as e:
                print(f"Error sending message to {connection_id}: {e}")
                await self.disconnect(connection_id)

    async def broadcast_to_room(
        self, message: dict[str, Any], room_id: str, exclude_connection_id: str | None = None
    ) -> None:
        """Broadcast message to all connections in a room."""
        connections = await self.redis_service.get_room_connections(room_id)
        
        for conn_str in connections:
            # Parse connection string (user_id:connection_id)
            parts = conn_str.split(":", 1)
            if len(parts) != 2:
                continue
            connection_id = parts[1]
            
            if connection_id == exclude_connection_id:
                continue
                
            if connection_id in self.active_connections:
                await self.send_personal_message(message, connection_id)

    async def handle_message(
        self, connection_id: str, message: dict[str, Any]
    ) -> None:
        """Handle incoming message from client."""
        try:
            client_message = ClientMessage(**message)
        except Exception as e:
            error_msg = ServerMessage(
                type=ServerMessageType.ERROR,
                error=f"Invalid message format: {str(e)}",
            )
            await self.send_personal_message(error_msg.model_dump(), connection_id)
            return

        # Validate room_id matches connection
        room_id = self.connection_rooms.get(connection_id)
        user_id = self.connection_users.get(connection_id)

        if not room_id or not user_id:
            error_msg = ServerMessage(
                type=ServerMessageType.ERROR,
                error="Connection not properly initialized",
            )
            await self.send_personal_message(error_msg.model_dump(), connection_id)
            return

        if client_message.room_id != room_id:
            error_msg = ServerMessage(
                type=ServerMessageType.ERROR,
                error="Room ID mismatch",
            )
            await self.send_personal_message(error_msg.model_dump(), connection_id)
            return

        # Send to Kafka
        try:
            print(f"[ConnectionManager] Sending client event to Kafka: action={client_message.action}, room_id={client_message.room_id}")
            self.kafka_service.send_client_event(client_message)
            print(f"[ConnectionManager] Client event sent to Kafka successfully")
        except Exception as e:
            print(f"[ConnectionManager] Error sending event to Kafka: {e}")
            import traceback
            traceback.print_exc()
            error_msg = ServerMessage(
                type=ServerMessageType.ERROR,
                error=f"Failed to process event: {str(e)}",
            )
            await self.send_personal_message(error_msg.model_dump(), connection_id)

    def start_kafka_consumer(self, loop: asyncio.AbstractEventLoop) -> None:
        """Start consuming sync messages from Kafka."""
        self.kafka_service.connect_consumer()

        def consume_loop():
            """Synchronous Kafka consumer loop running in thread."""
            import threading
            import time
            while True:
                try:
                    if not self.kafka_service.consumer:
                        time.sleep(1)
                        continue

                    # Poll for messages (blocking)
                    message_pack = self.kafka_service.consumer.poll(timeout_ms=1000)
                    
                    for topic_partition, messages in message_pack.items():
                        for message in messages:
                            try:
                                sync_data = message.value
                                print(f"[WS Gateway] ========== RECEIVED SYNC MESSAGE FROM KAFKA ==========")
                                print(f"[WS Gateway] Full sync_data: {sync_data}")
                                
                                room_id = sync_data.get("roomId") or sync_data.get("room_id")
                                state = sync_data.get("state")
                                
                                # Log state details
                                if state:
                                    # Check all possible field names for isPlaying
                                    state_is_playing = (
                                        state.get("isPlaying") or 
                                        state.get("is_playing") or 
                                        state.get("playing")
                                    )
                                    state_track = state.get("currentTrack") or state.get("current_track")
                                    print(f"[WS Gateway] State details:")
                                    print(f"[WS Gateway]   - roomId: {state.get('roomId') or state.get('room_id')}")
                                    print(f"[WS Gateway]   - isPlaying: {state_is_playing}")
                                    print(f"[WS Gateway]   - playing (raw): {state.get('playing')}")
                                    print(f"[WS Gateway]   - isPlaying (raw): {state.get('isPlaying')}")
                                    print(f"[WS Gateway]   - is_playing (raw): {state.get('is_playing')}")
                                    print(f"[WS Gateway]   - currentTrack: {state_track}")
                                    print(f"[WS Gateway]   - position: {state.get('position')}")
                                    
                                    # Normalize the field name to isPlaying for consistency
                                    if "playing" in state and "isPlaying" not in state and "is_playing" not in state:
                                        state["isPlaying"] = state["playing"]
                                        print(f"[WS Gateway] Normalized 'playing' to 'isPlaying': {state['isPlaying']}")
                                
                                print(f"[WS Gateway] Parsed sync message - room_id: {room_id}, has_state: {state is not None}")
                                
                                if room_id and state:
                                    # Convert timestamp from ISO string to float (Unix timestamp)
                                    timestamp = None
                                    timestamp_value = sync_data.get("timestamp")
                                    if timestamp_value:
                                        if isinstance(timestamp_value, str):
                                            # Parse ISO 8601 string to Unix timestamp
                                            from datetime import datetime
                                            try:
                                                dt = datetime.fromisoformat(timestamp_value.replace("Z", "+00:00"))
                                                timestamp = dt.timestamp()
                                            except Exception as e:
                                                print(f"[WS Gateway] Failed to parse timestamp {timestamp_value}: {e}")
                                                timestamp = None
                                        elif isinstance(timestamp_value, (int, float)):
                                            # Already a number (Unix timestamp in seconds)
                                            timestamp = float(timestamp_value)
                                    
                                    server_message = ServerMessage(
                                        type=ServerMessageType.SYNC_STATE,
                                        room_id=room_id,
                                        state=state,
                                        timestamp=timestamp,
                                    )
                                    
                                    # Log what we're sending to clients
                                    message_dict = server_message.model_dump()
                                    state_to_send = message_dict.get("state", {})
                                    # Normalize playing field to isPlaying if needed
                                    if isinstance(state_to_send, dict):
                                        if "playing" in state_to_send and "isPlaying" not in state_to_send and "is_playing" not in state_to_send:
                                            state_to_send["isPlaying"] = state_to_send["playing"]
                                            message_dict["state"] = state_to_send
                                            print(f"[WS Gateway] Normalized 'playing' to 'isPlaying' in message: {state_to_send['isPlaying']}")
                                    
                                    is_playing_in_message = (
                                        state_to_send.get("isPlaying") or 
                                        state_to_send.get("is_playing") or 
                                        state_to_send.get("playing")
                                    )
                                    print(f"[WS Gateway] ========== BROADCASTING TO CLIENTS ==========")
                                    print(f"[WS Gateway] Room: {room_id}")
                                    print(f"[WS Gateway] isPlaying in message: {is_playing_in_message}")
                                    print(f"[WS Gateway] State keys: {list(state_to_send.keys()) if isinstance(state_to_send, dict) else 'not a dict'}")
                                    print(f"[WS Gateway] Full state object: {state_to_send}")
                                    print(f"[WS Gateway] Broadcasting sync_state to room {room_id}")
                                    
                                    # Schedule coroutine in the main event loop
                                    asyncio.run_coroutine_threadsafe(
                                        self.broadcast_to_room(
                                            message_dict, room_id
                                        ),
                                        loop
                                    )
                                else:
                                    print(f"[WS Gateway] Skipping sync message - missing room_id or state: room_id={room_id}, state={state is not None}")
                            except Exception as e:
                                print(f"[WS Gateway] Error processing sync message: {e}")
                                import traceback
                                traceback.print_exc()

                except Exception as e:
                    print(f"Error in Kafka consumer loop: {e}")
                    time.sleep(1)

        # Run consumer in background thread
        import threading
        thread = threading.Thread(target=consume_loop, daemon=True)
        thread.start()


# Global connection manager instance
_manager: ConnectionManager | None = None


def get_manager() -> ConnectionManager:
    """Get global connection manager."""
    if _manager is None:
        raise RuntimeError("Connection manager not initialized")
    return _manager


def init_manager(
    redis_service: RedisService,
    kafka_service: KafkaService,
    jwt_service: JWTService,
    settings: Settings,
) -> ConnectionManager:
    """Initialize global connection manager."""
    global _manager
    _manager = ConnectionManager(
        redis_service, kafka_service, jwt_service, settings
    )
    return _manager

