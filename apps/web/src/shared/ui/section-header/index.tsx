import type { ReactNode } from 'react'
import { cn } from '@shared/lib/cn'

export interface SectionHeaderProps {
  title: ReactNode
  subtitle?: ReactNode
  action?: ReactNode
  className?: string
}

export const SectionHeader = ({ title, subtitle, action, className }: SectionHeaderProps) => (
  <div
    className={cn('flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between', className)}
  >
    <div>
      <h2 className="text-xl font-semibold text-foreground sm:text-2xl">{title}</h2>
      {subtitle && <p className="mt-1 text-sm text-muted-foreground sm:text-base">{subtitle}</p>}
    </div>
    {action}
  </div>
)
