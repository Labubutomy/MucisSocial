import { createApiClient } from '@shared/api/client'
import type { SearchHistoryItem } from '@features/search'
import type { SearchResult } from '@widgets/search'
import type { Track } from '@entities/track'
import type { Artist } from '@entities/artist'
import type { PlaylistSummary } from '@entities/playlist'

const client = createApiClient('http://localhost:8100')

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
  items: Array<{
    id: string
    name: string
    avatarUrl?: string
    genres?: string[]
  }>
}

interface PlaylistsResponse {
  filter: string
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
    avatarUrl: item.avatarUrl,
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
  const response = await client.get<TrendingResponse>('/tracks/search/trending')
  return response.data.items.map((item, index) => ({
    id: `trend-${index}`,
    label: item.query,
  }))
}

export const fetchSearchHistory = async () => {
  const response = await client.get<HistoryResponse>('/me/search-history')
  return response.data.items
}

export const addSearchHistoryEntry = async (query: string) => {
  const response = await client.post<SearchHistoryItem>('/me/search-history', { query })
  return response.data
}

export const fetchSearchResults = async (query: string) => {
  const [tracksResponse, artistsResponse, playlistsResponse] = await Promise.all([
    client.get<TrackSearchResponse>('/tracks/search', {
      params: { q: query, limit: 20 },
    }),
    client.get<ArtistSearchResponse>('/artists/search', {
      params: { q: query },
    }),
    client.get<PlaylistsResponse>('/playlists', {
      params: { filter: 'curated', limit: 24 },
    }),
  ])

  const lowered = query.trim().toLowerCase()

  const trackResults = tracksResponse.data.items.map(mapTrack)
  const artistResults = artistsResponse.data.items.map(mapArtist)
  const playlistResults = playlistsResponse.data.items
    .filter(item => {
      if (!lowered) return true
      const titleMatch = item.title.toLowerCase().includes(lowered)
      const descriptionMatch = item.description?.toLowerCase().includes(lowered)
      return titleMatch || Boolean(descriptionMatch)
    })
    .map(mapPlaylist)

  return [...trackResults, ...artistResults, ...playlistResults]
}
