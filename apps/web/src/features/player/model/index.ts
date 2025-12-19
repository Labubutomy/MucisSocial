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
import {
  $queueTrackIds,
  trackAddedToQueue,
  trackRemovedFromQueue,
  queueRefreshRequested,
  trackFromQueueSelected,
  loadQueueFx,
} from '@features/queue'
import { fetchTrackRecommendations, fetchTrackDetail } from '@entities/track/api'
import { $user } from '@features/auth'

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

// Событие для остановки воспроизведения без сброса состояния
const playbackPaused = createEvent()

// Effect для паузы без сброса состояния
const pausePlaybackOnlyFx = createEffect(async () => {
  await pauseStream()
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
  .on(pausePlaybackOnlyFx.done, () => false)
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
  target: playbackPaused, // Используем playbackPaused вместо stopPlaybackFx, чтобы не сбрасывать состояние
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

// Событие окончания трека (для автопереключения)
const trackEnded = createEvent()

// Effect для загрузки трека по ID
const loadTrackByIdFx = createEffect(async (trackId: string) => {
  return await fetchTrackDetail(trackId)
})

// Effect для получения рекомендаций
const fetchRecommendationsForQueueFx = createEffect(async () => {
  const recommendations = await fetchTrackRecommendations({ limit: 1 })
  return recommendations[0] || null // Первый трек из рекомендаций
})

// Когда трек поставлен на воспроизведение:
// 1. Удаляем его из очереди (если он там есть) - он начинает играть
// 2. НЕ добавляем его в очередь - он уже играет
sample({
  clock: trackQueued,
  source: { queue: $queueTrackIds, user: $user, currentTrack: $currentTrack },
  filter: ({ user, queue }, track) => {
    // Проверяем что пользователь авторизован
    if (!user) return false
    // Если трек в очереди - удаляем его (он начинает играть)
    return queue.includes(track.id)
  },
  fn: (_source, track) => track.id,
  target: trackRemovedFromQueue,
})

// Если очередь пуста при включении трека, добавляем рекомендацию
// НО только если это не трек, который пользователь выбрал вручную
sample({
  clock: trackQueued,
  source: { queue: $queueTrackIds, currentTrack: $currentTrack },
  filter: ({ queue, currentTrack }, track) => {
    // Очередь пуста И это не тот же трек, который уже играет
    return queue.length === 0 && currentTrack?.id !== track.id
  },
  fn: () => undefined,
  target: fetchRecommendationsForQueueFx,
})

// Когда получена рекомендация (при пустой очереди при включении трека), добавляем её в очередь
sample({
  clock: fetchRecommendationsForQueueFx.doneData,
  source: $queueTrackIds,
  filter: (queue, track) => track !== null && queue.length === 0,
  fn: (_queue, track) => track!.id,
  target: trackAddedToQueue,
})

// Событие для пропуска текущего трека
export const skipTrackRequested = createEvent()

// Флаги для отслеживания окончания трека и пропуска
const $trackEndedFlag = createStore(false)
  .on(trackEnded, () => true)
  .reset(loadTrackByIdFx.done)

const $skipRequested = createStore(false)
  .on(skipTrackRequested, () => true)
  .reset(loadTrackByIdFx.done)

// Когда трек закончился, удаляем его из очереди
sample({
  clock: trackEnded,
  source: { queue: $queueTrackIds, currentTrack: $currentTrack },
  filter: ({ currentTrack }) => Boolean(currentTrack),
  fn: ({ currentTrack }) => currentTrack!.id,
  target: trackRemovedFromQueue,
})

// Когда пользователь пропустил трек, останавливаем воспроизведение и удаляем его из очереди
sample({
  clock: skipTrackRequested,
  source: { queue: $queueTrackIds, currentTrack: $currentTrack },
  filter: ({ currentTrack }) => Boolean(currentTrack),
  fn: ({ currentTrack }) => currentTrack!.id,
  target: trackRemovedFromQueue,
})

// Останавливаем воспроизведение после удаления трека при пропуске (но не сбрасываем $currentTrack)
sample({
  clock: skipTrackRequested,
  source: $currentTrack,
  filter: Boolean,
  fn: () => undefined,
  target: pausePlaybackOnlyFx,
})

// После перезагрузки очереди (после удаления трека), берем первый трек из обновленной очереди
sample({
  clock: loadQueueFx.doneData,
  source: { trackEnded: $trackEndedFlag, skipRequested: $skipRequested },
  filter: ({ trackEnded, skipRequested }, queue) =>
    (trackEnded || skipRequested) && queue.length > 0,
  fn: (_source, queue) => queue[0], // Берем первый трек из очереди
  target: loadTrackByIdFx,
})

// Если после перезагрузки очереди она пуста, добавляем рекомендацию
sample({
  clock: loadQueueFx.doneData,
  source: { trackEnded: $trackEndedFlag, skipRequested: $skipRequested },
  filter: ({ trackEnded, skipRequested }, queue) =>
    (trackEnded || skipRequested) && queue.length === 0,
  fn: () => undefined,
  target: fetchRecommendationsForQueueFx,
})

// Обновляем UI очереди после удаления трека
sample({
  clock: trackRemovedFromQueue,
  fn: () => undefined,
  target: queueRefreshRequested,
})

// После загрузки трека, ставим его на воспроизведение
sample({
  clock: loadTrackByIdFx.doneData,
  fn: track => track,
  target: trackQueued,
})

// Когда пользователь выбрал трек из очереди, загружаем и играем его
sample({
  clock: trackFromQueueSelected,
  target: loadTrackByIdFx,
})

// Когда получена рекомендация после удаления трека (если очередь пуста), добавляем в очередь и играем
sample({
  clock: fetchRecommendationsForQueueFx.doneData,
  source: { trackEnded: $trackEndedFlag, skipRequested: $skipRequested },
  filter: ({ trackEnded, skipRequested }, track) => track !== null && (trackEnded || skipRequested),
  fn: (_source, track) => track!.id,
  target: trackAddedToQueue,
})

// После добавления рекомендации в очередь, загружаем и играем её
sample({
  clock: fetchRecommendationsForQueueFx.doneData,
  source: { trackEnded: $trackEndedFlag, skipRequested: $skipRequested },
  filter: ({ trackEnded, skipRequested }, track) => track !== null && (trackEnded || skipRequested),
  fn: (_source, track) => track!.id,
  target: loadTrackByIdFx,
})

if (typeof window !== 'undefined') {
  attachListener('ended', () => {
    trackEnded()
    // Не вызываем playbackStopped() здесь, чтобы не сбрасывать состояние
    // перед переключением на следующий трек
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
