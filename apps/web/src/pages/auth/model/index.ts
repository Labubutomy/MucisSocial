import { createEvent, createStore, sample } from 'effector'
import type { AuthFormValues, AuthMode } from '@features/auth'
import { authFlowStarted } from '@features/auth/model'

export const authFormSubmitted = createEvent<{
  mode: AuthMode
  values: AuthFormValues
}>()

const initialValues: AuthFormValues = {
  email: '',
  password: '',
  username: '',
}

export const modeChanged = createEvent<AuthMode>()
export const valuesChanged = createEvent<AuthFormValues>()
export const submitClicked = createEvent()

export const $mode = createStore<AuthMode>('signIn').on(modeChanged, (_, mode) => mode)

export const $values = createStore<AuthFormValues>(initialValues)
  .on(valuesChanged, (_, values) => values)
  .on(modeChanged, () => initialValues)

sample({
  clock: submitClicked,
  source: { mode: $mode, values: $values },
  target: [authFormSubmitted, authFlowStarted],
})
