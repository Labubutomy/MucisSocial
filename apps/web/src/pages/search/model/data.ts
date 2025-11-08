import type { SearchHistoryItem } from '@features/search'
import type { Artist } from '@entities/artist'
import type { PlaylistSummary } from '@entities/playlist'
import type { Track } from '@entities/track'
import type { SearchResult } from '@widgets/search'

export const searchHistory: SearchHistoryItem[] = [
  { id: '1', query: 'синтвейв полночь', createdAt: new Date().toISOString() },
  { id: '2', query: 'лоуфай для концентрации', createdAt: new Date().toISOString() },
  { id: '3', query: 'гиперпоп энергия', createdAt: new Date().toISOString() },
]

export const trendingQueries: { id: string; label: string }[] = [
  { id: 'trend-1', label: 'Эмбиент рассвет' },
  { id: 'trend-2', label: 'Неоновый фанк' },
  { id: 'trend-3', label: 'Фьючер-гараж' },
  { id: 'trend-4', label: 'Дрим-поп' },
  { id: 'trend-5', label: 'Органик-хаус' },
]

const resultsTracks: Track[] = [
  {
    id: 'search-track-1',
    title: 'City Halo',
    artist: { id: 'artist-shea', name: 'Shea Monarch' },
    coverUrl: 'https://images.unsplash.com/photo-1506619267508-51b1f1d95029?q=80&w=600',
    duration: 205,
  },
  {
    id: 'search-track-2',
    title: 'Glass Sky',
    artist: { id: 'artist-aleo', name: 'Aleo' },
    coverUrl: 'https://images.unsplash.com/photo-1452857297128-d9c29adba80b?q=80&w=600',
    duration: 218,
  },
]

const resultsArtists: Artist[] = [
  {
    id: 'artist-shea',
    name: 'Shea Monarch',
    avatarUrl: 'https://images.unsplash.com/photo-1524504388940-b1c1722653e1?q=80&w=200',
    genres: ['Синтвейв', 'Электро-поп'],
  },
  {
    id: 'artist-aleo',
    name: 'Aleo',
    avatarUrl: 'https://images.unsplash.com/photo-1544723795-3fb6469f5b39?q=80&w=200',
    genres: ['Гиперпоп', 'Инди'],
  },
]

const resultsPlaylists: PlaylistSummary[] = [
  {
    id: 'playlist-focus',
    title: 'Neon Focus',
    coverUrl: 'https://images.unsplash.com/photo-1511379938547-c1f69419868d?q=80&w=600',
    itemsCount: 42,
    description: 'Свечение синтвейва для глубокого погружения в работу.',
  },
  {
    id: 'playlist-late-night',
    title: 'Late Night Lights',
    coverUrl: 'https://images.unsplash.com/photo-1526470608268-f674ce90ebd4?q=80&w=600',
    itemsCount: 28,
    description: 'Фьючер R&B и неоновые вибрации поздней ночи.',
  },
]

export const searchResults: SearchResult[] = [
  ...resultsTracks.map(track => ({ type: 'track', data: track }) as const),
  ...resultsArtists.map(artist => ({ type: 'artist', data: artist }) as const),
  ...resultsPlaylists.map(playlist => ({ type: 'playlist', data: playlist }) as const),
]
