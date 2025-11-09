import { createEvent, createStore, sample } from 'effector'
import { createRoute } from 'atomic-router'
import { routes } from '@shared/router'
import { $session, $user } from './index'

export const privateRouteOpened = createEvent<{ route: keyof typeof routes }>()

export const $isAuthenticated = createStore(false)
  .on([$session, $user], () => true)
  .on($session.reinit!, () => true)
  .reset([$session.reinit!])
