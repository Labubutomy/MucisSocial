import type { ReactNode } from 'react'
import { cn } from '@shared/lib/cn'
import { AppHeader } from '@shared/ui/app-header'

export interface GlobalPlayerLayoutProps {
  children: ReactNode
  miniPlayer?: ReactNode
  showMiniPlayer?: boolean
  className?: string
}

export const GlobalPlayerLayout = ({
  children,
  miniPlayer,
  showMiniPlayer,
  className,
}: GlobalPlayerLayoutProps) => {
  const isMiniPlayerVisible = showMiniPlayer ?? (miniPlayer !== undefined && miniPlayer !== null)

  return (
    <div
      className={cn(
        'relative flex min-h-screen w-full flex-col bg-background text-foreground',
        className
      )}
    >
      <AppHeader />
      <main
        className={cn('flex-1 w-full pb-24 pt-8 md:pb-32', isMiniPlayerVisible ? 'md:pb-48' : '')}
      >
        {children}
      </main>

      {miniPlayer && isMiniPlayerVisible && (
        <div className="pointer-events-none sticky bottom-0 z-40 w-full px-4 pb-6">
          <div className="page-container">
            <div className="pointer-events-auto">{miniPlayer}</div>
          </div>
        </div>
      )}
    </div>
  )
}
