from __future__ import annotations

import uuid
from datetime import datetime

from sqlalchemy import DateTime, ForeignKey, String, Text, UniqueConstraint
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from .conversation import Base, Conversation


class Message(Base):
    __tablename__ = "messages"

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    conversation_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("conversations.id", ondelete="CASCADE")
    )
    sender_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True))
    content: Mapped[str] = mapped_column(Text)

    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), default=datetime.utcnow
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), default=datetime.utcnow, onupdate=datetime.utcnow
    )
    deleted_at: Mapped[datetime | None] = mapped_column(
        DateTime(timezone=True), nullable=True
    )

    conversation: Mapped[Conversation] = relationship(back_populates="messages")
    tracks: Mapped[list["MessageTrack"]] = relationship(
        back_populates="message", cascade="all, delete-orphan"
    )
    contexts: Mapped[list["MessageContext"]] = relationship(
        back_populates="message", cascade="all, delete-orphan"
    )
    read_statuses: Mapped[list["MessageReadStatus"]] = relationship(
        back_populates="message", cascade="all, delete-orphan"
    )


class MessageTrack(Base):
    __tablename__ = "message_tracks"

    message_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("messages.id", ondelete="CASCADE"),
        primary_key=True,
    )
    track_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True)

    message: Mapped[Message] = relationship(back_populates="tracks")


class MessageContext(Base):
    __tablename__ = "message_contexts"
    __table_args__ = (
        UniqueConstraint(
            "message_id",
            "context_type",
            "context_id",
            name="uq_message_context",
        ),
    )

    message_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("messages.id", ondelete="CASCADE"),
        primary_key=True,
    )
    context_type: Mapped[str] = mapped_column(String(50), primary_key=True)
    context_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True)

    message: Mapped[Message] = relationship(back_populates="contexts")


class MessageReadStatus(Base):
    __tablename__ = "message_read_status"

    message_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("messages.id", ondelete="CASCADE"),
        primary_key=True,
    )
    user_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True)
    read_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), default=datetime.utcnow
    )

    message: Mapped[Message] = relationship(back_populates="read_statuses")


