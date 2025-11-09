from pydantic import AnyHttpUrl, BaseModel, Field


class VariantStream(BaseModel):
    bitrate: int = Field(..., ge=1, description="Bitrate in bits per second")
    url: AnyHttpUrl


class StreamResponse(BaseModel):
    master_url: AnyHttpUrl
    variants: list[VariantStream]
    expires_in: int = Field(..., ge=1, description="Seconds until URLs expire")


class RefreshRequest(BaseModel):
    track_id: str = Field(..., description="Track identifier to refresh signed URLs")
    artist_id: str = Field(..., description="Artist identifier")
    available_bitrates: list[int] | None = Field(
        default=None, description="Available bitrates in bps. If not provided, uses default from settings"
    )

