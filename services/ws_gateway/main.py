from __future__ import annotations

import asyncio
import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI, WebSocket, WebSocketDisconnect, Query, status
from fastapi.responses import JSONResponse

from api.websocket import get_manager, init_manager
from core.config import Settings, get_settings
from services.jwt_service import JWTService
from services.kafka_service import KafkaService
from services.redis_service import RedisService

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)

settings = get_settings()

# Global services
redis_service: RedisService | None = None
kafka_service: KafkaService | None = None
jwt_service: JWTService | None = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifespan context manager for startup/shutdown."""
    global redis_service, kafka_service, jwt_service

    # Initialize services
    redis_service = RedisService(settings)
    kafka_service = KafkaService(settings)
    jwt_service = JWTService(settings)

    # Connect to services
    await redis_service.connect()
    kafka_service.connect_producer()

    # Initialize connection manager
    manager = init_manager(redis_service, kafka_service, jwt_service, settings)

    # Start Kafka consumer (get event loop first)
    import asyncio
    loop = asyncio.get_event_loop()
    manager.start_kafka_consumer(loop)

    logger.info("WebSocket Gateway started successfully")

    yield

    # Cleanup
    await redis_service.disconnect()
    kafka_service.disconnect_producer()
    kafka_service.disconnect_consumer()
    logger.info("WebSocket Gateway shut down")


app = FastAPI(
    title="WebSocket Gateway",
    description="WebSocket Gateway for Music Session Service",
    version=settings.app_version,
    lifespan=lifespan,
)


@app.get("/health")
async def health_check():
    """Health check endpoint."""
    health_status = {
        "status": "healthy",
        "redis": False,
        "kafka": False,
    }

    if redis_service:
        health_status["redis"] = await redis_service.ping()

    if kafka_service and kafka_service.producer:
        health_status["kafka"] = True

    overall_status = "healthy" if health_status["redis"] and health_status["kafka"] else "unhealthy"

    return JSONResponse(
        status_code=200 if overall_status == "healthy" else 503,
        content={"status": overall_status, **health_status},
    )


@app.websocket("/ws")
async def websocket_endpoint(
    websocket: WebSocket,
    token: str = Query(..., description="JWT token"),
    room_id: str = Query(..., description="Room ID"),
):
    """WebSocket endpoint for client connections."""
    try:
        logger.info(f"[WS Endpoint] Connection attempt - room_id: {room_id}, token present: {bool(token)}")
        
        # Validate token before accepting connection
        manager = get_manager()
        jwt_service = JWTService(settings)
        
        claims = jwt_service.validate_token(token)
        if not claims:
            logger.warning(f"[WS Endpoint] Invalid token - rejecting connection")
            await websocket.close(code=status.WS_1008_POLICY_VIOLATION, reason="Invalid token")
            return
            
        user_id = claims.get("user_id")
        if not user_id:
            logger.warning(f"[WS Endpoint] No user_id in token - rejecting connection")
            await websocket.close(code=status.WS_1008_POLICY_VIOLATION, reason="Missing user_id")
            return
            
        logger.info(f"[WS Endpoint] Token validated, user_id: {user_id}")
        
        # Connect client (will accept connection inside)
        result = await manager.connect(websocket, token, room_id)
        if not result:
            logger.warning(f"[WS Endpoint] Connection rejected by manager - room_id: {room_id}")
            return

        connection_id, user_id = result
        logger.info(f"[WS Endpoint] Client connected: {connection_id} (user: {user_id}, room: {room_id})")
    except Exception as e:
        logger.error(f"[WS Endpoint] Error during connection: {e}", exc_info=True)
        try:
            await websocket.close(code=status.WS_1011_INTERNAL_ERROR)
        except:
            pass
        return

    try:
        # Send ping periodically
        async def ping_task():
            while connection_id in manager.active_connections:
                try:
                    pong_msg = {
                        "type": "pong",
                        "timestamp": asyncio.get_event_loop().time(),
                    }
                    await websocket.send_json(pong_msg)
                except Exception as e:
                    logger.error(f"Error sending ping: {e}")
                    break
                await asyncio.sleep(settings.ping_interval)

        ping_task_handle = asyncio.create_task(ping_task())

        # Main message loop
        while True:
            try:
                logger.info(f"[WS Endpoint] Waiting for message from client {connection_id}...")
                data = await websocket.receive_json()
                logger.info(f"[WS Endpoint] Received message from client {connection_id}: {data}")
                await manager.handle_message(connection_id, data)
                logger.info(f"[WS Endpoint] Message handled successfully for {connection_id}")
            except WebSocketDisconnect:
                logger.info(f"Client disconnected: {connection_id}")
                break
            except Exception as e:
                logger.error(f"Error handling message: {e}")
                error_msg = {
                    "type": "error",
                    "error": f"Internal server error: {str(e)}",
                }
                try:
                    await websocket.send_json(error_msg)
                except Exception:
                    break

        ping_task_handle.cancel()

    except Exception as e:
        logger.error(f"WebSocket error: {e}")
    finally:
        await manager.disconnect(connection_id)
        logger.info(f"Connection closed: {connection_id}")


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "main:app",
        host=settings.host,
        port=settings.port,
        log_level=settings.log_level.lower(),
    )

