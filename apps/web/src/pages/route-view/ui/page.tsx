import { useEffect, useMemo, useState } from 'react'
import { useUnit } from 'effector-react'
import { Card } from '@shared/ui/card'
import { Button } from '@shared/ui/button'
import { routes } from '@shared/router'
import { $route, $loading, $error, routeLoaded } from '@pages/route-view/model'
import { RouteMapView } from './route-map-view'
import { fetchTrackDetail } from '@entities/track/api'
import type { Track } from '@entities/track/model/types'

export const RouteViewPage = () => {
  const { route, loading, error, routeId, navigateToRoutes } = useUnit({
    route: $route,
    loading: $loading,
    error: $error,
    routeId: routes.routeView.$params.map(params => params?.routeId || ''),
    navigateToRoutes: routes.routes.navigate,
  })

  const [tracks, setTracks] = useState<Record<string, Track | null>>({})
  const [tracksLoading, setTracksLoading] = useState(false)

  useEffect(() => {
    if (routeId) {
      routeLoaded(routeId)
    }
  }, [routeId])

  // Загружаем информацию о треках для каждой точки
  useEffect(() => {
    if (!route?.points || route.points.length === 0) return

    const loadTracks = async () => {
      setTracksLoading(true)
      const tracksMap: Record<string, Track | null> = {}

      await Promise.all(
        route.points.map(async point => {
          if (point.track_id) {
            try {
              const track = await fetchTrackDetail(point.track_id)
              tracksMap[point.track_id] = track
            } catch (error) {
              console.warn(`Failed to load track ${point.track_id}:`, error)
              tracksMap[point.track_id] = null
            }
          }
        })
      )

      setTracks(tracksMap)
      setTracksLoading(false)
    }

    loadTracks()
  }, [route?.points])

  const mapCenter: [number, number] = useMemo(
    () =>
      route?.points && route.points.length > 0
        ? ([route.points[0].latitude, route.points[0].longitude] as [number, number])
        : ([59.9343, 30.3351] as [number, number]), // Санкт-Петербург по умолчанию
    [route?.points]
  )

  const polylinePositions: [number, number][] = useMemo(
    () => route?.points?.map(point => [point.latitude, point.longitude] as [number, number]) || [],
    [route?.points]
  )

  if (loading) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center pb-16 pt-10">
        <p className="text-sm text-muted-foreground">Загрузка маршрута...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="page-container flex min-h-[60vh] flex-col items-center justify-center gap-4 pb-16 pt-10">
        <p className="text-lg font-semibold text-destructive">{error}</p>
        <Button variant="outline" onClick={() => navigateToRoutes({ params: {}, query: {} })}>
          Вернуться к списку маршрутов
        </Button>
      </div>
    )
  }

  if (!route) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center pb-16 pt-10">
        <p className="text-sm text-muted-foreground">Маршрут не найден</p>
      </div>
    )
  }

  return (
    <div className="page-container space-y-8 pb-20 pt-10">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 space-y-4">
          <div>
            <h1 className="text-3xl font-semibold md:text-4xl">{route.title}</h1>
            {route.description && <p className="mt-2 text-muted-foreground">{route.description}</p>}
          </div>

          <div className="flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
            {route.city && (
              <span className="flex items-center gap-1">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 24 24"
                  className="h-4 w-4"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7z" />
                  <circle cx="12" cy="9" r="2.5" />
                </svg>
                {route.city}
                {route.country && `, ${route.country}`}
              </span>
            )}
            {route.total_distance_km && (
              <span className="flex items-center gap-1">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 24 24"
                  className="h-4 w-4"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <path d="M12 2L2 7l10 5 10-5-10-5z" />
                  <path d="M2 17l10 5 10-5M2 12l10 5 10-5" />
                </svg>
                {route.total_distance_km.toFixed(1)} км
              </span>
            )}
            {route.estimated_minutes && (
              <span className="flex items-center gap-1">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 24 24"
                  className="h-4 w-4"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <circle cx="12" cy="12" r="10" />
                  <path d="M12 6v6l4 2" />
                </svg>
                {route.estimated_minutes} мин
              </span>
            )}
            {route.is_linear && (
              <span className="px-2 py-1 rounded bg-primary/20 text-primary text-xs">
                Линейный маршрут
              </span>
            )}
          </div>

          {route.tags && route.tags.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {route.tags.map(tag => (
                <span
                  key={tag}
                  className="px-3 py-1 rounded-full bg-primary/20 text-primary text-sm"
                >
                  {tag}
                </span>
              ))}
            </div>
          )}
        </div>

        {route.cover_image_url && (
          <div className="h-48 w-48 flex-shrink-0 rounded-lg overflow-hidden bg-secondary/50">
            <img
              src={route.cover_image_url}
              alt={route.title}
              className="w-full h-full object-cover"
            />
          </div>
        )}
      </div>

      {route.points && route.points.length > 0 && (
        <Card padding="lg">
          <div className="space-y-4">
            <h2 className="text-xl font-semibold">Точки маршрута</h2>
            <div className="h-96 w-full rounded-lg overflow-hidden">
              <RouteMapView
                key={`route-map-${route.id}`}
                routeId={route.id}
                points={route.points}
                center={mapCenter}
                polylinePositions={polylinePositions}
              />
            </div>

            <div className="space-y-3">
              {route.points.map((point, index) => {
                const track = point.track_id ? tracks[point.track_id] : null
                return (
                  <Card key={point.id || index} padding="md" className="bg-secondary/20">
                    <div className="flex items-start gap-4">
                      <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-primary/20 text-sm font-semibold text-primary">
                        {index + 1}
                      </div>
                      <div className="flex-1 space-y-2">
                        {point.title && <h3 className="font-semibold">{point.title}</h3>}
                        {point.description && (
                          <p className="text-sm text-muted-foreground">{point.description}</p>
                        )}
                        {track && (
                          <div className="flex items-center gap-2 rounded-lg bg-primary/10 px-3 py-2">
                            <svg
                              xmlns="http://www.w3.org/2000/svg"
                              viewBox="0 0 24 24"
                              className="h-4 w-4 text-primary"
                              fill="none"
                              stroke="currentColor"
                              strokeWidth="2"
                            >
                              <path d="M9 18V5l12-2v13" />
                              <circle cx="6" cy="18" r="3" />
                              <circle cx="18" cy="16" r="3" />
                            </svg>
                            <div className="flex-1">
                              <p className="text-sm font-medium">{track.title}</p>
                              {track.artist && (
                                <p className="text-xs text-muted-foreground">{track.artist.name}</p>
                              )}
                            </div>
                          </div>
                        )}
                        {tracksLoading && point.track_id && !track && (
                          <div className="text-xs text-muted-foreground">Загрузка трека...</div>
                        )}
                        {!tracksLoading && point.track_id && !track && (
                          <div className="text-xs text-muted-foreground">Трек не найден</div>
                        )}
                        <div className="flex items-center gap-4 text-xs text-muted-foreground">
                          <span>
                            Координаты: {point.latitude.toFixed(6)}, {point.longitude.toFixed(6)}
                          </span>
                          {point.radius_meters && <span>Радиус: {point.radius_meters} м</span>}
                        </div>
                      </div>
                    </div>
                  </Card>
                )
              })}
            </div>
          </div>
        </Card>
      )}

      <div className="flex justify-center">
        <Button variant="outline" onClick={() => navigateToRoutes({ params: {}, query: {} })}>
          Вернуться к списку маршрутов
        </Button>
      </div>
    </div>
  )
}
