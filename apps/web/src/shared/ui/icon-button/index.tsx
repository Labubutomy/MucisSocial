import type { ButtonHTMLAttributes } from 'react'
import { cn } from '@shared/lib/cn'

export interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  active?: boolean
  size?: 'sm' | 'md' | 'lg'
  variant?: 'ghost' | 'muted'
}

const sizeStyles: Record<NonNullable<IconButtonProps['size']>, string> = {
  sm: 'h-9 w-9',
  md: 'h-10 w-10',
  lg: 'h-12 w-12',
}

const variantStyles: Record<NonNullable<IconButtonProps['variant']>, string> = {
  ghost: 'bg-transparent hover:bg-muted/40',
  muted: 'bg-muted/60 hover:bg-muted/80',
}

export const IconButton = ({
  className,
  active,
  size = 'md',
  variant = 'ghost',
  ...props
}: IconButtonProps) => (
  <button
    className={cn(
      'inline-flex items-center justify-center rounded-full text-foreground transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:pointer-events-none disabled:opacity-60',
      sizeStyles[size],
      variantStyles[variant],
      active && 'bg-primary text-primary-foreground hover:bg-primary/90',
      className
    )}
    type="button"
    {...props}
  />
)
