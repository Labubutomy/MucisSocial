import { createApiClient } from '@shared/api/client'
import { API_CONFIG } from '@shared/config/api'

const apiClient = createApiClient(API_CONFIG.gateway)

export interface RoutePoint {
  id?: string
  latitude: number
  longitude: number
  radius_meters?: number
  track_id: string
  track_start_offset_sec?: number
  title?: string
  description?: string
  image_url?: string
  order_index?: number
}

export interface Route {
  id: string
  user_id: string
  title: string
  description?: string
  city?: string
  country?: string
  is_public: boolean
  is_linear: boolean
  total_distance_km?: number
  estimated_minutes?: number
  cover_image_url?: string
  tags: string[]
  points: RoutePoint[]
  created_at: string
  updated_at: string
}

export interface CreateRouteRequest {
  title: string
  description?: string
  city?: string
  country?: string
  is_public?: boolean
  is_linear?: boolean
  tags?: string[]
  points?: Omit<RoutePoint, 'id' | 'order_index'>[]
}

export type UpdateRouteRequest = Partial<CreateRouteRequest>

export const routesApi = {
  createRoute: async (data: CreateRouteRequest): Promise<Route> => {
    const response = await apiClient.post<Route>('/api/v1/routes', data)
    return response.data
  },

  getRoute: async (routeId: string, includePoints = false): Promise<Route> => {
    const response = await apiClient.get<Route>(`/api/v1/routes/${routeId}`, {
      params: { include_points: includePoints },
    })
    return response.data
  },

  updateRoute: async (routeId: string, data: UpdateRouteRequest): Promise<Route> => {
    const response = await apiClient.put<Route>(`/api/v1/routes/${routeId}`, data)
    return response.data
  },

  deleteRoute: async (routeId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/routes/${routeId}`)
  },

  listRoutes: async (params?: {
    user_id?: string
    is_public?: boolean
    city?: string
    limit?: number
    offset?: number
  }): Promise<{ routes: Route[]; total: number; limit: number; offset: number }> => {
    const response = await apiClient.get<{
      routes: Route[]
      total: number
      limit: number
      offset: number
    }>('/api/v1/routes', { params })
    return response.data
  },

  findNearbyRoutes: async (params: {
    latitude: number
    longitude: number
    radius_km?: number
    limit?: number
    offset?: number
  }): Promise<{ routes: Route[]; total: number; limit: number; offset: number }> => {
    const response = await apiClient.get<{
      routes: Route[]
      total: number
      limit: number
      offset: number
    }>('/api/v1/routes/nearby', { params })
    return response.data
  },

  addPoint: async (
    routeId: string,
    point: Omit<RoutePoint, 'id' | 'order_index'>
  ): Promise<RoutePoint> => {
    const response = await apiClient.post<RoutePoint>(`/api/v1/routes/${routeId}/points`, point)
    return response.data
  },

  getRoutePoints: async (routeId: string): Promise<RoutePoint[]> => {
    const response = await apiClient.get<RoutePoint[]>(`/api/v1/routes/${routeId}/points`)
    return response.data
  },

  updatePoint: async (
    routeId: string,
    pointId: string,
    point: Partial<RoutePoint>
  ): Promise<RoutePoint> => {
    const response = await apiClient.put<RoutePoint>(
      `/api/v1/routes/${routeId}/points/${pointId}`,
      point
    )
    return response.data
  },

  deletePoint: async (routeId: string, pointId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/routes/${routeId}/points/${pointId}`)
  },
}
