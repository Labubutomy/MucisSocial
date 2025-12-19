import { useUnit } from 'effector-react'
import { Avatar } from '@shared/ui/avatar'
import { Card } from '@shared/ui/card'
import { TrackRow } from '@entities/track'
import type { Track } from '@entities/track'
import { routes } from '@shared/router'
import { $artistDetail, $artistTracks, $artistPending, $tracksPending } from '../model'
import { $currentTrack, $isPlaying, playbackToggled, trackQueued } from '@features/player'
import { trackLikeToggled } from '@pages/track/model'

export const ArtistPage = () => {
  const {
    artist,
    tracks,
    artistPending,
    tracksPending,
    currentTrack,
    isPlaying,
    navigateToTrack,
    togglePlayback,
    enqueueTrack,
    toggleLike,
  } = useUnit({
    artist: $artistDetail,
    tracks: $artistTracks,
    artistPending: $artistPending,
    tracksPending: $tracksPending,
    currentTrack: $currentTrack,
    isPlaying: $isPlaying,
    navigateToTrack: routes.track.navigate,
    togglePlayback: playbackToggled,
    enqueueTrack: trackQueued,
    toggleLike: trackLikeToggled,
  })

  if (artistPending) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center pb-20 pt-10">
        <p className="text-sm text-muted-foreground">Загрузка артиста...</p>
      </div>
    )
  }

  if (!artist) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center pb-20 pt-10">
        <p className="text-sm text-muted-foreground">Артист не найден</p>
      </div>
    )
  }

  const handleTogglePlay = (track: Track) => {
    if (!currentTrack || currentTrack.id !== track.id) {
      enqueueTrack(track)
      return
    }
    togglePlayback()
  }

  const handleToggleLike = (track: Track) => {
    toggleLike(track.id)
  }

  const handleOpen = (track: Track) => {
    navigateToTrack({
      params: { trackId: track.id },
      query: {},
    })
  }

  return (
    <div className="page-container space-y-8 pb-20 pt-10">
      <div className="flex flex-col gap-6 sm:flex-row sm:items-end">
        <Avatar
          src={artist.avatarUrl}
          fallback={artist.name}
          size="xl"
          className="h-32 w-32 flex-shrink-0 sm:h-40 sm:w-40"
        />
        <div className="flex flex-col gap-2">
          <h1 className="text-3xl font-semibold sm:text-4xl md:text-5xl">{artist.name}</h1>
          {artist.genres && artist.genres.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {artist.genres.map(genre => (
                <span
                  key={genre}
                  className="rounded-full border border-border/60 bg-secondary/30 px-3 py-1 text-xs text-muted-foreground"
                >
                  {genre}
                </span>
              ))}
            </div>
          )}
          {artist.followers !== undefined && (
            <p className="text-sm text-muted-foreground">
              {artist.followers.toLocaleString()} подписчиков
            </p>
          )}
        </div>
      </div>

      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold">Треки</h2>
          {tracksPending && <p className="text-sm text-muted-foreground">Загрузка...</p>}
        </div>

        {tracks.length === 0 && !tracksPending ? (
          <Card padding="lg" className="text-center">
            <p className="text-sm text-muted-foreground">У артиста пока нет треков</p>
          </Card>
        ) : (
          <div className="space-y-2">
            {tracks.map((track, index) => (
              <TrackRow
                key={track.id}
                track={track}
                index={index}
                isPlaying={currentTrack?.id === track.id && isPlaying}
                onPlayToggle={handleTogglePlay}
                onLike={handleToggleLike}
                onShare={track => console.info('Поделиться треком', track.id)}
                onOpen={handleOpen}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
