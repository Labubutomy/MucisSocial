import { createEffect, createStore, createEvent, sample } from 'effector'
import {
  createRequestSession,
  getMyActiveSession,
  getIncomingRequests,
  handleTrackRequest,
  deactivateSession,
  type RequestSession,
  type TrackRequest,
} from '@entities/music-requests'
import { routes } from '@shared/router'

// Events
export const pageOpened = createEvent()
export const createSessionClicked = createEvent()
export const stopSessionClicked = createEvent()
export const acceptRequestClicked = createEvent<string>()
export const declineRequestClicked = createEvent<string>()
export const refreshRequestsClicked = createEvent()
export const newRequestReceived = createEvent<TrackRequest>()
export const requestStatusChanged = createEvent<{ requestId: string; status: 'accepted' | 'declined' }>()

// Effects
export const fetchActiveSessionFx = createEffect(async () => {
  return await getMyActiveSession()
})

export const createSessionFx = createEffect(async () => {
  return await createRequestSession()
})

export const stopSessionFx = createEffect(async (sessionId: string) => {
  await deactivateSession(sessionId)
})

export const fetchIncomingRequestsFx = createEffect(async () => {
  return await getIncomingRequests()
})

export const acceptRequestFx = createEffect(async (requestId: string) => {
  return await handleTrackRequest(requestId, 'accept')
})

export const declineRequestFx = createEffect(async (requestId: string) => {
  return await handleTrackRequest(requestId, 'decline')
})

// Stores
export const $activeSession = createStore<RequestSession | null>(null)
  .on(fetchActiveSessionFx.doneData, (_, session) => session)
  .on(createSessionFx.doneData, (_, session) => session)
  .on(stopSessionFx.done, () => null)

export const $incomingRequests = createStore<TrackRequest[]>([])
  .on(fetchIncomingRequestsFx.doneData, (_, requests) => requests)
  .on(acceptRequestFx.doneData, (requests, updatedRequest) =>
    requests.map(r => (r.id === updatedRequest.id ? updatedRequest : r))
  )
  .on(declineRequestFx.doneData, (requests, updatedRequest) =>
    requests.map(r => (r.id === updatedRequest.id ? updatedRequest : r))
  )
  .on(newRequestReceived, (requests, newRequest) => [newRequest, ...requests])

export const $isLoading = createStore(false)
  .on(fetchActiveSessionFx.pending, (_, pending) => pending)
  .on(createSessionFx.pending, (state, pending) => state || pending)

export const $pendingRequests = $incomingRequests.map(requests =>
  requests.filter(r => r.status === 'pending')
)

// Logic
sample({
  clock: routes.musicRequests.opened,
  target: [fetchActiveSessionFx, fetchIncomingRequestsFx],
})

sample({
  clock: pageOpened,
  target: [fetchActiveSessionFx, fetchIncomingRequestsFx],
})

sample({
  clock: createSessionClicked,
  target: createSessionFx,
})

sample({
  clock: stopSessionClicked,
  source: $activeSession,
  filter: (session): session is RequestSession => session !== null,
  fn: session => session.id,
  target: stopSessionFx,
})

sample({
  clock: acceptRequestClicked,
  target: acceptRequestFx,
})

sample({
  clock: declineRequestClicked,
  target: declineRequestFx,
})

sample({
  clock: refreshRequestsClicked,
  target: fetchIncomingRequestsFx,
})

// Refresh requests after accept/decline
sample({
  clock: [acceptRequestFx.done, declineRequestFx.done],
  target: fetchIncomingRequestsFx,
})
