import { createEffect, createStore, sample } from 'effector'
import type { PlaylistSummary } from '@entities/playlist'
import { fetchMyPlaylists } from '@entities/user/api'
import { routes } from '@shared/router'

export const fetchMyPlaylistsFx = createEffect(fetchMyPlaylists)

export const $myPlaylists = createStore<PlaylistSummary[]>([]).on(
  fetchMyPlaylistsFx.doneData,
  (_, items) => items
)

sample({
  clock: routes.profile.opened,
  fn: () => undefined,
  target: fetchMyPlaylistsFx,
})
