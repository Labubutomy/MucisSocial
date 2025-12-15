import { createEffect, createEvent, createStore, sample } from 'effector'
import { routesApi } from '@features/routes/api'
import type { Route } from '@features/routes/api'

export interface RoutesFilters {
  city?: string
  is_public?: boolean
  limit?: number
  offset?: number
}

export const $routes = createStore<Route[]>([])
export const $loading = createStore(false)
export const $total = createStore(0)
export const $filters = createStore<RoutesFilters>({
  limit: 20,
  offset: 0,
})

export const filtersChanged = createEvent<Partial<RoutesFilters>>()
export const loadRoutes = createEvent()
export const loadMoreRoutes = createEvent()

$filters.on(filtersChanged, (state, updates) => ({ ...state, ...updates }))

export const fetchRoutesFx = createEffect<RoutesFilters, { routes: Route[]; total: number }>(
  async filters => {
    const result = await routesApi.listRoutes(filters)
    return { routes: result.routes, total: result.total }
  }
)

export const fetchMoreRoutesFx = createEffect<RoutesFilters, Route[]>(async filters => {
  const result = await routesApi.listRoutes(filters)
  return result.routes
})

$routes
  .on(fetchRoutesFx.doneData, (_, data) => data.routes)
  .on(fetchMoreRoutesFx.doneData, (state, newRoutes) => [...state, ...newRoutes])

$total.on(fetchRoutesFx.doneData, (_, data) => data.total)
$loading.on(fetchRoutesFx.pending, (_, pending) => pending)

sample({
  clock: loadRoutes,
  source: $filters,
  target: fetchRoutesFx,
})

sample({
  clock: loadMoreRoutes,
  source: $filters,
  fn: filters => ({
    ...filters,
    offset: (filters.offset || 0) + (filters.limit || 20),
  }),
  target: fetchMoreRoutesFx,
})
