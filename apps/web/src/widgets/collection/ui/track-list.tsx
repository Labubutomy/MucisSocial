import type { Track } from '@entities/track'
import { TrackRow } from '@entities/track'
import { SectionHeader } from '@shared/ui/section-header'

export interface CollectionTrackListProps {
  tracks: Track[]
  onPlayToggle: (track: Track) => void
  onLike: (track: Track) => void
  onAddToPlaylist?: (track: Track) => void
  onShare?: (track: Track) => void
  onOpen?: (track: Track) => void
  activeTrackId?: string
}

export const CollectionTrackList = ({
  tracks,
  onPlayToggle,
  onLike,
  onAddToPlaylist,
  onShare,
  onOpen,
  activeTrackId,
}: CollectionTrackListProps) => (
  <section className="space-y-6">
    <SectionHeader title="Список треков" />
    <div className="space-y-2">
      {tracks.map((track, index) => (
        <TrackRow
          key={track.id}
          track={track}
          index={index}
          isPlaying={track.id === activeTrackId}
          onPlayToggle={onPlayToggle}
          onLike={onLike}
          onAddToPlaylist={onAddToPlaylist}
          onShare={onShare}
          onOpen={onOpen}
        />
      ))}
    </div>
  </section>
)
