import { useState } from 'react'
import { useUnit } from 'effector-react'
import { TrackHero, RecommendedList } from '@widgets/track'
import {
  $currentTrack,
  $isPlaying,
  $playbackError,
  $currentTime,
  $duration,
  $stream,
  $streamPending,
  playbackToggled,
  seekRequested,
  trackQueued,
  skipTrackRequested,
} from '@features/player'
import { QueueSidebar } from '@widgets/queue'
import type { Track } from '@entities/track'
import { routes } from '@shared/router'
import {
  $recommendedTracks,
  $trackDetail,
  $recommendationsPending,
  $recommendationsError,
  trackLikeToggled,
} from '@pages/track/model'

export const TrackPage = () => {
  const [isQueueOpen, setIsQueueOpen] = useState(false)

  const {
    trackDetail,
    recommended,
    recommendationsPending,
    recommendationsError,
    playerTrack,
    isPlaying,
    streamPending,
    playbackError,
    playbackTime,
    playbackDuration,
    enqueueTrack,
    togglePlayback,
    trackParams,
    navigateToTrack,
    toggleLike,
    seek,
    skipTrack,
  } = useUnit({
    trackDetail: $trackDetail,
    recommended: $recommendedTracks,
    recommendationsPending: $recommendationsPending,
    recommendationsError: $recommendationsError,
    playerTrack: $currentTrack,
    isPlaying: $isPlaying,
    stream: $stream,
    streamPending: $streamPending,
    playbackError: $playbackError,
    playbackTime: $currentTime,
    playbackDuration: $duration,
    enqueueTrack: trackQueued,
    togglePlayback: playbackToggled,
    trackParams: routes.track.$params,
    navigateToTrack: routes.track.navigate,
    toggleLike: trackLikeToggled,
    seek: seekRequested,
    skipTrack: skipTrackRequested,
  })

  if (!trackDetail) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center pb-16 pt-10">
        <p className="text-sm text-muted-foreground">Загрузка трека...</p>
      </div>
    )
  }

  const isActiveTrack = playerTrack?.id === trackDetail.id
  const effectiveDuration =
    isActiveTrack && playbackDuration > 0 ? playbackDuration : trackDetail.duration
  const effectiveTime = isActiveTrack ? playbackTime : 0

  const handleTogglePlay = (track: Track) => {
    if (!playerTrack || playerTrack.id !== track.id) {
      enqueueTrack(track)
      return
    }
    togglePlayback()
  }

  const handleToggleLike = (track: Track) => {
    toggleLike(track.id)
  }

  const handleOpen = (track: Track) => {
    // Если это другой трек, навигируем к нему (страница обновится автоматически)
    // Если это тот же трек, ничего не делаем
    if (track.id !== trackParams?.trackId) {
      navigateToTrack({
        params: { trackId: track.id },
        query: {},
      })
    }
  }

  const handleSeek = (seconds: number) => {
    if (!isActiveTrack) return
    seek(seconds)
  }

  return (
    <>
      <div className="page-container space-y-12 pb-16 pt-10">
        <TrackHero
          track={trackDetail}
          isPlaying={isActiveTrack && isPlaying}
          onTogglePlay={handleTogglePlay}
          onToggleLike={handleToggleLike}
          onShare={track => console.info('Поделиться треком', track.id)}
          onAddToPlaylist={track =>
            console.info('Добавить трек в плейлист', track.id, 'из', trackParams.trackId)
          }
          onGoToArtist={artistId => {
            routes.artist.navigate({
              params: { artistId },
              query: {},
            })
          }}
          onGoToAlbum={albumId => console.info('Открыть альбом', albumId)}
          onSkip={isActiveTrack ? skipTrack : undefined}
          onOpenQueue={() => setIsQueueOpen(true)}
          currentTime={effectiveTime}
          duration={effectiveDuration}
          isBuffering={isActiveTrack && streamPending}
          isSeekEnabled={isActiveTrack && !streamPending}
          onSeek={handleSeek}
        />
        <div className="space-y-4">
          {isActiveTrack && (
            <>
              {streamPending && (
                <div className="rounded-2xl border border-border/60 bg-secondary/20 px-4 py-3 text-sm text-muted-foreground">
                  Подготавливаем поток через CDN...
                </div>
              )}
              {playbackError && (
                <div className="rounded-2xl border border-destructive/40 bg-destructive/10 px-4 py-3 text-sm text-destructive">
                  {playbackError}
                </div>
              )}
            </>
          )}
          {recommendationsPending ? (
            <div className="rounded-2xl border border-border/60 bg-secondary/20 px-4 py-10 text-center text-sm text-muted-foreground">
              Загрузка рекомендаций...
            </div>
          ) : recommendationsError ? (
            <div className="rounded-2xl border border-destructive/40 bg-destructive/10 px-4 py-10 text-center text-sm text-destructive">
              {recommendationsError}
            </div>
          ) : recommended.length === 0 ? (
            <div className="rounded-2xl border border-dashed border-border/60 bg-secondary/30 px-4 py-10 text-center text-sm text-muted-foreground">
              Пока нет рекомендаций
            </div>
          ) : (
            <RecommendedList
              tracks={recommended}
              activeTrackId={playerTrack?.id}
              onPlayToggle={handleTogglePlay}
              onLike={handleToggleLike}
              onShare={track => console.info('Поделиться треком', track.id)}
              onOpen={handleOpen}
            />
          )}
        </div>
      </div>
      <QueueSidebar isOpen={isQueueOpen} onClose={() => setIsQueueOpen(false)} />
    </>
  )
}
