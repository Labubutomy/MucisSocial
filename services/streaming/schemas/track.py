from pydantic import BaseModel, Field


class TrackMetadata(BaseModel):
    track_id: str = Field(..., description="Track identifier")
    artist_id: str = Field(..., description="Artist identifier")
    available_bitrates: tuple[int, ...] | None = Field(
        default=None, description="Optional list of available bitrates in bits per second"
    )
    duration_seconds: float | None = Field(default=None, ge=0)
    sample_rate: int | None = Field(default=None, ge=1)
    channels: int | None = Field(default=None, ge=1)

