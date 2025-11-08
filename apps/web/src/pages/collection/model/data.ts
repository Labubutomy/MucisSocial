import type { PlaylistDetail } from '@entities/playlist'
import type { Track } from '@entities/track'

export const currentCollection: PlaylistDetail = {
  id: 'uplink',
  title: 'Uplink Sessions',
  description: 'Кураторское путешествие по неону даунтемпо и атмосферной электронике.',
  coverUrl: 'https://images.unsplash.com/photo-1485579149621-3123dd979885?q=80&w=900',
  itemsCount: 35,
  owner: {
    id: 'ava.wave',
    name: 'ava.wave',
  },
  type: 'playlist',
  liked: false,
  totalDuration: 8420,
}

export const collectionTracks: Track[] = [
  {
    id: 'uplink-1',
    title: 'Luminous Drift',
    artist: { id: 'solaria', name: 'Solaria' },
    coverUrl: 'https://images.unsplash.com/photo-1498050108023-c5249f4df085?q=80&w=600',
    duration: 230,
  },
  {
    id: 'uplink-2',
    title: 'Electric Haze',
    artist: { id: 'kyro', name: 'Kyro' },
    coverUrl: 'https://images.unsplash.com/photo-1526470608268-f674ce90ebd4?q=80&w=600',
    duration: 214,
  },
  {
    id: 'uplink-3',
    title: 'Silver Lines',
    artist: { id: 'luna-wave', name: 'Luna Wave' },
    coverUrl: 'https://images.unsplash.com/photo-1459749411175-04bf5292ceea?q=80&w=600',
    duration: 248,
  },
  {
    id: 'uplink-4',
    title: 'Magnetic Bloom',
    artist: { id: 'aviana', name: 'Aviana' },
    coverUrl: 'https://images.unsplash.com/photo-1464047736614-af63643285bf?q=80&w=600',
    duration: 198,
  },
]
