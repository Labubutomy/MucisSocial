import { createApiClient } from '@shared/api/client'
import { API_CONFIG } from '@shared/config/api'

const gatewayClient = createApiClient(API_CONFIG.gateway)

export interface Conversation {
  id: string
  conversationType: 'direct' | 'group'
  title: string | null
  unreadCount: number
  updatedAt: string
  otherParticipantId?: string | null
  other_participant_id?: string | null
}

// Тип для сырого сообщения от API (может быть в snake_case или camelCase)
type RawMessage = {
  id: string
  conversation_id?: string
  conversationId?: string
  sender_id?: string
  senderId?: string
  content: string
  created_at?: string
  createdAt?: string
  track_id?: string
  trackId?: string
}

// Интерфейс поддерживает оба формата (snake_case от бекенда и camelCase для фронтенда)
export interface Message {
  id: string
  conversationId?: string
  conversation_id?: string
  senderId?: string
  sender_id?: string
  content: string
  createdAt?: string
  created_at?: string
  trackId?: string
  track_id?: string
}

// Функция для нормализации сообщения из snake_case в camelCase
const normalizeMessage = (msg: RawMessage): Message => {
  return {
    id: msg.id,
    conversationId: msg.conversation_id || msg.conversationId,
    conversation_id: msg.conversation_id || msg.conversationId,
    senderId: msg.sender_id || msg.senderId,
    sender_id: msg.sender_id || msg.senderId,
    content: msg.content || '',
    createdAt: msg.created_at || msg.createdAt,
    created_at: msg.created_at || msg.createdAt,
    trackId: msg.track_id || msg.trackId,
    track_id: msg.track_id || msg.trackId,
  }
}

export interface SendMessagePayload {
  conversationId?: string
  recipientId?: string
  content: string
  trackId?: string
}

export const fetchConversations = async (): Promise<Conversation[]> => {
  const { data } = await gatewayClient.get<Conversation[]>('/api/v1/messaging/conversations')
  return data
}

export const fetchMessages = async (conversationId: string, limit = 50): Promise<Message[]> => {
  const { data } = await gatewayClient.get<RawMessage[]>(
    `/api/v1/messaging/conversations/${conversationId}/messages`,
    { params: { limit } }
  )
  // Нормализуем сообщения из snake_case в camelCase
  return data.map(normalizeMessage)
}

export const sendMessage = async (payload: SendMessagePayload): Promise<Message> => {
  const body: Record<string, unknown> = {
    content: payload.content || '',
  }
  if (payload.conversationId) body.conversation_id = payload.conversationId
  if (payload.recipientId) body.recipient_id = payload.recipientId
  if (payload.trackId) body.track_id = payload.trackId

  const { data } = await gatewayClient.post<RawMessage>('/api/v1/messaging/messages', body)
  // Нормализуем сообщение из snake_case в camelCase
  return normalizeMessage(data)
}

export const markConversationRead = async (conversationId: string): Promise<void> => {
  await gatewayClient.post(`/api/v1/messaging/conversations/${conversationId}/read`, {})
}

export const createDirectConversation = async (recipientId: string): Promise<Conversation> => {
  const { data } = await gatewayClient.post<Conversation>(
    '/api/v1/messaging/conversations/direct',
    null,
    {
      params: { recipient_id: recipientId },
    }
  )
  return data
}
