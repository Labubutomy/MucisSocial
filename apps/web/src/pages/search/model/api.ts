import { createApiClient } from '@shared/api/client'
import type { SearchHistoryItem } from '@features/search'
import type { SearchResult } from '@widgets/search'
import type { Track } from '@entities/track'
import type { Artist } from '@entities/artist'
import type { PlaylistSummary } from '@entities/playlist'

import { API_CONFIG } from '@shared/config/api'

// Use gateway for artists/search, tracks/search and me/search-history, mock-api for playlists
const gatewayClient = createApiClient(API_CONFIG.gateway)
const mockClient = createApiClient(API_CONFIG.mockApi)

interface TrendingResponse {
  items: Array<{ query: string }>
}

interface HistoryResponse {
  items: SearchHistoryItem[]
}

// Gateway track search response format
interface GatewayTrackSearchResponse {
  query: string
  items: Array<{
    id: string
    title: string
    duration_seconds: number
    cover_url: string
    artist_ids: string[] // Массив ID артистов
    genre?: string
  }>
  limit: number
  offset: number
}

interface ArtistSearchResponse {
  query?: string
  items: Array<{
    id: string
    name: string
    avatar_url?: string
    genres?: string[]
  }>
}

interface PlaylistsResponse {
  query?: string
  filter?: string
  items: Array<{
    id: string
    title: string
    coverUrl: string
    itemsCount: number
    description?: string
  }>
}

// Map gateway track to SearchResult
const mapGatewayTrack = (item: GatewayTrackSearchResponse['items'][number]): SearchResult => ({
  type: 'track',
  data: {
    id: item.id,
    title: item.title,
    artist:
      item.artist_ids && item.artist_ids.length > 0
        ? { id: item.artist_ids[0], name: 'Unknown' } // Имя будет получено отдельно при необходимости
        : { id: '', name: 'Unknown' },
    coverUrl: item.cover_url || '',
    duration: item.duration_seconds,
    liked: false, // Gateway doesn't provide like status
    stream: undefined, // Gateway doesn't provide stream info
  } satisfies Track,
})

const mapArtist = (item: ArtistSearchResponse['items'][number]): SearchResult => ({
  type: 'artist',
  data: {
    id: item.id,
    name: item.name,
    avatarUrl: item.avatar_url,
    genres: item.genres,
  } satisfies Artist,
})

const mapPlaylist = (item: PlaylistsResponse['items'][number]): SearchResult => ({
  type: 'playlist',
  data: {
    id: item.id,
    title: item.title,
    coverUrl: item.coverUrl,
    itemsCount: item.itemsCount,
    description: item.description,
  } satisfies PlaylistSummary,
})

export const fetchTrendingQueries = async () => {
  // Not available in gateway, use mock-api
  const response = await mockClient.get<TrendingResponse>('/api/v1/tracks/search/trending')
  return response.data.items.map((item, index) => ({
    id: `trend-${index}`,
    label: item.query,
  }))
}

export const fetchSearchHistory = async () => {
  const response = await gatewayClient.get<HistoryResponse>('/api/v1/me/search-history', {
    params: { limit: 5 },
  })
  return response.data.items
}

export const addSearchHistoryEntry = async (query: string) => {
  const response = await gatewayClient.post<{ item: SearchHistoryItem }>(
    '/api/v1/me/search-history',
    { query }
  )
  return response.data.item
}

export const clearSearchHistory = async () => {
  await gatewayClient.delete<{ success: boolean }>('/api/v1/me/search-history')
}

export const fetchSearchResults = async (query: string) => {
  // Execute all search requests in parallel, but handle errors gracefully
  const results = await Promise.allSettled([
    gatewayClient
      .get<GatewayTrackSearchResponse>('/api/v1/tracks/search', {
        params: { q: query, limit: 20 },
      })
      .catch(() => null),
    gatewayClient
      .get<ArtistSearchResponse>('/api/v1/artists/search', {
        params: { q: query },
      })
      .catch(() => null),
    mockClient
      .get<PlaylistsResponse>('/api/v1/playlists/search', {
        params: { q: query, limit: 20 },
      })
      .catch(() => null),
  ])

  // Process tracks response (ignore if failed)
  let trackResults: SearchResult[] = []
  if (results[0].status === 'fulfilled' && results[0].value?.data) {
    trackResults = (results[0].value.data.items || []).map(mapGatewayTrack)
  }

  // Process artists response (ignore if failed)
  let artistResults: SearchResult[] = []
  if (results[1].status === 'fulfilled' && results[1].value?.data) {
    artistResults = (results[1].value.data.items || []).map(mapArtist)
  }

  // Process playlists response (ignore if failed)
  let playlistResults: SearchResult[] = []
  if (results[2].status === 'fulfilled' && results[2].value?.data) {
    playlistResults = (results[2].value.data.items || []).map(mapPlaylist)
  }

  // Combine all results and limit to last 5
  const allResults = [...trackResults, ...artistResults, ...playlistResults]
  return allResults.slice(-5)
}
