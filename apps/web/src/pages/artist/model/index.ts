import { createEffect, createStore, sample } from 'effector'
import { routes } from '@shared/router'
import { fetchArtist } from '@entities/artist/api'
import type { ArtistDetail } from '@entities/artist/api'
import { fetchTracksByArtist } from '@entities/track/api'
import type { Track } from '@entities/track'

export const $artistDetail = createStore<ArtistDetail | null>(null)
export const $artistTracks = createStore<Track[]>([])

export const loadArtistFx = createEffect(async (artistId: string) => {
  return await fetchArtist(artistId)
})

export const loadArtistTracksFx = createEffect(async (artistId: string) => {
  return await fetchTracksByArtist(artistId)
})

export const $artistPending = loadArtistFx.pending
export const $tracksPending = loadArtistTracksFx.pending

$artistDetail.on(loadArtistFx.doneData, (_, artist) => artist)
$artistTracks.on(loadArtistTracksFx.doneData, (_, tracks) => tracks)

// Загружаем артиста и треки при открытии страницы
sample({
  clock: routes.artist.$params,
  filter: (params): params is { artistId: string } => Boolean(params?.artistId),
  fn: params => params.artistId,
  target: [loadArtistFx, loadArtistTracksFx],
})

// Сбрасываем данные при закрытии страницы
sample({
  clock: routes.artist.closed,
  fn: () => null,
  target: $artistDetail,
})

sample({
  clock: routes.artist.closed,
  fn: () => [],
  target: $artistTracks,
})
