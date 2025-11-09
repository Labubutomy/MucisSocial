import { useUnit } from 'effector-react'
import { MiniPlayer } from '@shared/ui/mini-player'
import { $currentTrack, $isPlaying, playbackToggled } from '@features/player'

export interface MiniPlayerControllerProps {
  onOpenTrack: (trackId: string) => void
}

export const MiniPlayerController = ({ onOpenTrack }: MiniPlayerControllerProps) => {
  const [track, isPlaying, togglePlayback] = useUnit([$currentTrack, $isPlaying, playbackToggled])

  if (!track) return null

  return (
    <MiniPlayer
      coverUrl={track.coverUrl}
      title={track.title}
      artist={track.artist.name}
      isPlaying={isPlaying}
      onTogglePlay={() => togglePlayback()}
      onOpenTrack={() => onOpenTrack(track.id)}
    />
  )
}
