import { combine, createEffect, createEvent, createStore, sample } from 'effector'
import type { SearchHistoryItem } from '@features/search'
import { routes } from '@shared/router'
import { toggleTrackLike } from '@entities/track/api'
import type { SearchResult } from '@widgets/search'
import {
  addSearchHistoryEntry,
  fetchSearchHistory,
  fetchSearchResults,
  fetchTrendingQueries,
} from './api'

export const queryChanged = createEvent<string>()
export const searchSubmitted = createEvent()
export const historyCleared = createEvent()
export const historyItemSelected = createEvent<SearchHistoryItem>()
export const trendingSelected = createEvent<{ id: string; label: string }>()
export const trackLikeToggled = createEvent<string>()

export const $query = createStore('')
  .on(queryChanged, (_, query) => query)
  .on(historyItemSelected, (_, item) => item.query)
  .on(trendingSelected, (_, item) => item.label)

const fetchTrendingFx = createEffect(fetchTrendingQueries)
const fetchHistoryFx = createEffect(fetchSearchHistory)
const addHistoryFx = createEffect(addSearchHistoryEntry)
const searchFx = createEffect(fetchSearchResults)
const toggleLikeFx = createEffect(
  async ({ trackId, isLiked }: { trackId: string; isLiked: boolean }) =>
    toggleTrackLike(trackId, isLiked)
)

const historyItemAdded = createEvent<SearchHistoryItem>()

export const $history = createStore<SearchHistoryItem[]>([])
  .on(fetchHistoryFx.doneData, (_, items) => items)
  .on(historyItemAdded, (history, item) => [item, ...history])
  .on(historyCleared, () => [])

sample({
  clock: addHistoryFx.doneData,
  target: historyItemAdded,
})

export const $likes = createStore<Record<string, boolean>>({})
  .on(searchFx.doneData, (state, results) => {
    const next = { ...state }
    results.forEach(result => {
      if (result.type === 'track') {
        next[result.data.id] = result.data.liked ?? false
      }
    })
    return next
  })
  .on(toggleLikeFx.doneData, (state, { trackId, isLiked }) => ({
    ...state,
    [trackId]: isLiked,
  }))

const $rawResults = createStore<SearchResult[]>([])
  .on(searchFx.doneData, (_, results) => results)
  .reset(historyCleared)

export const $results = combine($query, $likes, $rawResults, (query, likes, results) => {
  const lowered = query.trim().toLowerCase()
  const filtered = lowered
    ? results.filter(result => {
        switch (result.type) {
          case 'track':
            return (
              result.data.title.toLowerCase().includes(lowered) ||
              result.data.artist.name.toLowerCase().includes(lowered)
            )
          case 'artist':
            return result.data.name.toLowerCase().includes(lowered)
          case 'playlist':
            return (
              result.data.title.toLowerCase().includes(lowered) ||
              result.data.description?.toLowerCase().includes(lowered)
            )
          default:
            return false
        }
      })
    : results

  return filtered.map(result =>
    result.type === 'track'
      ? {
          ...result,
          data: {
            ...result.data,
            liked: likes[result.data.id] ?? result.data.liked ?? false,
          },
        }
      : result
  )
})

export const $trending = createStore<Array<{ id: string; label: string }>>([]).on(
  fetchTrendingFx.doneData,
  (_, items) => items
)

export const $suggestions = combine($query, $trending, (query, trending) => {
  const lowered = query.trim().toLowerCase()
  if (!lowered) return trending.map(item => item.label)
  return trending.map(item => item.label).filter(label => label.toLowerCase().includes(lowered))
})

sample({
  clock: routes.search.opened,
  target: [fetchTrendingFx, fetchHistoryFx],
})

sample({
  clock: routes.search.opened,
  source: routes.search.$query,
  fn: query => (typeof query.q === 'string' ? query.q : ''),
  target: queryChanged,
})

sample({
  clock: routes.search.opened,
  source: routes.search.$query,
  filter: query => Boolean(query.q),
  fn: query => String(query.q),
  target: [searchFx, addHistoryFx],
})

sample({
  clock: trendingSelected,
  fn: ({ label }) => label,
  target: queryChanged,
})

sample({
  clock: trendingSelected,
  fn: ({ label }) => label,
  target: [searchFx, addHistoryFx],
})

sample({
  clock: searchSubmitted,
  source: $query,
  filter: query => Boolean(query.trim()),
  fn: query => query.trim(),
  target: [searchFx, addHistoryFx],
})

sample({
  clock: historyItemSelected,
  fn: item => item.query,
  target: [queryChanged, searchFx],
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
