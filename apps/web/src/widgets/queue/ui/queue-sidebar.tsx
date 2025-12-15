import { useState, useEffect } from 'react'
import { useUnit } from 'effector-react'
import {
  $queueTrackIds,
  queueCleared,
  trackRemovedFromQueue,
  trackFromQueueSelected,
} from '@features/queue'
import { fetchTrackDetail } from '@entities/track/api'
import type { Track } from '@entities/track'
import { Button } from '@shared/ui/button'
import { IconButton } from '@shared/ui/icon-button'
import { cn } from '@shared/lib/cn'

export interface QueueSidebarProps {
  isOpen: boolean
  onClose: () => void
}

export const QueueSidebar = ({ isOpen, onClose }: QueueSidebarProps) => {
  const queueTrackIds = useUnit($queueTrackIds)
  const clearQueue = useUnit(queueCleared)
  const removeTrack = useUnit(trackRemovedFromQueue)
  const selectTrack = useUnit(trackFromQueueSelected)
  const [tracks, setTracks] = useState<Track[]>([])
  const [loading, setLoading] = useState(false)

  // Загружаем треки при открытии или изменении очереди
  useEffect(() => {
    if (!isOpen) return

    if (queueTrackIds.length > 0) {
      setLoading(true)
      Promise.all(queueTrackIds.map(id => fetchTrackDetail(id)))
        .then(setTracks)
        .catch(console.error)
        .finally(() => setLoading(false))
    } else {
      setTracks([])
      setLoading(false)
    }
  }, [isOpen, queueTrackIds])

  return (
    <>
      {/* Overlay */}
      {isOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm"
          onClick={onClose}
          aria-hidden="true"
        />
      )}

      {/* Sidebar */}
      <div
        className={cn(
          'fixed right-0 top-0 z-50 h-screen w-full max-w-md transform border-l border-border/60 bg-background shadow-2xl transition-transform duration-300 ease-in-out sm:w-96',
          isOpen ? 'translate-x-0' : 'translate-x-full'
        )}
      >
        <div className="flex h-full flex-col">
          {/* Header */}
          <div className="flex items-center justify-between border-b border-border/60 p-4">
            <h2 className="text-lg font-semibold">Очередь воспроизведения</h2>
            <div className="flex items-center gap-2">
              {queueTrackIds.length > 0 && (
                <Button variant="ghost" size="sm" onClick={() => clearQueue()}>
                  Очистить
                </Button>
              )}
              <IconButton aria-label="Закрыть" onClick={onClose} variant="ghost" size="sm">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 24 24"
                  className="h-5 w-5"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <path d="M18 6L6 18M6 6l12 12" />
                </svg>
              </IconButton>
            </div>
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto p-4">
            {loading ? (
              <div className="flex items-center justify-center py-8">
                <p className="text-sm text-muted-foreground">Загрузка очереди...</p>
              </div>
            ) : tracks.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 24 24"
                  className="h-16 w-16 text-muted-foreground/50"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.5"
                >
                  <path d="M4 5h16M4 9h16M4 13h16M4 17h16" />
                </svg>
                <p className="mt-4 text-sm font-medium text-foreground">Очередь пуста</p>
                <p className="mt-2 text-xs text-muted-foreground">
                  Добавьте треки в очередь для воспроизведения
                </p>
              </div>
            ) : (
              <div className="space-y-2">
                {tracks.map((track, index) => (
                  <div
                    key={track.id}
                    className="flex items-center gap-3 rounded-lg border border-border/60 bg-secondary/20 p-3 hover:bg-secondary/40 transition-colors cursor-pointer"
                    onClick={() => selectTrack(track.id)}
                  >
                    <span className="text-sm text-muted-foreground w-6">{index + 1}</span>
                    <div className="relative h-12 w-12 overflow-hidden rounded-lg flex-shrink-0">
                      <img
                        src={track.coverUrl}
                        alt={track.title}
                        className="h-full w-full object-cover"
                        loading="lazy"
                      />
                    </div>
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
            )}
          </div>
        </div>
      </div>
    </>
  )
}
