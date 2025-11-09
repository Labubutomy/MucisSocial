import { createEffect, createEvent, createStore, sample } from 'effector'
import type { AuthMode, AuthFormValues } from './types'
import { fetchMe, signIn, signUp, type UserProfile } from '../api'
import { routes } from '@shared/router'
import {
  clearSessionTokens,
  readSessionTokens,
  saveSessionTokens,
  type AuthSession,
} from '@shared/api/session'
import { appStarted } from '@shared/config/init'

export const signOut = createEvent()

export const signInFx = createEffect(signIn)
export const signUpFx = createEffect(signUp)
export const fetchMeFx = createEffect(fetchMe)

const persistSessionFx = createEffect((session: AuthSession) => {
  saveSessionTokens(session)
})

const clearSessionFx = createEffect(() => {
  clearSessionTokens()
})

const readSessionFx = createEffect(() => readSessionTokens())

const sessionHydrated = createEvent<AuthSession>()

export const $session = createStore<AuthSession | null>(null)
  .on(signInFx.doneData, (_, payload) => ({
    accessToken: payload.accessToken,
    refreshToken: payload.refreshToken,
  }))
  .on(signUpFx.doneData, (_, payload) => ({
    accessToken: payload.accessToken,
    refreshToken: payload.refreshToken,
  }))
  .on(sessionHydrated, (_, session) => session)
  .reset(signOut)

export const $isAuthenticated = $session.map(Boolean)

export const $user = createStore<UserProfile | null>(null)
  .on(fetchMeFx.doneData, (_, profile) => profile)
  .reset(signOut)

export const $authPending = createStore(false)
  .on(signInFx.pending, (_, pending) => pending)
  .on(signUpFx.pending, (_, pending) => pending)

export const $authError = createStore<string | null>(null)
  .on(signInFx.failData, (_, error) => (error as Error).message)
  .on(signUpFx.failData, (_, error) => (error as Error).message)
  .reset([signInFx.done, signUpFx.done, signOut])

export const authFlowStarted = createEvent<{ mode: AuthMode; values: AuthFormValues }>()

sample({
  clock: authFlowStarted,
  filter: ({ mode }) => mode === 'signIn',
  fn: ({ values }) => ({ email: values.email, password: values.password }),
  target: signInFx,
})

sample({
  clock: authFlowStarted,
  filter: ({ mode, values }) => mode === 'signUp' && Boolean(values.username),
  fn: ({ values }) => ({
    email: values.email,
    password: values.password,
    username: values.username ?? '',
  }),
  target: signUpFx,
})

sample({
  clock: appStarted,
  target: readSessionFx,
})

sample({
  clock: readSessionFx.doneData,
  filter: (session): session is AuthSession => Boolean(session),
  target: [sessionHydrated, fetchMeFx],
})

sample({
  clock: signOut,
  target: clearSessionFx,
})

const sessionFromSignIn = sample({
  clock: signInFx.doneData,
  fn: payload => ({
    accessToken: payload.accessToken,
    refreshToken: payload.refreshToken,
  }),
})

const sessionFromSignUp = sample({
  clock: signUpFx.doneData,
  fn: payload => ({
    accessToken: payload.accessToken,
    refreshToken: payload.refreshToken,
  }),
})

sample({
  clock: sessionFromSignIn,
  target: [sessionHydrated, persistSessionFx, fetchMeFx],
})

sample({
  clock: sessionFromSignUp,
  target: [sessionHydrated, persistSessionFx, fetchMeFx],
})

sample({
  clock: signOut,
  fn: () => ({
    params: {},
    query: {},
  }),
  target: routes.auth.navigate,
})

sample({
  clock: fetchMeFx.done,
  source: routes.auth.$isOpened,
  filter: Boolean,
  fn: () => ({
    params: {},
    query: {},
  }),
  target: routes.home.navigate,
})

sample({
  clock: routes.auth.opened,
  source: $isAuthenticated,
  filter: Boolean,
  fn: () => ({
    params: {},
    query: {},
  }),
  target: routes.home.navigate,
})

const protectedRouteNames: Array<keyof typeof routes> = [
  'search',
  'profile',
  'profilePlaylists',
  'curations',
  'playlistCreate',
  'playlistAddTracks',
  'collection',
]

const redirectToAuth = routes.auth.navigate.prepend(() => ({
  params: {},
  query: {},
}))

protectedRouteNames.forEach(routeName => {
  const route = routes[routeName]

  sample({
    clock: route.opened,
    source: $isAuthenticated,
    filter: (isAuthenticated: boolean) => !isAuthenticated,
    fn: () => undefined,
    target: redirectToAuth,
  })
})
