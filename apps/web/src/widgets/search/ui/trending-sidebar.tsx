import { SectionHeader } from '@shared/ui/section-header'

export interface TrendingSidebarProps {
  items: { id: string; label: string }[]
  onSelect: (item: { id: string; label: string }) => void
}

export const TrendingSidebar = ({ items, onSelect }: TrendingSidebarProps) => (
  <aside className="hidden w-64 flex-shrink-0 flex-col gap-4 rounded-3xl border border-border/60 bg-secondary/20 p-5 lg:flex">
    <SectionHeader title="Сейчас ищут" subtitle="Что ищет сообщество" />
    <ul className="space-y-3">
      {items.map((item, index) => (
        <li key={item.id}>
          <button
            type="button"
            onClick={() => onSelect(item)}
            className="flex w-full items-center justify-between rounded-2xl bg-secondary/40 px-4 py-3 text-left text-sm font-medium text-muted-foreground transition hover:bg-primary/20 hover:text-foreground"
          >
            <span className="truncate">{item.label}</span>
            <span className="text-xs text-muted-foreground/80">#{index + 1}</span>
          </button>
        </li>
      ))}
    </ul>
  </aside>
)
