import json
import logging
from typing import Optional

from aiokafka import AIOKafkaProducer, AIOKafkaConsumer
from aiokafka.errors import KafkaError

from core.config import get_settings

logger = logging.getLogger(__name__)
settings = get_settings()


class KafkaService:
    """Service for Kafka operations."""

    def __init__(self):
        self.producer: Optional[AIOKafkaProducer] = None
        self.consumer: Optional[AIOKafkaConsumer] = None

    async def connect_producer(self) -> None:
        """Connect to Kafka producer."""
        try:
            self.producer = AIOKafkaProducer(
                bootstrap_servers=settings.kafka_brokers,
                value_serializer=lambda v: json.dumps(v).encode("utf-8"),
            )
            await self.producer.start()
            logger.info("Kafka producer connected")
        except Exception as e:
            logger.error(f"Failed to connect Kafka producer: {e}")
            raise

    async def disconnect_producer(self) -> None:
        """Disconnect from Kafka producer."""
        if self.producer:
            await self.producer.stop()
            logger.info("Kafka producer disconnected")

    async def send_event(self, topic: str, event: dict) -> None:
        """Send event to Kafka topic."""
        if not self.producer:
            await self.connect_producer()

        try:
            await self.producer.send_and_wait(topic, event)
            logger.debug(f"Sent event to {topic}: {event}")
        except KafkaError as e:
            logger.error(f"Failed to send event to {topic}: {e}")
            raise

    async def connect_consumer(self, topics: list[str], group_id: str) -> None:
        """Connect to Kafka consumer."""
        try:
            self.consumer = AIOKafkaConsumer(
                *topics,
                bootstrap_servers=settings.kafka_brokers,
                group_id=group_id,
                value_deserializer=lambda m: json.loads(m.decode("utf-8")),
                auto_offset_reset="latest",
            )
            await self.consumer.start()
            logger.info(f"Kafka consumer connected to topics: {topics}")
        except Exception as e:
            logger.error(f"Failed to connect Kafka consumer: {e}")
            raise

    async def disconnect_consumer(self) -> None:
        """Disconnect from Kafka consumer."""
        if self.consumer:
            await self.consumer.stop()
            logger.info("Kafka consumer disconnected")

