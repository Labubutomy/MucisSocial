import type { MouseEventHandler } from 'react'
import { cn } from '@shared/lib/cn'
import { IconButton } from '@shared/ui/icon-button'

export interface MiniPlayerProps {
  coverUrl: string
  title: string
  artist: string
  isPlaying: boolean
  onTogglePlay: MouseEventHandler<HTMLButtonElement>
  onOpenTrack: () => void
  onSkip?: () => void
  onOpenQueue?: () => void
  className?: string
}

export const MiniPlayer = ({
  coverUrl,
  title,
  artist,
  isPlaying,
  onTogglePlay,
  onOpenTrack,
  onSkip,
  onOpenQueue,
  className,
}: MiniPlayerProps) => (
  <div
    className={cn(
      'flex items-center justify-between gap-4 rounded-2xl border border-border/60 bg-secondary/60 p-3 backdrop-blur-md shadow-inner shadow-black/30',
      className
    )}
  >
    <button
      type="button"
      onClick={onOpenTrack}
      className="flex flex-1 items-center gap-3 text-left"
    >
      <div className="relative h-14 w-14 overflow-hidden rounded-xl">
        <img src={coverUrl} alt={title} className="h-full w-full object-cover" loading="lazy" />
      </div>
      <div className="flex flex-1 flex-col">
        <span className="truncate text-sm font-semibold text-foreground">{title}</span>
        <span className="truncate text-xs text-muted-foreground">{artist}</span>
      </div>
    </button>
    <div className="flex items-center gap-2">
      {onOpenQueue && (
        <IconButton aria-label="Открыть очередь" onClick={onOpenQueue} variant="muted" size="sm">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            className="h-5 w-5"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <path d="M4 5h16M4 9h16M4 13h16M4 17h16" />
          </svg>
        </IconButton>
      )}
      {onSkip && (
        <IconButton aria-label="Пропустить трек" onClick={onSkip} variant="muted" size="sm">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            className="h-5 w-5"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <path d="m5 4 10 8-10 8V4Z" />
            <path d="M19 5v14" />
          </svg>
        </IconButton>
      )}
      <IconButton
        aria-label={isPlaying ? 'Поставить трек на паузу' : 'Воспроизвести трек'}
        onClick={onTogglePlay}
        variant="muted"
        size="lg"
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
      </IconButton>
    </div>
  </div>
)
