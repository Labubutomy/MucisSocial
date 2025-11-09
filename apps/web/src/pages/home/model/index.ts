import { combine, createEffect, createEvent, createStore, sample } from 'effector'
import type { Track } from '@entities/track'
import { routes } from '@shared/router'
import { fetchTracks, toggleTrackLike } from '@entities/track/api'

export type FeedTab = 'trending' | 'popular' | 'new'

export const tabChanged = createEvent<FeedTab>()
export const trackLikedToggled = createEvent<string>()

export const $activeTab = createStore<FeedTab>('trending').on(tabChanged, (_, tab) => tab)

const toggleTrackLikeFx = createEffect(
  async ({ trackId, isLiked }: { trackId: string; isLiked: boolean }) => {
    return toggleTrackLike(trackId, isLiked)
  }
)

export const fetchFeedFx = createEffect<
  { filter: FeedTab; limit?: number },
  { filter: FeedTab; tracks: Track[] }
>(async ({ filter, limit = 24 }) => {
  const tracks = await fetchTracks({ filter, limit })
  return { filter, tracks }
})

const initialFeeds: Record<FeedTab, Track[]> = {
  trending: [],
  popular: [],
  new: [],
}

const $feeds = createStore(initialFeeds).on(fetchFeedFx.doneData, (feeds, { filter, tracks }) => ({
  ...feeds,
  [filter]: tracks,
}))

export const $likes = createStore<Record<string, boolean>>({})
  .on(fetchFeedFx.doneData, (state, { tracks }) => {
    const next = { ...state }
    tracks.forEach(track => {
      next[track.id] = track.liked ?? false
    })
    return next
  })
  .on(toggleTrackLikeFx.doneData, (state, { trackId, isLiked }) => ({
    ...state,
    [trackId]: isLiked,
  }))

export const $tracks = combine($activeTab, $feeds, $likes, (tab, feeds, likes): Track[] =>
  feeds[tab].map(track => ({
    ...track,
    liked: likes[track.id] ?? track.liked ?? false,
  }))
)

const $feedLoaded = createStore<Record<FeedTab, boolean>>({
  trending: false,
  popular: false,
  new: false,
}).on(fetchFeedFx.doneData, (loaded, { filter }) => ({
  ...loaded,
  [filter]: true,
}))

sample({
  clock: routes.home.opened,
  fn: () => ({ filter: 'trending' as FeedTab }),
  target: fetchFeedFx,
})

sample({
  clock: tabChanged,
  source: $feedLoaded,
  filter: (loaded, filter) => !loaded[filter],
  fn: (_, filter) => ({ filter }),
  target: fetchFeedFx,
})

sample({
  clock: trackLikedToggled,
  source: $likes,
  fn: (likes, trackId) => ({
    trackId,
    isLiked: !(likes[trackId] ?? false),
  }),
  target: toggleTrackLikeFx,
})
