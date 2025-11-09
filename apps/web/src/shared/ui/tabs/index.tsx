import type { ReactNode } from 'react'
import { cn } from '@shared/lib/cn'

export interface TabItem {
  value: string
  label: ReactNode
  disabled?: boolean
}

export interface TabsProps {
  value: string
  onChange: (value: string) => void
  items: TabItem[]
  size?: 'sm' | 'md'
  className?: string
}

const sizeStyles: Record<NonNullable<TabsProps['size']>, string> = {
  sm: 'text-sm',
  md: 'text-base',
}

export const Tabs = ({ value, onChange, items, size = 'md', className }: TabsProps) => {
  return (
    <div
      className={cn(
        'inline-flex items-center justify-center rounded-full bg-secondary/40 p-1',
        className
      )}
    >
      {items.map(item => {
        const active = item.value === value
        return (
          <button
            key={item.value}
            type="button"
            onClick={() => onChange(item.value)}
            disabled={item.disabled}
            className={cn(
              'relative min-w-[110px] rounded-full px-4 py-2 font-semibold text-muted-foreground transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:opacity-60',
              sizeStyles[size],
              active && 'bg-background text-foreground shadow-lg shadow-primary/20'
            )}
          >
            {item.label}
          </button>
        )
      })}
    </div>
  )
}
