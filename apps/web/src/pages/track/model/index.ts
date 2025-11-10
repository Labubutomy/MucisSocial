import { combine, createEffect, createEvent, createStore, sample } from 'effector'
import type { Track } from '@entities/track'
import type { TrackDetail } from '@widgets/track'
import { routes } from '@shared/router'
import { fetchTrackDetail, fetchTrackRecommendations, toggleTrackLike } from '@entities/track/api'

export const trackLikeToggled = createEvent<string>()

const fetchTrackDetailFx = createEffect(async ({ trackId }: { trackId: string }) => {
  return fetchTrackDetail(trackId)
})

const fetchRecommendationsFx = createEffect(async () => {
  return fetchTrackRecommendations()
})

const toggleLikeFx = createEffect(
  async ({ trackId, isLiked }: { trackId: string; isLiked: boolean }) =>
    toggleTrackLike(trackId, isLiked)
)

const $trackDetailBase = createStore<TrackDetail | null>(null).on(
  fetchTrackDetailFx.doneData,
  (_, detail) => detail
)

const $recommendedBase = createStore<Track[]>([]).on(
  fetchRecommendationsFx.doneData,
  (_, tracks) => tracks
)

export const $likes = createStore<Record<string, boolean>>({})
  .on(fetchTrackDetailFx.doneData, (state, detail) => ({
    ...state,
    [detail.id]: detail.liked ?? false,
  }))
  .on(fetchRecommendationsFx.doneData, (state, tracks) => {
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

export const $trackDetail = combine($trackDetailBase, $likes, (detail, likes) => {
  if (!detail) return null
  const liked = likes[detail.id] ?? detail.liked ?? false
  return {
    ...detail,
    liked,
  }
})

export const $recommendedTracks = combine($recommendedBase, $likes, (tracks, likes) =>
  tracks.map(
    track =>
      ({
        ...track,
        liked: likes[track.id] ?? track.liked ?? false,
      }) satisfies Track
  )
)

// Обновляем трек при открытии маршрута или изменении параметра trackId
sample({
  clock: [routes.track.opened, routes.track.updated],
  fn: ({ params }) => ({ trackId: params.trackId }),
  target: fetchTrackDetailFx,
})

sample({
  clock: routes.track.opened,
  target: fetchRecommendationsFx,
})

sample({
  clock: trackLikeToggled,
  source: $likes,
  fn: (likes, trackId) => ({
    trackId,
    isLiked: !(likes[trackId] ?? false),
  }),
  target: toggleLikeFx,
})
