from datetime import datetime
from typing import Optional
from uuid import UUID, uuid4

import asyncpg

from core.database import get_connection
from schemas.route import RouteCreate, RouteUpdate
from services.point_service import PointService
from utils.geohash import encode_geohash


class RouteService:
    """Service for route operations."""

    def __init__(self):
        self.point_service = PointService()

    async def create_route(self, user_id: UUID, route_data: RouteCreate) -> dict:
        """Create a new route."""
        async with get_connection() as conn:
            route_id = uuid4()
            now = datetime.utcnow()

            # Insert route
            route = await conn.fetchrow(
                """
                INSERT INTO routes (
                    id, user_id, title, description, city, country,
                    is_public, is_linear, cover_image_url, created_at, updated_at
                ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
                RETURNING *
                """,
                route_id,
                user_id,
                route_data.title,
                route_data.description,
                route_data.city,
                route_data.country,
                route_data.is_public,
                route_data.is_linear,
                route_data.cover_image_url,
                now,
                now,
            )

            # Insert tags if provided
            if route_data.tags:
                tag_ids = await self._get_or_create_tags(conn, route_data.tags)
                for tag_id in tag_ids:
                    await conn.execute(
                        "INSERT INTO route_tag_relations (route_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
                        route_id,
                        tag_id,
                    )

            # Initialize route stats
            await conn.execute(
                "INSERT INTO route_stats (route_id) VALUES ($1) ON CONFLICT DO NOTHING",
                route_id,
            )

            route_dict = dict(route)

            # Get tags
            tags = await conn.fetch(
                """
                SELECT t.name
                FROM route_tags t
                JOIN route_tag_relations rtr ON t.id = rtr.tag_id
                WHERE rtr.route_id = $1
                """,
                route_id,
            )
            route_dict["tags"] = [tag["name"] for tag in tags]

            # Create points if provided
            points = []
            if route_data.points:
                for idx, point_data in enumerate(route_data.points):
                    # Set order_index if not provided
                    if point_data.order_index is None:
                        point_data.order_index = idx
                    # Create point using point service
                    point = await self.point_service.add_point(route_id, user_id, point_data)
                    # Convert Decimal fields to float for JSON serialization
                    point_dict = {
                        **dict(point),
                        "latitude": float(point["latitude"]),
                        "longitude": float(point["longitude"]),
                    }
                    points.append(point_dict)
            
            route_dict["points"] = points

            return route_dict

    async def get_route(self, route_id: UUID, include_points: bool = False) -> Optional[dict]:
        """Get route by ID."""
        async with get_connection() as conn:
            route = await conn.fetchrow(
                "SELECT * FROM routes WHERE id = $1 AND deleted_at IS NULL",
                route_id,
            )

            if not route:
                return None

            route_dict = dict(route)

            # Get tags
            tags = await conn.fetch(
                """
                SELECT t.name
                FROM route_tags t
                JOIN route_tag_relations rtr ON t.id = rtr.tag_id
                WHERE rtr.route_id = $1
                """,
                route_id,
            )
            route_dict["tags"] = [tag["name"] for tag in tags]

            # Get points if requested
            if include_points:
                points = await conn.fetch(
                    "SELECT * FROM route_points WHERE route_id = $1 ORDER BY order_index",
                    route_id,
                )
                # Convert Decimal fields to float for JSON serialization
                route_dict["points"] = [
                    {
                        **dict(point),
                        "latitude": float(point["latitude"]),
                        "longitude": float(point["longitude"]),
                    }
                    for point in points
                ]
            else:
                route_dict["points"] = []

            return route_dict

    async def update_route(self, route_id: UUID, user_id: UUID, route_data: RouteUpdate) -> Optional[dict]:
        """Update route."""
        async with get_connection() as conn:
            # Check ownership
            route = await conn.fetchrow(
                "SELECT user_id FROM routes WHERE id = $1 AND deleted_at IS NULL",
                route_id,
            )
            if not route:
                return None
            if route["user_id"] != user_id:
                raise PermissionError("You don't have permission to update this route")

            # Build update query
            updates = []
            values = []
            param_num = 1

            if route_data.title is not None:
                updates.append(f"title = ${param_num}")
                values.append(route_data.title)
                param_num += 1

            if route_data.description is not None:
                updates.append(f"description = ${param_num}")
                values.append(route_data.description)
                param_num += 1

            if route_data.city is not None:
                updates.append(f"city = ${param_num}")
                values.append(route_data.city)
                param_num += 1

            if route_data.country is not None:
                updates.append(f"country = ${param_num}")
                values.append(route_data.country)
                param_num += 1

            if route_data.is_public is not None:
                updates.append(f"is_public = ${param_num}")
                values.append(route_data.is_public)
                param_num += 1

            if route_data.is_linear is not None:
                updates.append(f"is_linear = ${param_num}")
                values.append(route_data.is_linear)
                param_num += 1

            if route_data.cover_image_url is not None:
                updates.append(f"cover_image_url = ${param_num}")
                values.append(route_data.cover_image_url)
                param_num += 1

            if not updates:
                return await self.get_route(route_id)

            values.append(route_id)
            query = f"UPDATE routes SET {', '.join(updates)}, updated_at = NOW() WHERE id = ${param_num} RETURNING *"
            updated_route = await conn.fetchrow(query, *values)

            # Update tags if provided
            if route_data.tags is not None:
                # Delete existing tags
                await conn.execute("DELETE FROM route_tag_relations WHERE route_id = $1", route_id)
                # Insert new tags
                tag_ids = await self._get_or_create_tags(conn, route_data.tags)
                for tag_id in tag_ids:
                    await conn.execute(
                        "INSERT INTO route_tag_relations (route_id, tag_id) VALUES ($1, $2)",
                        route_id,
                        tag_id,
                    )

            return dict(updated_route)

    async def delete_route(self, route_id: UUID, user_id: UUID) -> bool:
        """Soft delete route."""
        async with get_connection() as conn:
            # Check ownership
            route = await conn.fetchrow(
                "SELECT user_id FROM routes WHERE id = $1 AND deleted_at IS NULL",
                route_id,
            )
            if not route:
                return False
            if route["user_id"] != user_id:
                raise PermissionError("You don't have permission to delete this route")

            await conn.execute(
                "UPDATE routes SET deleted_at = NOW() WHERE id = $1",
                route_id,
            )
            return True

    async def list_routes(
        self,
        user_id: Optional[UUID] = None,
        is_public: Optional[bool] = None,
        city: Optional[str] = None,
        limit: int = 20,
        offset: int = 0,
    ) -> tuple[list[dict], int]:
        """List routes with filters."""
        async with get_connection() as conn:
            conditions = ["deleted_at IS NULL"]
            params = []
            param_num = 1

            if user_id:
                conditions.append(f"user_id = ${param_num}")
                params.append(user_id)
                param_num += 1

            if is_public is not None:
                conditions.append(f"is_public = ${param_num}")
                params.append(is_public)
                param_num += 1

            if city:
                conditions.append(f"city = ${param_num}")
                params.append(city)
                param_num += 1

            where_clause = " AND ".join(conditions)

            # Get total count
            count_query = f"SELECT COUNT(*) FROM routes WHERE {where_clause}"
            total = await conn.fetchval(count_query, *params)

            # Get routes
            query = f"""
                SELECT * FROM routes
                WHERE {where_clause}
                ORDER BY created_at DESC
                LIMIT ${param_num} OFFSET ${param_num + 1}
            """
            params.extend([limit, offset])
            routes = await conn.fetch(query, *params)

            # Get tags for each route
            route_dicts = []
            for route in routes:
                route_dict = dict(route)
                tags = await conn.fetch(
                    """
                    SELECT t.name
                    FROM route_tags t
                    JOIN route_tag_relations rtr ON t.id = rtr.tag_id
                    WHERE rtr.route_id = $1
                    """,
                    route["id"],
                )
                route_dict["tags"] = [tag["name"] for tag in tags]
                route_dicts.append(route_dict)

            return route_dicts, total

    async def find_nearby_routes(
        self,
        latitude: float,
        longitude: float,
        radius_km: float = 5.0,
        limit: int = 20,
        offset: int = 0,
    ) -> tuple[list[dict], int]:
        """Find routes near a location."""
        async with get_connection() as conn:
            # Use geohash for initial filtering
            geohash = encode_geohash(latitude, longitude, precision=7)  # ~150m precision

            # Get routes with points in nearby geohashes
            query = """
                SELECT DISTINCT r.*
                FROM routes r
                JOIN route_points rp ON r.id = rp.route_id
                WHERE r.deleted_at IS NULL
                  AND r.is_public = true
                  AND rp.geohash LIKE $1 || '%'
                ORDER BY r.created_at DESC
                LIMIT $2 OFFSET $3
            """
            routes = await conn.fetch(query, geohash[:6], limit, offset)

            # Filter by exact distance (Haversine formula)
            nearby_routes = []
            for route in routes:
                # Get closest point
                closest_point = await conn.fetchrow(
                    """
                    SELECT latitude, longitude,
                           (6371 * acos(
                               cos(radians($1)) * cos(radians(latitude)) *
                               cos(radians(longitude) - radians($2)) +
                               sin(radians($1)) * sin(radians(latitude))
                           )) AS distance_km
                    FROM route_points
                    WHERE route_id = $3
                    ORDER BY distance_km
                    LIMIT 1
                    """,
                    latitude,
                    longitude,
                    route["id"],
                )

                if closest_point and closest_point["distance_km"] <= radius_km:
                    route_dict = dict(route)
                    route_dict["distance_km"] = float(closest_point["distance_km"])
                    nearby_routes.append(route_dict)

            # Sort by distance
            nearby_routes.sort(key=lambda x: x["distance_km"])

            return nearby_routes[:limit], len(nearby_routes)

    async def _get_or_create_tags(self, conn: asyncpg.Connection, tag_names: list[str]) -> list[UUID]:
        """Get or create tags and return their IDs."""
        tag_ids = []
        for tag_name in tag_names:
            tag = await conn.fetchrow(
                "SELECT id FROM route_tags WHERE name = $1",
                tag_name.lower(),
            )
            if tag:
                tag_ids.append(tag["id"])
            else:
                tag_id = uuid4()
                await conn.execute(
                    "INSERT INTO route_tags (id, name) VALUES ($1, $2)",
                    tag_id,
                    tag_name.lower(),
                )
                tag_ids.append(tag_id)
        return tag_ids

