from __future__ import annotations

import logging
from typing import Annotated
from uuid import UUID

import jwt
from fastapi import Depends, FastAPI, Header, HTTPException, Query, status
from sqlalchemy import func, select
from sqlalchemy.ext.asyncio import AsyncSession

from core.config import get_settings
from core.database import get_db
from models import Conversation, Message, MessageContext, MessageTrack, UserConversation
from schemas import (
    ConversationOut,
    CreateGroupConversationRequest,
    MarkAsReadRequest,
    MessageOut,
    SendMessageRequest,
)
from services.kafka_service import (
    publish_conversation_read,
    publish_message_sent,
)


logger = logging.getLogger(__name__)
settings = get_settings()

app = FastAPI(
    title="Messaging Service",
    version=settings.app_version,
    description="Private messaging around music for Music Social",
)

# CORS обрабатывается на уровне gateway, не нужно дублировать здесь


async def get_current_user_id(
    authorization: Annotated[str | None, Header(alias="Authorization")] = None,
) -> UUID:
    if not authorization:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Missing Authorization header",
        )
    if not authorization.startswith("Bearer "):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid Authorization header format",
        )
    token = authorization.removeprefix("Bearer ").strip()
    try:
        payload = jwt.decode(
            token,
            settings.jwt_secret,
            algorithms=[settings.jwt_algorithm],
        )
        user_id = payload.get("user_id")
        if not user_id:
            raise ValueError("user_id missing in token")
        return UUID(user_id)
    except Exception as exc:
        logger.warning("Token validation failed: %s", exc)
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid token",
        ) from exc


@app.get("/health")
async def health_check() -> dict[str, str]:
    return {"status": "ok"}


async def _get_or_create_direct_conversation(
    db: AsyncSession,
    current_user_id: UUID,
    other_user_id: UUID,
) -> Conversation:
    """
    Найти или создать личный диалог между двумя пользователями.
    """
    # Найти диалоги, где ровно эти два участника
    uc_subq = (
        select(
            UserConversation.conversation_id,
            func.count(UserConversation.user_id).label("cnt"),
        )
        .where(UserConversation.user_id.in_([current_user_id, other_user_id]))
        .group_by(UserConversation.conversation_id)
        .having(func.count(UserConversation.user_id) == 2)
        .subquery()
    )

    res = await db.execute(
        select(Conversation)
        .join(uc_subq, Conversation.id == uc_subq.c.conversation_id)
        .where(Conversation.conversation_type == "direct")
    )
    conversation = res.scalar_one_or_none()

    if conversation:
        return conversation

    # Создать новый direct-диалог
    conversation = Conversation(conversation_type="direct")
    db.add(conversation)
    await db.flush()

    db.add_all(
        [
            UserConversation(conversation_id=conversation.id, user_id=current_user_id),
            UserConversation(conversation_id=conversation.id, user_id=other_user_id),
        ]
    )
    await db.flush()

    return conversation


