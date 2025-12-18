import { createEffect, createEvent, createStore, sample } from 'effector'
import { routes } from '@shared/router'
import {
  fetchConversations,
  fetchMessages,
  markConversationRead,
  sendMessage,
  createDirectConversation,
  type Conversation,
  type Message,
} from '../api'
import { MessagingWebSocket } from '../lib/websocket'

export const loadConversationsFx = createEffect(fetchConversations)
export const loadMessagesFx = createEffect(async ({ conversationId }: { conversationId: string }) =>
  fetchMessages(conversationId)
)
export const loadMessageByIdFx = createEffect(
  async ({ messageId, conversationId }: { messageId: string; conversationId: string }) => {
    // Загружаем все сообщения и находим нужное
    const messages = await fetchMessages(conversationId, 100)
    return messages.find(m => m.id === messageId) || null
  }
)
export const sendMessageFx = createEffect(sendMessage)
export const markConversationReadFx = createEffect(markConversationRead)
export const createDirectConversationFx = createEffect(createDirectConversation)
export const connectMessagingWsFx = createEffect(async () => {
  if (!wsClient) {
    throw new Error('WebSocket недоступен')
  }
  await wsClient.connect()
})

export const conversationOpened = createEvent<{ conversationId: string }>()
export const messageSendRequested = createEvent<{ content: string; trackId?: string }>()
export const wsMessageEvent = createEvent<unknown>()
export const wsConversationReadEvent = createEvent<unknown>()

export const $conversations = createStore<Conversation[]>([])
  .on(loadConversationsFx.doneData, (_, conversations) => conversations)
  .on(createDirectConversationFx.doneData, (state, conversation) => {
    // Добавляем новый диалог или обновляем существующий
    const existingIndex = state.findIndex(c => c.id === conversation.id)
    if (existingIndex >= 0) {
      return state.map((c, i) => (i === existingIndex ? conversation : c))
    }
    return [...state, conversation]
  })
  .on(sendMessageFx.doneData, (state, message) =>
    state.map(conv => {
      const conversationId = message.conversationId || message.conversation_id
      const createdAt = message.createdAt || message.created_at
      return conv.id === conversationId
        ? {
            ...conv,
            unreadCount: conv.unreadCount,
            updatedAt: createdAt || conv.updatedAt,
          }
        : conv
    })
  )

export const $activeConversationId = createStore<string | null>(null)
  .on(conversationOpened, (_, { conversationId }) => conversationId)
  .reset(routes.messages.opened)

export const $messages = createStore<Message[]>([])
  .on(loadMessagesFx.doneData, (_, messages) => {
    // При загрузке сообщений заменяем весь список (это нормально при открытии диалога)
    return messages
  })
  .on(sendMessageFx.doneData, (state, message) => {
    // При отправке сообщения добавляем его в конец, если его еще нет
    const exists = state.some(m => m.id === message.id)
    if (exists) return state
    return [...state, message]
  })
  .reset(routes.messages.opened)

// Тип для WebSocket события сообщения
// Событие приходит в формате: { type: 'message', conversation_id: '...', message: { event_type: 'message_sent', message_id: '...', ... } }
type WsMessageEvent = {
  type?: string
  conversation_id?: string
  message?: {
    event_type?: string
    message_id?: string
    conversation_id?: string
    sender_id?: string
    [key: string]: unknown
  }
  [key: string]: unknown
}

// Реакция на WebSocket-события: добавляем новое сообщение напрямую, если оно для активного диалога
// Это позволяет избежать полной перезагрузки списка сообщений
const wsMessageReceived = createEvent<Message>()

// Обрабатываем WebSocket события с полным content
sample({
  clock: wsMessageEvent,
  source: { activeConversationId: $activeConversationId, messages: $messages },
  filter: (
    {
      activeConversationId,
      messages,
    }: { activeConversationId: string | null; messages: Message[] },
    event: unknown
  ): boolean => {
    const e = event as WsMessageEvent
    const messageData = e.message || e
    if (!messageData || !messageData.message_id || !messageData.conversation_id) return false
    // Проверяем, что есть content
    if (!messageData.content) return false
    const convId = messageData.conversation_id || e.conversation_id
    // Обрабатываем только если это сообщение для активного диалога
    if (activeConversationId && String(convId) === String(activeConversationId)) {
      // Проверяем, нет ли уже такого сообщения
      const exists = messages.some((m: Message) => m.id === messageData.message_id)
      return !exists
    }
    return false
  },
  fn: (_: { activeConversationId: string | null; messages: Message[] }, event: unknown) => {
    const e = event as WsMessageEvent
    const messageData = e.message || e
    // Создаем объект сообщения из события
    const newMessage: Message = {
      id: String(messageData.message_id),
      conversationId: String(messageData.conversation_id || e.conversation_id),
      senderId: String(messageData.sender_id || ''),
      content: (messageData.content as string) || '',
      createdAt: messageData.created_at
        ? new Date(messageData.created_at as string).toISOString()
        : new Date().toISOString(),
      trackId: messageData.track_id ? String(messageData.track_id) : undefined,
    }
    return newMessage
  },
  target: wsMessageReceived,
})

