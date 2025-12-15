"""Kafka event producer for listening events."""

from __future__ import annotations

import asyncio
import json
import logging
from datetime import datetime, timezone
from typing import Any

from aiokafka import AIOKafkaProducer

logger = logging.getLogger(__name__)


class ListeningEventProducer:
    """
    Asynchronous Kafka producer for track_listened events.
    Sends events when media segments are served through CDN.
    """

    def __init__(
        self,
        bootstrap_servers: str,
        topic: str,
        enabled: bool = True,
    ):
        self._bootstrap_servers = bootstrap_servers
        self._topic = topic
        self._enabled = enabled
        self._producer: AIOKafkaProducer | None = None
        self._started = False
        # Deduplication: track recently sent events to avoid spam
        self._recent_events: dict[str, float] = {}
        self._dedup_window_seconds = (
            30.0  # Don't send same user+track more than once per 30s
        )

    async def start(self) -> None:
        """Start the Kafka producer."""
        if not self._enabled:
            logger.info("ListeningEventProducer is disabled")
            return

        try:
            self._producer = AIOKafkaProducer(
                bootstrap_servers=self._bootstrap_servers,
                value_serializer=lambda v: json.dumps(v).encode("utf-8"),
                acks="all",
                retry_backoff_ms=100,
                request_timeout_ms=30000,
            )
            await self._producer.start()
            self._started = True
            logger.info(
                f"ListeningEventProducer started (brokers: {self._bootstrap_servers}, topic: {self._topic})"
            )
        except Exception as e:
            logger.error(f"Failed to start Kafka producer: {e}")
            self._producer = None
            self._started = False

    async def stop(self) -> None:
        """Stop the Kafka producer."""
        if self._producer:
            try:
                await self._producer.stop()
                logger.info("ListeningEventProducer stopped")
            except Exception as e:
                logger.error(f"Error stopping Kafka producer: {e}")
            finally:
                self._producer = None
                self._started = False

    def _should_send_event(self, user_id: str, track_id: str) -> bool:
        """Check if event should be sent (deduplication)."""
        key = f"{user_id}:{track_id}"
        now = datetime.now(timezone.utc).timestamp()

        # Clean old entries
        cutoff = now - self._dedup_window_seconds
        self._recent_events = {
            k: v for k, v in self._recent_events.items() if v > cutoff
        }

        if key in self._recent_events:
            return False

        self._recent_events[key] = now
        return True

    async def publish_track_listened(
        self,
        user_id: str,
        track_id: str,
        artist_id: str | None = None,
        duration_sec: int | None = None,
        progress_pct: int | None = None,
        quality: str | None = None,
        source: str = "cdn",
    ) -> bool:
        """
        Publish a track_listened event to Kafka.

        Args:
            user_id: User who listened to the track
            track_id: Track being listened to
            artist_id: Optional artist ID for the track
            duration_sec: Optional duration of listening in seconds
            progress_pct: Optional playback progress percentage (0-100)
            quality: Optional quality level (low, medium, high)
            source: Source of the event (default: cdn)

        Returns:
            True if event was sent successfully, False otherwise
        """
        if not self._enabled or not self._started:
            return False

        # Deduplication check
        if not self._should_send_event(user_id, track_id):
            logger.debug(
                f"Skipping duplicate listening event for user={user_id}, track={track_id}"
            )
            return False

        # Convert to Unix timestamp (int64) for compatibility with recommendations service
        now = datetime.now(timezone.utc)
        ts = int(now.timestamp())
        
        event = {
            "event_type": "track_listened",
            "user_id": user_id,
            "track_id": track_id,
            "ts": ts,  # Unix timestamp as int64
            "listened_seconds": duration_sec if duration_sec is not None else 0,
        }

        # Keep additional fields for other consumers (optional)
        if artist_id:
            event["artist_id"] = artist_id
        if progress_pct is not None:
            event["progress_pct"] = progress_pct
        if quality:
            event["quality"] = quality
        event["source"] = source

        try:
            if self._producer:
                await self._producer.send_and_wait(
                    self._topic,
                    value=event,
                    key=track_id.encode("utf-8"),
                )
                logger.debug(
                    f"Published track_listened event: user={user_id}, track={track_id}"
                )
                return True
        except Exception as e:
            logger.error(f"Failed to publish track_listened event: {e}")
            return False

        return False

    @property
    def is_ready(self) -> bool:
        """Check if producer is ready to send events."""
        return self._enabled and self._started and self._producer is not None