@app.post(
    "/api/v1/messages",
    response_model=MessageOut,
    status_code=status.HTTP_201_CREATED,
)
async def send_message(
    payload: SendMessageRequest,
    current_user_id: UUID = Depends(get_current_user_id),
    db: AsyncSession = Depends(get_db),
) -> MessageOut:
    """
    Отправка сообщения.

    - если указан conversation_id — используем существующий диалог
    - иначе создаём (или находим) личный диалог с recipient_id
    """
    if not payload.conversation_id and not payload.recipient_id:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Either conversation_id or recipient_id is required",
        )

    # 1. Ensure conversation exists (for personal 1-1 chat)
    conversation: Conversation | None = None

    if payload.conversation_id:
        res = await db.execute(
            select(Conversation).where(Conversation.id == payload.conversation_id)
        )
        conversation = res.scalar_one_or_none()
        if not conversation:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND, detail="Conversation not found"
            )
    else:
        assert payload.recipient_id is not None
        conversation = await _get_or_create_direct_conversation(
            db=db,
            current_user_id=current_user_id,
            other_user_id=payload.recipient_id,
        )

    # 2. Create message
    # content может быть пустым, если есть track_id или context
    message_content = payload.content or ""
    if not message_content and not payload.track_id and not payload.context_id:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Either content, track_id, or context_id is required",
        )
    
    message = Message(
        conversation_id=conversation.id,
        sender_id=current_user_id,
        content=message_content,
    )
    db.add(message)
    await db.flush()

    # 3. Optional track and context bindings
    if payload.track_id:
        db.add(MessageTrack(message_id=message.id, track_id=payload.track_id))

    if payload.context_type and payload.context_id:
        db.add(
            MessageContext(
                message_id=message.id,
                context_type=payload.context_type,
                context_id=payload.context_id,
            )
        )

    # 4. Update conversation meta
    conversation.last_message_id = message.id
    conversation.last_message_at = message.created_at

    # 5. Increase unread counters for other participants
    res_uc = await db.execute(
        select(UserConversation).where(
            UserConversation.conversation_id == conversation.id,
            UserConversation.user_id != current_user_id,
        )
    )
    for uc in res_uc.scalars():
        uc.unread_count = (uc.unread_count or 0) + 1

    await db.commit()
    await db.refresh(message)

    # Получаем track_id из связанной таблицы (если есть)
    track_id = payload.track_id  # Используем track_id из payload, так как он уже сохранен в MessageTrack

    # Собираем список получателей для события
    res_all_uc = await db.execute(
        select(UserConversation.user_id).where(
            UserConversation.conversation_id == conversation.id,
            UserConversation.user_id != current_user_id,
        )
    )
    recipient_ids = [row[0] for row in res_all_uc.all()]

    # Публикуем domain-событие в Kafka (non-blocking best-effort)
    try:
        publish_message_sent(
            message_id=message.id,
            conversation_id=conversation.id,
            sender_id=current_user_id,
            recipient_ids=recipient_ids,
            created_at=message.created_at,
            track_id=payload.track_id,
            context_type=payload.context_type,
            context_id=payload.context_id,
        )
    except Exception as exc:
        logger.warning("Failed to publish message_sent event: %s", exc)

    # Создаем MessageOut с track_id
    message_out = MessageOut(
        id=message.id,
        conversation_id=message.conversation_id,
        sender_id=message.sender_id,
        content=message.content,
        created_at=message.created_at,
        track_id=track_id,
    )
    return message_out


@app.get(
    "/api/v1/conversations",
    response_model=list[ConversationOut],
)
async def list_conversations(
    current_user_id: UUID = Depends(get_current_user_id),
    db: AsyncSession = Depends(get_db),
) -> list[ConversationOut]:
    # Simplified: return bare conversations where user participates
    res = await db.execute(
        select(Conversation, UserConversation)
        .join(UserConversation, Conversation.id == UserConversation.conversation_id)
        .where(UserConversation.user_id == current_user_id)
        .order_by(Conversation.updated_at.desc())
    )
    items: list[ConversationOut] = []
    for conv, uc in res.all():
        # Для прямых диалогов находим ID другого участника
        other_participant_id = None
        if conv.conversation_type == "direct":
            res_other = await db.execute(
                select(UserConversation.user_id).where(
                    UserConversation.conversation_id == conv.id,
                    UserConversation.user_id != current_user_id,
                )
            )
            other_user = res_other.scalar_one_or_none()
            if other_user:
                other_participant_id = other_user
        
        items.append(
            ConversationOut(
                id=conv.id,
                conversation_type=conv.conversation_type,
                title=conv.title,
                unread_count=uc.unread_count or 0,
                updated_at=conv.updated_at,
                other_participant_id=other_participant_id,
            )
        )
    return items


@app.get(
    "/api/v1/conversations/{conversation_id}/messages",
    response_model=list[MessageOut],
)
async def list_messages(
    conversation_id: UUID,
    limit: int = 50,
    before_message_id: UUID | None = None,
    current_user_id: UUID = Depends(get_current_user_id),
    db: AsyncSession = Depends(get_db),
) -> list[MessageOut]:
    # Проверяем, что пользователь участник диалога
    res = await db.execute(
        select(UserConversation).where(
            UserConversation.conversation_id == conversation_id,
            UserConversation.user_id == current_user_id,
        )
    )
    if res.scalar_one_or_none() is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Conversation not found or access denied",
        )

    # Базовый запрос
    stmt = select(Message).where(Message.conversation_id == conversation_id)

    if before_message_id:
        # Найти created_at указанного сообщения и грузить более старые
        res_msg = await db.execute(
            select(Message).where(
                Message.id == before_message_id,
                Message.conversation_id == conversation_id,
            )
        )
        pivot = res_msg.scalar_one_or_none()
        if pivot:
            stmt = stmt.where(Message.created_at < pivot.created_at)

    stmt = stmt.order_by(Message.created_at.desc()).limit(limit)
    res_msgs = await db.execute(stmt)
    messages = list(res_msgs.scalars())

    # Возвращаем в обратном порядке (от старых к новым)
    messages.reverse()
    # Загружаем треки для каждого сообщения
    items: list[MessageOut] = []
    for msg in messages:
        # Получаем track_id из связанной таблицы
        res_track = await db.execute(
            select(MessageTrack.track_id).where(MessageTrack.message_id == msg.id).limit(1)
        )
        track_id = res_track.scalar_one_or_none()
        
        items.append(
            MessageOut(
                id=msg.id,
                conversation_id=msg.conversation_id,
                sender_id=msg.sender_id,
                content=msg.content,
                created_at=msg.created_at,
                track_id=track_id,
            )
        )
    return items


