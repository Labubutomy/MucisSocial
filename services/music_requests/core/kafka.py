from __future__ import annotations

import json
import logging
from typing import Any

from kafka import KafkaProducer
from kafka.errors import KafkaError

from core.config import get_settings

logger = logging.getLogger(__name__)
settings = get_settings()

_producer: KafkaProducer | None = None


def get_kafka_producer() -> KafkaProducer:
    """Get or create Kafka producer instance."""
    global _producer
    if _producer is None:
        _producer = KafkaProducer(
            bootstrap_servers=settings.kafka_brokers.split(","),
            value_serializer=lambda v: json.dumps(v).encode("utf-8"),
            key_serializer=lambda k: k.encode("utf-8") if k else None,
        )
        logger.info(f"Kafka producer connected to {settings.kafka_brokers}")
    return _producer


def send_music_request_event(event_type: str, data: dict[str, Any]) -> None:
    """Send music request event to Kafka."""
    producer = get_kafka_producer()

    event = {
        "event_type": event_type,
        **data,
    }

    try:
        future = producer.send(
            settings.kafka_music_requests_topic,
            key=str(data.get("artist_id", "")),
            value=event,
        )
        record_metadata = future.get(timeout=10)
        logger.info(
            f"Music request event '{event_type}' sent to topic {record_metadata.topic} "
            f"partition {record_metadata.partition}"
        )
    except KafkaError as e:
        logger.error(f"Failed to send music request event: {e}")
        # Don't fail the main request, just log
