import axios, { AxiosHeaders, type AxiosInstance, type AxiosRequestHeaders } from 'axios'
import { getAccessToken } from './session'

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

  return instance
}
