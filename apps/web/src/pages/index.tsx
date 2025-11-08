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
import { ProfilePage } from '@pages/profile'
import { UserPlaylistsPage } from '@pages/user-playlists'
import { CollectionPage } from '@pages/collection'
import { CreatePlaylistPage } from '@pages/create-playlist'
import { PlaylistAddTracksPage } from '@pages/playlist-add-tracks'
import { CurationsPage } from '@pages/curations'

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
    { route: routes.profile, view: ProfilePage },
    { route: routes.profilePlaylists, view: UserPlaylistsPage },
    { route: routes.curations, view: CurationsPage },
    { route: routes.playlistCreate, view: CreatePlaylistPage },
    { route: routes.playlistAddTracks, view: PlaylistAddTracksPage },
    { route: routes.collection, view: CollectionPage },
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
