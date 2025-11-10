import { createApiClient } from '@shared/api/client'
import { API_CONFIG } from '@shared/config/api'
import type { PlaylistDetail, PlaylistSummary } from './model/types'
import type { Track } from '@entities/track'

// Use gateway for playlist endpoints
const gatewayClient = createApiClient(API_CONFIG.gateway)

// Gateway playlist response format
interface GatewayPlaylist {
  id: string
  user_id: string
  name: string
  description: string
  is_private: boolean
  created_at: string
  updated_at: string
  tracks_count: number
}

interface GatewayPlaylistsResponse {
  playlists: GatewayPlaylist[]
  total: number
}

interface GatewayCreatePlaylistResponse {
  playlist_id: string
}

// Default cover URL for playlists (same as tracks)
const DEFAULT_COVER_URL =
  'https://mir-s3-cdn-cf.behance.net/projects/202/e2ba0e187042211.Y3JvcCw4MDgsNjMyLDAsMA.png'

// Map gateway playlist to PlaylistSummary
const mapGatewayPlaylist = (playlist: GatewayPlaylist): PlaylistSummary => ({
  id: playlist.id,
  title: playlist.name,
  coverUrl: DEFAULT_COVER_URL, // Use default cover URL
  itemsCount: playlist.tracks_count,
})

// Map gateway playlist to PlaylistDetail
const mapGatewayPlaylistDetail = (playlist: GatewayPlaylist): PlaylistDetail => ({
  id: playlist.id,
  title: playlist.name,
  description: playlist.description,
  coverUrl: DEFAULT_COVER_URL, // Use default cover URL
  itemsCount: playlist.tracks_count,
  owner: {
    id: playlist.user_id,
    name: '', // Gateway doesn't provide owner name
  },
  type: 'playlist',
  liked: false,
  totalDuration: undefined,
})

export const fetchPlaylistDetail = async (playlistId: string) => {
  const response = await gatewayClient.get<GatewayPlaylist>(`/api/v1/playlists/${playlistId}`)

  if (!response.data) {
    console.error('Playlist response is null or undefined', { playlistId, response })
    throw new Error('Playlist not found')
  }

  return mapGatewayPlaylistDetail(response.data)
}

// Gateway playlist tracks response format
interface GatewayPlaylistTracksResponse {
  tracks: Array<{
    track_id: string
    playlist_id: string
    position: number
    added_at?: string
  }>
  total: number
}

export const fetchPlaylistTracks = async (playlistId: string) => {
  // Gateway returns track_ids, need to fetch full track details
  const response = await gatewayClient.get<GatewayPlaylistTracksResponse>(
    `/api/v1/playlists/${playlistId}/tracks`
  )

  // Получаем полную информацию о треках через tracks API
  const trackPromises = response.data.tracks.map(async playlistTrack => {
    try {
      const trackResponse = await gatewayClient.get<{
        id: string
        title: string
        duration_seconds: number
        cover_url: string
        artist_ids: string[] // Gateway returns artist_ids, not artists
        genre?: string
      }>(`/api/v1/tracks/${playlistTrack.track_id}`)

      const track = trackResponse.data

      // Получаем информацию об артисте через artists-service (как в fetchTrackDetail)
      let artist = { id: '', name: 'Unknown' }
      if (track.artist_ids && track.artist_ids.length > 0) {
        try {
          const artistResponse = await gatewayClient.get<{
            id: string
            name: string
          }>(`/api/v1/artists/${track.artist_ids[0]}`)
          artist = {
            id: artistResponse.data.id,
            name: artistResponse.data.name,
          }
        } catch (error) {
          console.warn(`Failed to fetch artist ${track.artist_ids[0]}:`, error)
          // Используем ID без имени, если не удалось получить информацию об артисте
          artist = { id: track.artist_ids[0], name: 'Unknown' }
        }
      }

      return {
        id: track.id,
        title: track.title,
        artist: artist,
        coverUrl: track.cover_url || DEFAULT_COVER_URL,
        duration: track.duration_seconds,
        liked: false, // Gateway doesn't provide like status
        stream: undefined, // Gateway doesn't provide stream info
      } satisfies Track
    } catch (error) {
      console.error(`Failed to fetch track ${playlistTrack.track_id}:`, error)
      // Возвращаем минимальную информацию о треке
      return {
        id: playlistTrack.track_id,
        title: 'Unknown Track',
        artist: { id: '', name: 'Unknown' },
        coverUrl: DEFAULT_COVER_URL,
        duration: 0,
        liked: false,
        stream: undefined,
      } satisfies Track
    }
  })

  const tracks = await Promise.all(trackPromises)
  // Сортируем по позиции
  return tracks.sort((a, b) => {
    const posA = response.data.tracks.find(t => t.track_id === a.id)?.position ?? 0
    const posB = response.data.tracks.find(t => t.track_id === b.id)?.position ?? 0
    return posA - posB
  })
}

export const createPlaylist = async (payload: {
  title: string
  description?: string
  isPrivate?: boolean
  genres?: string[]
}) => {
  const response = await gatewayClient.post<GatewayCreatePlaylistResponse>('/api/v1/playlists', {
    name: payload.title,
    description: payload.description || '',
    is_private: payload.isPrivate || false,
  })
  return fetchPlaylistDetail(response.data.playlist_id)
}

export const addTracksToPlaylist = async (playlistId: string, trackIds: string[]) => {
  // Gateway accepts one track_id at a time, so we need to add them sequentially
  const results = await Promise.allSettled(
    trackIds.map(trackId =>
      gatewayClient.post<{ success: boolean }>(`/api/v1/playlists/${playlistId}/tracks`, {
        track_id: trackId,
      })
    )
  )

  // Проверяем результаты
  const failed = results.filter(r => r.status === 'rejected')
  if (failed.length > 0) {
    console.error('Some tracks failed to add:', failed)
    throw new Error(`Failed to add ${failed.length} track(s) to playlist`)
  }

  return {
    playlistId,
    added: trackIds.map((trackId, index) => ({
      trackId,
      position: index, // Примерная позиция, реальная будет определена сервером
    })),
  }
}

export const fetchPlaylists = async (params: { filter?: string; limit?: number }) => {
  // Use gateway for user playlists (GET /api/v1/playlists returns current user's playlists)
  const response = await gatewayClient.get<GatewayPlaylistsResponse>('/api/v1/playlists', {
    params: {
      limit: params.limit,
      offset: 0,
    },
  })
  return response.data.playlists.map(mapGatewayPlaylist)
}
