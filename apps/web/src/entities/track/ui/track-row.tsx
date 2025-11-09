import { memo, type KeyboardEvent, type MouseEvent } from 'react'
import type { Track } from '@entities/track/model/types'
import { cn } from '@shared/lib/cn'
import { IconButton } from '@shared/ui/icon-button'

export interface TrackRowProps {
  track: Track
  index?: number
  isPlaying?: boolean
  onPlayToggle: (track: Track) => void
  onLike: (track: Track) => void
  onAddToPlaylist?: (track: Track) => void
  onShare?: (track: Track) => void
  onOpen?: (track: Track) => void
  className?: string
}

export const TrackRow = memo(
  ({
    track,
    index,
    isPlaying,
    onPlayToggle,
    onLike,
    onAddToPlaylist,
    onShare,
    onOpen,
    className,
  }: TrackRowProps) => {
    const handlePlayToggle = (event: MouseEvent<HTMLButtonElement>) => {
      event.stopPropagation()
      onPlayToggle(track)
    }
    const handleLike = (event: MouseEvent<HTMLButtonElement>) => {
      event.stopPropagation()
      onLike(track)
    }
    const handleAdd = (event: MouseEvent<HTMLButtonElement>) => {
      event.stopPropagation()
      onAddToPlaylist?.(track)
    }
    const handleShare = (event: MouseEvent<HTMLButtonElement>) => {
      event.stopPropagation()
      onShare?.(track)
    }

    const handleOpen = () => onOpen?.(track)
    const handleOpenKey = (event: KeyboardEvent<HTMLButtonElement>) => {
      if (event.key === 'Enter' || event.key === ' ') {
        event.preventDefault()
        onOpen?.(track)
      }
    }

    const minutes = Math.floor(track.duration / 60)
    const seconds = String(track.duration % 60).padStart(2, '0')

    return (
      <div
        className={cn(
          'grid grid-cols-[auto,1.5fr,1fr,auto] items-center gap-3 rounded-xl border border-transparent bg-transparent px-3 py-4 text-sm transition hover:border-border/60 hover:bg-secondary/30 md:grid-cols-[auto,2fr,1fr,auto]',
          className
        )}
      >
        <button
          type="button"
          onClick={handlePlayToggle}
          aria-label={isPlaying ? 'Поставить трек на паузу' : 'Воспроизвести трек'}
          className="flex h-10 w-10 items-center justify-center rounded-full bg-muted/40 text-muted-foreground transition hover:bg-primary hover:text-primary-foreground"
        >
          {isPlaying ? (
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              className="h-4 w-4"
              fill="currentColor"
            >
              <path d="M9 5a1 1 0 0 1 1 1v12a1 1 0 1 1-2 0V6a1 1 0 0 1 1-1Zm6 0a1 1 0 0 1 1 1v12a1 1 0 1 1-2 0V6a1 1 0 0 1 1-1Z" />
            </svg>
          ) : (
            <>
              {typeof index === 'number' ? (
                <span className="text-xs font-semibold">{index + 1}</span>
              ) : (
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 24 24"
                  className="h-4 w-4"
                  fill="currentColor"
                >
                  <path d="M5 5.868a1 1 0 0 1 1.52-.854l12 6.132a1 1 0 0 1 0 1.708l-12 6.132A1 1 0 0 1 5 18.132V5.868Z" />
                </svg>
              )}
            </>
          )}
        </button>
        <button
          type="button"
          onClick={handleOpen}
          onKeyDown={handleOpenKey}
          disabled={!onOpen}
          className={cn(
            'flex items-center gap-3 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background',
            !onOpen && 'cursor-default'
          )}
        >
          <div className="relative aspect-square h-12 w-12 overflow-hidden rounded-lg">
            <img
              src={track.coverUrl}
              alt={track.title}
              className="h-full w-full object-cover"
              loading="lazy"
            />
          </div>
          <div className="flex flex-col">
            <span className="text-sm font-semibold text-foreground md:text-base">
              {track.title}
            </span>
            <span className="w-fit text-xs text-muted-foreground transition hover:text-foreground">
              {track.artist.name}
            </span>
          </div>
        </button>
        <div className="hidden items-center justify-end text-xs text-muted-foreground md:flex">
          {minutes}:{seconds}
        </div>
        <div className="flex items-center justify-end gap-1">
          <IconButton
            aria-label={track.liked ? 'Убрать из избранного' : 'Добавить в избранное'}
            onClick={handleLike}
            active={track.liked}
            size="sm"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              className="h-4 w-4"
              fill={track.liked ? 'currentColor' : 'none'}
              stroke="currentColor"
              strokeWidth="1.5"
            >
              <path d="m12 21-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3a4.5 4.5 0 0 1 3.57 1.75A4.5 4.5 0 0 1 14.64 3C17.72 3 20.14 5.42 20.14 8.5c0 3.78-3.4 6.86-8.55 11.18L12 21Z" />
            </svg>
          </IconButton>
          {onAddToPlaylist && (
            <IconButton aria-label="Добавить в плейлист" onClick={handleAdd} size="sm">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                className="h-4 w-4"
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
          )}
          {onShare && (
            <IconButton aria-label="Поделиться треком" onClick={handleShare} size="sm">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                className="h-4 w-4"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.5"
              >
                <path d="M4 12v7a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-7" />
                <path d="M16 6 12 2 8 6" />
                <path d="M12 2v14" />
              </svg>
            </IconButton>
          )}
        </div>
      </div>
    )
  }
)

TrackRow.displayName = 'TrackRow'
