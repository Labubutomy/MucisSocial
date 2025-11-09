import { resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const projectRootDir = fileURLToPath(new URL('.', import.meta.url))

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@app': resolve(projectRootDir, 'src/app'),
      '@pages': resolve(projectRootDir, 'src/pages'),
      '@widgets': resolve(projectRootDir, 'src/widgets'),
      '@features': resolve(projectRootDir, 'src/features'),
      '@entities': resolve(projectRootDir, 'src/entities'),
      '@shared': resolve(projectRootDir, 'src/shared'),
    },
  },
})
