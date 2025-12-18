import { useUnit } from 'effector-react'
import { createRoutesView } from 'atomic-router-react'
import { routes } from '@shared/router'
import { GlobalPlayerLayout } from '@shared/ui/global-player-layout'
import { MiniPlayerController } from '@widgets/player'
import { $currentTrack, $isPlaying } from '@features/player'
import { HomePage } from '@pages/home'
import { AuthPage } from '@pages/auth'
import { SearchPage } from '@pages/search'
import { TrackPage } from '@pages/track'
import { ArtistPage } from '@pages/artist'
import { ProfilePage } from '@pages/profile'
import { UserPlaylistsPage } from '@pages/user-playlists'
import { CollectionPage } from '@pages/collection'
import { SessionPage } from '@pages/session'
import { GroupsPage } from '@pages/groups'
import { CreatePlaylistPage } from '@pages/create-playlist'
import { PlaylistAddTracksPage } from '@pages/playlist-add-tracks'
import { CurationsPage } from '@pages/curations'
import { CreateRoutePage } from '@pages/create-route'
import { RoutesPage } from '@pages/routes'
import { RouteViewPage } from '@pages/route-view'
import { FriendsPage } from '@pages/friends'
import { MessagesPage } from '@pages/messages'

const NotFoundPage = () => (
  <div className="page-container flex min-h-screen flex-col items-center justify-center gap-4 text-center">
    <h1 className="text-4xl font-semibold">Тишина в эфире</h1>
    <p className="max-w-md text-muted-foreground">
      Кажется, эта страница выпала из сет-листа. Вернитесь на главную, чтобы музыка продолжила
      звучать.
    </p>
  </div>
)

const RoutesView = createRoutesView({
  routes: [
    { route: routes.home, view: HomePage },
    { route: routes.auth, view: AuthPage },
    { route: routes.search, view: SearchPage },
    { route: routes.track, view: TrackPage },
    { route: routes.artist, view: ArtistPage },
    { route: routes.profile, view: ProfilePage },
    { route: routes.profilePlaylists, view: UserPlaylistsPage },
    { route: routes.curations, view: CurationsPage },
    { route: routes.playlistCreate, view: CreatePlaylistPage },
    { route: routes.playlistAddTracks, view: PlaylistAddTracksPage },
    { route: routes.collection, view: CollectionPage },
    { route: routes.session, view: SessionPage },
    { route: routes.sessionRoom, view: SessionPage },
    { route: routes.groups, view: GroupsPage },
    { route: routes.routeCreate, view: CreateRoutePage },
    { route: routes.routes, view: RoutesPage },
    { route: routes.routeView, view: RouteViewPage },
    { route: routes.friends, view: FriendsPage },
    { route: routes.messages, view: MessagesPage },
  ],
  otherwise: NotFoundPage,
})

export const Pages = () => {
  const [currentTrack, isPlaying, navigateToTrack, trackPageOpened, trackRouteParams] = useUnit([
    $currentTrack,
    $isPlaying,
    routes.track.navigate,
    routes.track.$isOpened,
    routes.track.$params,
  ])

  const shouldShowMiniPlayer = Boolean(
    currentTrack && isPlaying && (!trackPageOpened || trackRouteParams?.trackId !== currentTrack.id)
  )

  return (
    <GlobalPlayerLayout
      miniPlayer={
        currentTrack ? (
          <MiniPlayerController
            onOpenTrack={trackId =>
              navigateToTrack({
                params: { trackId },
                query: {},
              })
            }
          />
        ) : null
      }
      showMiniPlayer={shouldShowMiniPlayer}
    >
      <RoutesView />
    </GlobalPlayerLayout>
  )
}
