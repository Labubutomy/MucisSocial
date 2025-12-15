import { createEffect, createEvent, createStore, sample } from 'effector'
import { $user } from '@features/auth'
import * as queueApi from '../api'

export type QueueContext =
  | { type: 'user'; userId: string }
  | { type: 'group'; groupId: string }
  | { type: 'session'; roomId: string }

// События
export const queueLoaded = createEvent()
export const trackAddedToQueue = createEvent<string>() // trackId
export const trackRemovedFromQueue = createEvent<string>() // trackId
export const nextTrackRequested = createEvent()
export const prevTrackRequested = createEvent()
export const queueCleared = createEvent()
export const trackFromQueueSelected = createEvent<string>() // trackId - когда пользователь кликает на трек в очереди

// Effects
const loadQueueFx = createEffect(async (context: QueueContext | null) => {
  const items = await queueApi.getQueue(100, context || undefined)
  return items.map(item => item.track_id)
})

const addTrackToQueueFx = createEffect(
  async ({ trackId, context }: { trackId: string; context: QueueContext | null }) => {
    await queueApi.addTrackToQueue(trackId, context || undefined)
    return trackId
  }
)

const removeTrackFromQueueFx = createEffect(
  async ({ trackId, context }: { trackId: string; context: QueueContext | null }) => {
    await queueApi.removeTrackFromQueue(trackId, context || undefined)
    return trackId
  }
)

const getNextTrackFx = createEffect(async (context: QueueContext | null) => {
  const next = await queueApi.getNextTrack(context || undefined)
  return next?.track_id || null
})

const getPrevTrackFx = createEffect(async (context: QueueContext | null) => {
  const prev = await queueApi.getPrevTrack(context || undefined)
  return prev?.track_id || null
})

// Экспортируем effects для использования в других модулях
export { getNextTrackFx, getPrevTrackFx, removeTrackFromQueueFx, loadQueueFx, addTrackToQueueFx }

const clearQueueFx = createEffect(async (context: QueueContext | null) => {
  await queueApi.clearQueue(context || undefined)
})

// Helper function to compare arrays
const arraysEqual = (a: string[], b: string[]): boolean => {
  if (a.length !== b.length) return false
  return a.every((val, index) => val === b[index])
}

// Stores
export const $queueTrackIds = createStore<string[]>([])
  .on(loadQueueFx.doneData, (currentTrackIds, newTrackIds) => {
    // Only update if data actually changed
    if (!arraysEqual(currentTrackIds, newTrackIds)) {
      return newTrackIds
    }
    return currentTrackIds // Return current state to prevent update
  })
  .on(addTrackToQueueFx.doneData, (state, trackId) => {
    // Only update if track is not already in queue
    if (state.includes(trackId)) {
      return state
    }
    return [...state, trackId]
  })
  .on(removeTrackFromQueueFx.doneData, (state, trackId) => {
    // Only update if track is actually in queue
    if (!state.includes(trackId)) {
      return state
    }
    return state.filter(id => id !== trackId)
  })
  .reset(clearQueueFx.done)

export const $queuePending = loadQueueFx.pending

export const $queueContext = createStore<QueueContext | null>(null)

export const queueContextSet = createEvent<QueueContext>()

sample({
  clock: queueContextSet,
  target: $queueContext,
})

// Загрузить очередь при входе пользователя
const userQueueContextReady = createEvent<QueueContext>()

sample({
  clock: $user,
  source: $queueContext,
  filter: (context, user) => {
    if (!user) return false
    const result = context || (user ? { type: 'user' as const, userId: user.id } : null)
    return result !== null
  },
  fn: (context, user) => {
    if (context) return context
    return { type: 'user' as const, userId: user!.id }
  },
  target: userQueueContextReady,
})

sample({
  clock: userQueueContextReady,
  target: queueLoaded,
})

sample({
  clock: queueLoaded,
  source: $queueContext,
  fn: context => context || null,
  target: loadQueueFx,
})

// Обработка событий
sample({
  clock: trackAddedToQueue,
  source: $queueContext,
  fn: (context, trackId) => ({ trackId, context: context || null }),
  target: addTrackToQueueFx,
})

sample({
  clock: trackRemovedFromQueue,
  source: $queueContext,
  fn: (context, trackId) => ({ trackId, context: context || null }),
  target: removeTrackFromQueueFx,
})

sample({
  clock: nextTrackRequested,
  source: $queueContext,
  fn: context => context,
  target: getNextTrackFx,
})

sample({
  clock: prevTrackRequested,
  source: $queueContext,
  fn: context => context,
  target: getPrevTrackFx,
})

sample({
  clock: queueCleared,
  source: $queueContext,
  fn: context => context,
  target: clearQueueFx,
})

// Когда пользователь выбрал трек из очереди, удаляем все треки до него и включаем его
const removeTracksUntilFx = createEffect(
  async ({ trackIds, context }: { trackIds: string[]; context: QueueContext | null }) => {
    // Удаляем все треки до выбранного (включительно)
    await Promise.all(
      trackIds.map(trackId => queueApi.removeTrackFromQueue(trackId, context || undefined))
    )
    return trackIds
  }
)

sample({
  clock: trackFromQueueSelected,
  source: { queue: $queueTrackIds, context: $queueContext },
  filter: ({ queue }, trackId) => queue.includes(trackId),
  fn: ({ queue, context }, trackId) => {
    // Находим индекс трека в очереди
    const trackIndex = queue.findIndex(id => id === trackId)
    // Удаляем все треки до выбранного (включительно)
    const tracksToRemove = queue.slice(0, trackIndex + 1)
    return { trackIds: tracksToRemove, context: context || null }
  },
  target: removeTracksUntilFx,
})

// Export event for session to use
export const trackFromQueueSelectedForSession = createEvent<string>()

// When track is selected from queue in session context, we need to handle it differently
sample({
  clock: trackFromQueueSelected,
  source: $queueContext,
  filter: context => context?.type === 'session',
  fn: (_, trackId) => trackId,
  target: trackFromQueueSelectedForSession,
})

// После добавления/удаления трека, перезагружаем очередь
sample({
  clock: [addTrackToQueueFx.done, removeTrackFromQueueFx.done, removeTracksUntilFx.done],
  source: $queueContext,
  fn: context => context,
  target: loadQueueFx,
})

// Экспортируем событие для обновления очереди извне
export const queueRefreshRequested = createEvent()

sample({
  clock: queueRefreshRequested,
  source: $queueContext,
  fn: context => context,
  target: loadQueueFx,
})
