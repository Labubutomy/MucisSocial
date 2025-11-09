from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


def test_get_artist():
    response = client.get("/artists/1")
    assert response.status_code == 200
    payload = response.json()
    assert payload["name"] == "Vulpes Vult"
    assert payload["topTracks"]


def test_trending_artists():
    response = client.get("/artists/trending")
    assert response.status_code == 200
    payload = response.json()
    assert payload["items"]


def test_search_artists():
    response = client.get("/artists/search", params={"q": "vulpes"})
    assert response.status_code == 200
    payload = response.json()
    assert payload["items"][0]["name"] == "Vulpes Vult"

