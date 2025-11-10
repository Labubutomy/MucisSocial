import { createApiClient } from '@shared/api/client'
import { API_CONFIG } from '@shared/config/api'
import type { Track } from './model/types'

// Use gateway for tracks endpoints that are available
const gatewayClient = createApiClient(API_CONFIG.gateway)
// Use mock-api for endpoints not yet in gateway (like)
const mockClient = createApiClient(API_CONFIG.mockApi)

// Default cover URL for tracks
const DEFAULT_COVER_URL =
  'https://mir-s3-cdn-cf.behance.net/projects/202/e2ba0e187042211.Y3JvcCw4MDgsNjMyLDAsMA.png'

// Gateway/Tracks service response format
interface GatewayTrackResponse {
  id: string
  title: string
  duration_seconds: number
  cover_url: string
  artist_ids: string[] // Массив ID артистов (информация об артистах хранится в artists-service)
  genre?: string
  audio_url?: string
  status?: string
}

// Gateway Artist response format
interface GatewayArtistResponse {
  id: string
  name: string
  avatar_url?: string
  genres?: string[]
}

interface GatewayTracksResponse {
  tracks: GatewayTrackResponse[]
  limit: number
  offset: number
}

// Gateway track detail response (same as GatewayTrackResponse)
type GatewayTrackDetailResponse = GatewayTrackResponse

interface ToggleLikeResponse {
  trackId: string
  isLiked: boolean
  likedAt: string | null
}

// Map gateway track response to Track (использует первый artist_id, если есть)
const mapGatewayTrack = (track: GatewayTrackResponse): Track => ({
  id: track.id,
  title: track.title,
  artist:
    track.artist_ids && track.artist_ids.length > 0
      ? { id: track.artist_ids[0], name: 'Unknown' } // Имя будет получено отдельно при необходимости
      : { id: '', name: 'Unknown' },
  coverUrl: track.cover_url || DEFAULT_COVER_URL,
  duration: track.duration_seconds,
  liked: false, // Gateway doesn't provide like status
  stream: undefined, // Gateway doesn't provide stream info
})

export const fetchTracks = async (params: { filter?: string; limit?: number }) => {
  const response = await gatewayClient.get<GatewayTracksResponse>('/api/v1/tracks', {
    params: {
      limit: params.limit,
      offset: 0,
    },
  })

  // Загружаем имена артистов для всех треков
  const tracksWithArtists = await Promise.all(
    response.data.tracks.map(async track => {
      const mappedTrack = mapGatewayTrack(track)

      // Загружаем имя артиста, если есть artist_id
      if (mappedTrack.artist.id) {
        try {
          const artistResponse = await gatewayClient.get<GatewayArtistResponse>(
            `/api/v1/artists/${mappedTrack.artist.id}`
          )
          mappedTrack.artist.name = artistResponse.data.name
        } catch (error) {
          console.warn(`Failed to fetch artist ${mappedTrack.artist.id}:`, error)
          // Оставляем 'Unknown', если не удалось загрузить
        }
      }

      return mappedTrack
    })
  )

  return tracksWithArtists
}

export const fetchTrackDetail = async (trackId: string) => {
  try {
    const response = await gatewayClient.get<GatewayTrackDetailResponse>(
      `/api/v1/tracks/${trackId}`
    )
    const track = response.data

    // Получаем информацию об артисте через artists-service
    let artist = { id: '', name: 'Unknown' }
    if (track.artist_ids && track.artist_ids.length > 0) {
      try {
        const artistResponse = await gatewayClient.get<GatewayArtistResponse>(
          `/api/v1/artists/${track.artist_ids[0]}`
        )
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
      album: {
        id: `album-${track.id}`,
        title: 'Неизвестный альбом',
      },
      coverUrl: track.cover_url || DEFAULT_COVER_URL,
      duration: track.duration_seconds,
      credits: [],
      bpm: 0,
      liked: false,
      progress: 0,
      stream: undefined,
    }
  } catch (error) {
    // Не делаем fallback на mock-api, чтобы избежать путаницы с данными
    console.error('Failed to fetch track detail from gateway:', error)
    throw error
  }
}

export const fetchTrackRecommendations = async () => {
  // Временно возвращаем список всех треков вместо рекомендаций
  const response = await gatewayClient.get<GatewayTracksResponse>('/api/v1/tracks', {
    params: {
      limit: 12,
      offset: 0,
    },
  })

  // Загружаем имена артистов для всех треков
  const tracksWithArtists = await Promise.all(
    response.data.tracks.map(async track => {
      const mappedTrack = mapGatewayTrack(track)

      // Загружаем имя артиста, если есть artist_id
      if (mappedTrack.artist.id) {
        try {
          const artistResponse = await gatewayClient.get<GatewayArtistResponse>(
            `/api/v1/artists/${mappedTrack.artist.id}`
          )
          mappedTrack.artist.name = artistResponse.data.name
        } catch (error) {
          console.warn(`Failed to fetch artist ${mappedTrack.artist.id}:`, error)
          // Оставляем 'Unknown', если не удалось загрузить
        }
      }

      return mappedTrack
    })
  )

  return tracksWithArtists
}

export const toggleTrackLike = async (trackId: string, isLiked: boolean) => {
  // Not available in gateway, use mock-api
  const response = await mockClient.post<ToggleLikeResponse>(`/api/v1/tracks/${trackId}/like`, {
    isLiked,
  })
  return response.data
}

// Gateway search response format
interface GatewaySearchTracksResponse {
  query: string
  items: GatewayTrackResponse[]
  limit: number
  offset: number
}

export const searchTracks = async (query: string, limit = 20) => {
  // Use gateway for track search
  const response = await gatewayClient.get<GatewaySearchTracksResponse>('/api/v1/tracks/search', {
    params: {
      q: query,
      limit,
      offset: 0,
    },
  })
  return response.data.items.map(mapGatewayTrack)
}
