from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


def test_master_playlist():
    response = client.get("/hls/1/master.m3u8")
    assert response.status_code == 200
    body = response.text
    assert "#EXTM3U" in body
    assert "aac_96" in body


def test_variant_playlist_contains_segments():
    response = client.get("/hls/1/aac_96/index.m3u8")
    assert response.status_code == 200
    body = response.text.splitlines()
    assert body[0] == "#EXTM3U"
    segments = [line for line in body if line.endswith(".aac")]
    assert len(segments) == 3


def test_segment_served():
    response = client.get("/hls/1/aac_96/segment-001.aac")
    assert response.status_code == 200
    assert response.content == b"AACDATA"

