import type { FormEventHandler } from 'react'
import { useUnit } from 'effector-react'
import { SearchBar, SearchHistory } from '@features/search'
import { TrendingSidebar, SearchResultsList } from '@widgets/search'
import { $currentTrack, $isPlaying, playbackToggled, trackQueued } from '@features/player'
import type { Track } from '@entities/track'
import type { Artist } from '@entities/artist'
import type { PlaylistSummary } from '@entities/playlist'
import { routes } from '@shared/router'
import {
  $history,
  $query,
  $results,
  $trending,
  historyCleared,
  historyItemSelected,
  queryChanged,
  searchSubmitted,
  trackLikeToggled,
  trendingSelected,
} from '@pages/search/model'

export const SearchPage = () => {
  const {
    query,
    history,
    results,
    trending,
    currentTrack,
    isPlaying,
    enqueueTrack,
    togglePlayback,
    navigateToTrack,
    changeQuery,
    submitSearch,
    selectHistory,
    clearHistory,
    selectTrending,
    toggleTrackLike,
  } = useUnit({
    query: $query,
    history: $history,
    results: $results,
    trending: $trending,
    currentTrack: $currentTrack,
    isPlaying: $isPlaying,
    enqueueTrack: trackQueued,
    togglePlayback: playbackToggled,
    navigateToTrack: routes.track.navigate,
    changeQuery: queryChanged,
    submitSearch: searchSubmitted,
    selectHistory: historyItemSelected,
    clearHistory: historyCleared,
    selectTrending: trendingSelected,
    toggleTrackLike: trackLikeToggled,
  })

  const handleSearchSubmit: FormEventHandler<HTMLFormElement> = event => {
    event.preventDefault()
    submitSearch()
  }

  const handleTrackPlayToggle = (track: Track) => {
    if (!currentTrack || currentTrack.id !== track.id) {
      enqueueTrack(track)
      return
    }
    togglePlayback()
  }

  const handleTrackLike = (track: Track) => {
    toggleTrackLike(track.id)
  }

  const handlePlaylistOpen = (playlist: PlaylistSummary) => {
    console.info('Открыть плейлист', playlist.id)
  }

  const handleArtistOpen = (artist: Artist) => {
    console.info('Открыть артиста', artist.id)
  }

  const handleTrackOpen = (track: Track) => {
    navigateToTrack({
      params: { trackId: track.id },
      query: {},
    })
  }

  const handleTrendingSelect = (item: { id: string; label: string }) => {
    selectTrending(item)
  }

  return (
    <div className="page-container flex flex-col gap-8 pb-24 pt-6 lg:flex-row">
      <TrendingSidebar items={trending} onSelect={handleTrendingSelect} />
      <div className="flex-1 space-y-8">
        <div className="space-y-5 rounded-3xl border border-border/60 bg-secondary/20 p-5 shadow-inner shadow-black/20 lg:p-8">
          <div className="space-y-2">
            <p className="text-xs uppercase tracking-[0.4em] text-primary">Поиск</p>
            <h1 className="text-3xl font-semibold md:text-4xl">
              Находите треки, артистов и плейлисты в один клик
            </h1>
          </div>
          <SearchBar
            value={query}
            onChange={event => changeQuery(event.target.value)}
            onSubmit={handleSearchSubmit}
          />
          <div className="flex flex-wrap gap-2 lg:hidden">
            {trending.map(item => (
              <button
                key={item.id}
                type="button"
                onClick={() => handleTrendingSelect(item)}
                className="rounded-full bg-secondary/40 px-3 py-1 text-xs font-semibold text-muted-foreground transition hover:bg-primary/20 hover:text-foreground"
              >
                {item.label}
              </button>
            ))}
          </div>
          <SearchHistory items={history} onSelect={selectHistory} onClearAll={clearHistory} />
        </div>

        <SearchResultsList
          results={results}
          activeTrackId={isPlaying ? currentTrack?.id : undefined}
          onTrackPlayToggle={handleTrackPlayToggle}
          onTrackLike={handleTrackLike}
          onTrackAdd={track => console.info('Добавить в плейлист', track.id)}
          onTrackShare={track => console.info('Поделиться треком', track.id)}
          onTrackOpen={handleTrackOpen}
          onArtistOpen={handleArtistOpen}
          onPlaylistOpen={handlePlaylistOpen}
        />
      </div>
    </div>
  )
}
