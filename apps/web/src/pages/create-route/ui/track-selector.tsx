import { useState, useEffect } from 'react'
import { Card } from '@shared/ui/card'
import { Input } from '@shared/ui/input'
import { Button } from '@shared/ui/button'
import { searchTracks } from '@entities/track/api'
import type { Track } from '@entities/track/model/types'

interface TrackSelectorProps {
  onSelect: (trackId: string) => void
  onClose: () => void
  currentTrackId?: string
}

export const TrackSelector = ({ onSelect, onClose, currentTrackId }: TrackSelectorProps) => {
  const [searchQuery, setSearchQuery] = useState('')
  const [tracks, setTracks] = useState<Track[]>([])
  const [loading, setLoading] = useState(false)

  // Debounce для поиска треков (1 секунда)
  useEffect(() => {
    if (!searchQuery.trim()) {
      setTracks([])
      return
    }

    const timeoutId = setTimeout(async () => {
      setLoading(true)
      try {
        const results = await searchTracks(searchQuery, 20)
        setTracks(results)
      } catch (error) {
        console.error('Failed to search tracks:', error)
        setTracks([])
      } finally {
        setLoading(false)
      }
    }, 1000)

    return () => clearTimeout(timeoutId)
  }, [searchQuery])

  const handleSearch = async () => {
    if (!searchQuery.trim()) return
    setLoading(true)
    try {
      const results = await searchTracks(searchQuery, 20)
      setTracks(results)
    } catch (error) {
      console.error('Failed to search tracks:', error)
      setTracks([])
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <Card padding="lg" className="w-full max-w-2xl max-h-[80vh] flex flex-col">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold">Выберите трек</h2>
          <Button type="button" variant="outline" size="sm" onClick={onClose}>
            Закрыть
          </Button>
        </div>
        <div className="flex gap-2 mb-4">
          <Input
            placeholder="Поиск треков..."
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && handleSearch()}
            className="flex-1"
          />
          <Button type="button" onClick={handleSearch} disabled={loading}>
            {loading ? 'Поиск...' : 'Найти'}
          </Button>
        </div>
        <div className="flex-1 overflow-y-auto space-y-2">
          {tracks.length === 0 && !loading && (
            <p className="text-center text-muted-foreground py-8">
              Введите запрос и нажмите "Найти" для поиска треков
            </p>
          )}
          {tracks.map(track => (
            <button
              key={track.id}
              type="button"
              onClick={() => {
                onSelect(track.id)
                onClose()
              }}
              className={`w-full text-left p-3 rounded-lg border transition hover:bg-secondary/50 ${
                currentTrackId === track.id ? 'border-primary bg-primary/10' : 'border-border/60'
              }`}
            >
              <div className="flex items-center gap-3">
                <div className="h-12 w-12 rounded-lg bg-secondary/50 flex items-center justify-center text-xs text-muted-foreground overflow-hidden">
                  {track.coverUrl ? (
                    <img
                      src={track.coverUrl}
                      alt={track.title}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    '🎵'
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="font-semibold truncate">{track.title}</p>
                  <p className="text-sm text-muted-foreground truncate">{track.artist.name}</p>
                </div>
              </div>
            </button>
          ))}
        </div>
      </Card>
    </div>
  )
}
