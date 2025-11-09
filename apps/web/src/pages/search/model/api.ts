import { createApiClient } from '@shared/api/client'
import type { SearchHistoryItem } from '@features/search'
import type { SearchResult } from '@widgets/search'
import type { Track } from '@entities/track'
import type { Artist } from '@entities/artist'
import type { PlaylistSummary } from '@entities/playlist'

import { API_CONFIG } from '@shared/config/api'

// Use gateway for artists/search and me/search-history, mock-api for tracks
const gatewayClient = createApiClient(API_CONFIG.gateway)
const mockClient = createApiClient(API_CONFIG.mockApi)

interface TrendingResponse {
  items: Array<{ query: string }>
}

interface HistoryResponse {
  items: SearchHistoryItem[]
}

interface TrackSearchResponse {
  query: string
  items: Array<{
    type: 'track'
    data: {
      id: string
      title: string
      durationSec: number
      coverUrl: string
      artist: {
        id: string
        name: string
      }
      isLiked?: boolean
      stream?: {
        quality?: string[]
        hlsMasterUrl?: string
      }
    }
  }>
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

const mapTrack = (item: TrackSearchResponse['items'][number]): SearchResult => ({
  type: 'track',
  data: {
    id: item.data.id,
    title: item.data.title,
    artist: item.data.artist,
    coverUrl: item.data.coverUrl,
    duration: item.data.durationSec,
    liked: item.data.isLiked ?? false,
    stream: item.data.stream?.hlsMasterUrl
      ? {
          masterUrl: item.data.stream.hlsMasterUrl,
          qualities: item.data.stream.quality ?? [],
        }
      : undefined,
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
    mockClient
      .get<TrackSearchResponse>('/api/v1/tracks/search', {
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
    trackResults = (results[0].value.data.items || []).map(mapTrack)
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
