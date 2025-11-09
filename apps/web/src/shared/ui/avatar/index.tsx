import type { HTMLAttributes } from 'react'
import { cn } from '@shared/lib/cn'

export interface AvatarProps extends HTMLAttributes<HTMLDivElement> {
  src?: string
  alt?: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  fallback?: string
}

const sizeStyles: Record<NonNullable<AvatarProps['size']>, string> = {
  sm: 'h-10 w-10',
  md: 'h-14 w-14',
  lg: 'h-20 w-20',
  xl: 'h-28 w-28',
}

export const Avatar = ({ src, alt, size = 'md', fallback, className, ...props }: AvatarProps) => {
  return (
    <div
      className={cn(
        'relative overflow-hidden rounded-full bg-muted/40',
        sizeStyles[size],
        className
      )}
      {...props}
    >
      {src ? (
        <img src={src} alt={alt} className="h-full w-full object-cover" loading="lazy" />
      ) : (
        <span className="flex h-full w-full items-center justify-center text-lg font-semibold uppercase text-muted-foreground">
          {fallback?.slice(0, 2)}
        </span>
      )}
    </div>
  )
}
