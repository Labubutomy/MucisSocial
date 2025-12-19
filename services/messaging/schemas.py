from __future__ import annotations

from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, Field, model_validator


class SendMessageRequest(BaseModel):
    content: str = Field(default="", max_length=4000, description="Текст сообщения (может быть пустым, если есть track_id)")
    recipient_id: UUID | None = Field(
        default=None, description="ID второго участника, если нет conversation_id"
    )
    conversation_id: UUID | None = Field(default=None)
    track_id: UUID | None = Field(default=None)
    context_type: str | None = Field(default=None)
    context_id: UUID | None = Field(default=None)
    
    @model_validator(mode='after')
    def validate_content_or_attachment(self):
        # Если content пустой, но есть track_id или context, разрешаем
        if not self.content and not self.track_id and not self.context_id:
            raise ValueError("Either content, track_id, or context_id is required")
        return self


class MessageOut(BaseModel):
    id: UUID
    conversation_id: UUID
    sender_id: UUID
    content: str
    created_at: datetime
    track_id: UUID | None = None

    class Config:
        from_attributes = True


class ConversationOut(BaseModel):
    id: UUID
    conversation_type: str
    title: str | None = None
    unread_count: int = 0
    updated_at: datetime
    other_participant_id: UUID | None = Field(
        default=None, description="ID другого участника для прямых диалогов"
    )


class CreateGroupConversationRequest(BaseModel):
    title: str = Field(min_length=1, max_length=255)
    participant_ids: list[UUID] = Field(
        default_factory=list, description="Список участников (кроме текущего пользователя)"
    )


class MarkAsReadRequest(BaseModel):
    message_id: UUID | None = Field(
        default=None,
        description="До какого сообщения помечать как прочитанные (если не указано — до последнего)",
    )


