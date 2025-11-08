import type { Track } from '@entities/track'
import type { TrackDetail } from '@widgets/track'

export const currentTrackDetail: TrackDetail = {
  id: 'nebula-night',
  title: 'Nebula Night',
  artist: { id: 'aviana', name: 'Aviana' },
  album: { id: 'album-nebula', title: 'Starlight Bloom' },
  coverUrl: 'https://images.unsplash.com/photo-1498050108023-c5249f4df085?q=80&w=900',
  duration: 214,
  progress: 0.32,
  liked: true,
}

export const recommendedTracks: Track[] = [
  {
    id: 'horizon-sway',
    title: 'Horizon Sway',
    artist: { id: 'lunaric', name: 'Lunaric' },
    coverUrl: 'https://images.unsplash.com/photo-1493225457124-a3eb161ffa5f?q=80&w=600',
    duration: 205,
  },
  {
    id: 'echo-parade',
    title: 'Echo Parade',
    artist: { id: 'kyro', name: 'Kyro' },
    coverUrl: 'https://images.unsplash.com/photo-1446057032654-9d8885db76c6?q=80&w=600',
    duration: 192,
  },
  {
    id: 'mystic-line',
    title: 'Mystic Line',
    artist: { id: 'solaria', name: 'Solaria' },
    coverUrl: 'https://images.unsplash.com/photo-1446185250204-f94591f7d702?q=80&w=600',
    duration: 238,
  },
]
