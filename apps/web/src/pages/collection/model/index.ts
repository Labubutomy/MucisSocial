import { combine, createEffect, createEvent, createStore, sample } from 'effector'
import type { PlaylistDetail } from '@entities/playlist'
import type { Track } from '@entities/track'
import { routes } from '@shared/router'
import { fetchPlaylistDetail, fetchPlaylistTracks } from '@entities/playlist/api'
import { toggleTrackLike } from '@entities/track/api'

export const collectionLikeToggled = createEvent()
export const trackLikeToggled = createEvent<string>()

const fetchCollectionFx = createEffect(async ({ playlistId }: { playlistId: string }) =>
  fetchPlaylistDetail(playlistId)
)

const fetchCollectionTracksFx = createEffect(async ({ playlistId }: { playlistId: string }) =>
  fetchPlaylistTracks(playlistId)
)

const toggleLikeFx = createEffect(
  async ({ trackId, isLiked }: { trackId: string; isLiked: boolean }) =>
    toggleTrackLike(trackId, isLiked)
)

const $collectionBase = createStore<PlaylistDetail | null>(null).on(
  fetchCollectionFx.doneData,
  (_, detail) => detail
)

export const $collectionLiked = createStore(false)
  .on(fetchCollectionFx.doneData, (_, detail) => detail.liked ?? false)
  .on(collectionLikeToggled, liked => !liked)

export const $collection = combine($collectionBase, $collectionLiked, (collection, liked) => {
  if (!collection) return null
  return { ...collection, liked }
})

const $tracksBase = createStore<Track[]>([]).on(
  fetchCollectionTracksFx.doneData,
  (_, tracks) => tracks
)

export const $trackLikes = createStore<Record<string, boolean>>({})
  .on(fetchCollectionTracksFx.doneData, (state, tracks) => {
    const next = { ...state }
    tracks.forEach(track => {
      next[track.id] = track.liked ?? false
    })
    return next
  })
  .on(toggleLikeFx.doneData, (state, { trackId, isLiked }) => ({
    ...state,
    [trackId]: isLiked,
  }))

export const $tracks = combine($tracksBase, $trackLikes, (tracks, likes) =>
  tracks.map(track => ({
    ...track,
    liked: likes[track.id] ?? track.liked ?? false,
  }))
)

sample({
  clock: routes.collection.opened,
  fn: ({ params }) => ({ playlistId: params.collectionId }),
  target: [fetchCollectionFx, fetchCollectionTracksFx],
})

sample({
  clock: trackLikeToggled,
  source: $trackLikes,
  fn: (likes, trackId) => ({
    trackId,
    isLiked: !(likes[trackId] ?? false),
  }),
  target: toggleLikeFx,
})
