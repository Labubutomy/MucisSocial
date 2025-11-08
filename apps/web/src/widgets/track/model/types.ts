import type { Track } from '@entities/track'

export interface TrackDetail extends Track {
  album: {
    id: string
    title: string
  }
  progress?: number
}
