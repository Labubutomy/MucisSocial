import { createBrowserHistory } from 'history'
import { appStarted } from '../config/init'
import { sample } from 'effector'
import { router } from './config'

sample({
  clock: appStarted,
  fn: () => createBrowserHistory(),
  target: router.setHistory,
})
