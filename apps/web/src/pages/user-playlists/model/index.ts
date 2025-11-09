import { sample } from 'effector'
import { routes } from '@shared/router'
import { fetchMyPlaylistsFx } from '@pages/profile/model'

sample({
  clock: routes.profilePlaylists.opened,
  fn: () => 48,
  target: fetchMyPlaylistsFx,
})
