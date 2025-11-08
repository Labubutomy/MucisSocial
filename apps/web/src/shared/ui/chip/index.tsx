import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { cn } from '@shared/lib/cn'

export interface ChipProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  selected?: boolean
  icon?: ReactNode
}

export const Chip = ({ className, children, selected, icon, ...props }: ChipProps) => (
  <button
    type="button"
    className={cn(
      'inline-flex items-center gap-2 rounded-full border border-border/60 bg-secondary/40 px-4 py-2 text-sm font-medium text-muted-foreground transition hover:bg-secondary/70 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background',
      selected && 'bg-primary/10 text-foreground border-primary/40',
      className
    )}
    {...props}
  >
    {icon}
    <span>{children}</span>
  </button>
)
