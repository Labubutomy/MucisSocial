import { useState } from 'react'
import { useUnit } from 'effector-react'
import { MiniPlayer } from '@shared/ui/mini-player'
import { QueueSidebar } from '@widgets/queue'
import { $currentTrack, $isPlaying, playbackToggled, skipTrackRequested } from '@features/player'

export interface MiniPlayerControllerProps {
  onOpenTrack: (trackId: string) => void
}

export const MiniPlayerController = ({ onOpenTrack }: MiniPlayerControllerProps) => {
  const [isQueueOpen, setIsQueueOpen] = useState(false)
  const [track, isPlaying, togglePlayback, skipTrack] = useUnit([
    $currentTrack,
    $isPlaying,
    playbackToggled,
    skipTrackRequested,
  ])

  if (!track) return null

  return (
    <>
      <MiniPlayer
        coverUrl={track.coverUrl}
        title={track.title}
        artist={track.artist.name}
        isPlaying={isPlaying}
        onTogglePlay={() => togglePlayback()}
        onOpenTrack={() => onOpenTrack(track.id)}
        onSkip={skipTrack}
        onOpenQueue={() => setIsQueueOpen(true)}
      />
      <QueueSidebar isOpen={isQueueOpen} onClose={() => setIsQueueOpen(false)} />
    </>
  )
}
