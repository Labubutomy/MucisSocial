import { useUnit } from 'effector-react'
import { useState, useEffect } from 'react'
import {
  $queueTrackIds,
  trackRemovedFromQueue,
  queueCleared,
  trackFromQueueSelected,
  $queueContext,
} from '@features/queue'
import { fetchTrackDetail } from '@entities/track/api'
import type { Track } from '@entities/track'
import { Button } from '@shared/ui/button'
import { IconButton } from '@shared/ui/icon-button'

export const QueueList = () => {
  const queueTrackIds = useUnit($queueTrackIds)
  const removeTrack = useUnit(trackRemovedFromQueue)
  const clearQueue = useUnit(queueCleared)
  const selectTrackFromQueue = useUnit(trackFromQueueSelected)
  const queueContext = useUnit($queueContext)
  const [tracks, setTracks] = useState<Track[]>([])
  const [loading, setLoading] = useState(false)

  const isSessionContext = queueContext?.type === 'session'

  useEffect(() => {
    if (queueTrackIds.length === 0) {
      setTracks([])
      return
    }

    setLoading(true)
    Promise.all(queueTrackIds.map(id => fetchTrackDetail(id)))
      .then(setTracks)
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [queueTrackIds])

  if (loading) {
    return (
      <div className="space-y-2">
        <p className="text-sm text-muted-foreground">Загрузка очереди...</p>
      </div>
    )
  }

  if (tracks.length === 0) {
    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold">Очередь воспроизведения</h3>
        </div>
        <p className="text-sm text-muted-foreground">Очередь пуста</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Очередь воспроизведения</h3>
        <Button variant="ghost" size="sm" onClick={() => clearQueue()}>
          Очистить
        </Button>
      </div>
      <div className="space-y-2">
        {tracks.map((track, index) => (
          <div
            key={track.id}
            className={`flex items-center gap-3 rounded-lg border border-border/60 bg-secondary/20 p-3 transition-colors ${
              isSessionContext ? 'hover:bg-secondary/40 cursor-pointer' : 'hover:bg-secondary/40'
            }`}
            onClick={
              isSessionContext
                ? () => {
                    // In session, clicking a track should play it and remove all tracks before it
                    selectTrackFromQueue(track.id)
                  }
                : undefined
            }
          >
            <span className="text-sm text-muted-foreground w-6">{index + 1}</span>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium truncate">{track.title}</p>
              <p className="text-xs text-muted-foreground truncate">{track.artist.name}</p>
            </div>
            <IconButton
              aria-label="Удалить из очереди"
              onClick={e => {
                e.stopPropagation()
                removeTrack(track.id)
              }}
              variant="ghost"
              size="sm"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                className="h-4 w-4"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path d="M18 6L6 18M6 6l12 12" />
              </svg>
            </IconButton>
          </div>
        ))}
      </div>
    </div>
  )
}
