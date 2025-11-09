import type { ChangeEvent } from 'react'
import type { TrackDetail } from '@widgets/track/model/types'
import { IconButton } from '@shared/ui/icon-button'
import { cn } from '@shared/lib/cn'

export interface TrackHeroProps {
  track: TrackDetail
  isPlaying: boolean
  onTogglePlay: (track: TrackDetail) => void
  onToggleLike: (track: TrackDetail) => void
  onShare: (track: TrackDetail) => void
  onAddToPlaylist: (track: TrackDetail) => void
  onGoToArtist: (artistId: string) => void
  onGoToAlbum: (albumId: string) => void
  currentTime: number
  duration: number
  isBuffering?: boolean
  isSeekEnabled?: boolean
  onSeek: (seconds: number) => void
}

const formatDuration = (seconds: number) => {
  const minutes = Math.floor(seconds / 60)
  const remaining = String(Math.floor(seconds % 60)).padStart(2, '0')
  return `${minutes}:${remaining}`
}

export const TrackHero = ({
  track,
  isPlaying,
  onTogglePlay,
  onToggleLike,
  onShare,
  onAddToPlaylist,
  onGoToArtist,
  onGoToAlbum,
  currentTime,
  duration,
  isBuffering = false,
  isSeekEnabled = true,
  onSeek,
}: TrackHeroProps) => {
  const effectiveDuration = Math.max(duration, track.duration ?? 0)
  const safeDuration = effectiveDuration > 0 ? effectiveDuration : (track.duration ?? 0)
  const safeTime = Math.min(Math.max(currentTime, 0), safeDuration)
  const progress = safeDuration > 0 ? safeTime / safeDuration : 0
  const playedSeconds = Math.floor(safeTime)
  const sliderMax = Math.max(1, Math.round(safeDuration || track.duration || 1))

  const handleSeek = (event: ChangeEvent<HTMLInputElement>) => {
    if (!isSeekEnabled) return
    const next = Number(event.target.value)
    if (Number.isFinite(next)) {
      onSeek(next)
    }
  }

  return (
    <section className="grid gap-8 lg:grid-cols-[minmax(0,0.8fr),minmax(0,1fr)]">
      <div className="relative mx-auto aspect-square w-full max-w-sm overflow-hidden rounded-[2.5rem] border border-border/60 bg-secondary/20 p-4 shadow-2xl shadow-black/40 lg:mx-0">
        <div className="overflow-hidden rounded-[2rem]">
          <img
            src={track.coverUrl}
            alt={track.title}
            className="h-full w-full object-cover"
            loading="lazy"
          />
        </div>
        <div className="pointer-events-none absolute inset-0 rounded-[2.5rem] bg-gradient-to-b from-transparent via-transparent to-black/35" />
      </div>

      <div className="flex flex-col gap-8 rounded-3xl border border-border/60 bg-secondary/20 p-6 backdrop-blur md:p-8">
        <div className="space-y-2">
          <span className="text-xs font-semibold uppercase tracking-[0.4em] text-primary">
            Сейчас играет
          </span>
          <h1 className="text-3xl font-semibold md:text-5xl">{track.title}</h1>
          <div className="flex flex-wrap items-center gap-3 text-sm text-muted-foreground md:text-base">
            <button
              type="button"
              onClick={() => onGoToArtist(track.artist.id)}
              className="text-foreground transition hover:text-primary"
            >
              {track.artist.name}
            </button>
            <span className="h-1 w-1 rounded-full bg-muted-foreground/50" />
            <button
              type="button"
              onClick={() => onGoToAlbum(track.album.id)}
              className="text-foreground transition hover:text-primary"
            >
              {track.album.title}
            </button>
          </div>
        </div>

        <div className="space-y-4">
          <div className="relative h-2 rounded-full bg-muted/50">
            <div
              className={cn('absolute inset-y-0 left-0 rounded-full bg-primary transition-[width]')}
              style={{ width: `${progress * 100}%` }}
            />
            <input
              type="range"
              min={0}
              max={sliderMax}
              step={1}
              value={Math.floor(safeTime)}
              onChange={handleSeek}
              disabled={!isSeekEnabled}
              aria-label="Позиция трека"
              className="absolute inset-0 h-2 w-full cursor-pointer appearance-none bg-transparent"
            />
          </div>
          <div className="flex items-center justify-between text-xs font-medium uppercase tracking-wider text-muted-foreground">
            <span>{formatDuration(playedSeconds)}</span>
            <span>{formatDuration(Math.floor(safeDuration || track.duration))}</span>
          </div>
          {isBuffering && (
            <p className="text-xs text-muted-foreground/80">Буферизация потока через CDN…</p>
          )}
        </div>

        <div className="flex flex-wrap items-center gap-4">
          <IconButton
            size="lg"
            variant="muted"
            onClick={() => onTogglePlay(track)}
            aria-label={isPlaying ? 'Остановить воспроизведение' : 'Включить трек'}
          >
            {isPlaying ? (
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                className="h-6 w-6"
                fill="currentColor"
              >
                <rect x="6" y="6" width="12" height="12" rx="2" />
              </svg>
            ) : (
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                className="h-6 w-6"
                fill="currentColor"
              >
                <path d="M5 5.868a1 1 0 0 1 1.52-.854l12 6.132a1 1 0 0 1 0 1.708l-12 6.132A1 1 0 0 1 5 18.132V5.868Z" />
              </svg>
            )}
          </IconButton>
          <IconButton
            size="lg"
            active={track.liked}
            onClick={() => onToggleLike(track)}
            aria-label={track.liked ? 'Убрать из избранного' : 'Добавить в избранное'}
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              className="h-5 w-5"
              fill={track.liked ? 'currentColor' : 'none'}
              stroke="currentColor"
              strokeWidth="1.5"
            >
              <path d="m12 21-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3a4.5 4.5 0 0 1 3.57 1.75A4.5 4.5 0 0 1 14.64 3C17.72 3 20.14 5.42 20.14 8.5c0 3.78-3.4 6.86-8.55 11.18L12 21Z" />
            </svg>
          </IconButton>
          <IconButton size="lg" onClick={() => onShare(track)} aria-label="Поделиться треком">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              className="h-5 w-5"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
            >
              <path d="M4 12v7a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-7" />
              <path d="M16 6 12 2 8 6" />
              <path d="M12 2v14" />
            </svg>
          </IconButton>
          <IconButton
            size="lg"
            onClick={() => onAddToPlaylist(track)}
            aria-label="Добавить трек в плейлист"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              className="h-5 w-5"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
            >
              <path d="M4 5h16" />
              <path d="M4 9h16" />
              <path d="M4 19h10" />
              <path d="M16 15h4v4h-4z" />
            </svg>
          </IconButton>
        </div>
      </div>
    </section>
  )
}
