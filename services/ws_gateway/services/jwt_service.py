from __future__ import annotations

from jose import JWTError, jwt

from core.config import Settings


class JWTService:
    """Service for JWT token validation."""

    def __init__(self, settings: Settings):
        self.settings = settings
        self.secret_key = settings.jwt_secret
        self.algorithm = settings.jwt_algorithm

    def validate_token(self, token: str) -> dict[str, str] | None:
        """Validate JWT token and return claims."""
        try:
            payload = jwt.decode(
                token, self.secret_key, algorithms=[self.algorithm]
            )
            return payload
        except JWTError:
            return None

    def extract_user_id(self, token: str) -> str | None:
        """Extract user_id from JWT token."""
        claims = self.validate_token(token)
        if claims:
            return claims.get("user_id")
        return None

