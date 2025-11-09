from fastapi.testclient import TestClient

from app.main import app
from app.data import store

client = TestClient(app)


def test_get_playlist():
    response = client.get("/playlists/uplink")
    assert response.status_code == 200
    payload = response.json()
    assert payload["id"] == "uplink"
    assert payload["itemsCount"] == 1


def test_create_playlist_and_add_tracks():
    response = client.post(
        "/playlists",
        json={
            "title": "Ночной драйв",
            "description": "Синтвейв для поздних поездок",
            "isPrivate": False,
            "genres": ["Синтвейв", "Электроника"],
        },
    )
    assert response.status_code == 200
    playlist = response.json()
    playlist_id = playlist["id"]

    response = client.post(f"/playlists/{playlist_id}/tracks", json={"trackIds": ["1"]})
    assert response.status_code == 200
    payload = response.json()
    assert payload["playlistId"] == playlist_id
    assert payload["added"][0]["trackId"] == "1"


def test_list_playlists():
    response = client.get("/playlists")
    assert response.status_code == 200
    payload = response.json()
    assert payload["items"]

