import { useUnit } from 'effector-react'
import { ProfileHeader, TasteCloud } from '@entities/user'
import { Card } from '@shared/ui/card'
import { Button } from '@shared/ui/button'
import { routes } from '@shared/router'
import { $user, signOut } from '@features/auth/model'
import { tabChanged } from '@pages/home/model'
import { searchSubmitted } from '@pages/search/model'
import { searchArtists } from '@entities/artist/api'
import {
  $myPlaylists,
  fetchMyPlaylistsFx,
  $userTaste,
  fetchUserTasteFx,
} from '@pages/profile/model'

export const ProfilePage = () => {
  const {
    user,
    playlists,
    playlistsPending,
    userTaste,
    tastePending,
    goToPlaylists,
    goToCurations,
    goToSession,
    goToRoutes,
    goToFriends,
    goToHome,
    handleSignOut,
  } = useUnit({
    user: $user,
    playlists: $myPlaylists,
    playlistsPending: fetchMyPlaylistsFx.pending,
    userTaste: $userTaste,
    tastePending: fetchUserTasteFx.pending,
    goToPlaylists: routes.profilePlaylists.navigate,
    goToCurations: routes.curations.navigate,
    goToSession: routes.session.navigate,
    goToRoutes: routes.routes.navigate,
    goToFriends: routes.friends.navigate,
    goToHome: routes.home.navigate,
    handleSignOut: signOut,
  })

  if (!user) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center pb-20 pt-10">
        <p className="text-sm text-muted-foreground">Загрузка профиля...</p>
      </div>
    )
  }

  const recentPlaylists = playlists.slice(0, 3)

  return (
    <div className="page-container space-y-8 pb-20 pt-10">
      <div className="grid gap-8 lg:grid-cols-[minmax(0,0.9fr),minmax(0,1.1fr)]">
        <ProfileHeader user={user} />
        <TasteCloud
          user={{
            ...user,
            musicTasteSummary: userTaste
              ? {
                  topGenres: userTaste.topGenres,
                  topArtists: userTaste.topArtists,
                }
              : user.musicTasteSummary,
          }}
          onSelectGenre={genre => {
            // Открываем страницу поиска с фильтром по жанру
            routes.search.navigate({ params: {}, query: { q: genre } })
            // После навигации выполняем поиск
            setTimeout(() => {
              searchSubmitted()
            }, 100)
          }}
          onSelectArtist={async artist => {
            // Ищем артиста по имени, чтобы получить его ID
            try {
              const artists = await searchArtists(artist, 1)
              if (artists.length > 0) {
                routes.artist.navigate({
                  params: { artistId: artists[0].id },
                  query: {},
                })
              } else {
                // Если не нашли, открываем страницу поиска
                routes.search.navigate({ params: {}, query: { q: artist, type: 'artist' } })
              }
            } catch (error) {
              console.error('Failed to find artist:', error)
              // Fallback: открываем страницу поиска
              routes.search.navigate({ params: {}, query: { q: artist, type: 'artist' } })
            }
          }}
        />
      </div>

      <Card padding="lg" className="space-y-4 bg-secondary/20">
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <Button
            variant="outline"
            onClick={() => goToSession({ params: {}, query: {} })}
            className="w-full"
          >
            Слушать вместе
          </Button>
          <Button
            variant="outline"
            onClick={() => goToRoutes({ params: {}, query: {} })}
            className="w-full"
          >
            Маршруты
          </Button>
          <Button
            variant="outline"
            onClick={() => goToFriends({ params: {}, query: {} })}
            className="w-full"
          >
            Найти друзей
          </Button>
        </div>
      </Card>

      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {/* Мои плейлисты */}
        <Card
          padding="lg"
          className="cursor-pointer space-y-4 bg-secondary/20 transition hover:bg-secondary/30 hover:shadow-lg"
          onClick={() => goToPlaylists({ params: {}, query: {} })}
        >
          <div className="space-y-2">
            <p className="text-xs uppercase tracking-[0.4em] text-primary">Мои плейлисты</p>
            <h3 className="text-xl font-semibold">
              {playlistsPending ? 'Загрузка...' : `${playlists.length} плейлистов`}
            </h3>
            <p className="text-sm text-muted-foreground">
              {recentPlaylists.length > 0
                ? `Недавние: ${recentPlaylists.map(p => p.title).join(', ')}`
                : 'Создайте свой первый плейлист'}
            </p>
          </div>
        </Card>

        {/* Рекомендации */}
        <Card
          padding="lg"
          className="cursor-pointer space-y-4 bg-secondary/20 transition hover:bg-secondary/30 hover:shadow-lg"
          onClick={() => {
            // Открываем главную страницу и переключаемся на вкладку "Рекомендации"
            goToHome({ params: {}, query: {} })
            // Используем setTimeout, чтобы навигация произошла перед переключением вкладки
            setTimeout(() => {
              tabChanged('recommended')
            }, 100)
          }}
        >
          <div className="space-y-2">
            <p className="text-xs uppercase tracking-[0.4em] text-primary">Рекомендации</p>
            <h3 className="text-xl font-semibold">Персональные треки</h3>
            <p className="text-sm text-muted-foreground">
              Откройте для себя новую музыку на основе ваших предпочтений
            </p>
          </div>
        </Card>

        {/* Любимые артисты */}
        <Card
          padding="lg"
          className="cursor-pointer space-y-4 bg-secondary/20 transition hover:bg-secondary/30 hover:shadow-lg"
          onClick={() =>
            goToCurations({
              params: {},
              query: { tab: 'artists' },
            })
          }
        >
          <div className="space-y-2">
            <p className="text-xs uppercase tracking-[0.4em] text-primary">Любимые артисты</p>
            <h3 className="text-xl font-semibold">
              {tastePending
                ? 'Загрузка...'
                : `${(userTaste?.topArtists ?? user.musicTasteSummary?.topArtists ?? []).length} артистов`}
            </h3>
            <p className="text-sm text-muted-foreground">
              {(userTaste?.topArtists ?? user.musicTasteSummary?.topArtists ?? []).length > 0
                ? `Топ: ${(userTaste?.topArtists ?? user.musicTasteSummary?.topArtists ?? [])
                    .slice(0, 3)
                    .join(', ')}`
                : 'Начните слушать музыку, чтобы увидеть ваших любимых артистов'}
            </p>
          </div>
        </Card>
      </div>

      <Button
        variant="outline"
        onClick={handleSignOut}
        className="w-full border-destructive/40 text-destructive hover:bg-destructive/10"
      >
        Выйти
      </Button>
    </div>
  )
}
