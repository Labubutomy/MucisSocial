import type { HTMLAttributes } from 'react'
import { cn } from '@shared/lib/cn'

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  padding?: 'none' | 'sm' | 'md' | 'lg'
  border?: boolean
}

const paddingStyles: Record<NonNullable<CardProps['padding']>, string> = {
  none: '',
  sm: 'p-4 md:p-6',
  md: 'p-6 md:p-8',
  lg: 'p-8 md:p-10',
}

export const Card = ({
  className,
  children,
  padding = 'md',
  border = true,
  ...props
}: CardProps) => (
  <div
    className={cn(
      'rounded-2xl bg-card/80 backdrop-blur-md shadow-lg shadow-black/25',
      border && 'border border-border/60',
      paddingStyles[padding],
      className
    )}
    {...props}
  >
    {children}
  </div>
)
