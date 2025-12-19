import { useCallback, useEffect, useRef, useState } from 'react'
import { Marker, Source, Layer } from 'react-map-gl/maplibre'
import type { RoutePoint } from '@features/routes'
import { SafeMapContainer, MapStyleToggle, type MapStyleType } from '@features/map'
import type { MapRef, MapLayerMouseEvent } from 'react-map-gl/maplibre'

interface RouteMapProps {
  points: RoutePoint[]
  onPointAdd: (lat: number, lng: number) => void
  onPointClick: (index: number) => void
  selectedPointIndex?: number | null
}

function MapBounds({
  points,
  mapRef,
}: {
  points: RoutePoint[]
  mapRef: React.RefObject<MapRef | null>
}) {
  useEffect(() => {
    if (!mapRef.current) return

    // При одной точке - просто центрируем без сильного приближения
    if (points.length === 1) {
      try {
        mapRef.current.setCenter({
          lng: points[0].longitude,
          lat: points[0].latitude,
        })
        // Устанавливаем умеренный зум (не слишком близко)
        mapRef.current.setZoom(14)
      } catch (error) {
        console.warn('Failed to center map:', error)
      }
      return
    }

    // При двух и более точках - используем fitBounds
    if (points.length > 1) {
      try {
        const lngs = points.map(p => p.longitude)
        const lats = points.map(p => p.latitude)
        const minLng = Math.min(...lngs)
        const maxLng = Math.max(...lngs)
        const minLat = Math.min(...lats)
        const maxLat = Math.max(...lats)

        mapRef.current.fitBounds(
          [
            [minLng, minLat],
            [maxLng, maxLat],
          ],
          {
            padding: { top: 50, bottom: 50, left: 50, right: 50 },
            duration: 500,
          }
        )
      } catch (error) {
        console.warn('Failed to fit bounds:', error)
      }
    }
  }, [points, mapRef])
  return null
}

export const RouteMap = ({
  points,
  onPointAdd,
  onPointClick,
  selectedPointIndex,
}: RouteMapProps) => {
  const mapRef = useRef<MapRef>(null)
  const [mapStyleType, setMapStyleType] = useState<MapStyleType>('hybrid')

  const handleMapClick = useCallback(
    (event: MapLayerMouseEvent) => {
      onPointAdd(event.lngLat.lat, event.lngLat.lng)
    },
    [onPointAdd]
  )

  // Санкт-Петербург по умолчанию
  return (
    <div className="relative h-full w-full">
      <div className="absolute top-4 right-4 z-10">
        <MapStyleToggle styleType={mapStyleType} onChange={setMapStyleType} />
      </div>
      <SafeMapContainer
        center={[59.9343, 30.3351]}
        zoom={13}
        mapStyleType={mapStyleType}
        className="h-full w-full rounded-xl overflow-hidden border border-border/60"
        onMapReady={ref => {
          mapRef.current = ref.current
        }}
        onMapClick={handleMapClick}
      >
        <MapBounds points={points} mapRef={mapRef} />
        {/* Полилиния между точками */}
        {points.length > 1 && (
          <Source
            id="route-line"
            type="geojson"
            data={{
              type: 'Feature',
              properties: {},
              geometry: {
                type: 'LineString',
                coordinates: points.map(p => [p.longitude, p.latitude]),
              },
            }}
          >
            <Layer
              id="route-line-layer"
              type="line"
              paint={{
                'line-color': '#3b82f6',
                'line-width': 3,
              }}
            />
          </Source>
        )}
        {points.map((point, index) => (
          <Marker
            key={point.id || `point-${index}`}
            longitude={point.longitude}
            latitude={point.latitude}
            anchor="bottom"
            onClick={(e: { originalEvent: Event }) => {
              e.originalEvent.stopPropagation()
              onPointClick(index)
            }}
          >
            <div
              style={{
                backgroundColor: selectedPointIndex === index ? '#3b82f6' : '#ef4444',
                width: '28px',
                height: '28px',
                borderRadius: '50%',
                border: '2px solid white',
                cursor: 'pointer',
                boxShadow: '0 2px 4px rgba(0,0,0,0.3)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'white',
                fontSize: '12px',
                fontWeight: 'bold',
              }}
            >
              {index + 1}
            </div>
          </Marker>
        ))}
      </SafeMapContainer>
    </div>
  )
}
