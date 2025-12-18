from __future__ import annotations

import json
from datetime import datetime
from typing import Any
from uuid import UUID

from kafka import KafkaProducer
from kafka.errors import KafkaError

from core.config import get_settings


class MessagingKafkaProducer:
    """Thin wrapper around Kafka producer for messaging events."""

    def __init__(self) -> None:
        settings = get_settings()
        self._topic = settings.kafka_events_topic
        self._producer = KafkaProducer(
            bootstrap_servers=settings.kafka_brokers.split(","),
            value_serializer=lambda v: json.dumps(v).encode("utf-8"),
            key_serializer=lambda k: k.encode("utf-8") if k else None,
        )

    def send_event(self, key: str | None, value: dict[str, Any]) -> None:
        future = self._producer.send(
            self._topic,
            key=key or "",
            value=value,
        )
        try:
            future.get(timeout=5)
        except KafkaError as exc:
            # Для первой версии просто логируем, но не роняем запрос
            print(f"[MessagingKafkaProducer] Failed to send event: {exc}")


_producer: MessagingKafkaProducer | None = None


def get_producer() -> MessagingKafkaProducer:
    global _producer
    if _producer is None:
        _producer = MessagingKafkaProducer()
    return _producer


def publish_message_sent(
    message_id: UUID,
    conversation_id: UUID,
    sender_id: UUID,
    recipient_ids: list[UUID],
    created_at: datetime,
    track_id: UUID | None,
    context_type: str | None,
    context_id: UUID | None,
) -> None:
    event: dict[str, Any] = {
        "event_type": "message_sent",
        "message_id": str(message_id),
        "conversation_id": str(conversation_id),
        "sender_id": str(sender_id),
        "recipient_ids": [str(rid) for rid in recipient_ids],
        "created_at": created_at.isoformat(),
    }
    if track_id:
        event["track_id"] = str(track_id)
    if context_type and context_id:
        event["context_type"] = context_type
        event["context_id"] = str(context_id)

    producer = get_producer()
    producer.send_event(key=str(conversation_id), value=event)


def publish_conversation_read(
    conversation_id: UUID,
    user_id: UUID,
    message_id: UUID,
    read_at: datetime | None = None,
) -> None:
    event: dict[str, Any] = {
        "event_type": "conversation_read",
        "conversation_id": str(conversation_id),
        "user_id": str(user_id),
        "message_id": str(message_id),
        "read_at": (read_at or datetime.utcnow()).isoformat(),
    }
    producer = get_producer()
    producer.send_event(key=str(conversation_id), value=event)


