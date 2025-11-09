from __future__ import annotations

import os
import re
from typing import Iterable

from core.config import Settings
from services.signing import URLSigner

_MAP_PATTERN = re.compile(r'URI="(?P<uri>[^"]+)"')


def _ensure_leading_slash(path: str) -> str:
    return path if path.startswith("/") else f"/{path}"


def _join_resource_path(base_directory: str, relative_path: str) -> str:
    return _ensure_leading_slash(os.path.join(base_directory, relative_path))


def rewrite_master_playlist(content: str, resource_path: str, signer: URLSigner, settings: Settings) -> str:
    """Append fresh signatures to variant playlists in the master playlist body."""

    directory = os.path.dirname(resource_path)
    lines: list[str] = []

    for line in _iterate_lines_preserve(content):
        stripped = line.strip()
        if stripped and not stripped.startswith("#"):
            variant_resource = _join_resource_path(directory, stripped)
            signed_path, signature = signer.sign(variant_resource, settings.playlist_ttl_seconds)
            lines.append(f"{stripped}?{signed_path.as_query(signature)}")
        else:
            lines.append(line)

    return "\n".join(lines)


def rewrite_variant_playlist(content: str, resource_path: str, signer: URLSigner, settings: Settings) -> str:
    """Append signatures to init segment and media segments inside variant playlist."""

    directory = os.path.dirname(resource_path)
    lines: list[str] = []

    for line in _iterate_lines_preserve(content):
        stripped = line.strip()
        if stripped.startswith("#EXT-X-MAP:"):
            lines.append(_rewrite_map_line(line, directory, signer, settings))
        elif stripped and not stripped.startswith("#"):
            segment_resource = _join_resource_path(directory, stripped)
            signed_path, signature = signer.sign(segment_resource, settings.segment_ttl_seconds)
            lines.append(f"{stripped}?{signed_path.as_query(signature)}")
        else:
            lines.append(line)

    return "\n".join(lines)


def _rewrite_map_line(line: str, directory: str, signer: URLSigner, settings: Settings) -> str:
    match = _MAP_PATTERN.search(line)
    if not match:
        return line

    uri = match.group("uri")
    segment_resource = _join_resource_path(directory, uri)
    signed_path, signature = signer.sign(segment_resource, settings.segment_ttl_seconds)
    signed_uri = f'{uri}?{signed_path.as_query(signature)}'
    return _MAP_PATTERN.sub(f'URI="{signed_uri}"', line, count=1)


def _iterate_lines_preserve(content: str) -> Iterable[str]:
    if not content:
        return []
    lines = content.splitlines()
    if content.endswith("\n"):
        lines.append("")
    return lines

