import { createApiClient } from '@shared/api/client'
import { API_CONFIG } from '@shared/config/api'
import type { Artist } from './model/types'

const gatewayClient = createApiClient(API_CONFIG.gateway)

export interface ArtistDetail extends Artist {
  followers?: number
  topTracks?: Array<{
    id: string
    title: string
    coverUrl?: string
  }>
}

export const searchArtists = async (query: string, limit = 20): Promise<Artist[]> => {
  const { data } = await gatewayClient.get<{
    query?: string
    items: Array<{
      id: string
      name: string
      avatar_url?: string
      genres?: string[]
    }>
  }>('/api/v1/artists/search', {
    params: { q: query, limit },
  })

  return data.items.map(item => ({
    id: item.id,
    name: item.name,
    avatarUrl: item.avatar_url,
    genres: item.genres,
  }))
}

export const fetchArtist = async (artistId: string): Promise<ArtistDetail> => {
  const { data } = await gatewayClient.get<{
    id: string
    name: string
    avatar_url?: string
    genres?: string[]
    followers?: number
    top_tracks?: Array<{
      id: string
      title: string
      cover_url?: string
    }>
  }>(`/api/v1/artists/${artistId}`)

  return {
    id: data.id,
    name: data.name,
    avatarUrl: data.avatar_url,
    genres: data.genres,
    followers: data.followers,
    topTracks: data.top_tracks?.map(track => ({
      id: track.id,
      title: track.title,
      coverUrl: track.cover_url,
    })),
  }
}
