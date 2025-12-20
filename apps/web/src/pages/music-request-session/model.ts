import { createEffect, createStore, createEvent, sample } from 'effector'
import { createGate } from 'effector-react'
import {
  getSessionByCode,
  createTrackRequest,
  getUserCoins,
  type RequestSession,
  type UserCoins,
} from '@entities/music-requests'
import { routes } from '@shared/router'

// Gate
export const RequestSessionGate = createGate<{ sessionCode: string }>()

// Events
export const trackSelected = createEvent<string>()
export const messageChanged = createEvent<string>()
export const submitRequestClicked = createEvent()
export const resetForm = createEvent()

// Effects
export const fetchSessionFx = createEffect(async (sessionCode: string) => {
  return await getSessionByCode(sessionCode)
})

export const fetchUserCoinsFx = createEffect(async () => {
  return await getUserCoins()
})

export const submitRequestFx = createEffect(
  async ({
    sessionCode,
    trackId,
    message,
  }: {
    sessionCode: string
    trackId: string
    message?: string
  }) => {
    return await createTrackRequest({
      session_code: sessionCode,
      track_id: trackId,
      message,
    })
  }
)

// Stores
export const $session = createStore<RequestSession | null>(null)
  .on(fetchSessionFx.doneData, (_, session) => session)

export const $userCoins = createStore<UserCoins | null>(null)
  .on(fetchUserCoinsFx.doneData, (_, coins) => coins)

export const $selectedTrackId = createStore<string>('')
  .on(trackSelected, (_, trackId) => trackId)
  .reset(resetForm)

export const $message = createStore<string>('')
  .on(messageChanged, (_, message) => message)
  .reset(resetForm)

export const $isLoading = createStore(false)
  .on(fetchSessionFx.pending, (_, pending) => pending)

export const $isSubmitting = createStore(false)
  .on(submitRequestFx.pending, (_, pending) => pending)

export const $submitSuccess = createStore(false)
  .on(submitRequestFx.done, () => true)
  .reset(resetForm)

export const $submitError = createStore<string | null>(null)
  .on(submitRequestFx.failData, (_, error) => {
    if (error instanceof Error) {
      if (error.message.includes('Not enough coins')) {
        return 'Недостаточно монет для заказа'
      }
      return error.message
    }
    return 'Произошла ошибка'
  })
  .reset([resetForm, submitRequestFx])

export const $sessionError = createStore<string | null>(null)
  .on(fetchSessionFx.failData, () => 'Сессия не найдена или неактивна')
  .reset(fetchSessionFx)

// Load session when page opens
sample({
  clock: RequestSessionGate.open,
  fn: ({ sessionCode }) => sessionCode,
  target: fetchSessionFx,
})

// Load user coins when session loaded
sample({
  clock: fetchSessionFx.doneData,
  target: fetchUserCoinsFx,
})

// Submit request
sample({
  clock: submitRequestClicked,
  source: {
    session: $session,
    trackId: $selectedTrackId,
    message: $message,
  },
  filter: ({ session, trackId }) => session !== null && trackId.length > 0,
  fn: ({ session, trackId, message }) => ({
    sessionCode: session!.session_code,
    trackId,
    message: message || undefined,
  }),
  target: submitRequestFx,
})

// Refresh coins after successful submit
sample({
  clock: submitRequestFx.done,
  target: fetchUserCoinsFx,
})
