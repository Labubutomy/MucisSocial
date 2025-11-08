import { createEvent } from 'effector'
import type { AuthFormValues, AuthMode } from '@features/auth'

export const authFormSubmitted = createEvent<{
  mode: AuthMode
  values: AuthFormValues
}>()
