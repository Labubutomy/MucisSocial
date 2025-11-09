import { createEffect, createEvent, createStore, sample } from 'effector'
import type { Track } from '@entities/track'
import { fetchStreamMetadata, type StreamMetadata } from '@features/player/api'
import {
  attachListener,
  pauseStream,
  playStream,
  resumeStream,
  stopStream,
  getCurrentTime,
  getDuration,
  seekTo,
} from '@features/player/lib/audio'

export const trackQueued = createEvent<Track>()
export const playbackToggled = createEvent()
export const playbackStopped = createEvent()
export const playbackFailed = createEvent<string>()
export const seekRequested = createEvent<number>()

const playbackTimeChanged = createEvent<number>()
const playbackDurationChanged = createEvent<number>()

const parseBitrates = (qualities?: string[]) => {
  if (!qualities) return undefined
  const values = Array.from(
    new Set(
      qualities
        .map(item => {
          const digits = item.replace(/\D/g, '')
          if (!digits) return null
          const numeric = Number(digits)
          if (!numeric) return null
          return numeric < 1000 ? numeric * 1000 : numeric
        })
        .filter((value): value is number => Boolean(value))
    )
  )
  return values.length ? values : undefined
}

const fetchStreamFx = createEffect(async ({ track }: { track: Track }) => {
  const bitrates = parseBitrates(track.stream?.qualities)
  const metadata = await fetchStreamMetadata({
    trackId: track.id,
    artistId: track.artist.id,
    bitrates,
  })
  return {
    trackId: track.id,
    metadata,
  }
})

const startPlaybackFx = createEffect(async (masterUrl: string) => {
  await playStream(masterUrl)
})

const pausePlaybackFx = createEffect(async () => {
  await pauseStream()
})

const resumePlaybackFx = createEffect(async () => {
  await resumeStream()
})

const stopPlaybackFx = createEffect(async () => {
  await stopStream()
})

const seekPlaybackFx = createEffect(async (seconds: number) => {
  await seekTo(seconds)
})

export const $currentTrack = createStore<Track | null>(null)
  .on(trackQueued, (_, track) => track)
  .reset(playbackStopped)

interface StreamState extends StreamMetadata {
  trackId: string
  fetchedAt: number
  expiresAt: number
}

export const $stream = createStore<StreamState | null>(null)
  .on(trackQueued, () => null)
  .on(fetchStreamFx.doneData, (_, payload) => ({
    trackId: payload.trackId,
    masterUrl: payload.metadata.masterUrl,
    variants: payload.metadata.variants,
    expiresIn: payload.metadata.expiresIn,
    fetchedAt: Date.now(),
    expiresAt: Date.now() + payload.metadata.expiresIn * 1000,
  }))
  .reset(playbackStopped)

export const $isPlaying = createStore(false)
  .on(startPlaybackFx.done, () => true)
  .on(resumePlaybackFx.done, () => true)
  .on(pausePlaybackFx.done, () => false)
  .on(stopPlaybackFx.done, () => false)
  .on(fetchStreamFx.fail, () => false)
  .reset(playbackStopped)

export const $playbackError = createStore<string | null>(null)
  .on(playbackFailed, (_, message) => message)
  .reset([trackQueued, fetchStreamFx.done, playbackStopped])

export const $streamPending = fetchStreamFx.pending

export const $currentTime = createStore(0)
  .on(playbackTimeChanged, (_, time) => time)
  .on(trackQueued, () => 0)
  .reset(playbackStopped)

export const $duration = createStore(0)
  .on(playbackDurationChanged, (_, duration) => duration)
  .on(trackQueued, (_, track) => track.duration ?? 0)
  .reset(playbackStopped)

sample({
  clock: trackQueued,
  filter: track => Boolean(track.artist?.id),
  fn: track => ({ track }),
  target: fetchStreamFx,
})

sample({
  clock: trackQueued,
  source: $currentTrack,
  filter: current => Boolean(current),
  target: stopPlaybackFx,
})

sample({
  clock: fetchStreamFx.doneData,
  fn: ({ metadata }) => metadata.masterUrl,
  target: startPlaybackFx,
})

sample({
  clock: fetchStreamFx.failData,
  fn: error => (error instanceof Error ? error.message : 'Не удалось загрузить поток'),
  target: playbackFailed,
})

sample({
  clock: startPlaybackFx.failData,
  fn: error => (error instanceof Error ? error.message : 'Не удалось запустить воспроизведение'),
  target: playbackFailed,
})

sample({
  clock: playbackToggled,
  source: { isPlaying: $isPlaying, hasStream: $stream.map(Boolean) },
  filter: ({ isPlaying, hasStream }) => isPlaying && hasStream,
  fn: () => undefined,
  target: pausePlaybackFx,
})

sample({
  clock: playbackToggled,
  source: { isPlaying: $isPlaying, hasStream: $stream.map(Boolean) },
  filter: ({ isPlaying, hasStream }) => !isPlaying && hasStream,
  fn: () => undefined,
  target: resumePlaybackFx,
})

sample({
  clock: playbackStopped,
  fn: () => undefined,
  target: stopPlaybackFx,
})

sample({
  clock: seekRequested,
  source: $duration,
  fn: (duration, seconds) => {
    const safeSeconds = Math.max(0, seconds)
    if (!Number.isFinite(duration) || duration <= 0) {
      return safeSeconds
    }
    return Math.min(safeSeconds, duration)
  },
  target: seekPlaybackFx,
})

sample({
  clock: seekPlaybackFx.done,
  fn: ({ params }) => params,
  target: playbackTimeChanged,
})

sample({
  clock: stopPlaybackFx.done,
  fn: () => 0,
  target: playbackTimeChanged,
})

sample({
  clock: startPlaybackFx.done,
  fn: () => getDuration(),
  target: playbackDurationChanged,
})

sample({
  clock: seekPlaybackFx.failData,
  fn: error => (error instanceof Error ? error.message : 'Не удалось перемотать трек'),
  target: playbackFailed,
})

if (typeof window !== 'undefined') {
  attachListener('ended', () => {
    playbackStopped()
  })
  attachListener('timeupdate', () => {
    playbackTimeChanged(getCurrentTime())
  })
  attachListener('loadedmetadata', () => {
    playbackDurationChanged(getDuration())
  })
  attachListener('durationchange', () => {
    playbackDurationChanged(getDuration())
  })
  attachListener('seeked', () => {
    playbackTimeChanged(getCurrentTime())
  })
}
