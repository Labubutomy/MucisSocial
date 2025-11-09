import type { Artist } from '@entities/artist'
import type { PlaylistSummary } from '@entities/playlist'
import type { Track } from '@entities/track'

export type SearchResult =
  | { type: 'track'; data: Track }
  | { type: 'artist'; data: Artist }
  | { type: 'playlist'; data: PlaylistSummary }
