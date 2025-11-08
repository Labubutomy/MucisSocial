import { useMemo, useState } from 'react'
import { useUnit } from 'effector-react'
import { SearchBar, SearchHistory, type SearchHistoryItem } from '@features/search'
import { TrendingSidebar, SearchResultsList } from '@widgets/search'
import {
  searchHistory as initialHistory,
  searchResults,
  trendingQueries,
} from '@pages/search/model/data'
import { $currentTrack, $isPlaying, playbackToggled, trackQueued } from '@features/player'
import type { Track } from '@entities/track'
import type { Artist } from '@entities/artist'
import type { PlaylistSummary } from '@entities/playlist'
import { routes } from '@shared/router'

export const SearchPage = () => {
  const [query, setQuery] = useState('')
  const [history, setHistory] = useState<SearchHistoryItem[]>(initialHistory)
  const [likes, setLikes] = useState<Record<string, boolean>>({})

  const [currentTrack, isPlaying, enqueueTrack, togglePlayback, navigateToTrack] = useUnit([
    $currentTrack,
    $isPlaying,
    trackQueued,
    playbackToggled,
    routes.track.navigate,
  ])

  const filteredResults = useMemo(() => {
    if (!query) return searchResults
    const lower = query.toLowerCase()
    return searchResults.filter(result => {
      switch (result.type) {
        case 'track':
          return (
            result.data.title.toLowerCase().includes(lower) ||
            result.data.artist.name.toLowerCase().includes(lower)
          )
        case 'artist':
          return result.data.name.toLowerCase().includes(lower)
        case 'playlist':
          return (
            result.data.title.toLowerCase().includes(lower) ||
            result.data.description?.toLowerCase().includes(lower)
          )
        default:
          return false
      }
    })
  }, [query])

  const handleSearchSubmit: React.FormEventHandler<HTMLFormElement> = event => {
    event.preventDefault()
    if (!query.trim()) return
    const exists = history.some(item => item.query.toLowerCase() === query.toLowerCase())
    if (!exists) {
      setHistory(prev => [
        { id: crypto.randomUUID(), query, createdAt: new Date().toISOString() },
        ...prev,
      ])
    }
  }

  const handleTrackPlayToggle = (track: Track) => {
    if (!currentTrack || currentTrack.id !== track.id) {
      enqueueTrack(track)
      return
    }
    togglePlayback()
  }

  const handleTrackLike = (track: Track) => {
    setLikes(prev => ({ ...prev, [track.id]: !prev[track.id] }))
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

  const likedResults = filteredResults.map(result =>
    result.type === 'track'
      ? {
          ...result,
          data: {
            ...result.data,
            liked: likes[result.data.id] ?? result.data.liked ?? false,
          },
        }
      : result
  )

  const handleHistorySelect = (item: SearchHistoryItem) => {
    setQuery(item.query)
  }

  const handleTrendingSelect = (id: string) => {
    const item = trendingQueries.find(entry => entry.id === id)
    if (item) {
      setQuery(item.label)
    }
  }

  return (
    <div className="page-container flex flex-col gap-8 pb-24 pt-6 lg:flex-row">
      <TrendingSidebar items={trendingQueries} onSelect={handleTrendingSelect} />
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
            onChange={event => setQuery(event.target.value)}
            onSubmit={handleSearchSubmit}
          />
          <div className="flex flex-wrap gap-2 lg:hidden">
            {trendingQueries.map(item => (
              <button
                key={item.id}
                type="button"
                onClick={() => handleTrendingSelect(item.id)}
                className="rounded-full bg-secondary/40 px-3 py-1 text-xs font-semibold text-muted-foreground transition hover:bg-primary/20 hover:text-foreground"
              >
                {item.label}
              </button>
            ))}
          </div>
          <SearchHistory
            items={history}
            onSelect={handleHistorySelect}
            onClearAll={() => setHistory([])}
          />
        </div>

        <SearchResultsList
          results={likedResults}
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
