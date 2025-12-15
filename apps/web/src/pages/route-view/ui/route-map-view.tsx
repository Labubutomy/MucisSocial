import { useState } from 'react'
import { Marker, Source, Layer } from 'react-map-gl/maplibre'
import type { RoutePoint } from '@features/routes/api'
import { SafeMapContainer, MapStyleToggle, type MapStyleType } from '@features/map'

interface RouteMapViewProps {
  points: RoutePoint[]
  center: [number, number] | { lat: number; lng: number }
  polylinePositions: [number, number][]
  routeId: string
}

export const RouteMapView = ({ points, center, polylinePositions }: RouteMapViewProps) => {
  const centerArray: [number, number] = Array.isArray(center) ? center : [center.lat, center.lng]
  const [mapStyleType, setMapStyleType] = useState<MapStyleType>('hybrid')

  // Преобразуем polylinePositions в GeoJSON формат для Source
  const routeGeoJson = {
    type: 'Feature' as const,
    properties: {},
    geometry: {
      type: 'LineString' as const,
      coordinates: polylinePositions.map(([lat, lng]) => [lng, lat]), // MapLibre использует [lng, lat]
    },
  }

  return (
    <div className="relative h-full w-full">
      <div className="absolute top-4 right-4 z-10">
        <MapStyleToggle styleType={mapStyleType} onChange={setMapStyleType} />
      </div>
      <SafeMapContainer
        center={centerArray}
        zoom={13}
        mapStyleType={mapStyleType}
        scrollWheelZoom={false}
        style={{ height: '100%', width: '100%' }}
      >
        {polylinePositions.length > 1 && (
          <Source id="route" type="geojson" data={routeGeoJson}>
            <Layer
              id="route-line"
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
          >
            <div
              style={{
                backgroundColor: '#ef4444',
                width: '28px',
                height: '28px',
                borderRadius: '50%',
                border: '2px solid white',
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
