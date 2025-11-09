import type { FormEventHandler, ChangeEventHandler } from 'react'
import { cn } from '@shared/lib/cn'

export interface SearchBarProps {
  value: string
  placeholder?: string
  onChange: ChangeEventHandler<HTMLInputElement>
  onSubmit: FormEventHandler<HTMLFormElement>
  className?: string
}

export const SearchBar = ({
  value,
  placeholder = 'Ищите треки, артистов и плейлисты...',
  onChange,
  onSubmit,
  className,
}: SearchBarProps) => (
  <form
    onSubmit={onSubmit}
    className={cn(
      'relative flex items-center rounded-3xl border border-border/60 bg-secondary/40 px-4 py-2 shadow-sm shadow-black/10',
      className
    )}
  >
    <span className="mr-3 text-muted-foreground">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        className="h-5 w-5"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
      >
        <path d="m21 21-4.35-4.35" />
        <circle cx={11} cy={11} r={8} />
      </svg>
    </span>
    <input
      type="search"
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      className="w-full bg-transparent text-base text-foreground placeholder:text-muted-foreground focus:outline-none"
    />
    <button
      type="submit"
      className="ml-3 rounded-full bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground transition hover:bg-primary/90"
    >
      Найти
    </button>
  </form>
)
