import { createEffect, createEvent, createStore, sample } from 'effector'
import { routesApi } from '@features/routes/api'
import type { Route } from '@features/routes/api'

export const $route = createStore<Route | null>(null)
export const $loading = createStore(false)
export const $error = createStore<string | null>(null)

export const routeLoaded = createEvent<string>()

export const fetchRouteFx = createEffect<string, Route>(async routeId => {
  const route = await routesApi.getRoute(routeId, true) // Include points
  return route
})

$route.on(fetchRouteFx.doneData, (_, route) => route)
$loading.on(fetchRouteFx.pending, (_, pending) => pending)
$error
  .on(fetchRouteFx.failData, (_, error) => error.message || 'Ошибка загрузки маршрута')
  .on(fetchRouteFx, () => null)

sample({
  clock: routeLoaded,
  target: fetchRouteFx,
})
