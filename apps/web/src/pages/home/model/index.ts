import { combine, createEffect, createEvent, createStore, sample } from 'effector'
import type { Track } from '@entities/track'
import { routes } from '@shared/router'
import {
  fetchTopCharts,
  fetchNewReleases,
  fetchTrackRecommendations,
  toggleTrackLike,
} from '@entities/track/api'

export type FeedTab = 'trending' | 'new' | 'recommended'

export const tabChanged = createEvent<FeedTab>()
export const trackLikedToggled = createEvent<string>()

export const $activeTab = createStore<FeedTab>('trending').on(tabChanged, (_, tab) => tab)

const toggleTrackLikeFx = createEffect(
  async ({ trackId, isLiked }: { trackId: string; isLiked: boolean }) => {
    return toggleTrackLike(trackId, isLiked)
  }
)

// Effect для загрузки топ чартов (в тренде)
const fetchTrendingFx = createEffect(async () => {
  const tracks = await fetchTopCharts(24)
  return { filter: 'trending' as FeedTab, tracks }
})

// Effect для загрузки новинок
const fetchNewFx = createEffect(async () => {
  const tracks = await fetchNewReleases(24, 30) // Последние 30 дней
  return { filter: 'new' as FeedTab, tracks }
})

// Effect для загрузки рекомендаций
const fetchRecommendationsFx = createEffect(async () => {
  const tracks = await fetchTrackRecommendations({ limit: 24 })
  return { filter: 'recommended' as FeedTab, tracks }
})

const initialFeeds: Record<FeedTab, Track[]> = {
  trending: [],
  new: [],
  recommended: [],
}

const $feeds = createStore(initialFeeds)
  .on(fetchTrendingFx.doneData, (feeds, { filter, tracks }) => ({
    ...feeds,
    [filter]: tracks,
  }))
  .on(fetchNewFx.doneData, (feeds, { filter, tracks }) => ({
    ...feeds,
    [filter]: tracks,
  }))
  .on(fetchRecommendationsFx.doneData, (feeds, { filter, tracks }) => ({
    ...feeds,
    [filter]: tracks,
  }))

export const $likes = createStore<Record<string, boolean>>({})
  .on(fetchTrendingFx.doneData, (state, { tracks }) => {
    const next = { ...state }
    tracks.forEach(track => {
      next[track.id] = track.liked ?? false
    })
    return next
  })
  .on(fetchNewFx.doneData, (state, { tracks }) => {
    const next = { ...state }
    tracks.forEach(track => {
      next[track.id] = track.liked ?? false
    })
    return next
  })
  .on(fetchRecommendationsFx.doneData, (state, { tracks }) => {
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
  new: false,
  recommended: false,
})
  .on(fetchTrendingFx.doneData, (loaded, { filter }) => ({
    ...loaded,
    [filter]: true,
  }))
  .on(fetchNewFx.doneData, (loaded, { filter }) => ({
    ...loaded,
    [filter]: true,
  }))
  .on(fetchRecommendationsFx.doneData, (loaded, { filter }) => ({
    ...loaded,
    [filter]: true,
  }))

// Загружать топ чарты при открытии главной страницы
sample({
  clock: routes.home.opened,
  fn: () => undefined,
  target: fetchTrendingFx,
})

// Загружать данные при переключении на вкладку
sample({
  clock: tabChanged,
  source: $feedLoaded,
  filter: (loaded, filter) => !loaded[filter] && filter === 'trending',
  fn: () => undefined,
  target: fetchTrendingFx,
})

sample({
  clock: tabChanged,
  source: $feedLoaded,
  filter: (loaded, filter) => !loaded[filter] && filter === 'new',
  fn: () => undefined,
  target: fetchNewFx,
})

sample({
  clock: tabChanged,
  source: $feedLoaded,
  filter: (loaded, filter) => !loaded[filter] && filter === 'recommended',
  fn: () => undefined,
  target: fetchRecommendationsFx,
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

// Состояния загрузки и ошибок
export const $trendingPending = fetchTrendingFx.pending
export const $newPending = fetchNewFx.pending
export const $recommendationsPending = fetchRecommendationsFx.pending

export const $trendingError = createStore<string | null>(null)
  .on(fetchTrendingFx.failData, (_, error) =>
    error instanceof Error ? error.message : 'Не удалось загрузить топ чарты'
  )
  .reset(fetchTrendingFx)

export const $newError = createStore<string | null>(null)
  .on(fetchNewFx.failData, (_, error) =>
    error instanceof Error ? error.message : 'Не удалось загрузить новинки'
  )
  .reset(fetchNewFx)

export const $recommendationsError = createStore<string | null>(null)
  .on(fetchRecommendationsFx.failData, (_, error) =>
    error instanceof Error ? error.message : 'Не удалось загрузить рекомендации'
  )
  .reset(fetchRecommendationsFx)
