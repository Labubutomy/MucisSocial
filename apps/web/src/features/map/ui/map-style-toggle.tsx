import { Button } from '@shared/ui/button'
import type { MapStyleType } from './map-container'

interface MapStyleToggleProps {
  styleType: MapStyleType
  onChange: (styleType: MapStyleType) => void
  className?: string
}

export const MapStyleToggle = ({ styleType, onChange, className }: MapStyleToggleProps) => {
  return (
    <div className={`flex gap-2 ${className || ''}`}>
      <Button
        variant={styleType === 'hybrid' ? 'primary' : 'outline'}
        size="sm"
        onClick={() => onChange('hybrid')}
        className="text-xs"
      >
        Гибрид
      </Button>
      <Button
        variant={styleType === 'schematic' ? 'primary' : 'outline'}
        size="sm"
        onClick={() => onChange('schematic')}
        className="text-xs"
      >
        Схема
      </Button>
    </div>
  )
}
