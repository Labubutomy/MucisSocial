import { createApiClient } from '@shared/api/client'
import type { PlaylistDetail, PlaylistSummary } from './model/types'
import type { Track } from '@entities/track'

const client = createApiClient('http://localhost:8100')

interface PlaylistResponse extends PlaylistSummary {
  description: string
  owner: {
    id: string
    name: string
  }
  type: string
  isPrivate?: boolean
  totalDurationSec?: number
  liked?: boolean
}

interface PlaylistTracksResponse {
  items: Array<{
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
  }>
}

interface PlaylistsResponse {
  filter: string
  items: PlaylistSummary[]
}

interface AddTracksResponse {
  playlistId: string
  added: Array<{
    trackId: string
    position: number
  }>
}

const mapTrack = (track: PlaylistTracksResponse['items'][number]): Track => ({
  id: track.id,
  title: track.title,
  artist: track.artist,
  coverUrl: track.coverUrl,
  duration: track.durationSec,
  liked: track.isLiked ?? false,
  stream: track.stream?.hlsMasterUrl
    ? {
        masterUrl: track.stream.hlsMasterUrl,
        qualities: track.stream.quality ?? [],
      }
    : undefined,
})

export const fetchPlaylistDetail = async (playlistId: string) => {
  const response = await client.get<PlaylistResponse>(`/playlists/${playlistId}`)
  const data = response.data
  return {
    id: data.id,
    title: data.title,
    description: data.description,
    coverUrl: data.coverUrl,
    itemsCount: data.itemsCount,
    owner: data.owner,
    type: data.type === 'album' ? 'album' : 'playlist',
    liked: data.liked ?? false,
    totalDuration: data.totalDurationSec,
  } satisfies PlaylistDetail
}

export const fetchPlaylistTracks = async (playlistId: string) => {
  const response = await client.get<PlaylistTracksResponse>(`/playlists/${playlistId}/tracks`)
  return response.data.items.map(mapTrack)
}

export const createPlaylist = async (payload: {
  title: string
  description?: string
  isPrivate?: boolean
  genres?: string[]
}) => {
  const response = await client.post<PlaylistResponse>('/playlists', payload)
  return fetchPlaylistDetail(response.data.id)
}

export const addTracksToPlaylist = async (playlistId: string, trackIds: string[]) => {
  const response = await client.post<AddTracksResponse>(`/playlists/${playlistId}/tracks`, {
    trackIds,
  })
  return response.data
}

export const fetchPlaylists = async (params: { filter?: string; limit?: number }) => {
  const response = await client.get<PlaylistsResponse>('/playlists', { params })
  return response.data.items
}
