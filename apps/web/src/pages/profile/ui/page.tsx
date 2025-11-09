import { useUnit } from 'effector-react'
import { ProfileHeader, TasteCloud } from '@entities/user'
import { Card } from '@shared/ui/card'
import { Button } from '@shared/ui/button'
import { PlaylistCard } from '@entities/playlist'
import { routes } from '@shared/router'
import { $user } from '@features/auth/model'
import { $myPlaylists, fetchMyPlaylistsFx } from '@pages/profile/model'

export const ProfilePage = () => {
  const { user, playlists, playlistsPending, goToPlaylists, goToCollection, goToCurations } =
    useUnit({
      user: $user,
      playlists: $myPlaylists,
      playlistsPending: fetchMyPlaylistsFx.pending,
      goToPlaylists: routes.profilePlaylists.navigate,
      goToCollection: routes.collection.navigate,
      goToCurations: routes.curations.navigate,
    })

  if (!user) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center pb-20 pt-10">
        <p className="text-sm text-muted-foreground">Загрузка профиля...</p>
      </div>
    )
  }

  const recentPlaylists = playlists.slice(0, 3)
  const curatedPlaylists = playlists.slice(0, 3).map((playlist, index) => ({
    ...playlist,
    originalId: playlist.id,
    id: `${playlist.id}-curated-${index}`,
  }))

  return (
    <div className="page-container space-y-8 pb-20 pt-10">
      <div className="grid gap-8 lg:grid-cols-[minmax(0,0.9fr),minmax(0,1.1fr)]">
        <ProfileHeader user={user} />
        <TasteCloud
          user={user}
          onSelectGenre={genre => console.info('Выбрать жанр', genre)}
          onSelectArtist={artist => console.info('Выбрать артиста', artist)}
        />
      </div>

      <Card padding="lg" className="space-y-6 bg-secondary/20">
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <p className="text-xs uppercase tracking-[0.4em] text-primary">Мои плейлисты</p>
            <h2 className="text-2xl font-semibold md:text-3xl">Недавние плейлисты</h2>
          </div>
          <Button variant="outline" onClick={() => goToPlaylists({ params: {}, query: {} })}>
            Смотреть всё
          </Button>
        </div>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {playlistsPending && recentPlaylists.length === 0 ? (
            <p className="text-sm text-muted-foreground">Загрузка плейлистов...</p>
          ) : (
            recentPlaylists.map(playlist => (
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
            ))
          )}
        </div>
      </Card>

      <Card padding="lg" className="space-y-6 bg-secondary/20">
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <p className="text-xs uppercase tracking-[0.4em] text-primary">Подборка плейлистов</p>
            <h2 className="text-2xl font-semibold md:text-3xl">
              Плейлисты, которые стоит послушать
            </h2>
          </div>
          <Button
            variant="outline"
            onClick={() =>
              goToCurations({
                params: {},
                query: { tab: 'playlists' },
              })
            }
          >
            Смотреть всё
          </Button>
        </div>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {curatedPlaylists.map(playlist => (
            <PlaylistCard
              key={playlist.id}
              playlist={playlist}
              onClick={() =>
                goToCollection({
                  params: {
                    collectionId: playlist.originalId,
                  },
                  query: {},
                })
              }
            />
          ))}
        </div>
      </Card>

      <Card padding="lg" className="space-y-6 bg-secondary/20">
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <p className="text-xs uppercase tracking-[0.4em] text-primary">
              Треки от любимых артистов
            </p>
            <h2 className="text-2xl font-semibold md:text-3xl">
              Подборка музыки от артистов, вдохновляющих вас
            </h2>
          </div>
          <Button
            variant="outline"
            onClick={() =>
              goToCurations({
                params: {},
                query: { tab: 'artists' },
              })
            }
          >
            Смотреть всё
          </Button>
        </div>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {(user.musicTasteSummary?.topArtists ?? []).length > 0 ? (
            user.musicTasteSummary?.topArtists?.map(artist => (
              <div
                key={artist}
                className="flex flex-col gap-3 rounded-2xl border border-border/60 bg-secondary/40 p-4 text-left transition hover:border-primary/40 hover:shadow-lg hover:shadow-primary/20"
              >
                <div className="h-28 w-full overflow-hidden rounded-xl bg-gradient-to-br from-primary/30 via-accent/20 to-secondary/30" />
                <h3 className="text-lg font-semibold text-foreground">{artist}</h3>
              </div>
            ))
          ) : (
            <div className="rounded-2xl border border-dashed border-border/60 bg-secondary/30 px-6 py-10 text-center text-sm text-muted-foreground">
              Музыкальные рекомендации появятся после того, как начнёте слушать музыку.
            </div>
          )}
        </div>
      </Card>
    </div>
  )
}
