import axios from 'axios'
import { combine, createEffect, createEvent, createStore, sample } from 'effector'
import { routes } from '@shared/router'
import { appStarted } from '@shared/config/init'
import { API_CONFIG } from '@shared/config/api'

const seeds = [
  'Новые релизы',
  'Неоновый фанк',
  'Синтвейв ночь',
  'Лоуфай для работы',
  'Электронный чилл',
  'Инди-поп плейлист',
]

export const queryChanged = createEvent<string>()
export const focusChanged = createEvent<boolean>()
export const hoverChanged = createEvent<boolean>()
export const searchSubmitted = createEvent()
export const suggestionSelected = createEvent<string>()

export const $query = createStore('')
  .on(queryChanged, (_, query) => query)
  .on(suggestionSelected, (_, value) => value)

export const $isFocused = createStore(false)
  .on(focusChanged, (_, value) => value)
  .on(searchSubmitted, () => false)
  .on(suggestionSelected, () => false)

export const $isHoveringList = createStore(false)
  .on(hoverChanged, (_, value) => value)
  .on(searchSubmitted, () => false)
  .on(suggestionSelected, () => false)

const fetchSuggestionSeedsFx = createEffect(async () => {
  const response = await axios.get<{ items: Array<{ query: string }> }>(
    `${API_CONFIG.mockApi}/api/v1/tracks/search/trending`
  )
  return response.data.items.map(item => item.query)
})

export const $suggestionSeeds = createStore(seeds).on(
  fetchSuggestionSeedsFx.doneData,
  (_, values) => values
)

export const $suggestions = combine($query, $suggestionSeeds, (query, currentSeeds) => {
  const lowered = query.trim().toLowerCase()
  if (!lowered) return currentSeeds
  return currentSeeds.filter(item => item.toLowerCase().includes(lowered))
})

export const $showDropdown = combine(
  $isFocused,
  $isHoveringList,
  $suggestions,
  (isFocused, isHovering, suggestions) => (isFocused || isHovering) && suggestions.length > 0
)

sample({
  clock: searchSubmitted,
  source: $query,
  fn: query => ({
    params: {},
    query: query.trim() ? { q: query.trim() } : {},
  }),
  target: routes.search.navigate,
})

sample({
  clock: suggestionSelected,
  fn: value => ({
    params: {},
    query: { q: value },
  }),
  target: routes.search.navigate,
})

sample({
  clock: appStarted,
  target: fetchSuggestionSeedsFx,
})
