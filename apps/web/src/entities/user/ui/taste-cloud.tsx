import type { UserProfile } from '@entities/user/model/types'
import { Chip } from '@shared/ui/chip'
import { SectionHeader } from '@shared/ui/section-header'

export interface TasteCloudProps {
  user: UserProfile
  onSelectGenre?: (genre: string) => void
  onSelectArtist?: (artist: string) => void
}

export const TasteCloud = ({ user, onSelectGenre, onSelectArtist }: TasteCloudProps) => (
  <div className="space-y-6 rounded-3xl border border-border/60 bg-secondary/20 p-6 md:p-8">
    <SectionHeader title="Музыкальный ДНК-профиль" subtitle="Ваше актуальное звучание" />
    <div className="space-y-4">
      <div>
        <p className="mb-3 text-xs uppercase tracking-[0.35em] text-muted-foreground">
          Любимые жанры
        </p>
        <div className="flex flex-wrap gap-2">
          {user.musicTasteSummary.topGenres.map(genre => (
            <Chip key={genre} onClick={() => onSelectGenre?.(genre)}>
              {genre}
            </Chip>
          ))}
        </div>
      </div>
      <div>
        <p className="mb-3 text-xs uppercase tracking-[0.35em] text-muted-foreground">
          Любимые артисты
        </p>
        <div className="flex flex-wrap gap-2">
          {user.musicTasteSummary.topArtists.map(artist => (
            <Chip key={artist} onClick={() => onSelectArtist?.(artist)}>
              {artist}
            </Chip>
          ))}
        </div>
      </div>
    </div>
  </div>
)
