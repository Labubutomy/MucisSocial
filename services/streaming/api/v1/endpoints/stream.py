from fastapi import APIRouter, Depends, Query

from core.dependencies import get_streaming_service
from schemas.stream import RefreshRequest, StreamResponse
from services.streaming import StreamingService

router = APIRouter(tags=["streaming"])


@router.get("/stream/{track_id}", response_model=StreamResponse)
async def get_stream(
    track_id: str,
    artist_id: str = Query(..., description="Artist identifier"),
    available_bitrates: str | None = Query(
        default=None, description="Comma-separated list of available bitrates in bps (e.g., 256000,160000,96000)"
    ),
    service: StreamingService = Depends(get_streaming_service),
) -> StreamResponse:
    bitrates_tuple: tuple[int, ...] | None = None
    if available_bitrates:
        try:
            bitrates_list = [int(b.strip()) for b in available_bitrates.split(",") if b.strip()]
            if bitrates_list:
                bitrates_tuple = tuple(bitrates_list)
        except ValueError:
            # Если не удалось распарсить, используем None (значения по умолчанию)
            bitrates_tuple = None
    
    return service.get_stream(track_id, artist_id, bitrates_tuple)


@router.post("/stream/refresh", response_model=StreamResponse)
async def refresh_stream(
    payload: RefreshRequest, service: StreamingService = Depends(get_streaming_service)
) -> StreamResponse:
    bitrates_tuple = tuple(payload.available_bitrates) if payload.available_bitrates else None
    return service.refresh_stream(payload.track_id, payload.artist_id, bitrates_tuple)