// Обрабатываем WebSocket события без content - загружаем через API
sample({
  clock: wsMessageEvent,
  source: { activeConversationId: $activeConversationId, messages: $messages },
  filter: (
    {
      activeConversationId,
      messages,
    }: { activeConversationId: string | null; messages: Message[] },
    event: unknown
  ): boolean => {
    const e = event as WsMessageEvent
    const messageData = e.message || e
    if (!messageData || !messageData.message_id || !messageData.conversation_id) return false
    // Проверяем, что нет content
    if (messageData.content) return false
    const convId = messageData.conversation_id || e.conversation_id
    // Обрабатываем только если это сообщение для активного диалога
    if (activeConversationId && String(convId) === String(activeConversationId)) {
      // Проверяем, нет ли уже такого сообщения
      const exists = messages.some((m: Message) => m.id === messageData.message_id)
      return !exists
    }
    return false
  },
  fn: (_: { activeConversationId: string | null; messages: Message[] }, event: unknown) => {
    const e = event as WsMessageEvent
    const messageData = e.message || e
    return {
      messageId: String(messageData.message_id),
      conversationId: String(messageData.conversation_id || e.conversation_id),
    }
  },
  target: loadMessageByIdFx,
})

// Добавляем сообщение из WebSocket напрямую в store
$messages.on(wsMessageReceived, (state, message) => {
  // Проверяем, нет ли уже такого сообщения
  const exists = state.some(m => m.id === message.id)
  if (exists) return state
  return [...state, message]
})

// Добавляем загруженное сообщение в store
$messages.on(loadMessageByIdFx.doneData, (state, message) => {
  if (!message) return state
  // Проверяем, нет ли уже такого сообщения
  const exists = state.some(m => m.id === message.id)
  if (exists) return state
  return [...state, message]
})

// Добавляем сообщение из WebSocket напрямую в store
$messages.on(wsMessageReceived, (state, message) => {
  // Проверяем, нет ли уже такого сообщения
  const exists = state.some(m => m.id === message.id)
  if (exists) return state
  return [...state, message]
})

// Для других диалогов просто обновляем список диалогов
sample({
  clock: wsMessageEvent,
  source: $activeConversationId,
  filter: (activeConversationId, event: unknown): boolean => {
    const e = event as WsMessageEvent
    const messageData = e.message || e
    const convId = messageData.conversation_id || e.conversation_id
    if (!convId) return false
    // Обновляем список диалогов, если это не активный диалог
    return !activeConversationId || String(convId) !== String(activeConversationId)
  },
  target: loadConversationsFx,
})

sample({
  clock: wsConversationReadEvent,
  target: loadConversationsFx,
})

sample({
  clock: routes.messages.opened,
  target: [loadConversationsFx, connectMessagingWsFx],
})

// Загружаем друзей при открытии страницы сообщений
import { loadFriendsFx } from '@features/friends/model'
sample({
  clock: routes.messages.opened,
  target: loadFriendsFx,
})

sample({
  clock: conversationOpened,
  fn: ({ conversationId }) => ({ conversationId }),
  target: loadMessagesFx,
})

// Помечаем диалог прочитанным при открытии
sample({
  clock: conversationOpened,
  filter: ({ conversationId }) => Boolean(conversationId),
  fn: ({ conversationId }) => conversationId,
  target: markConversationReadFx,
})

sample({
  clock: messageSendRequested,
  source: $activeConversationId,
  filter: (conversationId): conversationId is string => Boolean(conversationId),
  fn: (conversationId, payload) => ({
    conversationId: conversationId as string,
    content: payload.content,
    trackId: payload.trackId,
  }),
  target: sendMessageFx,
})

sample({
  clock: sendMessageFx.doneData,
  filter: (message): message is Message & { conversationId: string } => {
    const conversationId = message.conversationId || message.conversation_id
    return Boolean(conversationId)
  },
  fn: (message): string => {
    const conversationId = message.conversationId || message.conversation_id
    return conversationId!
  },
  target: markConversationReadFx,
})

// После пометки прочитанным обновляем список диалогов (сброс unread)
sample({
  clock: markConversationReadFx.done,
  target: loadConversationsFx,
})

// После создания диалога открываем его
sample({
  clock: createDirectConversationFx.doneData,
  fn: conversation => ({ conversationId: conversation.id }),
  target: conversationOpened,
})

// После создания диалога обновляем список диалогов
sample({
  clock: createDirectConversationFx.done,
  target: loadConversationsFx,
})

// Инициализация WebSocket клиента (только в браузере)
const wsClient = typeof window !== 'undefined' ? new MessagingWebSocket() : null

if (wsClient) {
  wsClient.onMessage('message', message => wsMessageEvent(message))
  wsClient.onMessage('conversation_read', message => wsConversationReadEvent(message))
}
