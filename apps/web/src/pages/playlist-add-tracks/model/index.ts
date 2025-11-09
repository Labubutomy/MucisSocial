import { createEffect, createEvent, createStore, sample } from 'effector'
import type { Track } from '@entities/track'
import { fetchTracks } from '@entities/track/api'
import { addTracksToPlaylist } from '@entities/playlist/api'
import { routes } from '@shared/router'

export const searchChanged = createEvent<string>()
export const trackSelected = createEvent<Track>()
export const selectionCleared = createEvent()
export const saveRequested = createEvent<{ playlistId: string }>()

export const $search = createStore('').on(searchChanged, (_, query) => query)

export const $selectedTracks = createStore<Track[]>([])
  .on(trackSelected, (tracks, track) => {
    if (tracks.some(item => item.id === track.id)) {
      return tracks
    }
    return [...tracks, track]
  })
  .reset(selectionCleared)

export const fetchSuggestionsFx = createEffect(async () =>
  fetchTracks({ filter: 'new', limit: 20 })
)

export const $suggestions = createStore<Track[]>([]).on(
  fetchSuggestionsFx.doneData,
  (_, tracks) => tracks
)

const saveTracksFx = createEffect(
  async ({ playlistId, trackIds }: { playlistId: string; trackIds: string[] }) =>
    addTracksToPlaylist(playlistId, trackIds)
)

sample({
  clock: routes.playlistAddTracks.opened,
  target: fetchSuggestionsFx,
})

sample({
  clock: saveRequested,
  source: $selectedTracks,
  filter: (tracks, { playlistId }) => tracks.length > 0 && Boolean(playlistId),
  fn: (tracks, { playlistId }) => ({
    playlistId,
    trackIds: tracks.map(track => track.id),
  }),
  target: saveTracksFx,
})

$selectedTracks.reset(saveTracksFx.done)

export { saveTracksFx }
