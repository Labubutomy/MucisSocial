export type AuthMode = 'signIn' | 'signUp'

export interface AuthFormValues {
  email: string
  password: string
  username?: string
}
