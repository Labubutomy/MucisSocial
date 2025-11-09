import { useUnit } from 'effector-react'
import { Card } from '@shared/ui/card'
import { Button } from '@shared/ui/button'
import { PlaylistCard } from '@entities/playlist'
import { routes } from '@shared/router'
import { $user } from '@features/auth/model'
import { $myPlaylists, fetchMyPlaylistsFx } from '@pages/profile/model'

export const CurationsPage = () => {
  const [navigate, query, goToCollection] = useUnit([
    routes.curations.navigate,
    routes.curations.$query,
    routes.collection.navigate,
  ])
  const { user, playlists, playlistsPending } = useUnit({
    user: $user,
    playlists: $myPlaylists,
    playlistsPending: fetchMyPlaylistsFx.pending,
  })

  const tab = query.tab === 'artists' ? 'artists' : 'playlists'

  const handleTabChange = (nextTab: 'playlists' | 'artists') => {
    navigate({
      params: {},
      query: { tab: nextTab },
    })
  }

  const curatedPlaylists = playlists.slice(0, 6)

  return (
    <div className="page-container space-y-8 pb-20 pt-10">
      <header className="space-y-4">
        <div className="flex flex-wrap items-center gap-3">
          <Button
            type="button"
            variant={tab === 'playlists' ? 'primary' : 'outline'}
            onClick={() => handleTabChange('playlists')}
          >
            Плейлисты
          </Button>
          <Button
            type="button"
            variant={tab === 'artists' ? 'primary' : 'outline'}
            onClick={() => handleTabChange('artists')}
          >
            Артисты
          </Button>
        </div>
        <div>
          <h1 className="text-3xl font-semibold md:text-4xl">
            {tab === 'playlists'
              ? 'Подборки плейлистов, которые стоит услышать'
              : 'Музыка от артистов, вдохновляющих сообщество'}
          </h1>
          <p className="mt-2 max-w-2xl text-sm text-muted-foreground md:text-base">
            Выбирайте плейлист под настроение или откройте новых артистов, которые зарядят энергией.
          </p>
        </div>
      </header>

      {tab === 'playlists' ? (
        <Card padding="lg" className="space-y-6 bg-secondary/20">
          {playlistsPending && curatedPlaylists.length === 0 ? (
            <p className="text-sm text-muted-foreground">Загрузка подборок...</p>
          ) : (
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {curatedPlaylists.map(playlist => (
                <PlaylistCard
                  key={playlist.id}
                  playlist={playlist}
                  onClick={() =>
                    goToCollection({
                      params: { collectionId: playlist.id },
                      query: {},
                    })
                  }
                />
              ))}
            </div>
          )}
        </Card>
      ) : (
        <Card padding="lg" className="space-y-6 bg-secondary/20">
          {user ? (
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
              {user.musicTasteSummary.topArtists.map(artist => (
                <div
                  key={artist}
                  className="flex flex-col gap-3 rounded-2xl border border-border/60 bg-secondary/40 p-4 text-left transition hover:border-primary/40 hover:shadow-lg hover:shadow-primary/20"
                >
                  <div className="h-28 w-full overflow-hidden rounded-xl bg-gradient-to-br from-primary/30 via-accent/20 to-secondary/30" />
                  <h3 className="text-lg font-semibold text-foreground">{artist}</h3>
                  <p className="text-xs text-muted-foreground">
                    Добавьте треки любимого артиста в свой плейлист.
                  </p>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">Загрузка артистов...</p>
          )}
        </Card>
      )}
    </div>
  )
}
