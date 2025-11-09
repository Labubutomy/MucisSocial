import { createApiClient } from '@shared/api/client'
import { API_CONFIG } from '@shared/config/api'
import type { PlaylistSummary } from '@entities/playlist'

const client = createApiClient(API_CONFIG.mockApi)

interface MyPlaylistsResponse {
  items: PlaylistSummary[]
}

export const fetchMyPlaylists = async (limit = 12) => {
  const response = await client.get<MyPlaylistsResponse>('/api/v1/me/playlists', {
    params: { limit },
  })
  return response.data.items
}

export const fetchUserPlaylists = async (userId: string, limit = 24) => {
  const response = await client.get<MyPlaylistsResponse>(`/api/v1/users/${userId}/playlists`, {
    params: { limit },
  })
  return response.data.items
}
