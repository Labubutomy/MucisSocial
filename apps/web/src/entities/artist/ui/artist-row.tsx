import type { Artist } from '@entities/artist/model/types'
import { Avatar } from '@shared/ui/avatar'
import { cn } from '@shared/lib/cn'

export interface ArtistRowProps {
  artist: Artist
  onOpen: (artist: Artist) => void
  className?: string
}

export const ArtistRow = ({ artist, onOpen, className }: ArtistRowProps) => (
  <button
    type="button"
    onClick={() => onOpen(artist)}
    className={cn(
      'flex w-full items-center gap-4 rounded-2xl border border-transparent bg-secondary/30 p-3 text-left transition hover:border-border/60 hover:bg-secondary/60',
      className
    )}
  >
    <Avatar src={artist.avatarUrl} fallback={artist.name} size="sm" className="flex-shrink-0" />
    <div className="flex flex-col">
      <span className="text-sm font-semibold text-foreground md:text-base">{artist.name}</span>
      {artist.genres && artist.genres.length > 0 && (
        <span className="text-xs text-muted-foreground">
          {artist.genres.slice(0, 3).join(' • ')}
        </span>
      )}
    </div>
  </button>
)
