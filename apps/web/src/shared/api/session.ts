export interface AuthSession {
  accessToken: string
  refreshToken: string
}

let memorySession: AuthSession | null = null

const ACCESS_TOKEN_KEY = 'access_token'
const REFRESH_TOKEN_KEY = 'refresh_token'
const DEFAULT_MAX_AGE_SECONDS = 60 * 60 * 24 * 7 // 7 days

const isBrowser = () => typeof document !== 'undefined'

const setCookie = (key: string, value: string, maxAge = DEFAULT_MAX_AGE_SECONDS) => {
  if (!isBrowser()) {
    memorySession = {
      ...(memorySession ?? { accessToken: '', refreshToken: '' }),
      [key === ACCESS_TOKEN_KEY ? 'accessToken' : 'refreshToken']: value,
    }
    return
  }

  const encoded = encodeURIComponent(value)
  const secure = window.location.protocol === 'https:' ? '; Secure' : ''
  document.cookie = `${key}=${encoded}; Path=/; Max-Age=${maxAge}; SameSite=Lax${secure}`
}

const deleteCookie = (key: string) => {
  if (!isBrowser()) {
    if (!memorySession) return
    memorySession = {
      accessToken: key === ACCESS_TOKEN_KEY ? '' : memorySession.accessToken,
      refreshToken: key === REFRESH_TOKEN_KEY ? '' : memorySession.refreshToken,
    }
    return
  }
  document.cookie = `${key}=; Path=/; Max-Age=0; SameSite=Lax`
}

const getCookie = (key: string): string | null => {
  if (!isBrowser()) {
    if (!memorySession) return null
    return key === ACCESS_TOKEN_KEY ? memorySession.accessToken : memorySession.refreshToken
  }

  const cookies = document.cookie.split('; ').filter(Boolean)
  for (const cookie of cookies) {
    const [name, ...rest] = cookie.split('=')
    if (name === key) {
      return decodeURIComponent(rest.join('='))
    }
  }
  return null
}

export const saveSessionTokens = (session: AuthSession) => {
  setCookie(ACCESS_TOKEN_KEY, session.accessToken)
  setCookie(REFRESH_TOKEN_KEY, session.refreshToken)
  memorySession = session
}

export const clearSessionTokens = () => {
  deleteCookie(ACCESS_TOKEN_KEY)
  deleteCookie(REFRESH_TOKEN_KEY)
  memorySession = null
}

export const readSessionTokens = (): AuthSession | null => {
  const accessToken = getCookie(ACCESS_TOKEN_KEY)
  const refreshToken = getCookie(REFRESH_TOKEN_KEY)

  if (!accessToken || !refreshToken) {
    return null
  }

  const session = { accessToken, refreshToken }
  memorySession = session
  return session
}

export const getAccessToken = () => getCookie(ACCESS_TOKEN_KEY)
