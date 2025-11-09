import { createApiClient } from '@shared/api/client'
import type { PlaylistSummary } from '@entities/playlist'

const client = createApiClient('http://localhost:8100')

interface MyPlaylistsResponse {
  items: PlaylistSummary[]
}

export const fetchMyPlaylists = async (limit = 12) => {
  const response = await client.get<MyPlaylistsResponse>('/me/playlists', {
    params: { limit },
  })
  return response.data.items
}

export const fetchUserPlaylists = async (userId: string, limit = 24) => {
  const response = await client.get<MyPlaylistsResponse>(`/users/${userId}/playlists`, {
    params: { limit },
  })
  return response.data.items
}
