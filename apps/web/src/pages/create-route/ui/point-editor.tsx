import { useState } from 'react'
import { Card } from '@shared/ui/card'
import { Input } from '@shared/ui/input'
import { Button } from '@shared/ui/button'
import type { RoutePoint } from '@features/routes'

interface PointEditorProps {
  point: RoutePoint
  index: number
  onUpdate: (point: Partial<RoutePoint>) => void
  onRemove: () => void
  onSelectTrack: () => void
}

export const PointEditor = ({
  point,
  index,
  onUpdate,
  onRemove,
  onSelectTrack,
}: PointEditorProps) => {
  const [isExpanded, setIsExpanded] = useState(false)

  return (
    <Card padding="md" className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/20 text-sm font-semibold text-primary">
            {index + 1}
          </div>
          <div>
            <p className="font-semibold">{point.title || `Точка ${index + 1}`}</p>
            <p className="text-xs text-muted-foreground">
              {point.latitude.toFixed(6)}, {point.longitude.toFixed(6)}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => setIsExpanded(!isExpanded)}
          >
            {isExpanded ? 'Свернуть' : 'Развернуть'}
          </Button>
          <Button type="button" variant="outline" size="sm" onClick={onRemove}>
            Удалить
          </Button>
        </div>
      </div>

      {isExpanded && (
        <div className="space-y-4 border-t border-border/60 pt-4">
          <Input
            label="Название точки"
            placeholder="Например, Площадь Революции"
            value={point.title || ''}
            onChange={e => onUpdate({ title: e.target.value })}
          />
          <div className="flex flex-col gap-2">
            <label className="text-sm font-medium text-muted-foreground">Описание</label>
            <textarea
              value={point.description || ''}
              onChange={e => onUpdate({ description: e.target.value })}
              placeholder="Описание места или события"
              className="min-h-[80px] rounded-lg border border-input bg-secondary/40 px-4 py-3 text-sm text-foreground placeholder:text-muted-foreground/70 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Радиус активации (м)"
              type="number"
              value={point.radius_meters || 50}
              onChange={e => onUpdate({ radius_meters: parseInt(e.target.value) || 50 })}
              min={10}
              max={1000}
            />
            <Input
              label="Смещение трека (сек)"
              type="number"
              value={point.track_start_offset_sec || 0}
              onChange={e => onUpdate({ track_start_offset_sec: parseInt(e.target.value) || 0 })}
              min={0}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-muted-foreground">Трек</label>
            <Button type="button" variant="outline" className="w-full" onClick={onSelectTrack}>
              {point.track_id ? `Трек выбран (${point.track_id.slice(0, 8)}...)` : 'Выбрать трек'}
            </Button>
          </div>
        </div>
      )}
    </Card>
  )
}
