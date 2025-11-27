import axios from 'axios'
import { createApiClient } from '@shared/api/client'
import { API_CONFIG } from '@shared/config/api'

const sessionApiClient = createApiClient(API_CONFIG.gateway)

export interface RoomState {
  roomId: string
  currentTrack: {
    trackId: string
    title: string
    artist: string
    duration: number
    cdnUrl: string
  } | null
  position: number
  isPlaying: boolean
  participants: Array<{
    userId: string
    username: string
    isOnline: boolean
    joinedAt: string
  }>
  queue: Array<{
    trackId: string
    title: string
    artist: string
    duration: number
    cdnUrl: string
  }>
  lastAction: {
    actionId: string
    type: string
    userId: string
    timestamp: string
  } | null
  createdAt: string
  updatedAt: string
}

export const createRoom = async (roomId: string): Promise<RoomState> => {
  const response = await sessionApiClient.post<RoomState>(`/api/rooms/${roomId}`)
  return response.data
}

export const getRoom = async (roomId: string): Promise<RoomState | null> => {
  try {
    const response = await sessionApiClient.get<RoomState>(`/api/rooms/${roomId}`)
    return response.data
  } catch (error: unknown) {
    if (axios.isAxiosError(error) && error.response?.status === 404) {
      return null
    }
    throw error
  }
}

export const addParticipant = async (
  roomId: string,
  userId: string,
  username: string
): Promise<RoomState> => {
  const response = await sessionApiClient.post<RoomState>(
    `/api/rooms/${roomId}/participants?userId=${userId}&username=${encodeURIComponent(username)}`
  )
  return response.data
}

export const removeParticipant = async (roomId: string, userId: string): Promise<RoomState> => {
  const response = await sessionApiClient.delete<RoomState>(
    `/api/rooms/${roomId}/participants/${userId}`
  )
  return response.data
}
