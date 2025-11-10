import { createApiClient } from '@shared/api/client'
import { API_CONFIG } from '@shared/config/api'

const client = createApiClient(API_CONFIG.cdn)

interface StreamVariantDto {
  bitrate: number
  url: string
}

interface StreamResponseDto {
  master_url: string
  variants: StreamVariantDto[]
  expires_in: number
}

export interface StreamMetadata {
  masterUrl: string
  variants: StreamVariantDto[]
  expiresIn: number
}

export const fetchStreamMetadata = async (params: {
  trackId: string
  artistId: string
  bitrates?: number[]
}) => {
  const { trackId, artistId, bitrates } = params

  const response = await client.get<StreamResponseDto>(`/api/stream/${trackId}`, {
    params: {
      artist_id: artistId,
      available_bitrates: bitrates?.length ? bitrates.join(',') : undefined,
    },
  })

  const data = response.data
  return {
    masterUrl: data.master_url,
    variants: data.variants,
    expiresIn: data.expires_in,
  } satisfies StreamMetadata
}

export const refreshStreamMetadata = async (payload: {
  trackId: string
  artistId: string
  bitrates?: number[]
}) => {
  const response = await client.post<StreamResponseDto>('/api/stream/refresh', {
    track_id: payload.trackId,
    artist_id: payload.artistId,
    available_bitrates: payload.bitrates,
  })

  const data = response.data
  return {
    masterUrl: data.master_url,
    variants: data.variants,
    expiresIn: data.expires_in,
  } satisfies StreamMetadata
}
