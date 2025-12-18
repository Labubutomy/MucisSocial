import axios, {
  AxiosError,
  AxiosHeaders,
  type AxiosInstance,
  type AxiosRequestHeaders,
} from 'axios'
import { getAccessToken, getRefreshToken, saveSessionTokens } from './session'

let isRefreshing = false
let failedQueue: Array<{
  resolve: (value?: unknown) => void
  reject: (error?: unknown) => void
}> = []

const processQueue = (error: AxiosError | null, token: string | null = null) => {
  failedQueue.forEach(prom => {
    if (error) {
      prom.reject(error)
    } else {
      prom.resolve(token)
    }
  })
  failedQueue = []
}

export const createApiClient = (baseURL: string): AxiosInstance => {
  const instance = axios.create({
    baseURL,
  })

  instance.interceptors.request.use(config => {
    const token = getAccessToken()
    if (token) {
      const headers =
        config.headers instanceof AxiosHeaders
          ? config.headers
          : AxiosHeaders.from(config.headers as AxiosRequestHeaders | undefined)
      headers.set('Authorization', `Bearer ${token}`)
      config.headers = headers
    }
    return config
  })

  instance.interceptors.response.use(
    response => response,
    async (error: AxiosError) => {
      const originalRequest = error.config as any

      // Если ошибка 401 и это не запрос на refresh
      if (
        error.response?.status === 401 &&
        !originalRequest._retry &&
        originalRequest.url !== '/api/v1/auth/refresh'
      ) {
        if (isRefreshing) {
          // Если уже идет обновление токена, добавляем запрос в очередь
          return new Promise((resolve, reject) => {
            failedQueue.push({ resolve, reject })
          })
            .then(token => {
              originalRequest.headers['Authorization'] = `Bearer ${token}`
              return instance(originalRequest)
            })
            .catch(err => {
              return Promise.reject(err)
            })
        }

        originalRequest._retry = true
        isRefreshing = true

        const refreshToken = getRefreshToken()
        if (!refreshToken) {
          processQueue(error, null)
          isRefreshing = false
          return Promise.reject(error)
        }

        try {
          // Пытаемся обновить токен
          const refreshResponse = await axios.post<{ access_token: string; refresh_token: string }>(
            `${baseURL}/api/v1/auth/refresh`,
            { refresh_token: refreshToken },
            { headers: { 'Content-Type': 'application/json' } }
          )

          const { access_token, refresh_token } = refreshResponse.data
          saveSessionTokens({ accessToken: access_token, refreshToken: refresh_token })

          // Обновляем токен в оригинальном запросе
          originalRequest.headers['Authorization'] = `Bearer ${access_token}`

          processQueue(null, access_token)
          isRefreshing = false

          // Повторяем оригинальный запрос
          return instance(originalRequest)
        } catch (refreshError) {
          processQueue(refreshError as AxiosError, null)
          isRefreshing = false
          return Promise.reject(refreshError)
        }
      }

      return Promise.reject(error)
    }
  )

  return instance
}
