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

// User taste statistics API response format
interface GatewayUserTasteResponse {
  user_id: string
  top_genres: Array<{
    genre: string
    count: number
  }>
  top_artists: Array<{
    artist_id: string
    count: number
  }>
}

// Gateway Artist response format
interface GatewayArtistResponse {
  id: string
  name: string
  avatar_url?: string
  genres?: string[]
}

/**
 * Fetches user taste statistics (top genres and artists)
 * @param userId User ID
 * @returns Object with topGenres (array of genre names) and topArtists (array of artist names)
 */
interface GatewayUserResponse {
  id: string
  username: string
  avatar_url?: string
}

/**
 * Fetches user information by ID
 * @param userId User ID
 * @returns User profile with id, username, and avatarUrl
 */
export const fetchUserById = async (
  userId: string
): Promise<{
  id: string
  username: string
  avatarUrl?: string
}> => {
  try {
    const response = await gatewayClient.get<GatewayUserResponse>(`/api/v1/users/${userId}`)
    return {
      id: response.data.id,
      username: response.data.username,
      avatarUrl: response.data.avatar_url,
    }
  } catch (error) {
    console.error(`Failed to fetch user ${userId}:`, error)
    throw error
  }
}

export const fetchUserTaste = async (
  userId: string
): Promise<{
  topGenres: string[]
  topArtists: string[]
}> => {
  try {
    const response = await gatewayClient.get<GatewayUserTasteResponse>(
      `/api/v1/users/${userId}/taste`
    )

    // Extract genre names
    const topGenres = response.data.top_genres.map(g => g.genre)

    // Fetch artist names for top artists
    const topArtists = await Promise.all(
      response.data.top_artists.map(async artistStat => {
        try {
          const artistResponse = await gatewayClient.get<GatewayArtistResponse>(
            `/api/v1/artists/${artistStat.artist_id}`
          )
          return artistResponse.data.name
        } catch (error) {
          console.warn(`Failed to fetch artist ${artistStat.artist_id}:`, error)
          return `Unknown Artist (${artistStat.artist_id})`
        }
      })
    )

    return {
      topGenres,
      topArtists,
    }
  } catch (error) {
    console.error('Failed to fetch user taste:', error)
    // Return empty arrays on error
    return {
      topGenres: [],
      topArtists: [],
    }
  }
}
