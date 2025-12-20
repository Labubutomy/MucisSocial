import { createApiClient } from '@shared/api/client'
import { API_CONFIG } from '@shared/config/api'

const gatewayClient = createApiClient(API_CONFIG.gateway)

export interface RequestSession {
  id: string
  artist_id: string
  session_code: string
  is_active: boolean
  qr_code_url: string | null
  created_at: string
  updated_at: string
}

export interface TrackRequest {
  id: string
  session_id: string
  requester_id: string
  artist_id: string
  track_id: string
  status: 'pending' | 'accepted' | 'declined'
  message: string | null
  created_at: string
  updated_at: string
}

export interface CreateTrackRequestData {
  session_code: string
  track_id: string
  message?: string
}

export interface TrackSearchResult {
  id: string
  title: string
  artist_name: string
}

export interface UserCoins {
  user_id: string
  coins: number
}

// Sessions API
export const createRequestSession = async (): Promise<RequestSession> => {
  const response = await gatewayClient.post('/api/v1/music-requests/sessions', {})
  return response.data
}

export const getMyActiveSession = async (): Promise<RequestSession | null> => {
  try {
    const response = await gatewayClient.get('/api/v1/music-requests/sessions/my')
    return response.data
  } catch (error: any) {
    if (error.response?.status === 404) {
      return null
    }
    throw error
  }
}

export const getSessionByCode = async (sessionCode: string): Promise<RequestSession | null> => {
  try {
    const response = await gatewayClient.get(`/api/v1/music-requests/sessions/code/${sessionCode}`)
    return response.data
  } catch (error: any) {
    if (error.response?.status === 404) {
      return null
    }
    throw error
  }
}

export const deactivateSession = async (sessionId: string): Promise<void> => {
  await gatewayClient.delete(`/api/v1/music-requests/sessions/${sessionId}`)
}

export const getQrCodeUrl = (sessionCode: string): string => {
  return `${API_CONFIG.gateway}/api/v1/music-requests/sessions/${sessionCode}/qr`
}

// Track Requests API
export const createTrackRequest = async (data: CreateTrackRequestData): Promise<TrackRequest> => {
  const response = await gatewayClient.post('/api/v1/music-requests/requests', data)
  return response.data
}

export const getIncomingRequests = async (status?: string): Promise<TrackRequest[]> => {
  const params = status ? { status } : {}
  const response = await gatewayClient.get('/api/v1/music-requests/requests/incoming', { params })
  return response.data
}

export const getOutgoingRequests = async (): Promise<TrackRequest[]> => {
  const response = await gatewayClient.get('/api/v1/music-requests/requests/outgoing')
  return response.data
}

export const handleTrackRequest = async (
  requestId: string,
  action: 'accept' | 'decline'
): Promise<TrackRequest> => {
  const response = await gatewayClient.patch(`/api/v1/music-requests/requests/${requestId}`, {
    action,
  })
  return response.data
}

// Coins API
export const getUserCoins = async (): Promise<UserCoins> => {
  const response = await gatewayClient.get('/api/v1/me/coins')
  return response.data
}
