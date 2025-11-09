import { createEvent, createStore } from 'effector'
import { routes } from '@shared/router'
import { $session, $user } from './index'

export const privateRouteOpened = createEvent<{ route: keyof typeof routes }>()

export const $isAuthenticated = createStore(false)
  .on([$session, $user], () => true)
  .on($session.reinit!, () => true)
  .reset([$session.reinit!])
