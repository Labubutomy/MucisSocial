from __future__ import annotations

import json

from schemas.track import TrackMetadata
from services.exceptions import ObjectNotFound, TrackNotFound
from services.storage import MinioStorageService


# NOTE: This service is currently not used. Client provides track metadata directly via API.
# Kept for potential future use when track catalog/index is implemented.
class TrackCatalogService:
    """Resolve track metadata stored in MinIO."""

    def __init__(self, storage: MinioStorageService) -> None:
        self._storage = storage

    async def get_track(self, track_id: str) -> TrackMetadata:
        index_key = f"index/{track_id}.json"  # Bucket уже называется "tracks"
        try:
            payload = await self._storage.read_text(index_key)
        except ObjectNotFound as exc:
            raise TrackNotFound(f"Track {track_id} not found") from exc

        data = json.loads(payload)
        metadata = TrackMetadata.model_validate(data)

        if metadata.track_id != track_id:
            raise TrackNotFound(f"Track {track_id} has invalid metadata")

        return metadata

