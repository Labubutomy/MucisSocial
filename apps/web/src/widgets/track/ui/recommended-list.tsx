import type { Track } from '@entities/track'
import { TrackCard } from '@entities/track'
import { SectionHeader } from '@shared/ui/section-header'

export interface RecommendedListProps {
  tracks: Track[]
  onPlayToggle: (track: Track) => void
  onLike: (track: Track) => void
  onShare: (track: Track) => void
  onOpen: (track: Track) => void
  activeTrackId?: string
}

export const RecommendedList = ({
  tracks,
  onPlayToggle,
  onLike,
  onShare,
  onOpen,
  activeTrackId,
}: RecommendedListProps) => (
  <section className="space-y-6">
    <SectionHeader title="Рекомендации для вас" />
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {tracks.map(track => (
        <TrackCard
          key={track.id}
          track={track}
          isPlaying={track.id === activeTrackId}
          onPlayToggle={onPlayToggle}
          onLike={onLike}
          onShare={onShare}
          onOpen={onOpen}
        />
      ))}
    </div>
  </section>
)
