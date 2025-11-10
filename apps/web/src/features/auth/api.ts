import { createApiClient } from '@shared/api/client'
import { API_CONFIG } from '@shared/config/api'

const client = createApiClient(API_CONFIG.gateway)

interface GatewayAuthResponse {
  access_token: string
  refresh_token: string
  user: {
    id: string
    username: string
    avatar_url?: string
  }
}

interface GatewayUserProfile {
  id: string
  username: string
  avatar_url?: string
  music_taste_summary?: {
    top_genres?: string[]
    top_artists?: string[]
  }
}

interface GatewayMeResponse {
  user: GatewayUserProfile
}

export interface AuthResponse {
  accessToken: string
  refreshToken: string
  user: {
    id: string
    username: string
    avatarUrl: string
  }
}

export interface UserProfile {
  id: string
  username: string
  avatarUrl: string
  musicTasteSummary?: {
    topGenres?: string[]
    topArtists?: string[]
  }
}

const mapAuthResponse = (data: GatewayAuthResponse): AuthResponse => ({
  accessToken: data.access_token,
  refreshToken: data.refresh_token,
  user: {
    id: data.user.id,
    username: data.user.username,
    avatarUrl: data.user.avatar_url || '',
  },
})

const mapUserProfile = (data: GatewayUserProfile): UserProfile => ({
  id: data.id,
  username: data.username,
  avatarUrl: data.avatar_url || '',
  musicTasteSummary: data.music_taste_summary
    ? {
        topGenres: data.music_taste_summary.top_genres,
        topArtists: data.music_taste_summary.top_artists,
      }
    : undefined,
})

export const signIn = async (payload: { email: string; password: string }) => {
  const response = await client.post<GatewayAuthResponse>('/api/v1/auth/sign-in', payload)
  return mapAuthResponse(response.data)
}

export const signUp = async (payload: { email: string; password: string; username: string }) => {
  const response = await client.post<GatewayAuthResponse>('/api/v1/auth/sign-up', payload)
  return mapAuthResponse(response.data)
}

export const fetchMe = async () => {
  const response = await client.get<GatewayMeResponse>('/api/v1/me')
  return mapUserProfile(response.data.user)
}
