import { useCallback, useRef } from 'react'
import Map from 'react-map-gl/maplibre'
import type { MapRef, MapLayerMouseEvent } from 'react-map-gl/maplibre'
import 'maplibre-gl/dist/maplibre-gl.css'

export type MapStyleType = 'hybrid' | 'schematic'

interface MapContainerProps {
  center: [number, number]
  zoom: number
  children?: React.ReactNode
  className?: string
  style?: React.CSSProperties
  scrollWheelZoom?: boolean
  mapStyleType?: MapStyleType
  onMapReady?: (map: React.RefObject<MapRef>) => void
  onMapClick?: (event: MapLayerMouseEvent) => void
}

/**
 * Безопасный контейнер для MapLibre карты
 * Работает корректно в React Strict Mode
 */
export const SafeMapContainer = ({
  center,
  zoom,
  children,
  className,
  style,
  scrollWheelZoom = true,
  mapStyleType = 'hybrid',
  onMapReady,
  onMapClick,
}: MapContainerProps) => {
  const mapRef = useRef<MapRef>(null)

  // Обновляем ref при изменении onMapReady
  const handleMapLoad = useCallback(() => {
    if (mapRef.current && onMapReady) {
      onMapReady(mapRef as React.RefObject<MapRef>)
    }
  }, [onMapReady])

  // Схематичный стиль карты (OpenStreetMap)
  const schematicMapStyle = {
    version: 8 as const,
    sources: {
      osm: {
        type: 'raster' as const,
        tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'],
        tileSize: 256,
        attribution:
          '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
      },
    },
    layers: [
      {
        id: 'osm-layer',
        type: 'raster' as const,
        source: 'osm',
        minzoom: 0,
        maxzoom: 19,
      },
    ],
  }

  // Гибридный стиль карты (спутниковый с подписями)
  const hybridMapStyle = {
    version: 8 as const,
    sources: {
      'esri-satellite': {
        type: 'raster' as const,
        tiles: [
          'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
        ],
        tileSize: 256,
        attribution:
          '&copy; <a href="https://www.esri.com/">Esri</a> &copy; <a href="https://www.esri.com/">Esri</a>',
      },
      'esri-labels': {
        type: 'raster' as const,
        tiles: [
          'https://server.arcgisonline.com/ArcGIS/rest/services/Reference/World_Boundaries_and_Places/MapServer/tile/{z}/{y}/{x}',
        ],
        tileSize: 256,
      },
    },
    layers: [
      {
        id: 'esri-satellite-layer',
        type: 'raster' as const,
        source: 'esri-satellite',
        minzoom: 0,
        maxzoom: 22,
      },
      {
        id: 'esri-labels-layer',
        type: 'raster' as const,
        source: 'esri-labels',
        minzoom: 0,
        maxzoom: 22,
      },
    ],
  }

  const currentMapStyle = mapStyleType === 'hybrid' ? hybridMapStyle : schematicMapStyle

  return (
    <div className={className} style={style}>
      <Map
        ref={mapRef}
        mapStyle={currentMapStyle}
        initialViewState={{
          longitude: center[1],
          latitude: center[0],
          zoom: zoom,
          pitch: 0,
          bearing: 0,
        }}
        onLoad={handleMapLoad}
        onClick={onMapClick}
        scrollZoom={scrollWheelZoom}
        dragRotate={false}
        touchZoomRotate={false}
        doubleClickZoom={false}
        style={{ width: '100%', height: '100%' }}
        attributionControl={false}
      >
        {children}
      </Map>
    </div>
  )
}
