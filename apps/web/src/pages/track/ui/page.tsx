import { useMemo, useState } from 'react'
import { useUnit } from 'effector-react'
import { TrackHero, RecommendedList } from '@widgets/track'
import { currentTrackDetail, recommendedTracks } from '@pages/track/model/data'
import { $currentTrack, $isPlaying, playbackToggled, trackQueued } from '@features/player'
import type { Track } from '@entities/track'
import type { TrackDetail } from '@widgets/track'
import { routes } from '@shared/router'

export const TrackPage = () => {
  const [likes, setLikes] = useState<Record<string, boolean>>({
    [currentTrackDetail.id]: Boolean(currentTrackDetail.liked),
  })

  const [playerTrack, isPlaying, enqueueTrack, togglePlayback, trackParams, navigateToTrack] =
    useUnit([
      $currentTrack,
      $isPlaying,
      trackQueued,
      playbackToggled,
      routes.track.$params,
      routes.track.navigate,
    ])

  const trackDetail = useMemo<TrackDetail>(
    () => ({
      ...currentTrackDetail,
      liked: likes[currentTrackDetail.id] ?? currentTrackDetail.liked ?? false,
    }),
    [likes]
  )

  const recommended = useMemo(
    () =>
      recommendedTracks.map(track => ({
        ...track,
        liked: likes[track.id] ?? track.liked ?? false,
      })),
    [likes]
  )

  const handleTogglePlay = (track: Track) => {
    if (!playerTrack || playerTrack.id !== track.id) {
      enqueueTrack(track)
      return
    }
    togglePlayback()
  }

  const handleToggleLike = (track: Track) => {
    setLikes(prev => ({ ...prev, [track.id]: !prev[track.id] }))
  }

  const handleOpen = (track: Track) => {
    navigateToTrack({
      params: { trackId: track.id },
      query: {},
    })
  }

  return (
    <div className="page-container space-y-12 pb-16 pt-10">
      <TrackHero
        track={trackDetail}
        isPlaying={playerTrack?.id === trackDetail.id && isPlaying}
        onTogglePlay={handleTogglePlay}
        onToggleLike={handleToggleLike}
        onShare={track => console.info('Поделиться треком', track.id)}
        onAddToPlaylist={track =>
          console.info('Добавить трек в плейлист', track.id, 'из', trackParams.trackId)
        }
        onGoToArtist={artistId => console.info('Открыть артиста', artistId)}
        onGoToAlbum={albumId => console.info('Открыть альбом', albumId)}
      />
      <RecommendedList
        tracks={recommended}
        activeTrackId={playerTrack?.id}
        onPlayToggle={handleTogglePlay}
        onLike={handleToggleLike}
        onShare={track => console.info('Поделиться треком', track.id)}
        onOpen={handleOpen}
      />
    </div>
  )
}
