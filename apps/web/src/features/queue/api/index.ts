import { createApiClient } from '@shared/api/client'
import { API_CONFIG } from '@shared/config/api'
import type { QueueContext } from '../model'

const client = createApiClient(API_CONFIG.gateway)

export interface QueueItem {
  track_id: string
}

export interface QueueResponse {
  items: QueueItem[]
}

export interface CurrentTrackResponse {
  current: QueueItem | null
}

export interface NextTrackResponse {
  next: QueueItem | null
}

export interface PrevTrackResponse {
  previous: QueueItem | null
}

// Получить текущий трек из очереди
export const getCurrentTrack = async (context?: QueueContext): Promise<QueueItem | null> => {
  const endpoint =
    context?.type === 'session'
      ? `/api/v1/sessions/${context.roomId}/queue/current`
      : context?.type === 'group'
        ? `/api/v1/groups/${context.groupId}/queue/current`
        : '/api/v1/me/queue/current'

  const response = await client.get<CurrentTrackResponse>(endpoint)
  return response.data.current
}

// Получить список треков в очереди
export const getQueue = async (limit = 50, context?: QueueContext): Promise<QueueItem[]> => {
  const endpoint =
    context?.type === 'session'
      ? `/api/v1/sessions/${context.roomId}/queue`
      : context?.type === 'group'
        ? `/api/v1/groups/${context.groupId}/queue`
        : '/api/v1/me/queue'

  const response = await client.get<QueueResponse>(endpoint, {
    params: { limit },
  })
  return response.data.items
}

// Добавить трек в очередь
export const addTrackToQueue = async (trackId: string, context?: QueueContext): Promise<void> => {
  const endpoint =
    context?.type === 'session'
      ? `/api/v1/sessions/${context.roomId}/queue/tracks`
      : context?.type === 'group'
        ? `/api/v1/groups/${context.groupId}/queue/tracks`
        : '/api/v1/me/queue/tracks'

  await client.post(endpoint, {
    track_id: trackId,
  })
}

// Удалить трек из очереди
export const removeTrackFromQueue = async (
  trackId: string,
  context?: QueueContext
): Promise<void> => {
  const endpoint =
    context?.type === 'session'
      ? `/api/v1/sessions/${context.roomId}/queue/tracks/${trackId}`
      : context?.type === 'group'
        ? `/api/v1/groups/${context.groupId}/queue/tracks/${trackId}`
        : `/api/v1/me/queue/tracks/${trackId}`

  await client.delete(endpoint)
}

// Получить следующий трек (и сдвинуть курсор)
export const getNextTrack = async (context?: QueueContext): Promise<QueueItem | null> => {
  const endpoint =
    context?.type === 'session'
      ? `/api/v1/sessions/${context.roomId}/queue/next`
      : context?.type === 'group'
        ? `/api/v1/groups/${context.groupId}/queue/next`
        : '/api/v1/me/queue/next'

  const response = await client.post<NextTrackResponse>(endpoint)
  return response.data.next
}

// Получить предыдущий трек (и сдвинуть курсор)
export const getPrevTrack = async (context?: QueueContext): Promise<QueueItem | null> => {
  const endpoint =
    context?.type === 'session'
      ? `/api/v1/sessions/${context.roomId}/queue/prev`
      : context?.type === 'group'
        ? `/api/v1/groups/${context.groupId}/queue/prev`
        : '/api/v1/me/queue/prev'

  const response = await client.post<PrevTrackResponse>(endpoint)
  return response.data.previous
}

// Очистить очередь
export const clearQueue = async (context?: QueueContext): Promise<void> => {
  const endpoint =
    context?.type === 'session'
      ? `/api/v1/sessions/${context.roomId}/queue`
      : context?.type === 'group'
        ? `/api/v1/groups/${context.groupId}/queue`
        : '/api/v1/me/queue'

  await client.delete(endpoint)
}
