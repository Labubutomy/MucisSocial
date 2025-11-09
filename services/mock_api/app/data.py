from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Dict, List, Optional


def utc_now_iso() -> str:
    return datetime.now(tz=timezone.utc).isoformat()


CDN_IMAGE_URL = "https://cdn.pixabay.com/photo/2016/11/29/05/08/adult-1868667_640.jpg"


@dataclass
class Artist:
    id: str
    name: str
    avatar_url: str
    genres: List[str]
    followers: int
    top_tracks: List[str] = field(default_factory=list)


@dataclass
class Track:
    id: str
    title: str
    duration_sec: int
    cover_url: str
    artist_id: str
    album_id: str
    album_title: str
    album_cover_url: str
    album_release_date: str
    credits: List[str]
    bpm: int
    is_liked: bool = False
    liked_at: Optional[str] = None
    stream_quality: List[str] = field(default_factory=lambda: ["aac_128", "aac_256"])
    master_url: str = ""


@dataclass
class User:
    id: str
    username: str
    email: str
    avatar_url: str
    password: str
    top_genres: List[str]
    top_artists: List[str]
    playlists: List[str] = field(default_factory=list)
    search_history: List[Dict[str, str]] = field(default_factory=list)


@dataclass
class Playlist:
    id: str
    title: str
    description: str
    cover_url: str
    owner_id: str
    is_private: bool
    genres: List[str]
    tracks: List[str] = field(default_factory=list)


