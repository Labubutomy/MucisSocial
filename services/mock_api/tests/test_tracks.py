from fastapi.testclient import TestClient

from app.main import app
from app.data import store

client = TestClient(app)


def test_list_tracks_respects_limit():
    store.tracks["extra"] = store.tracks["1"]
    try:
        response = client.get("/tracks?limit=2")
        assert response.status_code == 200
        payload = response.json()
        assert payload["filter"] == "all"
        assert len(payload["items"]) == 2
        ids = {item["id"] for item in payload["items"]}
        assert ids == {"1", "extra"}
    finally:
        store.tracks.pop("extra", None)


def test_get_track():
    response = client.get("/tracks/1")
    assert response.status_code == 200
    payload = response.json()
    assert payload["id"] == "1"
    assert payload["artist"]["name"] == "Vulpes Vult"
    assert payload["stream"]["hlsMasterUrl"]


def test_recommendations_exclude_current():
    response = client.get("/tracks/1/recommendations")
    assert response.status_code == 200
    payload = response.json()
    assert payload["trackId"] == "1"
    assert payload["items"] == []


def test_track_like_toggle():
    response = client.post("/tracks/1/like", json={"isLiked": True})
    assert response.status_code == 200
    payload = response.json()
    assert payload["isLiked"] is True
    assert payload["likedAt"]

    response = client.post("/tracks/1/like", json={"isLiked": False})
    assert response.status_code == 200
    payload = response.json()
    assert payload["isLiked"] is False
    assert payload["likedAt"] is None


def test_trending_queries():
    response = client.get("/tracks/search/trending")
    assert response.status_code == 200
    payload = response.json()
    assert payload["items"]
    assert "query" in payload["items"][0]

