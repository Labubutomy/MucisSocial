from __future__ import annotations

from dataclasses import dataclass

from core.config import Settings
from schemas.stream import StreamResponse, VariantStream
from services.signing import URLSigner


@dataclass(slots=True)
class StreamUrls:
    master_url: str
    variants: list[VariantStream]
    expires_in: int


class StreamingService:
    """Business logic for generating signed playlists."""

    def __init__(self, settings: Settings, signer: URLSigner) -> None:
        self._settings = settings
        self._signer = signer

    def get_stream(
        self, track_id: str, artist_id: str, available_bitrates: tuple[int, ...] | None = None
    ) -> StreamResponse:
        bitrates = available_bitrates or self._settings.available_bitrates
        urls = self._generate_urls(track_id, artist_id, bitrates)
        return StreamResponse(master_url=urls.master_url, variants=urls.variants, expires_in=urls.expires_in)

    def refresh_stream(
        self, track_id: str, artist_id: str, available_bitrates: tuple[int, ...] | None = None
    ) -> StreamResponse:
        bitrates = available_bitrates or self._settings.available_bitrates
        urls = self._generate_urls(track_id, artist_id, bitrates)
        return StreamResponse(master_url=urls.master_url, variants=urls.variants, expires_in=urls.expires_in)

    def _generate_urls(self, track_id: str, artist_id: str, bitrates: tuple[int, ...]) -> StreamUrls:
        # Структура MinIO: tracks/{artist_id}/{track_id}/transcoded/...
        base_path = f"/tracks/{artist_id}/{track_id}/transcoded"

        # Используем base_url или cdn_base_url если указан
        service_base_url = self._settings.cdn_base_url or self._settings.base_url

        master_path = f"{base_path}/master.m3u8"
        master_signed, master_signature = self._signer.sign(master_path, self._settings.playlist_ttl_seconds)
        # Origin endpoint находится по пути /origin/{resource_path}
        origin_master_path = f"/origin{master_path}"
        master_url = self._signer.build_url(service_base_url, master_signed, master_signature)
        # Заменяем путь на origin endpoint
        master_url = master_url.replace(master_path, origin_master_path)

        variants: list[VariantStream] = []
        for bitrate in bitrates:
            bitrate_folder = f"aac_{bitrate // 1000}"
            variant_path = f"{base_path}/{bitrate_folder}/index.m3u8"
            signed_path, signature = self._signer.sign(variant_path, self._settings.playlist_ttl_seconds)
            url = self._signer.build_url(service_base_url, signed_path, signature)
            # Заменяем путь на origin endpoint
            origin_variant_path = f"/origin{variant_path}"
            url = url.replace(variant_path, origin_variant_path)
            variants.append(VariantStream(bitrate=bitrate, url=url))

        return StreamUrls(master_url=master_url, variants=variants, expires_in=self._settings.playlist_ttl_seconds)

