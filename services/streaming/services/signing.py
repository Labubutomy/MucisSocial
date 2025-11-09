from __future__ import annotations

from typing import Any

import hashlib
import hmac
import time
from dataclasses import dataclass


@dataclass(slots=True)
class SignedPath:
    resource_path: str
    expires_at: int

    def as_query(self, signature: str) -> str:
        return f"exp={self.expires_at}&sig={signature}"


class URLSigner:
    """Generate and verify short-lived HMAC signatures for resource paths."""

    def __init__(self, secret: str) -> None:
        self._secret = secret.encode("utf-8")

    def sign(self, resource_path: str, expires_in: int) -> tuple[SignedPath, str]:
        expires_at = int(time.time()) + expires_in
        data = f"{resource_path}{expires_at}".encode("utf-8")
        signature = hmac.new(self._secret, data, hashlib.sha256).hexdigest()
        return SignedPath(resource_path=resource_path, expires_at=expires_at), signature

    def build_url(self, base_url: str | Any, signed_path: SignedPath, signature: str) -> str:
        # Преобразуем AnyHttpUrl в строку, если необходимо
        base_url_str = str(base_url).rstrip('/')
        url = f"{base_url_str}{signed_path.resource_path}"
        query = signed_path.as_query(signature)
        return f"{url}?{query}"

    def verify(self, resource_path: str, expires_at: int, signature: str) -> bool:
        if expires_at <= int(time.time()):
            return False
        data = f"{resource_path}{expires_at}".encode("utf-8")
        expected = hmac.new(self._secret, data, hashlib.sha256).hexdigest()
        return hmac.compare_digest(expected, signature)

