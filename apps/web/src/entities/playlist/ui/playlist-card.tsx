import type { PlaylistSummary } from '@entities/playlist/model/types'
import { Card } from '@shared/ui/card'
import { cn } from '@shared/lib/cn'

export interface PlaylistCardProps {
  playlist: PlaylistSummary
  onClick: (playlist: PlaylistSummary) => void
  className?: string
}

export const PlaylistCard = ({ playlist, onClick, className }: PlaylistCardProps) => (
  <button type="button" onClick={() => onClick(playlist)} className={cn('text-left', className)}>
    <Card padding="sm" className="flex flex-col gap-4 hover:bg-card/95">
      <div className="relative aspect-square w-full overflow-hidden rounded-2xl">
        <img
          src={playlist.coverUrl}
          alt={playlist.title}
          className="h-full w-full object-cover transition duration-700 hover:scale-105"
          loading="lazy"
        />
      </div>
      <div className="flex flex-col gap-2 px-2 pb-2">
        <span className="text-base font-semibold text-foreground md:text-lg">{playlist.title}</span>
        {playlist.description && (
          <p className="line-clamp-2 text-sm text-muted-foreground">{playlist.description}</p>
        )}
        <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
          {playlist.itemsCount ?? 0} треков
        </span>
      </div>
    </Card>
  </button>
)
