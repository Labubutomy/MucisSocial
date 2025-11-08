import { createEvent, createStore } from 'effector'
import type { Track } from '@entities/track'

export const trackQueued = createEvent<Track>()
export const playbackToggled = createEvent()
export const playbackStopped = createEvent()

export const $currentTrack = createStore<Track | null>(null)
  .on(trackQueued, (_, track) => track)
  .reset(playbackStopped)

export const $isPlaying = createStore(false)
  .on(trackQueued, () => true)
  .on(playbackToggled, state => !state)
  .reset(playbackStopped)
