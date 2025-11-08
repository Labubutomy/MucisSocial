import { createHistoryRouter, createRoute } from 'atomic-router'

export const routes = {
  home: createRoute(),
  auth: createRoute(),
  search: createRoute(),
  track: createRoute<{ trackId: string }>(),
  profile: createRoute(),
  profilePlaylists: createRoute(),
  curations: createRoute(),
  playlistCreate: createRoute(),
  playlistAddTracks: createRoute<{ playlistId: string }>(),
  collection: createRoute<{ collectionId: string }>(),
}

export const mappedRoutes = [
  { route: routes.home, path: '/' },
  { route: routes.auth, path: '/auth' },
  { route: routes.search, path: '/search' },
  { route: routes.track, path: '/track/:trackId' },
  { route: routes.profile, path: '/profile' },
  { route: routes.profilePlaylists, path: '/profile/playlists' },
  { route: routes.curations, path: '/curations' },
  { route: routes.playlistCreate, path: '/playlists/create' },
  { route: routes.playlistAddTracks, path: '/playlists/:playlistId/add-tracks' },
  { route: routes.collection, path: '/collection/:collectionId' },
]

export const router = createHistoryRouter({
  routes: mappedRoutes,
})
