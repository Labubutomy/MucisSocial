import type { Track } from '@entities/track'

const createTrack = (data: Track): Track => data

export const homeFeed: Record<'trending' | 'popular' | 'new', Track[]> = {
  trending: [
    createTrack({
      id: 'nebula-night',
      title: 'Nebula Night',
      artist: { id: 'aviana', name: 'Aviana' },
      coverUrl: 'https://images.unsplash.com/photo-1453090927415-5f45085b65c0?q=80&w=600',
      duration: 214,
    }),
    createTrack({
      id: 'pulse-signal',
      title: 'Pulse Signal',
      artist: { id: 'kyro', name: 'Kyro' },
      coverUrl: 'https://images.unsplash.com/photo-1498050108023-c5249f4df085?q=80&w=600',
      duration: 202,
    }),
    createTrack({
      id: 'midnight-drive',
      title: 'Midnight Drive',
      artist: { id: 'luna-wave', name: 'Luna Wave' },
      coverUrl: 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?q=80&w=600',
      duration: 236,
    }),
    createTrack({
      id: 'aurora-echoes',
      title: 'Aurora Echoes',
      artist: { id: 'solaria', name: 'Solaria' },
      coverUrl: 'https://images.unsplash.com/photo-1511379938547-c1f69419868d?q=80&w=600',
      duration: 188,
    }),
  ],
  popular: [
    createTrack({
      id: 'neon-silhouette',
      title: 'Neon Silhouette',
      artist: { id: 'valerie-storm', name: 'Valerie Storm' },
      coverUrl: 'https://images.unsplash.com/photo-1497032205916-ac775f0649ae?q=80&w=600',
      duration: 240,
    }),
    createTrack({
      id: 'tidal-fragments',
      title: 'Tidal Fragments',
      artist: { id: 'atlas', name: 'Atlas' },
      coverUrl: 'https://images.unsplash.com/photo-1511671782779-c97d3d27a1d4?q=80&w=600',
      duration: 265,
    }),
    createTrack({
      id: 'afterglow',
      title: 'Afterglow',
      artist: { id: 'violet-city', name: 'Violet City' },
      coverUrl: 'https://images.unsplash.com/photo-1459749411175-04bf5292ceea?q=80&w=600',
      duration: 210,
    }),
    createTrack({
      id: 'spectrum-rush',
      title: 'Spectrum Rush',
      artist: { id: 'nexus', name: 'Nexus' },
      coverUrl: 'https://images.unsplash.com/photo-1546443046-ed1ce6ffd1ab?q=80&w=600',
      duration: 198,
    }),
  ],
  new: [
    createTrack({
      id: 'daybreak',
      title: 'Daybreak',
      artist: { id: 'hiro', name: 'Hiro' },
      coverUrl: 'https://images.unsplash.com/photo-1524679576720-d009c7e83ee0?q=80&w=600',
      duration: 186,
    }),
    createTrack({
      id: 'gravity',
      title: 'Gravity',
      artist: { id: 'moonrise', name: 'Moonrise' },
      coverUrl: 'https://images.unsplash.com/photo-1485579149621-3123dd979885?q=80&w=600',
      duration: 208,
    }),
    createTrack({
      id: 'luminous',
      title: 'Luminous',
      artist: { id: 'stellar-kid', name: 'Stellar Kid' },
      coverUrl: 'https://images.unsplash.com/photo-1487215078519-e21cc028cb29?q=80&w=600',
      duration: 192,
    }),
    createTrack({
      id: 'magnetic',
      title: 'Magnetic',
      artist: { id: 'echo-drive', name: 'Echo Drive' },
      coverUrl: 'https://images.unsplash.com/photo-1527596428170-1a96b1f1c3c4?q=80&w=600',
      duration: 234,
    }),
  ],
}
