import type { UserProfile } from '@entities/user'

export const userProfile: UserProfile = {
  id: 'user-ava',
  username: 'ava.wave',
  avatarUrl: 'https://images.unsplash.com/photo-1524504388940-b1c1722653e1?q=80&w=400',
  musicTasteSummary: {
    topGenres: ['Синтвейв', 'Дрим-поп', 'Инди-электроника', 'Гиперпоп'],
    topArtists: ['Aviana', 'Kyro', 'Luna Wave', 'Solaria'],
  },
}
