from datetime import datetime
from typing import Optional
from uuid import UUID, uuid4

import asyncpg

from core.database import get_connection
from schemas.point import RoutePointCreate, RoutePointUpdate
from utils.geohash import encode_geohash


class PointService:
    """Service for route point operations."""

    async def add_point(self, route_id: UUID, user_id: UUID, point_data: RoutePointCreate) -> dict:
        """Add a point to a route."""
        async with get_connection() as conn:
            # Check route ownership
            route = await conn.fetchrow(
                "SELECT user_id FROM routes WHERE id = $1 AND deleted_at IS NULL",
                route_id,
            )
            if not route:
                raise ValueError("Route not found")
            if route["user_id"] != user_id:
                raise PermissionError("You don't have permission to modify this route")

            # Determine order_index
            if point_data.order_index is None:
                max_order = await conn.fetchval(
                    "SELECT COALESCE(MAX(order_index), -1) FROM route_points WHERE route_id = $1",
                    route_id,
                )
                order_index = max_order + 1
            else:
                order_index = point_data.order_index
                # Shift existing points
                await conn.execute(
                    "UPDATE route_points SET order_index = order_index + 1 WHERE route_id = $1 AND order_index >= $2",
                    route_id,
                    order_index,
                )

            # Generate geohash
            geohash = encode_geohash(point_data.latitude, point_data.longitude)

            # Insert point
            point_id = uuid4()
            point = await conn.fetchrow(
                """
                INSERT INTO route_points (
                    id, route_id, latitude, longitude, geohash,
                    radius_meters, track_id, track_start_offset_sec,
                    order_index, title, description, image_url, created_at
                ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
                RETURNING *
                """,
                point_id,
                route_id,
                point_data.latitude,
                point_data.longitude,
                geohash,
                point_data.radius_meters,
                point_data.track_id,
                point_data.track_start_offset_sec,
                order_index,
                point_data.title,
                point_data.description,
                point_data.image_url,
                datetime.utcnow(),
            )

            # Recalculate route distance
            await self._recalculate_route_distance(conn, route_id)

            return dict(point)

    async def update_point(self, point_id: UUID, user_id: UUID, point_data: RoutePointUpdate) -> Optional[dict]:
        """Update a route point."""
        async with get_connection() as conn:
            # Check point and route ownership
            point = await conn.fetchrow(
                """
                SELECT rp.*, r.user_id as route_user_id
                FROM route_points rp
                JOIN routes r ON rp.route_id = r.id
                WHERE rp.id = $1 AND r.deleted_at IS NULL
                """,
                point_id,
            )
            if not point:
                return None
            if point["route_user_id"] != user_id:
                raise PermissionError("You don't have permission to modify this route")

            # Build update query
            updates = []
            values = []
            param_num = 1

            if point_data.latitude is not None:
                updates.append(f"latitude = ${param_num}")
                values.append(point_data.latitude)
                param_num += 1

            if point_data.longitude is not None:
                updates.append(f"longitude = ${param_num}")
                values.append(point_data.longitude)
                param_num += 1

            if point_data.radius_meters is not None:
                updates.append(f"radius_meters = ${param_num}")
                values.append(point_data.radius_meters)
                param_num += 1

            if point_data.track_id is not None:
                updates.append(f"track_id = ${param_num}")
                values.append(point_data.track_id)
                param_num += 1

            if point_data.track_start_offset_sec is not None:
                updates.append(f"track_start_offset_sec = ${param_num}")
                values.append(point_data.track_start_offset_sec)
                param_num += 1

            if point_data.title is not None:
                updates.append(f"title = ${param_num}")
                values.append(point_data.title)
                param_num += 1

            if point_data.description is not None:
                updates.append(f"description = ${param_num}")
                values.append(point_data.description)
                param_num += 1

            if point_data.image_url is not None:
                updates.append(f"image_url = ${param_num}")
                values.append(point_data.image_url)
                param_num += 1

            if point_data.order_index is not None:
                updates.append(f"order_index = ${param_num}")
                values.append(point_data.order_index)
                param_num += 1

            if not updates:
                updated_point = await conn.fetchrow("SELECT * FROM route_points WHERE id = $1", point_id)
                return dict(updated_point) if updated_point else None

            # Recalculate geohash if location changed
            if point_data.latitude is not None or point_data.longitude is not None:
                # Get current values
                current = await conn.fetchrow("SELECT latitude, longitude FROM route_points WHERE id = $1", point_id)
                lat = point_data.latitude if point_data.latitude is not None else current["latitude"]
                lng = point_data.longitude if point_data.longitude is not None else current["longitude"]
                geohash = encode_geohash(lat, lng)
                updates.append(f"geohash = ${param_num}")
                values.append(geohash)
                param_num += 1

            values.append(point_id)
            query = f"UPDATE route_points SET {', '.join(updates)} WHERE id = ${param_num} RETURNING *"
            updated_point = await conn.fetchrow(query, *values)

            # Recalculate route distance
            route_id = updated_point["route_id"]
            await self._recalculate_route_distance(conn, route_id)

            return dict(updated_point)

    async def delete_point(self, point_id: UUID, user_id: UUID) -> bool:
        """Delete a route point."""
        async with get_connection() as conn:
            # Check point and route ownership
            point = await conn.fetchrow(
                """
                SELECT rp.route_id, r.user_id as route_user_id
                FROM route_points rp
                JOIN routes r ON rp.route_id = r.id
                WHERE rp.id = $1 AND r.deleted_at IS NULL
                """,
                point_id,
            )
            if not point:
                return False
            if point["route_user_id"] != user_id:
                raise PermissionError("You don't have permission to modify this route")

            route_id = point["route_id"]
            order_index = await conn.fetchval("SELECT order_index FROM route_points WHERE id = $1", point_id)

            # Delete point
            await conn.execute("DELETE FROM route_points WHERE id = $1", point_id)

            # Shift remaining points
            await conn.execute(
                "UPDATE route_points SET order_index = order_index - 1 WHERE route_id = $1 AND order_index > $2",
                route_id,
                order_index,
            )

            # Recalculate route distance
            await self._recalculate_route_distance(conn, route_id)

            return True

    async def get_route_points(self, route_id: UUID) -> list[dict]:
        """Get all points for a route."""
        async with get_connection() as conn:
            points = await conn.fetch(
                "SELECT * FROM route_points WHERE route_id = $1 ORDER BY order_index",
                route_id,
            )
            return [dict(point) for point in points]

    async def reorder_points(self, route_id: UUID, user_id: UUID, point_ids: list[UUID]) -> bool:
        """Reorder route points."""
        async with get_connection() as conn:
            # Check route ownership
            route = await conn.fetchrow(
                "SELECT user_id FROM routes WHERE id = $1 AND deleted_at IS NULL",
                route_id,
            )
            if not route:
                raise ValueError("Route not found")
            if route["user_id"] != user_id:
                raise PermissionError("You don't have permission to modify this route")

            # Verify all points belong to this route
            point_count = await conn.fetchval(
                "SELECT COUNT(*) FROM route_points WHERE route_id = $1 AND id = ANY($2::uuid[])",
                route_id,
                point_ids,
            )
            if point_count != len(point_ids):
                raise ValueError("Some points don't belong to this route")

            # Update order_index for each point
            for index, point_id in enumerate(point_ids):
                await conn.execute(
                    "UPDATE route_points SET order_index = $1 WHERE id = $2",
                    index,
                    point_id,
                )

            return True

    async def _recalculate_route_distance(self, conn: asyncpg.Connection, route_id: UUID) -> None:
        """Recalculate total distance for a route."""
        points = await conn.fetch(
            "SELECT latitude, longitude FROM route_points WHERE route_id = $1 ORDER BY order_index",
            route_id,
        )

        if len(points) < 2:
            await conn.execute(
                "UPDATE routes SET total_distance_km = 0 WHERE id = $1",
                route_id,
            )
            return

        # Calculate distance using Haversine formula
        total_distance = 0.0
        for i in range(len(points) - 1):
            lat1, lon1 = points[i]["latitude"], points[i]["longitude"]
            lat2, lon2 = points[i + 1]["latitude"], points[i + 1]["longitude"]

            # Haversine formula
            from math import radians, cos, sin, asin, sqrt

            R = 6371  # Earth radius in km
            dlat = radians(lat2 - lat1)
            dlon = radians(lon2 - lon1)
            a = sin(dlat / 2) ** 2 + cos(radians(lat1)) * cos(radians(lat2)) * sin(dlon / 2) ** 2
            c = 2 * asin(sqrt(a))
            distance = R * c

            total_distance += distance

        await conn.execute(
            "UPDATE routes SET total_distance_km = $1 WHERE id = $2",
            round(total_distance, 2),
            route_id,
        )

