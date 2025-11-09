import type { ReactNode } from 'react'
import { Card } from '@shared/ui/card'
import { cn } from '@shared/lib/cn'

export interface AuthCardProps {
  header: {
    title: string
    subtitle: string
  }
  tabs: ReactNode
  form: ReactNode
  illustration?: ReactNode
  className?: string
}

export const AuthCard = ({ header, tabs, form, illustration, className }: AuthCardProps) => (
  <div
    className={cn(
      'grid w-full max-w-5xl gap-8 rounded-3xl border border-border/60 bg-secondary/20 p-6 shadow-2xl shadow-black/30 md:grid-cols-[1.1fr,0.9fr] md:p-8',
      className
    )}
  >
    <Card
      padding="lg"
      className="flex flex-col gap-8 bg-background/80 shadow-inner shadow-black/40"
    >
      <div className="space-y-3">
        <p className="text-xs uppercase tracking-[0.4em] text-primary">{header.subtitle}</p>
        <h1 className="text-3xl font-semibold md:text-4xl">{header.title}</h1>
      </div>
      <div className="flex flex-col gap-6">
        {tabs}
        {form}
      </div>
    </Card>
    <div className="hidden md:flex md:flex-col md:justify-between md:gap-6">
      {illustration}
      <div className="rounded-3xl border border-border/40 bg-gradient-to-br from-primary/50 via-accent/40 to-primary/20 p-6 text-sm text-primary-foreground shadow-xl shadow-primary/30">
        Слушайте вместе. Открывайте вместе. Ваш музыкальный дом.
      </div>
    </div>
  </div>
)
