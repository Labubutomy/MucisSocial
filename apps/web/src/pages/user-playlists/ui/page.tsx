import { useUnit } from 'effector-react'
import { PlaylistGrid } from '@widgets/playlists'
import { routes } from '@shared/router'
import { $myPlaylists, fetchMyPlaylistsFx } from '@pages/profile/model'

export const UserPlaylistsPage = () => {
  const { playlists, pending, navigateToCollection, navigateToCreate } = useUnit({
    playlists: $myPlaylists,
    pending: fetchMyPlaylistsFx.pending,
    navigateToCollection: routes.collection.navigate,
    navigateToCreate: routes.playlistCreate.navigate,
  })

  return (
    <div className="page-container space-y-10 pb-20 pt-10">
      <header className="space-y-4">
        <p className="text-xs uppercase tracking-[0.4em] text-primary">Медиатека</p>
        <h1 className="text-3xl font-semibold md:text-4xl">Плейлисты, созданные вами</h1>
        <p className="max-w-2xl text-base text-muted-foreground">
          Организуйте звуковые миры, собирайте подборки вместе с друзьями и создавайте атмосферу под
          каждый момент.
        </p>
      </header>
      {pending && playlists.length === 0 ? (
        <p className="text-sm text-muted-foreground">Загрузка плейлистов...</p>
      ) : (
        <PlaylistGrid
          title="Ваши плейлисты"
          playlists={playlists}
          onCreate={() => navigateToCreate({ params: {}, query: {} })}
          onOpen={playlist =>
            navigateToCollection({
              params: { collectionId: playlist.id },
              query: {},
            })
          }
        />
      )}
    </div>
  )
}
