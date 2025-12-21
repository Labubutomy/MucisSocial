from __future__ import annotations

import json
import uuid
from datetime import datetime
from typing import Any

from kafka import KafkaProducer, KafkaConsumer
from kafka.errors import KafkaError

from core.config import Settings
from models.messages import ClientMessage


class KafkaService:
    """Service for Kafka message bus operations."""

    def __init__(self, settings: Settings):
        self.settings = settings
        self.producer: KafkaProducer | None = None
        self.consumer: KafkaConsumer | None = None

    def connect_producer(self) -> None:
        """Connect Kafka producer."""
        print(
            f"[KafkaService] Connecting producer to brokers: {self.settings.kafka_brokers}"
        )
        self.producer = KafkaProducer(
            bootstrap_servers=self.settings.kafka_brokers.split(","),
            value_serializer=lambda v: json.dumps(v).encode("utf-8"),
            key_serializer=lambda k: k.encode("utf-8") if k else None,
        )
        print(f"[KafkaService] Producer connected successfully")

    def connect_consumer(self, group_id: str = "ws-gateway-group") -> None:
        """Connect Kafka consumer for session sync and messaging events."""
        topics = [
            self.settings.kafka_sync_topic,
            self.settings.kafka_messaging_topic,
            self.settings.kafka_music_requests_topic,
        ]
        self.consumer = KafkaConsumer(
            *topics,
            bootstrap_servers=self.settings.kafka_brokers.split(","),
            group_id=group_id,
            value_deserializer=lambda m: json.loads(m.decode("utf-8")),
            key_deserializer=lambda k: k.decode("utf-8") if k else None,
            auto_offset_reset="latest",
            enable_auto_commit=True,
        )

    def disconnect_producer(self) -> None:
        """Disconnect Kafka producer."""
        if self.producer:
            self.producer.close()
            self.producer = None

    def disconnect_consumer(self) -> None:
        """Disconnect Kafka consumer."""
        if self.consumer:
            self.consumer.close()
            self.consumer = None

    def send_client_event(self, message: ClientMessage) -> None:
        """Send client event to Kafka."""
        if not self.producer:
            raise RuntimeError("Kafka producer not connected")

        event = {
            "event_id": str(uuid.uuid4()),
            "type": message.type.value,
            "room_id": message.room_id,
            "user_id": message.user_id,
            "action": message.action.value,
            "payload": message.payload,
            "client_timestamp": message.timestamp or datetime.now().timestamp(),
            "server_timestamp": datetime.now().timestamp(),
        }

        print(
            f"[KafkaService] Sending event to topic {self.settings.kafka_events_topic}: {event}"
        )

        future = self.producer.send(
            self.settings.kafka_events_topic,
            key=message.room_id,
            value=event,
        )

        try:
            record_metadata = future.get(timeout=10)
            print(
                f"[KafkaService] Event sent successfully to topic {record_metadata.topic} "
                f"partition {record_metadata.partition} "
                f"offset {record_metadata.offset}"
            )
        except KafkaError as e:
            print(f"[KafkaService] Failed to send event: {e}")
            import traceback

            traceback.print_exc()
            raise
