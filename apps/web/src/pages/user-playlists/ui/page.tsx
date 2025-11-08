import { useUnit } from 'effector-react'
import { PlaylistGrid } from '@widgets/playlists'
import { userPlaylists } from '@pages/user-playlists/model/data'
import { routes } from '@shared/router'

export const UserPlaylistsPage = () => {
  const navigateToCollection = useUnit(routes.collection.navigate)
  const navigateToCreate = useUnit(routes.playlistCreate.navigate)

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
      <PlaylistGrid
        title="Ваши плейлисты"
        playlists={userPlaylists}
        onCreate={() => navigateToCreate({ params: {}, query: {} })}
        onOpen={playlist =>
          navigateToCollection({
            params: { collectionId: playlist.id },
            query: {},
          })
        }
      />
    </div>
  )
}
