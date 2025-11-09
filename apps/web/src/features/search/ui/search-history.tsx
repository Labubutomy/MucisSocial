import type { SearchHistoryItem } from '@features/search/model/types'
import { Chip } from '@shared/ui/chip'
import { SectionHeader } from '@shared/ui/section-header'

export interface SearchHistoryProps {
  items: SearchHistoryItem[]
  onSelect: (item: SearchHistoryItem) => void
  onClearAll?: () => void
}

export const SearchHistory = ({ items, onSelect, onClearAll }: SearchHistoryProps) => {
  if (items.length === 0) return null

  return (
    <div className="space-y-4">
      <SectionHeader
        title="Недавние запросы"
        action={
          onClearAll && (
            <button
              type="button"
              className="text-sm font-semibold text-muted-foreground transition hover:text-destructive"
              onClick={onClearAll}
            >
              Очистить всё
            </button>
          )
        }
      />
      <div className="flex flex-wrap gap-2">
        {items.map(item => (
          <Chip key={item.id} onClick={() => onSelect(item)}>
            {item.query}
          </Chip>
        ))}
      </div>
    </div>
  )
}
