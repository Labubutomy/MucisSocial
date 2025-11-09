import { useUnit } from 'effector-react'
import { PlaylistHero } from '@entities/playlist'
import { CollectionTrackList } from '@widgets/collection'
import { $currentTrack, playbackToggled, trackQueued } from '@features/player'
import type { PlaylistDetail } from '@entities/playlist'
import type { Track } from '@entities/track'
import { routes } from '@shared/router'
import { Button } from '@shared/ui/button'
import {
  $collection,
  $tracks,
  collectionLikeToggled,
  trackLikeToggled,
} from '@pages/collection/model'

export const CollectionPage = () => {
  const {
    collection,
    tracks,
    playerTrack,
    enqueueTrack,
    togglePlayback,
    navigateToTrack,
    navigateToAddTracks,
    toggleCollectionLike,
    toggleTrackLike,
  } = useUnit({
    collection: $collection,
    tracks: $tracks,
    playerTrack: $currentTrack,
    enqueueTrack: trackQueued,
    togglePlayback: playbackToggled,
    navigateToTrack: routes.track.navigate,
    navigateToAddTracks: routes.playlistAddTracks.navigate,
    toggleCollectionLike: collectionLikeToggled,
    toggleTrackLike: trackLikeToggled,
  })

  const handlePlayCollection = (playlist: PlaylistDetail) => {
    const firstTrack = tracks[0]
    if (!firstTrack) return
    if (playerTrack?.id === firstTrack.id) {
      togglePlayback()
    } else {
      enqueueTrack(firstTrack)
    }
    console.info('Воспроизвести подборку', playlist.id)
  }

  const handleTrackPlayToggle = (track: Track) => {
    if (!playerTrack || playerTrack.id !== track.id) {
      enqueueTrack(track)
      return
    }
    togglePlayback()
  }

  const handleTrackLike = (track: Track) => {
    toggleTrackLike(track.id)
  }

  const handleTrackOpen = (track: Track) => {
    navigateToTrack({
      params: { trackId: track.id },
      query: {},
    })
  }

  if (!collection) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center pb-20 pt-10">
        <p className="text-sm text-muted-foreground">Загрузка подборки...</p>
      </div>
    )
  }

  return (
    <div className="page-container space-y-12 pb-20 pt-10">
      <PlaylistHero
        playlist={collection}
        onPlay={handlePlayCollection}
        onToggleLike={toggleCollectionLike}
        onShare={playlist => console.info('Поделиться плейлистом', playlist.id)}
      />
      <CollectionTrackList
        tracks={tracks}
        activeTrackId={playerTrack?.id}
        onPlayToggle={handleTrackPlayToggle}
        onLike={handleTrackLike}
        onAddToPlaylist={track => console.info('Добавить в плейлист из коллекции', track.id)}
        onShare={track => console.info('Поделиться треком', track.id)}
        onOpen={handleTrackOpen}
      />
      <div className="flex justify-end">
        <Button
          variant="outline"
          onClick={() =>
            navigateToAddTracks({
              params: { playlistId: collection.id },
              query: {},
            })
          }
        >
          Добавить треки в плейлист
        </Button>
      </div>
    </div>
  )
}
