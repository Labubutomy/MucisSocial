import { createRoutesView } from 'atomic-router-react'
import { routes } from '@shared/router'

export const Pages = createRoutesView({
  routes: [{ view: () => <div>Hello World</div>, route: routes.index }],
})
