import { createApiClient } from '@shared/api/client'
import type { Track } from './model/types'

const client = createApiClient('http://localhost:8100')

interface StreamResponsePayload {
  quality?: string[]
  hlsMasterUrl?: string
}

interface TrackResponse {
  id: string
  title: string
  durationSec: number
  coverUrl: string
  artist: {
    id: string
    name: string
  }
  isLiked?: boolean
  stream?: StreamResponsePayload
}

interface TracksResponse {
  filter: string
  items: TrackResponse[]
}

interface TrackDetailResponse extends TrackResponse {
  album?: {
    id: string
    title: string
    coverUrl?: string
    releasedAt?: string
  }
  credits?: string[]
  bpm?: number
  liked?: boolean
  progress?: number
}

interface ToggleLikeResponse {
  trackId: string
  isLiked: boolean
  likedAt: string | null
}

const mapStream = (stream?: StreamResponsePayload) => {
  if (!stream || !stream.hlsMasterUrl) {
    return undefined
  }

  return {
    masterUrl: stream.hlsMasterUrl,
    qualities: stream.quality ?? [],
  }
}

const mapTrack = (track: TrackResponse): Track => ({
  id: track.id,
  title: track.title,
  artist: track.artist,
  coverUrl: track.coverUrl,
  duration: track.durationSec,
  liked: track.isLiked ?? false,
  stream: mapStream(track.stream),
})

export const fetchTracks = async (params: { filter?: string; limit?: number }) => {
  const response = await client.get<TracksResponse>('/tracks', {
    params,
  })
  return response.data.items.map(mapTrack)
}

export const fetchTrackDetail = async (trackId: string) => {
  const response = await client.get<TrackDetailResponse>(`/tracks/${trackId}`)
  const track = response.data
  const album = track.album ?? {
    id: `album-${track.id}`,
    title: 'Неизвестный альбом',
  }
  return {
    id: track.id,
    title: track.title,
    artist: track.artist,
    album: {
      id: album.id,
      title: album.title,
    },
    coverUrl: track.coverUrl,
    duration: track.durationSec,
    credits: track.credits ?? [],
    bpm: track.bpm ?? 0,
    liked: track.liked ?? track.isLiked ?? false,
    progress: track.progress ?? 0,
    stream: mapStream(track.stream),
  }
}

export const fetchTrackRecommendations = async (trackId: string) => {
  const response = await client.get<{ trackId: string; items: TrackResponse[] }>(
    `/tracks/${trackId}/recommendations`,
    {
      params: { limit: 12 },
    }
  )

  return response.data.items.map(mapTrack)
}

export const toggleTrackLike = async (trackId: string, isLiked: boolean) => {
  const response = await client.post<ToggleLikeResponse>(`/tracks/${trackId}/like`, {
    isLiked,
  })
  return response.data
}

export const searchTracks = async (query: string, limit = 20) => {
  const response = await client.get<{ query: string; items: TrackResponse[] }>('/tracks/search', {
    params: {
      q: query,
      limit,
    },
  })
  return response.data.items.map(mapTrack)
}
