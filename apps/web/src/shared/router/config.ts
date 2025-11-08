import { createHistoryRouter, createRoute } from 'atomic-router'

export const routes = {
  index: createRoute(),
}

export const mappedRoutes = [
  {
    route: routes.index,
    path: '/',
  },
]

export const router = createHistoryRouter({
  routes: mappedRoutes,
})
