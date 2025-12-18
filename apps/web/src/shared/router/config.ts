import { createHistoryRouter, createRoute } from 'atomic-router'

export const routes = {
  home: createRoute(),
  auth: createRoute(),
  search: createRoute(),
  track: createRoute<{ trackId: string }>(),
  artist: createRoute<{ artistId: string }>(),
  profile: createRoute(),
  profilePlaylists: createRoute(),
  curations: createRoute(),
  playlistCreate: createRoute(),
  playlistAddTracks: createRoute<{ playlistId: string }>(),
  collection: createRoute<{ collectionId: string }>(),
  session: createRoute(),
  sessionRoom: createRoute<{ roomId: string }>(),
  groups: createRoute(),
  routeCreate: createRoute(),
  routes: createRoute(),
  routeView: createRoute<{ routeId: string }>(),
  friends: createRoute(),
  messages: createRoute(),
}

export const mappedRoutes = [
  { route: routes.home, path: '/' },
  { route: routes.auth, path: '/auth' },
  { route: routes.search, path: '/search' },
  { route: routes.track, path: '/track/:trackId' },
  { route: routes.artist, path: '/artist/:artistId' },
  { route: routes.profile, path: '/profile' },
  { route: routes.profilePlaylists, path: '/profile/playlists' },
  { route: routes.curations, path: '/curations' },
  { route: routes.playlistCreate, path: '/playlists/create' },
  { route: routes.playlistAddTracks, path: '/playlists/:playlistId/add-tracks' },
  { route: routes.collection, path: '/collection/:collectionId' },
  { route: routes.session, path: '/listen' },
  { route: routes.sessionRoom, path: '/listen/:roomId' },
  { route: routes.groups, path: '/groups' },
  { route: routes.routeCreate, path: '/routes/create' },
  { route: routes.routes, path: '/routes' },
  { route: routes.routeView, path: '/routes/:routeId' },
  { route: routes.friends, path: '/friends' },
  { route: routes.messages, path: '/messages' },
]

export const router = createHistoryRouter({
  routes: mappedRoutes,
})
