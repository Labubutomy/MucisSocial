import { createEffect, createStore, sample } from 'effector'
import type { PlaylistSummary } from '@entities/playlist'
import { fetchMyPlaylists, fetchUserTaste } from '@entities/user/api'
import { routes } from '@shared/router'
import { $user } from '@features/auth/model'

export const fetchMyPlaylistsFx = createEffect(fetchMyPlaylists)

export const $myPlaylists = createStore<PlaylistSummary[]>([]).on(
  fetchMyPlaylistsFx.doneData,
  (_, items) => items
)

export const fetchUserTasteFx = createEffect(async (userId: string) => {
  return fetchUserTaste(userId)
})

export const $userTaste = createStore<{
  topGenres: string[]
  topArtists: string[]
} | null>(null)
  .on(fetchUserTasteFx.doneData, (_, data) => data)
  .reset(fetchUserTasteFx.fail)

sample({
  clock: routes.profile.opened,
  fn: () => undefined,
  target: fetchMyPlaylistsFx,
})

// Load user taste when profile page opens
sample({
  clock: routes.profile.opened,
  source: $user,
  filter: (user): user is NonNullable<typeof user> => user !== null,
  fn: user => {
    // TypeScript guard: user is guaranteed to be non-null after filter
    if (!user) return ''
    return user.id
  },
  target: fetchUserTasteFx,
})
