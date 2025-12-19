import { createEffect, createEvent, createStore, sample } from 'effector'
import { routes } from '@shared/router'
import {
  fetchFriends,
  fetchIncomingRequests,
  respondToFriendRequest,
  searchUsers,
  type Friend,
  type FriendRequest,
  type UserSearchResult,
} from '../api'

export const loadFriendsFx = createEffect(fetchFriends)
export const loadIncomingRequestsFx = createEffect(async () => fetchIncomingRequests())
export const respondToRequestFx = createEffect(
  async ({ requestId, accept }: { requestId: string; accept: boolean }) => {
    await respondToFriendRequest(requestId, accept)
  }
)

export const $friends = createStore<Friend[]>([])
  .on(loadFriendsFx.doneData, (_, friends) => friends)
  .reset(routes.friends.opened)

export const $incomingRequests = createStore<FriendRequest[]>([])
  .on(loadIncomingRequestsFx.doneData, (_, requests) => requests)
  .reset(routes.friends.opened)

export const friendRequestResponded = createEvent<{ requestId: string; accept: boolean }>()

sample({
  clock: routes.friends.opened,
  fn: () => undefined,
  target: [loadFriendsFx, loadIncomingRequestsFx],
})

sample({
  clock: friendRequestResponded,
  fn: ({ requestId, accept }) => ({ requestId, accept }),
  target: respondToRequestFx,
})

// После ответа на запрос перезагружаем список запросов и друзей
sample({
  clock: respondToRequestFx.done,
  fn: () => undefined,
  target: [loadFriendsFx, loadIncomingRequestsFx],
})

// Search users
export const searchUsersFx = createEffect(searchUsers)
export const $searchResults = createStore<UserSearchResult[]>([])
  .on(searchUsersFx.doneData, (_, results) => results)
  .reset(routes.friends.closed)
