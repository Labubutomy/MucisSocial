from __future__ import annotations

from typing import Any, Dict, List, Optional

from fastapi import APIRouter, Body, FastAPI, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field, EmailStr

from .data import store

app = FastAPI(
    title="Mock API Service",
    version="0.1.0",
    description="Mock implementation of the API surface required by the web client.",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Create router with /api/v1 prefix
router = APIRouter(prefix="/api/v1")


def track_to_response(track, include_album: bool = True) -> Dict[str, Any]:
    artist = store.get_artist(track.artist_id)
    artist_payload = {"id": artist.id, "name": artist.name} if artist else {"id": track.artist_id, "name": "Unknown"}
    album_payload = None
    if include_album:
        album_payload = {
            "id": track.album_id,
            "title": track.album_title,
            "coverUrl": track.album_cover_url,
            "releasedAt": track.album_release_date,
        }
    return {
        "id": track.id,
        "title": track.title,
        "durationSec": track.duration_sec,
        "coverUrl": track.cover_url,
        "artist": artist_payload,
        "album": album_payload,
        "credits": track.credits,
        "bpm": track.bpm,
        "isLiked": track.is_liked,
        "stream": {
            "quality": track.stream_quality,
            "hlsMasterUrl": track.master_url,
        },
    }


def playlist_to_response(playlist) -> Dict[str, Any]:
    owner = store.get_user(playlist.owner_id)
    return {
        "id": playlist.id,
        "title": playlist.title,
        "description": playlist.description,
        "coverUrl": playlist.cover_url,
        "itemsCount": len(playlist.tracks),
        "owner": {"id": owner.id, "name": owner.username} if owner else None,
        "type": "playlist",
        "isLiked": False,
        "totalDurationSec": sum(store.get_track(tid).duration_sec for tid in playlist.tracks if store.get_track(tid)),
    }


# -------------------------- Tracks service -------------------------- #


@router.get("/tracks")
def list_tracks(filter: Optional[str] = Query(default=None), limit: int = Query(default=24, ge=1, le=100)) -> Dict[str, Any]:
    items = [track_to_response(track) for track in store.list_tracks(filter, limit)]
    return {"filter": filter or "all", "items": items}


@router.get("/tracks/{track_id}")
def get_track(track_id: str) -> Dict[str, Any]:
    track = store.get_track(track_id)
    if not track:
        raise HTTPException(status_code=404, detail="Track not found")
    payload = track_to_response(track)
    payload["album"] = {
        "id": track.album_id,
        "title": track.album_title,
        "coverUrl": track.album_cover_url,
        "releasedAt": track.album_release_date,
    }
    return payload


@router.get("/tracks/{track_id}/recommendations")
def get_recommendations(track_id: str, limit: int = Query(default=12, ge=1, le=50)) -> Dict[str, Any]:
    items = [track_to_response(track, include_album=False) for track in store.add_recommendations(track_id, limit)]
    return {"trackId": track_id, "items": items}


@router.get("/tracks/search")
def search_tracks(q: str = Query(...), limit: int = Query(default=20, ge=1, le=100)) -> Dict[str, Any]:
    results = []
    for track in store.search_tracks(q, limit):
        results.append(
            {
                "type": "track",
                "data": {
                    "id": track.id,
                    "title": track.title,
                    "durationSec": track.duration_sec,
                    "coverUrl": track.cover_url,
                    "artist": {
                        "id": track.artist_id,
                        "name": store.get_artist(track.artist_id).name if store.get_artist(track.artist_id) else "Unknown",
                    },
                    "isLiked": track.is_liked,
                },
            }
        )
    return {"query": q, "items": results}


class LikePayload(BaseModel):
    isLiked: Optional[bool] = Field(default=True)


@router.post("/tracks/{track_id}/like")
def like_track(track_id: str, payload: LikePayload = Body(default_factory=LikePayload)) -> Dict[str, Any]:
    track = store.set_track_like(track_id, payload.isLiked if payload.isLiked is not None else True)
    if not track:
        raise HTTPException(status_code=404, detail="Track not found")
    return {"trackId": track_id, "isLiked": track.is_liked, "likedAt": track.liked_at}


@router.get("/tracks/search/trending")
def trending_search_queries() -> Dict[str, Any]:
    return {"items": [{"query": q} for q in store.trending_queries]}


# -------------------------- CDN service -------------------------- #
# CDN endpoints don't use /api/v1 prefix


@app.get("/hls/{track_id}/master.m3u8")
def get_master_playlist(track_id: str) -> str:
    track = store.get_track(track_id)
    if not track:
        raise HTTPException(status_code=404, detail="Track not found")

    lines = ["#EXTM3U", "#EXT-X-VERSION:3"]
    for quality in track.stream_quality:
        bitrate = int(quality.split("_")[-1]) * 1000

        lines.append(f"#EXT-X-STREAM-INF:BANDWIDTH={bitrate},CODECS=\"mp4a.40.2\"")
        lines.append(f"/hls/{track_id}/{quality}/index.m3u8")
    return "\n".join(lines)


@app.get("/hls/{track_id}/{quality}/index.m3u8")
def get_variant_playlist(track_id: str, quality: str) -> str:
    if quality not in {"aac_96", "aac_128", "aac_160", "aac_256", "flac"}:
        raise HTTPException(status_code=404, detail="Quality not supported")

    segments = [f"/hls/{track_id}/{quality}/segment-{idx:03d}.aac" for idx in range(1, 4)]
    body = ["#EXTM3U", "#EXT-X-VERSION:3", "#EXT-X-TARGETDURATION:4", "#EXT-X-MEDIA-SEQUENCE:0"]
    for segment in segments:
        body.append("#EXTINF:4.0,")
        body.append(segment)
    body.append("#EXT-X-ENDLIST")
    return "\n".join(body)


@app.get("/hls/{track_id}/{quality}/segment-{segment_id}.aac")
def get_segment(track_id: str, quality: str, segment_id: str) -> bytes:
    return b"AACDATA"


# -------------------------- Users service -------------------------- #


class SignInPayload(BaseModel):
    email: EmailStr
    password: str


class SignUpPayload(SignInPayload):
    username: str


def _auth_response(user) -> Dict[str, Any]:
    return {
        "accessToken": "mock-access-token",
        "refreshToken": "mock-refresh-token",
        "user": {
            "id": user.id,
            "username": user.username,
            "avatarUrl": user.avatar_url,
        },
    }


@router.post("/auth/sign-in")
def sign_in(payload: SignInPayload) -> Dict[str, Any]:
    user = store.authenticate(payload.email, payload.password)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid credentials")
    return _auth_response(user)


@router.post("/auth/sign-up")
def sign_up(payload: SignUpPayload) -> Dict[str, Any]:
    try:
        user = store.create_user(payload.email, payload.password, payload.username)
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    return _auth_response(user)


@router.get("/me")
def get_me() -> Dict[str, Any]:
    user = store.get_user("user-ava")
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return {
        "id": user.id,
        "username": user.username,
        "avatarUrl": user.avatar_url,
        "musicTasteSummary": {
            "topGenres": user.top_genres,
            "topArtists": user.top_artists,
        },
    }


@router.get("/me/playlists")
def get_my_playlists(limit: int = Query(default=50, ge=1, le=100)) -> Dict[str, Any]:
    user = store.get_user("user-ava")
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    items = [playlist_to_response(store.playlists[pid]) for pid in user.playlists[:limit] if pid in store.playlists]
    return {"items": items}


@router.get("/me/search-history")
def get_search_history() -> Dict[str, Any]:
    user = store.get_user("user-ava")
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return {"items": user.search_history}


class SearchHistoryPayload(BaseModel):
    query: str


@router.post("/me/search-history")
def add_search_history(payload: SearchHistoryPayload) -> Dict[str, Any]:
    entry = store.add_search_history("user-ava", payload.query)
    return entry


# -------------------------- Artists service -------------------------- #


@router.get("/artists/{artist_id}")
def get_artist(artist_id: str) -> Dict[str, Any]:
    artist = store.get_artist(artist_id)
    if not artist:
        raise HTTPException(status_code=404, detail="Artist not found")
    return {
        "id": artist.id,
        "name": artist.name,
        "avatarUrl": artist.avatar_url,
        "genres": artist.genres,
        "followers": artist.followers,
        "topTracks": [
            {"id": track_id, "title": store.get_track(track_id).title, "coverUrl": store.get_track(track_id).cover_url}
            for track_id in artist.top_tracks
            if store.get_track(track_id)
        ],
    }


@router.get("/artists/trending")
def trending_artists() -> Dict[str, Any]:
    items = [
        {
            "id": artist.id,
            "name": artist.name,
            "avatarUrl": artist.avatar_url,
            "genres": artist.genres,
        }
        for artist in store.artists.values()
    ]
    return {"items": items}


@router.get("/artists/search")
def search_artists(q: str = Query(...)) -> Dict[str, Any]:
    matches = [
        {
            "id": artist.id,
            "name": artist.name,
            "avatarUrl": artist.avatar_url,
            "genres": artist.genres,
        }
        for artist in store.search_artists(q)
    ]
    return {"items": matches}


# -------------------------- Playlists service -------------------------- #


class CreatePlaylistPayload(BaseModel):
    title: str
    description: Optional[str] = None
    isPrivate: bool = False
    coverImageId: Optional[str] = None
    genres: Optional[List[str]] = None


class AddTracksPayload(BaseModel):
    trackIds: List[str]


@router.get("/playlists/{playlist_id}")
def get_playlist(playlist_id: str) -> Dict[str, Any]:
    playlist = store.playlists.get(playlist_id)
    if not playlist:
        raise HTTPException(status_code=404, detail="Playlist not found")
    return playlist_to_response(playlist)


@router.get("/playlists/{playlist_id}/tracks")
def get_playlist_tracks(playlist_id: str) -> Dict[str, Any]:
    tracks = [track_to_response(track, include_album=False) for track in store.get_playlist_tracks(playlist_id)]
    return {"items": tracks}


@router.post("/playlists")
def create_playlist(payload: CreatePlaylistPayload) -> Dict[str, Any]:
    playlist = store.create_playlist("user-ava", payload.model_dump())
    response = playlist_to_response(playlist)
    return response


@router.post("/playlists/{playlist_id}/tracks")
def add_tracks_to_playlist(playlist_id: str, payload: AddTracksPayload) -> Dict[str, Any]:
    result = store.add_tracks_to_playlist(playlist_id, payload.trackIds)
    return result


@router.get("/playlists")
def list_playlists(filter: Optional[str] = Query(default=None), limit: int = Query(default=24, ge=1, le=100)) -> Dict[str, Any]:
    items = []
    for playlist in store.playlists.values():
        items.append(
            {
                "id": playlist.id,
                "title": playlist.title,
                "coverUrl": playlist.cover_url,
                "itemsCount": len(playlist.tracks),
                "description": playlist.description,
            }
        )
    return {"filter": filter or "all", "items": items[:limit]}


@router.get("/playlists/search")
def search_playlists(q: str = Query(...), limit: int = Query(default=20, ge=1, le=100)) -> Dict[str, Any]:
    items = []
    for playlist in store.search_playlists(q, limit):
        items.append(
            {
                "id": playlist.id,
                "title": playlist.title,
                "coverUrl": playlist.cover_url,
                "itemsCount": len(playlist.tracks),
                "description": playlist.description,
            }
        )
    return {"query": q, "items": items}


@router.get("/users/{user_id}/playlists")
def list_user_playlists(user_id: str) -> Dict[str, Any]:
    user = store.get_user(user_id)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    items = [
        {
            "id": playlist_id,
            "title": store.playlists[playlist_id].title,
            "coverUrl": store.playlists[playlist_id].cover_url,
            "itemsCount": len(store.playlists[playlist_id].tracks),
            "description": store.playlists[playlist_id].description,
        }
        for playlist_id in user.playlists
        if playlist_id in store.playlists
    ]
    return {"items": items}


# -------------------------- Health -------------------------- #


@app.get("/health")
def health() -> Dict[str, Any]:
    return {"status": "healthy", "service": "Mock API Service", "version": "0.1.0"}


# Include router with /api/v1 prefix
app.include_router(router)

