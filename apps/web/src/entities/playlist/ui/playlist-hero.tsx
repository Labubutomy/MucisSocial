import type { PlaylistDetail } from '@entities/playlist/model/types'
import { Button } from '@shared/ui/button'
import { IconButton } from '@shared/ui/icon-button'

export interface PlaylistHeroProps {
  playlist: PlaylistDetail
  onPlay: (playlist: PlaylistDetail) => void
  onToggleLike: (playlist: PlaylistDetail) => void
  onShare: (playlist: PlaylistDetail) => void
}

const formatDuration = (seconds?: number) => {
  if (!seconds) return null
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (hours > 0) {
    return `${hours} hr ${minutes} min`
  }
  return `${minutes} min`
}

export const PlaylistHero = ({ playlist, onPlay, onToggleLike, onShare }: PlaylistHeroProps) => {
  const typeTitle = playlist.type === 'album' ? 'Альбом' : 'Плейлист'
  const playAction = playlist.type === 'album' ? 'альбом' : 'плейлист'

  return (
    <section className="flex flex-col gap-8 md:flex-row md:items-end">
      <div className="mx-auto aspect-square w-44 overflow-hidden rounded-3xl shadow-2xl shadow-black/40 md:mx-0 md:w-60">
        <img
          src={playlist.coverUrl}
          alt={playlist.title}
          className="h-full w-full object-cover"
          loading="lazy"
        />
      </div>
      <div className="flex flex-1 flex-col gap-6">
        <div className="flex flex-col gap-2">
          <span className="text-xs font-semibold uppercase tracking-[0.3em] text-primary">
            {typeTitle}
          </span>
          <h1 className="text-3xl font-semibold md:text-5xl">{playlist.title}</h1>
          {playlist.description && (
            <p className="max-w-2xl text-base text-muted-foreground md:text-lg">
              {playlist.description}
            </p>
          )}
        </div>
        <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div className="flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
            <button type="button" className="text-foreground transition hover:text-primary">
              {playlist.owner.name}
            </button>
            <span className="h-1 w-1 rounded-full bg-muted-foreground/60" />
            <span>{playlist.itemsCount} треков</span>
            {playlist.totalDuration && (
              <>
                <span className="h-1 w-1 rounded-full bg-muted-foreground/60" />
                <span>{formatDuration(playlist.totalDuration)}</span>
              </>
            )}
          </div>
          <div className="flex items-center gap-3">
            <Button size="lg" onClick={() => onPlay(playlist)}>
              Воспроизвести {playAction}
            </Button>
            <IconButton
              aria-label={playlist.liked ? 'Убрать из избранного' : 'Добавить в избранное'}
              onClick={() => onToggleLike(playlist)}
              active={playlist.liked}
              size="lg"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                className="h-5 w-5"
                fill={playlist.liked ? 'currentColor' : 'none'}
                stroke="currentColor"
                strokeWidth="1.5"
              >
                <path d="m12 21-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3a4.5 4.5 0 0 1 3.57 1.75A4.5 4.5 0 0 1 14.64 3C17.72 3 20.14 5.42 20.14 8.5c0 3.78-3.4 6.86-8.55 11.18L12 21Z" />
              </svg>
            </IconButton>
            <IconButton
              aria-label="Поделиться плейлистом"
              onClick={() => onShare(playlist)}
              size="lg"
            >
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
        </div>
      </div>
    </section>
  )
}
