import type { Track } from '@entities/track'
import { TrackCard } from '@entities/track'
import { SectionHeader } from '@shared/ui/section-header'

export interface TrackFeedProps {
  title: string
  subtitle?: string
  tracks: Track[]
  activeTrackId?: string
  onPlayToggle: (track: Track) => void
  onLike: (track: Track) => void
  onShare: (track: Track) => void
  onOpen: (track: Track) => void
}

export const TrackFeed = ({
  title,
  subtitle,
  tracks,
  activeTrackId,
  onPlayToggle,
  onLike,
  onShare,
  onOpen,
}: TrackFeedProps) => (
  <section className="space-y-6">
    <SectionHeader title={title} subtitle={subtitle} />
    <div className="grid gap-5 sm:grid-cols-2 md:grid-cols-3">
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
