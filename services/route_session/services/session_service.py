from datetime import datetime
from typing import Optional
from uuid import UUID, uuid4
import json

import asyncpg
import httpx

from core.database import get_connection
from core.config import get_settings
from schemas.session import RouteSessionStart, RouteSessionUpdate

settings = get_settings()


class RouteSessionService:
    """Service for route session operations."""

    async def start_session(self, user_id: UUID, route_id: UUID) -> dict:
        """Start a new route session."""
        async with get_connection() as conn:
            session_id = uuid4()
            now = datetime.utcnow()

            # Create session in database
            session = await conn.fetchrow(
                """
                INSERT INTO route_sessions (
                    id, user_id, route_id, started_at,
                    current_point_index, visited_points, status
                ) VALUES ($1, $2, $3, $4, $5, $6, $7)
                RETURNING *
                """,
                session_id,
                user_id,
                route_id,
                now,
                0,
                json.dumps([]),
                "active",
            )

            return dict(session)

    async def get_session(self, session_id: UUID) -> Optional[dict]:
        """Get session by ID."""
        async with get_connection() as conn:
            session = await conn.fetchrow(
                "SELECT * FROM route_sessions WHERE id = $1",
                session_id,
            )
            if not session:
                return None

            session_dict = dict(session)
            # Parse visited_points JSON
            if session_dict.get("visited_points"):
                session_dict["visited_points"] = json.loads(session_dict["visited_points"])
            else:
                session_dict["visited_points"] = []

            return session_dict

    async def update_location(
        self,
        session_id: UUID,
        latitude: float,
        longitude: float,
    ) -> Optional[dict]:
        """Update session location and check for triggered points."""
        async with get_connection() as conn:
            # Get session
            session = await conn.fetchrow(
                "SELECT * FROM route_sessions WHERE id = $1 AND status = 'active'",
                session_id,
            )
            if not session:
                return None

            # Get route points from routes service
            async with httpx.AsyncClient() as client:
                try:
                    response = await client.get(
                        f"{settings.routes_service_url}/api/v1/routes/{session['route_id']}/points",
                        timeout=5.0,
                    )
                    if response.status_code != 200:
                        return None
                    points = response.json()
                except Exception:
                    return None

            # Parse visited points
            visited_points = json.loads(session["visited_points"]) if session["visited_points"] else []

            # Check if we're near any unvisited points
            triggered_point = None
            for point in points:
                point_id = str(point["id"])
                if point_id in visited_points:
                    continue

                # Calculate distance (Haversine formula)
                from math import radians, cos, sin, asin, sqrt

                lat1, lon1 = radians(latitude), radians(longitude)
                lat2, lon2 = radians(float(point["latitude"])), radians(float(point["longitude"]))

                R = 6371000  # Earth radius in meters
                dlat = lat2 - lat1
                dlon = lon2 - lon1
                a = sin(dlat / 2) ** 2 + cos(lat1) * cos(lat2) * sin(dlon / 2) ** 2
                c = 2 * asin(sqrt(a))
                distance = R * c

                if distance <= point["radius_meters"]:
                    triggered_point = point
                    break

            # Update session if point triggered
            if triggered_point:
                visited_points.append(str(triggered_point["id"]))
                current_index = max(
                    (i for i, p in enumerate(points) if str(p["id"]) == str(triggered_point["id"])),
                    default=0,
                )

                await conn.execute(
                    """
                    UPDATE route_sessions
                    SET visited_points = $1, current_point_index = $2, updated_at = NOW()
                    WHERE id = $3
                    """,
                    json.dumps(visited_points),
                    current_index,
                    session_id,
                )

                # Get updated session
                updated_session = await conn.fetchrow(
                    "SELECT * FROM route_sessions WHERE id = $1",
                    session_id,
                )
                session_dict = dict(updated_session)
                session_dict["visited_points"] = visited_points
                session_dict["triggered_point"] = triggered_point
                return session_dict

            return None

    async def pause_session(self, session_id: UUID, user_id: UUID) -> Optional[dict]:
        """Pause a route session."""
        async with get_connection() as conn:
            session = await conn.fetchrow(
                "SELECT * FROM route_sessions WHERE id = $1 AND user_id = $2",
                session_id,
                user_id,
            )
            if not session:
                return None

            await conn.execute(
                "UPDATE route_sessions SET status = 'paused', updated_at = NOW() WHERE id = $1",
                session_id,
            )

            updated = await conn.fetchrow(
                "SELECT * FROM route_sessions WHERE id = $1",
                session_id,
            )
            return dict(updated)

    async def resume_session(self, session_id: UUID, user_id: UUID) -> Optional[dict]:
        """Resume a paused route session."""
        async with get_connection() as conn:
            session = await conn.fetchrow(
                "SELECT * FROM route_sessions WHERE id = $1 AND user_id = $2",
                session_id,
                user_id,
            )
            if not session:
                return None

            await conn.execute(
                "UPDATE route_sessions SET status = 'active', updated_at = NOW() WHERE id = $1",
                session_id,
            )

            updated = await conn.fetchrow(
                "SELECT * FROM route_sessions WHERE id = $1",
                session_id,
            )
            return dict(updated)

    async def end_session(self, session_id: UUID, user_id: UUID, completed: bool = False) -> Optional[dict]:
        """End a route session."""
        async with get_connection() as conn:
            session = await conn.fetchrow(
                "SELECT * FROM route_sessions WHERE id = $1 AND user_id = $2",
                session_id,
                user_id,
            )
            if not session:
                return None

            status = "completed" if completed else "abandoned"
            await conn.execute(
                """
                UPDATE route_sessions
                SET status = $1, completed_at = NOW(), updated_at = NOW()
                WHERE id = $2
                """,
                status,
                session_id,
            )

            # Update route stats
            if completed:
                await conn.execute(
                    """
                    INSERT INTO route_stats (route_id, total_completions, last_completed_at)
                    VALUES ($1, 1, NOW())
                    ON CONFLICT (route_id) DO UPDATE
                    SET total_completions = route_stats.total_completions + 1,
                        last_completed_at = NOW(),
                        updated_at = NOW()
                    """,
                    session["route_id"],
                )

            updated = await conn.fetchrow(
                "SELECT * FROM route_sessions WHERE id = $1",
                session_id,
            )
            return dict(updated)

