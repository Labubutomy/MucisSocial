import { createApiClient } from '@shared/api/client'

const client = createApiClient('http://localhost:8100')

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

export const signIn = async (payload: { email: string; password: string }) => {
  const response = await client.post<AuthResponse>('/auth/sign-in', payload)
  return response.data
}

export const signUp = async (payload: { email: string; password: string; username: string }) => {
  const response = await client.post<AuthResponse>('/auth/sign-up', payload)
  return response.data
}

export const fetchMe = async () => {
  const response = await client.get<UserProfile>('/me')
  return response.data
}
