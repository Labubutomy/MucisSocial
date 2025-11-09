import { combine, createEffect, createEvent, createStore, sample } from 'effector'
import { createPlaylist } from '@entities/playlist/api'
import { routes } from '@shared/router'

export const titleChanged = createEvent<string>()
export const descriptionChanged = createEvent<string>()
export const genreToggled = createEvent<string>()
export const privacyToggled = createEvent()
export const formSubmitted = createEvent()

export const $title = createStore('').on(titleChanged, (_, value) => value)
export const $description = createStore('').on(descriptionChanged, (_, value) => value)

export const $genres = createStore<string[]>([]).on(genreToggled, (genres, genre) => {
  if (genres.includes(genre)) {
    return genres.filter(item => item !== genre)
  }
  return [...genres, genre]
})

export const $isPrivate = createStore(false).on(privacyToggled, value => !value)

export const $form = combine({
  title: $title,
  description: $description,
  genres: $genres,
  isPrivate: $isPrivate,
})

export const createPlaylistFx = createEffect(
  async (payload: { title: string; description?: string; genres: string[]; isPrivate: boolean }) =>
    createPlaylist(payload)
)

sample({
  clock: formSubmitted,
  source: $form,
  filter: form => Boolean(form.title.trim()),
  fn: form => ({
    title: form.title.trim(),
    description: form.description?.trim() || undefined,
    genres: form.genres,
    isPrivate: form.isPrivate,
  }),
  target: createPlaylistFx,
})

sample({
  clock: createPlaylistFx.doneData,
  fn: playlist => ({
    params: { playlistId: playlist.id },
    query: {},
  }),
  target: routes.playlistAddTracks.navigate,
})

$title.reset(createPlaylistFx.done)
$description.reset(createPlaylistFx.done)
$genres.reset(createPlaylistFx.done)
$isPrivate.reset(createPlaylistFx.done)