@app.post(
    "/api/v1/conversations/{conversation_id}/read",
)
async def mark_conversation_read(
    conversation_id: UUID,
    payload: MarkAsReadRequest | None = None,
    current_user_id: UUID = Depends(get_current_user_id),
    db: AsyncSession = Depends(get_db),
) -> dict[str, bool]:
    # Найти связь user <-> conversation
    res_uc = await db.execute(
        select(UserConversation).where(
            UserConversation.conversation_id == conversation_id,
            UserConversation.user_id == current_user_id,
        )
    )
    uc = res_uc.scalar_one_or_none()
    if uc is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Conversation not found or access denied",
        )

    # Определяем, до какого сообщения помечать прочитанными
    target_message_id: UUID | None = None
    if payload and payload.message_id:
        target_message_id = payload.message_id
    else:
        # Берем последнее сообщение
        res_msg = await db.execute(
            select(Message.id)
            .where(Message.conversation_id == conversation_id)
            .order_by(Message.created_at.desc())
            .limit(1)
        )
        target_message_id = res_msg.scalar_one_or_none()

    if target_message_id is None:
        return {"success": True}

    uc.last_read_message_id = target_message_id
    uc.unread_count = 0

    await db.commit()

    # Публикуем событие о прочтении
    try:
        publish_conversation_read(
            conversation_id=conversation_id,
            user_id=current_user_id,
            message_id=target_message_id,
        )
    except Exception as exc:
        logger.warning("Failed to publish conversation_read event: %s", exc)

    return {"success": True}


@app.post(
    "/api/v1/conversations",
    response_model=ConversationOut,
    status_code=status.HTTP_201_CREATED,
)
async def create_group_conversation(
    payload: CreateGroupConversationRequest,
    current_user_id: UUID = Depends(get_current_user_id),
    db: AsyncSession = Depends(get_db),
) -> ConversationOut:
    """
    Создать групповой чат с указанными участниками.
    """
    participant_ids = set(payload.participant_ids or [])
    participant_ids.add(current_user_id)

    conversation = Conversation(
        conversation_type="group",
        title=payload.title,
    )
    db.add(conversation)
    await db.flush()

    db.add_all(
        [
            UserConversation(conversation_id=conversation.id, user_id=user_id)
            for user_id in participant_ids
        ]
    )
    await db.commit()
    await db.refresh(conversation)

    return ConversationOut(
        id=conversation.id,
        conversation_type=conversation.conversation_type,
        title=conversation.title,
        unread_count=0,
        updated_at=conversation.updated_at,
        other_participant_id=None,  # Для групповых чатов не применимо
    )


@app.post(
    "/api/v1/conversations/direct",
    response_model=ConversationOut,
    status_code=status.HTTP_201_CREATED,
)
async def create_direct_conversation(
    recipient_id: UUID = Query(..., description="ID пользователя для создания диалога"),
    current_user_id: UUID = Depends(get_current_user_id),
    db: AsyncSession = Depends(get_db),
) -> ConversationOut:
    """
    Создать или найти личный диалог с указанным пользователем.
    """
    conversation = await _get_or_create_direct_conversation(
        db=db,
        current_user_id=current_user_id,
        other_user_id=recipient_id,
    )
    await db.commit()
    await db.refresh(conversation)

    # Получаем unread_count для текущего пользователя
    res_uc = await db.execute(
        select(UserConversation).where(
            UserConversation.conversation_id == conversation.id,
            UserConversation.user_id == current_user_id,
        )
    )
    uc = res_uc.scalar_one_or_none()
    unread_count = uc.unread_count if uc else 0

    # Для прямых диалогов находим ID другого участника
    other_participant_id = None
    if conversation.conversation_type == "direct":
        res_other = await db.execute(
            select(UserConversation.user_id).where(
                UserConversation.conversation_id == conversation.id,
                UserConversation.user_id != current_user_id,
            )
        )
        other_user = res_other.scalar_one_or_none()
        if other_user:
            other_participant_id = other_user
    
    return ConversationOut(
        id=conversation.id,
        conversation_type=conversation.conversation_type,
        title=conversation.title,
        unread_count=unread_count,
        updated_at=conversation.updated_at,
        other_participant_id=other_participant_id,
    )



if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "main:app",
        host=settings.host,
        port=settings.port,
        reload=False,
    )


