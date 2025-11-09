import type { Artist } from '@entities/artist'
import type { PlaylistSummary } from '@entities/playlist'
import type { Track } from '@entities/track'
import type { SearchResult } from '@widgets/search/model/types'
import { ArtistRow } from '@entities/artist'
import { TrackRow } from '@entities/track'
import { Card } from '@shared/ui/card'
import { SectionHeader } from '@shared/ui/section-header'

export interface SearchResultsListProps {
  results: SearchResult[]
  activeTrackId?: string
  onTrackPlayToggle: (track: Track) => void
  onTrackLike: (track: Track) => void
  onTrackAdd?: (track: Track) => void
  onTrackShare?: (track: Track) => void
  onTrackOpen: (track: Track) => void
  onArtistOpen: (artist: Artist) => void
  onPlaylistOpen: (playlist: PlaylistSummary) => void
}

export const SearchResultsList = ({
  results,
  activeTrackId,
  onTrackPlayToggle,
  onTrackLike,
  onTrackAdd,
  onTrackShare,
  onTrackOpen,
  onArtistOpen,
  onPlaylistOpen,
}: SearchResultsListProps) => (
  <section className="space-y-6">
    <SectionHeader
      title="Результаты поиска"
      subtitle="Результаты поиска из треков, артистов и плейлистов"
    />
    <div className="space-y-4">
      {results.map((result, index) => {
        switch (result.type) {
          case 'track':
            return (
              <TrackRow
                key={`track-${result.data.id}`}
                track={result.data}
                index={index}
                isPlaying={result.data.id === activeTrackId}
                onPlayToggle={onTrackPlayToggle}
                onLike={onTrackLike}
                onAddToPlaylist={onTrackAdd}
                onShare={onTrackShare}
                onOpen={onTrackOpen}
              />
            )
          case 'artist':
            return (
              <ArtistRow
                key={`artist-${result.data.id}`}
                artist={result.data}
                onOpen={onArtistOpen}
                className="bg-secondary/40"
              />
            )
          case 'playlist':
            return (
              <Card
                key={`playlist-${result.data.id}`}
                padding="sm"
                className="flex items-center justify-between gap-4 bg-secondary/30 hover:bg-secondary/50"
              >
                <button
                  type="button"
                  className="flex flex-1 items-center gap-4 text-left"
                  onClick={() => onPlaylistOpen(result.data)}
                >
                  <div className="h-16 w-16 overflow-hidden rounded-xl">
                    <img
                      src={result.data.coverUrl}
                      alt={result.data.title}
                      className="h-full w-full object-cover"
                      loading="lazy"
                    />
                  </div>
                  <div className="flex flex-col">
                    <span className="text-base font-semibold text-foreground">
                      {result.data.title}
                    </span>
                    {result.data.description && (
                      <span className="text-sm text-muted-foreground">
                        {result.data.description}
                      </span>
                    )}
                    <span className="text-xs text-muted-foreground/80">
                      {result.data.itemsCount} треков
                    </span>
                  </div>
                </button>
                <button
                  type="button"
                  className="rounded-full border border-primary px-4 py-2 text-sm font-semibold text-primary transition hover:bg-primary hover:text-primary-foreground"
                  onClick={() => onPlaylistOpen(result.data)}
                >
                  Посмотреть
                </button>
              </Card>
            )
          default:
            return null
        }
      })}
    </div>
  </section>
)