class InMemoryStore:
    def __init__(self) -> None:
        self.artists: Dict[str, Artist] = {}
        self.tracks: Dict[str, Track] = {}
        self.users: Dict[str, User] = {}
        self.playlists: Dict[str, Playlist] = {}
        self.trending_queries: List[str] = ["синтвейв ночь", "неоновый фанк", "dream pop"]
        self.search_suggestions: List[Dict[str, str]] = [
            {"id": "history-1", "query": "синтвейв полночь", "createdAt": "2025-11-07T12:00:00.000Z"},
            {"id": "history-2", "query": "лоуфай для концентрации", "createdAt": "2025-11-05T09:30:00.000Z"},
        ]
        self._seed()

    def _seed(self) -> None:
        artist = Artist(
            id="1",
            name="Vulpes Vult",
            avatar_url=CDN_IMAGE_URL,
            genres=["Синтвейв", "Электроника"],
            followers=1204300,
            top_tracks=["1"],
        )
        self.artists[artist.id] = artist

        track = Track(
            id="1",
            title="Терновый куст",
            duration_sec=214,
            cover_url=CDN_IMAGE_URL,
            artist_id=artist.id,
            album_id="album-1",
            album_title="Starlight Bloom",
            album_cover_url=CDN_IMAGE_URL,
            album_release_date="2024-02-10",
            credits=["Vulpes Vult"],
            bpm=102,
            stream_quality=["aac_96", "aac_160", "aac_256"],
            master_url="http://localhost:8000/origin/tracks/1/1/transcoded/master.m3u8",
        )
        self.tracks[track.id] = track

        user = User(
            id="user-ava",
            username="ava.wave",
            email="hello@music.social",
            avatar_url=CDN_IMAGE_URL,
            password="password",
            top_genres=["Синтвейв", "Дрим-поп", "Инди-электроника"],
            top_artists=["Vulpes Vult", "Kyro", "Luna Wave", "Solaria"],
        )
        self.users[user.id] = user

        playlist = Playlist(
            id="uplink",
            title="Uplink Sessions",
            description="Кураторское путешествие по неону даунтемпо и атмосферной электронике.",
            cover_url=CDN_IMAGE_URL,
            owner_id=user.id,
            is_private=False,
            genres=["Синтвейв", "Электроника"],
            tracks=[track.id],
        )
        self.playlists[playlist.id] = playlist
        user.playlists.append(playlist.id)

    # Tracks
    def list_tracks(self, filter_type: Optional[str], limit: int) -> List[Track]:
        return list(self.tracks.values())[:limit]

    def get_track(self, track_id: str) -> Optional[Track]:
        return self.tracks.get(track_id)

    def set_track_like(self, track_id: str, is_liked: bool) -> Optional[Track]:
        track = self.tracks.get(track_id)
        if not track:
            return None
        track.is_liked = is_liked
        track.liked_at = utc_now_iso() if is_liked else None
        return track

    def add_recommendations(self, track_id: str, limit: int) -> List[Track]:
        return [t for tid, t in self.tracks.items() if tid != track_id][:limit]

    def search_tracks(self, query: str, limit: int) -> List[Track]:
        query_lower = query.lower()
        matches = [
            track for track in self.tracks.values() if query_lower in track.title.lower()
            or query_lower in track.id.lower()
        ]
        return matches[:limit]

    # Users
    def authenticate(self, email: str, password: str) -> Optional[User]:
        for user in self.users.values():
            if user.email == email and user.password == password:
                return user
        return None

    def create_user(self, email: str, password: str, username: str) -> User:
        if any(u.email == email for u in self.users.values()):
            raise ValueError("user already exists")
        user_id = f"user-{len(self.users) + 1}"
        user = User(
            id=user_id,
            username=username,
            email=email,
            avatar_url=CDN_IMAGE_URL,
            password=password,
            top_genres=["Синтвейв"],
            top_artists=["Vulpes Vult"],
        )
        self.users[user_id] = user
        return user

    def get_user(self, user_id: str) -> Optional[User]:
        return self.users.get(user_id)

    def add_search_history(self, user_id: str, query: str) -> Dict[str, str]:
        user = self.get_user(user_id)
        if not user:
            raise ValueError("user not found")
        entry = {"id": f"history-{len(user.search_history) + 1}", "query": query, "createdAt": utc_now_iso()}
        user.search_history.insert(0, entry)
        return entry

    # Playlists
    def create_playlist(self, owner_id: str, payload: Dict[str, Optional[str]]) -> Playlist:
        playlist_id = payload.get("id") or f"playlist-{len(self.playlists) + 1}"
        playlist = Playlist(
            id=playlist_id,
            title=payload["title"],
            description=payload.get("description") or "",
            cover_url=CDN_IMAGE_URL,
            owner_id=owner_id,
            is_private=payload.get("isPrivate", False),
            genres=payload.get("genres") or [],
        )
        self.playlists[playlist.id] = playlist
        owner = self.get_user(owner_id)
        if owner:
            owner.playlists.append(playlist.id)
        return playlist

    def add_tracks_to_playlist(self, playlist_id: str, track_ids: List[str]) -> Dict[str, List[Dict[str, str]]]:
        playlist = self.playlists.get(playlist_id)
        if not playlist:
            raise ValueError("playlist not found")
        start_position = len(playlist.tracks) + 1
        added = []
        for index, track_id in enumerate(track_ids):
            if track_id not in self.tracks:
                continue
            playlist.tracks.append(track_id)
            added.append({"trackId": track_id, "position": start_position + index})
        return {"playlistId": playlist.id, "added": added}

    # Artists
    def search_artists(self, query: str) -> List[Artist]:
        query_lower = query.lower()
        return [
            artist for artist in self.artists.values()
            if query_lower in artist.name.lower() or any(query_lower in genre.lower() for genre in artist.genres)
        ]

    def get_artist(self, artist_id: str) -> Optional[Artist]:
        return self.artists.get(artist_id)

    # Playlists
    def search_playlists(self, query: str, limit: int) -> List[Playlist]:
        query_lower = query.lower()
        matches = [
            playlist for playlist in self.playlists.values()
            if query_lower in playlist.title.lower()
            or (playlist.description and query_lower in playlist.description.lower())
        ]
        return matches[:limit]

    def get_playlist_tracks(self, playlist_id: str) -> List[Track]:
        playlist = self.playlists.get(playlist_id)
        if not playlist:
            return []
        return [self.tracks[track_id] for track_id in playlist.tracks if track_id in self.tracks]


store = InMemoryStore()

