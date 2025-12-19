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
        """Start consuming sync and messaging events from Kafka."""
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
                                topic = topic_partition.topic
                                value = message.value
                                print(f"[WS Gateway] Kafka message from topic={topic}: {value}")
                                
                                if topic == self.settings.kafka_sync_topic:
                                    # старый путь синхронизации сессий
                                    self._handle_session_sync_message(loop, value)
                                elif topic == self.settings.kafka_messaging_topic:
                                    self._handle_messaging_event(loop, value)
                            except Exception as e:
                                print(f"[WS Gateway] Error processing Kafka message: {e}")
                                import traceback
                                traceback.print_exc()

                except Exception as e:
                    print(f"Error in Kafka consumer loop: {e}")
                    time.sleep(1)

        # Run consumer in background thread
        import threading
        thread = threading.Thread(target=consume_loop, daemon=True)
        thread.start()

    def _handle_session_sync_message(self, loop: asyncio.AbstractEventLoop, sync_data: dict[str, Any]) -> None:
        from datetime import datetime

        print(f"[WS Gateway] ========== RECEIVED SYNC MESSAGE FROM KAFKA ==========")
        print(f"[WS Gateway] Full sync_data: {sync_data}")

        room_id = sync_data.get("roomId") or sync_data.get("room_id")
        state = sync_data.get("state")

        if not room_id or not state:
            print(f"[WS Gateway] Skipping sync message - missing room_id or state: room_id={room_id}, state={state is not None}")
            return

        timestamp = None
        timestamp_value = sync_data.get("timestamp")
        if timestamp_value:
            if isinstance(timestamp_value, str):
                try:
                    dt = datetime.fromisoformat(timestamp_value.replace("Z", "+00:00"))
                    timestamp = dt.timestamp()
                except Exception as e:
                    print(f"[WS Gateway] Failed to parse timestamp {timestamp_value}: {e}")
            elif isinstance(timestamp_value, (int, float)):
                timestamp = float(timestamp_value)

        server_message = ServerMessage(
            type=ServerMessageType.SYNC_STATE,
            room_id=room_id,
            state=state,
            timestamp=timestamp,
        )

        message_dict = server_message.model_dump()
        asyncio.run_coroutine_threadsafe(
            self.broadcast_to_room(message_dict, room_id),
            loop,
        )

    def _handle_messaging_event(self, loop: asyncio.AbstractEventLoop, event: dict[str, Any]) -> None:
        """Handle messaging events from messaging-service."""
        event_type = event.get("event_type")
        print(f"[WS Gateway] Handling messaging event: {event_type}")

        if event_type == "message_sent":
            conv_id = event.get("conversation_id")
            if not conv_id:
                return
            server_message = ServerMessage(
                type=ServerMessageType.MESSAGE,
                conversation_id=conv_id,
                message=event,
            )
            message_dict = server_message.model_dump()
            # Для начала рассылаем всем активным подключениям (фронт сам отфильтрует)
            asyncio.run_coroutine_threadsafe(
                self._broadcast_to_all(message_dict),
                loop,
            )
        elif event_type == "conversation_read":
            conv_id = event.get("conversation_id")
            if not conv_id:
                return
            server_message = ServerMessage(
                type=ServerMessageType.CONVERSATION_READ,
                conversation_id=conv_id,
                message=event,
            )
            message_dict = server_message.model_dump()
            asyncio.run_coroutine_threadsafe(
                self._broadcast_to_all(message_dict),
                loop,
            )

    async def _broadcast_to_all(self, message: dict[str, Any]) -> None:
        """Broadcast message to all active WebSocket connections (for messaging events)."""
        for connection_id in list(self.active_connections.keys()):
            await self.send_personal_message(message, connection_id)


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

