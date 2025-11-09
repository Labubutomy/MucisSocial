import { memo, type KeyboardEvent, type MouseEvent } from 'react'
import type { Track } from '@entities/track/model/types'
import { cn } from '@shared/lib/cn'
import { Card } from '@shared/ui/card'
import { IconButton } from '@shared/ui/icon-button'

export interface TrackCardProps {
  track: Track
  isPlaying?: boolean
  onPlayToggle: (track: Track) => void
  onLike: (track: Track) => void
  onShare: (track: Track) => void
  onOpen: (track: Track) => void
  className?: string
}

export const TrackCard = memo(
  ({ track, isPlaying, onPlayToggle, onLike, onShare, onOpen, className }: TrackCardProps) => {
    const handleToggle = (event: MouseEvent<HTMLButtonElement>) => {
      event.stopPropagation()
      onPlayToggle(track)
    }

    const handleLike = (event: MouseEvent<HTMLButtonElement>) => {
      event.stopPropagation()
      onLike(track)
    }

    const handleShare = (event: MouseEvent<HTMLButtonElement>) => {
      event.stopPropagation()
      onShare(track)
    }

    const handleOpen = () => onOpen(track)
    const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
      if (event.key === 'Enter' || event.key === ' ') {
        event.preventDefault()
        onOpen(track)
      }
    }

    return (
      <Card
        padding="sm"
        className={cn(
          'group flex flex-col gap-4 bg-card/80 transition hover:bg-card/95',
          className
        )}
      >
        <div
          role="button"
          tabIndex={0}
          onClick={handleOpen}
          onKeyDown={handleKeyDown}
          className="flex flex-col gap-3 text-left cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
        >
          <div className="relative aspect-square overflow-hidden rounded-2xl">
            <img
              src={track.coverUrl}
              alt={track.title}
              className="h-full w-full object-cover transition duration-500 group-hover:scale-105"
              loading="lazy"
            />
            <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-black/10 to-transparent opacity-0 transition group-hover:opacity-100" />
            <button
              type="button"
              onClick={handleToggle}
              aria-label={isPlaying ? 'Поставить трек на паузу' : 'Воспроизвести трек'}
              className="absolute bottom-4 right-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-xl shadow-primary/40 transition hover:scale-105"
            >
              {isPlaying ? (
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 24 24"
                  className="h-6 w-6"
                  fill="currentColor"
                >
                  <path d="M9 5a1 1 0 0 1 1 1v12a1 1 0 1 1-2 0V6a1 1 0 0 1 1-1Zm6 0a1 1 0 0 1 1 1v12a1 1 0 1 1-2 0V6a1 1 0 0 1 1-1Z" />
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
            </button>
          </div>
          <div className="flex flex-col gap-1 px-2 pb-1">
            <span className="text-base font-semibold text-foreground md:text-lg">
              {track.title}
            </span>
            <span className="w-fit text-sm text-muted-foreground transition hover:text-foreground">
              {track.artist.name}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-2 px-2 pb-1">
          <IconButton
            aria-label={track.liked ? 'Убрать из избранного' : 'Добавить в избранное'}
            onClick={handleLike}
            active={track.liked}
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
          <IconButton aria-label="Поделиться треком" onClick={handleShare}>
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
        </div>
      </Card>
    )
  }
)

TrackCard.displayName = 'TrackCard'
