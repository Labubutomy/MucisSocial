import type { PlaylistSummary } from '@entities/playlist'
import { PlaylistCard } from '@entities/playlist'
import { SectionHeader } from '@shared/ui/section-header'
import { Button } from '@shared/ui/button'

export interface PlaylistGridProps {
  title: string
  playlists: PlaylistSummary[]
  onCreate: () => void
  onOpen: (playlist: PlaylistSummary) => void
}

export const PlaylistGrid = ({ title, playlists, onCreate, onOpen }: PlaylistGridProps) => (
  <section className="space-y-6">
    <SectionHeader
      title={title}
      action={
        <Button variant="outline" onClick={onCreate}>
          Создать плейлист
        </Button>
      }
    />
    <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
      {playlists.map(playlist => (
        <PlaylistCard key={playlist.id} playlist={playlist} onClick={onOpen} />
      ))}
    </div>
  </section>
)
