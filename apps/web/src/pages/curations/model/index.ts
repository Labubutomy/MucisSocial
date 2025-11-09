import { sample } from 'effector'
import { routes } from '@shared/router'
import { fetchMyPlaylistsFx } from '@pages/profile/model'

sample({
  clock: routes.curations.opened,
  fn: () => 12,
  target: fetchMyPlaylistsFx,
})
