from fastapi.testclient import TestClient

from app.main import app
from app.data import store

client = TestClient(app)


def test_sign_in_success():
    response = client.post("/auth/sign-in", json={"email": "hello@music.social", "password": "password"})
    assert response.status_code == 200
    payload = response.json()
    assert payload["user"]["username"] == "ava.wave"


def test_sign_in_failure():
    response = client.post("/auth/sign-in", json={"email": "hello@music.social", "password": "wrong"})
    assert response.status_code == 401


def test_sign_up_creates_user():
    response = client.post(
        "/auth/sign-up",
        json={"email": "new@music.social", "password": "secret", "username": "new.user"},
    )
    assert response.status_code == 200
    payload = response.json()
    assert payload["user"]["username"] == "new.user"


def test_sign_up_duplicate_email_fails():
    response = client.post(
        "/auth/sign-up",
        json={"email": "hello@music.social", "password": "secret", "username": "dup.user"},
    )
    assert response.status_code == 400


def test_me_returns_profile():
    response = client.get("/me")
    assert response.status_code == 200
    payload = response.json()
    assert payload["username"] == "ava.wave"


def test_search_history_append():
    # clear existing to ensure deterministic
    user = store.get_user("user-ava")
    user.search_history.clear()

    response = client.post("/me/search-history", json={"query": "synthwave"})
    assert response.status_code == 200
    entry = response.json()
    assert entry["query"] == "synthwave"

    response = client.get("/me/search-history")
    history = response.json()
    assert len(history["items"]) == 1
    assert history["items"][0]["query"] == "synthwave"

