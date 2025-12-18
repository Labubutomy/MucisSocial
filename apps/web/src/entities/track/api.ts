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

export const fetchTracksByArtist = async (artistId: string, limit = 50) => {
  const response = await gatewayClient.get<GatewayTracksResponse>('/api/v1/tracks', {
    params: {
      artist_id: artistId,
      limit,
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

// Recommendation service response format
interface RecommendationsResponse {
  tracks: string[] // Array of track IDs
}

export interface RecommendationOptions {
  limit?: number
  excludeExplicit?: boolean
}

export const fetchTrackRecommendations = async (
  options?: RecommendationOptions
): Promise<Track[]> => {
  try {
    // Вызов Gateway API для получения рекомендаций
    const response = await gatewayClient.get<RecommendationsResponse>('/api/v1/recommendations', {
      params: {
        limit: options?.limit ?? 12,
        exclude_explicit: options?.excludeExplicit ?? false,
      },
    })

    const trackIds = response.data.tracks

    if (trackIds.length === 0) {
      return []
    }

    // Загружаем полную информацию о каждом треке
    const tracksWithArtists = await Promise.all(
      trackIds.map(async trackId => {
        try {
          const trackResponse = await gatewayClient.get<GatewayTrackDetailResponse>(
            `/api/v1/tracks/${trackId}`
          )
          const track = trackResponse.data

          // Получаем информацию об артисте
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
            }
          }

          const mappedTrack = mapGatewayTrack(track)
          mappedTrack.artist = artist

          return mappedTrack
        } catch (error) {
          console.warn(`Failed to fetch track ${trackId}:`, error)
          return null
        }
      })
    )

    // Фильтруем null значения (треки, которые не удалось загрузить)
    return tracksWithArtists.filter((track): track is Track => track !== null)
  } catch (error) {
    console.error('Failed to fetch recommendations:', error)
    // Возвращаем пустой массив при ошибке, чтобы не ломать UI
    return []
  }
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

export const searchTracks = async (query: string, limit = 20): Promise<Track[]> => {
  // Use gateway for track search
  const response = await gatewayClient.get<GatewaySearchTracksResponse>('/api/v1/tracks/search', {
    params: {
      q: query,
      limit,
      offset: 0,
    },
  })

  // Загружаем имена артистов для всех треков (как в fetchTracks)
  const tracksWithArtists = await Promise.all(
    response.data.items.map(async track => {
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

// Charts API response format
interface ChartTrack {
  track_id: string
  position: number
  play_count: number
}

interface ChartsResponse {
  period: string
  updated_at: string
  tracks: ChartTrack[]
}

// New releases API response format
interface NewReleasesResponse {
  tracks: string[] // Array of track IDs
  updated_at: string
}

/**
 * Fetches top charts for the current month
 * @param limit Maximum number of tracks to return (default: 50, max: 100)
 * @returns Array of tracks with chart positions
 */
export const fetchTopCharts = async (limit: number = 50): Promise<Track[]> => {
  try {
    const response = await gatewayClient.get<ChartsResponse>('/api/v1/charts/top', {
      params: {
        limit: Math.min(limit, 100), // Cap at 100
      },
    })

    const chartTracks = response.data.tracks
    if (chartTracks.length === 0) {
      return []
    }

    // Load full track details for each chart entry
    const tracksWithArtists = await Promise.all(
      chartTracks.map(async chartEntry => {
        try {
          const trackResponse = await gatewayClient.get<GatewayTrackDetailResponse>(
            `/api/v1/tracks/${chartEntry.track_id}`
          )
          const track = trackResponse.data

          // Get artist information
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
            }
          }

          const mappedTrack = mapGatewayTrack(track)
          mappedTrack.artist = artist

          return mappedTrack
        } catch (error) {
          console.warn(`Failed to fetch track ${chartEntry.track_id}:`, error)
          return null
        }
      })
    )

    return tracksWithArtists.filter((track): track is Track => track !== null)
  } catch (error) {
    console.error('Failed to fetch top charts:', error)
    return []
  }
}

/**
 * Fetches new releases (recently released tracks)
 * @param limit Maximum number of tracks to return (default: 50, max: 100)
 * @param days Number of days to look back (default: 30, max: 365)
 * @returns Array of tracks sorted by release date (newest first)
 */
export const fetchNewReleases = async (limit: number = 50, days: number = 30): Promise<Track[]> => {
  try {
    const response = await gatewayClient.get<NewReleasesResponse>('/api/v1/tracks/new', {
      params: {
        limit: Math.min(limit, 100), // Cap at 100
        days: Math.min(days, 365), // Cap at 365
      },
    })

    const trackIds = response.data.tracks
    if (trackIds.length === 0) {
      return []
    }

    // Load full track details for each new release
    const tracksWithArtists = await Promise.all(
      trackIds.map(async trackId => {
        try {
          const trackResponse = await gatewayClient.get<GatewayTrackDetailResponse>(
            `/api/v1/tracks/${trackId}`
          )
          const track = trackResponse.data

          // Get artist information
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
            }
          }

          const mappedTrack = mapGatewayTrack(track)
          mappedTrack.artist = artist

          return mappedTrack
        } catch (error) {
          console.warn(`Failed to fetch track ${trackId}:`, error)
          return null
        }
      })
    )

    return tracksWithArtists.filter((track): track is Track => track !== null)
  } catch (error) {
    console.error('Failed to fetch new releases:', error)
    return []
  }
}
