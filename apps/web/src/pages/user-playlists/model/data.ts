import type { PlaylistSummary } from '@entities/playlist'

export const userPlaylists: PlaylistSummary[] = [
  {
    id: 'uplink',
    title: 'Uplink Sessions',
    coverUrl: 'https://images.unsplash.com/photo-1485579149621-3123dd979885?q=80&w=600',
    itemsCount: 35,
    description: 'Даунтемпо и синтовые текстуры для поздней ночи.',
  },
  {
    id: 'pulse',
    title: 'Pulse Run',
    coverUrl: 'https://images.unsplash.com/photo-1511379938547-c1f69419868d?q=80&w=600',
    itemsCount: 24,
    description: 'Высокоэнергетическое электронное топливо для тренировок.',
  },
  {
    id: 'oceanic',
    title: 'Oceanic Glow',
    coverUrl: 'https://images.unsplash.com/photo-1459749411175-04bf5292ceea?q=80&w=600',
    itemsCount: 18,
    description: 'Плавные эмбиент-оттенки для концентрации и размышлений.',
  },
  {
    id: 'hush',
    title: 'Hush Bloom',
    coverUrl: 'https://images.unsplash.com/photo-1464047736614-af63643285bf?q=80&w=600',
    itemsCount: 29,
    description: 'Нежный вокал и воздушные биты для ночных прослушиваний.',
  },
]
