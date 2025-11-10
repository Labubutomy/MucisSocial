import { createApiClient } from '@shared/api/client'
import { API_CONFIG } from '@shared/config/api'
import type { PlaylistSummary } from '@entities/playlist'

// Use gateway for my playlists (GET /api/v1/playlists returns current user's playlists)
const gatewayClient = createApiClient(API_CONFIG.gateway)
// Use mock-api for other user's playlists (not available in gateway)
const mockClient = createApiClient(API_CONFIG.mockApi)

// Gateway response format
interface GatewayPlaylistsResponse {
  playlists: Array<{
    id: string
    user_id: string
    name: string
    description: string
    is_private: boolean
    created_at: string
    updated_at: string
    tracks_count: number
  }>
  total: number
}

// Mock API response format
interface MockPlaylistsResponse {
  items: PlaylistSummary[]
}

// Default cover URL for playlists (same as tracks)
const DEFAULT_COVER_URL =
  'https://mir-s3-cdn-cf.behance.net/projects/202/e2ba0e187042211.Y3JvcCw4MDgsNjMyLDAsMA.png'

// Map gateway playlist to PlaylistSummary
const mapGatewayPlaylist = (
  playlist: GatewayPlaylistsResponse['playlists'][number]
): PlaylistSummary => ({
  id: playlist.id,
  title: playlist.name,
  coverUrl: DEFAULT_COVER_URL, // Use default cover URL
  itemsCount: playlist.tracks_count,
})

export const fetchMyPlaylists = async (limit = 12) => {
  // Use gateway: GET /api/v1/playlists returns current user's playlists
  const response = await gatewayClient.get<GatewayPlaylistsResponse>('/api/v1/playlists', {
    params: { limit, offset: 0 },
  })
  return response.data.playlists.map(mapGatewayPlaylist)
}

export const fetchUserPlaylists = async (userId: string, limit = 24) => {
  // Not available in gateway, use mock-api
  const response = await mockClient.get<MockPlaylistsResponse>(
    `/api/v1/users/${userId}/playlists`,
    {
      params: { limit },
    }
  )
  return response.data.items
}
