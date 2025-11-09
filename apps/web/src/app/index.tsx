import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'atomic-router-react'
import { router } from '@shared/router'
import { Pages } from '@pages/index'
import { appStarted } from '@shared/config/init'
import '@shared/styles/global.css'

appStarted()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router}>
      <Pages />
    </RouterProvider>
  </StrictMode>
)
