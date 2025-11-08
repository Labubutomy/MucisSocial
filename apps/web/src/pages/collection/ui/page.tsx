import { useMemo, useState } from 'react'
import { useUnit } from 'effector-react'
import { PlaylistHero } from '@entities/playlist'
import { CollectionTrackList } from '@widgets/collection'
import { currentCollection, collectionTracks } from '@pages/collection/model/data'
import { $currentTrack, playbackToggled, trackQueued } from '@features/player'
import type { PlaylistDetail } from '@entities/playlist'
import type { Track } from '@entities/track'
import { routes } from '@shared/router'
import { Button } from '@shared/ui/button'

export const CollectionPage = () => {
  const [likedTracks, setLikedTracks] = useState<Record<string, boolean>>({})
  const [collectionLiked, setCollectionLiked] = useState(Boolean(currentCollection.liked))

  const [playerTrack, enqueueTrack, togglePlayback, navigateToTrack, navigateToAddTracks] = useUnit(
    [
      $currentTrack,
      trackQueued,
      playbackToggled,
      routes.track.navigate,
      routes.playlistAddTracks.navigate,
    ]
  )

  const tracks = useMemo(
    () =>
      collectionTracks.map(track => ({
        ...track,
        liked: likedTracks[track.id] ?? track.liked ?? false,
      })),
    [likedTracks]
  )

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
    setLikedTracks(prev => ({ ...prev, [track.id]: !prev[track.id] }))
  }

  const handleTrackOpen = (track: Track) => {
    navigateToTrack({
      params: { trackId: track.id },
      query: {},
    })
  }

  return (
    <div className="page-container space-y-12 pb-20 pt-10">
      <PlaylistHero
        playlist={{ ...currentCollection, liked: collectionLiked }}
        onPlay={handlePlayCollection}
        onToggleLike={() => setCollectionLiked(prev => !prev)}
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
              params: { playlistId: currentCollection.id },
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
