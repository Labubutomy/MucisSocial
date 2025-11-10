from __future__ import annotations

from contextlib import asynccontextmanager
from typing import AsyncIterator
from urllib.parse import urlparse, parse_qs

import pytest
from fastapi.testclient import TestClient

from core.config import Settings, get_settings
from core.dependencies import (
    get_signer,
    get_storage,
    get_streaming_service,
)
from main import create_app
from services.exceptions import ObjectNotFound
from services.signing import URLSigner
from services.streaming import StreamingService


class FakeStorage:
    def __init__(self, objects: dict[str, bytes | str]) -> None:
        self._objects = objects

    async def read_text(self, key: str, encoding: str = "utf-8") -> str:
        try:
            obj = self._objects[key]
        except KeyError as exc:
            raise ObjectNotFound(f"{key} not found") from exc
        if isinstance(obj, bytes):
            return obj.decode(encoding)
        return obj

    @asynccontextmanager
    async def stream(self, key: str, chunk_size: int = 1 << 20) -> AsyncIterator[AsyncIterator[bytes]]:
        try:
            obj = self._objects[key]
        except KeyError as exc:
            raise ObjectNotFound(f"{key} not found") from exc

        data = obj.encode("utf-8") if isinstance(obj, str) else obj

        async def iterator() -> AsyncIterator[bytes]:
            yield data

        yield iterator()


@pytest.fixture()
def settings() -> Settings:
    get_settings.cache_clear()
    return Settings(
        signing_secret="integration-secret",
        cdn_base_url="https://cdn.test",
        playlist_ttl_seconds=120,
        segment_ttl_seconds=60,
    )


@pytest.fixture()
def signer(settings: Settings) -> URLSigner:
    return URLSigner(settings.signing_secret)




@pytest.fixture()
def storage_objects() -> dict[str, bytes | str]:
    return {
        "artist-1/track-1/transcoded/master.m3u8": "\n".join(
            [
                "#EXTM3U",
                "#EXT-X-STREAM-INF:BANDWIDTH=256000",
                "aac_256/index.m3u8",
                "#EXT-X-STREAM-INF:BANDWIDTH=160000",
                "aac_160/index.m3u8",
                "#EXT-X-STREAM-INF:BANDWIDTH=96000",
                "aac_96/index.m3u8",
                "",
            ]
        ),
        "artist-1/track-1/transcoded/aac_256/index.m3u8": "\n".join(
            [
                "#EXTM3U",
                "#EXT-X-TARGETDURATION:4",
                "#EXT-X-VERSION:7",
                '#EXT-X-MAP:URI="init.mp4"',
                "#EXTINF:4.0,",
                "chunk_0001.m4s",
                "#EXTINF:4.0,",
                "chunk_0002.m4s",
                "",
            ]
        ),
        "artist-1/track-1/transcoded/aac_256/init.mp4": b"init-segment",
        "artist-1/track-1/transcoded/aac_256/chunk_0001.m4s": b"segment-1",
        "artist-1/track-1/transcoded/aac_256/chunk_0002.m4s": b"segment-2",
        "artist-1/track-1/transcoded/aac_160/index.m3u8": "#EXTM3U\n",
        "artist-1/track-1/transcoded/aac_96/index.m3u8": "#EXTM3U\n",
    }


@pytest.fixture()
def app(settings: Settings, signer: URLSigner, storage_objects: dict[str, bytes | str]):
    fake_storage = FakeStorage(storage_objects)
    streaming_service = StreamingService(settings, signer)

    application = create_app()
    application.dependency_overrides[get_settings] = lambda: settings
    application.dependency_overrides[get_signer] = lambda: signer
    application.dependency_overrides[get_storage] = lambda: fake_storage
    application.dependency_overrides[get_streaming_service] = lambda: streaming_service
    return application


@pytest.fixture()
def client(app) -> TestClient:
    return TestClient(app)


def _extract_signature_parts(url: str) -> tuple[str, int, str]:
    parsed = urlparse(url)
    query = parse_qs(parsed.query)
    exp = int(query["exp"][0])
    sig = query["sig"][0]
    return parsed.path, exp, sig


def test_stream_endpoint_returns_signed_urls(client: TestClient, settings: Settings) -> None:
    response = client.get("/api/stream/track-1", params={"artist_id": "artist-1", "available_bitrates": [256000, 160000, 96000]})
    assert response.status_code == 200

    payload = response.json()
    assert payload["expires_in"] == settings.playlist_ttl_seconds

    master_path, exp, sig = _extract_signature_parts(payload["master_url"])
    assert master_path == "/artist-1/track-1/transcoded/master.m3u8"  # Bucket уже называется "tracks"
    assert exp > 0
    assert sig

    variants = payload["variants"]
    assert len(variants) == 3
    assert {variant["bitrate"] for variant in variants} == {256000, 160000, 96000}


def test_refresh_endpoint_returns_new_signature(client: TestClient) -> None:
    first = client.post(
        "/api/stream/refresh", json={"track_id": "track-1", "artist_id": "artist-1", "available_bitrates": [256000, 160000, 96000]}
    ).json()
    second = client.post(
        "/api/stream/refresh", json={"track_id": "track-1", "artist_id": "artist-1", "available_bitrates": [256000, 160000, 96000]}
    ).json()

    assert first["master_url"] != second["master_url"]


def test_master_playlist_rewrite(client: TestClient) -> None:
    response = client.get("/api/stream/track-1", params={"artist_id": "artist-1", "available_bitrates": [256000, 160000, 96000]})
    payload = response.json()
    master_url = payload["master_url"]
    master_path, exp, sig = _extract_signature_parts(master_url)

    origin_response = client.get(f"/origin{master_path}", params={"exp": exp, "sig": sig})
    assert origin_response.status_code == 200

    body = origin_response.text.splitlines()
    variant_lines = [line for line in body if line.endswith("index.m3u8") or "index.m3u8?" in line]
    assert all("sig=" in line for line in variant_lines)


def test_variant_playlist_rewrite_and_segment_delivery(client: TestClient) -> None:
    # Request master to get updated variant path with signature
    master_payload = client.get("/api/stream/track-1", params={"artist_id": "artist-1", "available_bitrates": [256000, 160000, 96000]}).json()
    variant_url = next(item["url"] for item in master_payload["variants"] if item["bitrate"] == 256000)
    variant_path, exp, sig = _extract_signature_parts(variant_url)

    variant_response = client.get(f"/origin{variant_path}", params={"exp": exp, "sig": sig})
    assert variant_response.status_code == 200
    lines = variant_response.text.splitlines()

    map_line = next(line for line in lines if line.startswith("#EXT-X-MAP"))
    assert "sig=" in map_line and "exp=" in map_line

    segment_line = next(line for line in lines if line.startswith("chunk_0001.m4s"))
    variant_dir = variant_path.rsplit("/", 1)[0]
    segment_path, seg_exp, seg_sig = _extract_signature_parts(f"{variant_dir}/{segment_line}")

    segment_response = client.get(f"/origin{segment_path}", params={"exp": seg_exp, "sig": seg_sig})
    assert segment_response.status_code == 200
    assert segment_response.content == b"segment-1"

